package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/integration/seatunnel"
	"dataease/backend/internal/repository"

	"gorm.io/gorm"
)

var jdbcHostPortPattern = regexp.MustCompile(`(?i)://([^:/?#]+):(\d+)`)

var repeatCheckSchemaTypes = map[string]struct{}{
	"sqlserver": {},
	"db2":       {},
	"oracle":    {},
	"pg":        {},
	"redshift":  {},
}

type DatasourceService struct {
	repo             *repository.DatasourceRepository
	excelService     *ExcelService
	seatunnelAddress string
	seatunnelTimeout time.Duration
	seatunnelRetries int
	seatunnelClient  *seatunnel.Client
	seatunnelMu      sync.Mutex
}

func NewDatasourceService(repo *repository.DatasourceRepository) *DatasourceService {
	return &DatasourceService{
		repo:             repo,
		excelService:     NewExcelService(),
		seatunnelAddress: "",
		seatunnelTimeout: 15 * time.Second,
		seatunnelRetries: 1,
	}
}

func (s *DatasourceService) SetSeatunnelConfig(address string, timeout time.Duration, retries int) {
	s.seatunnelAddress = strings.TrimSpace(address)
	if timeout > 0 {
		s.seatunnelTimeout = timeout
	}
	if retries >= 0 {
		s.seatunnelRetries = retries
	}

	s.seatunnelMu.Lock()
	if s.seatunnelClient != nil {
		_ = s.seatunnelClient.Close()
		s.seatunnelClient = nil
	}
	s.seatunnelMu.Unlock()
}

func (s *DatasourceService) List(req *datasource.ListRequest) (*datasource.ListResponse, error) {
	list, total, err := s.repo.Query(req)
	if err != nil {
		return nil, err
	}

	current := req.Current
	if current < 1 {
		current = 1
	}
	size := req.Size
	if size < 1 {
		size = 10
	}

	return &datasource.ListResponse{
		List:    list,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (s *DatasourceService) Validate(req *datasource.ValidateRequest) (*datasource.ValidateResponse, error) {
	dsType, cfgRaw, err := s.resolveConfig(req)
	if err != nil {
		return &datasource.ValidateResponse{Status: datasource.StatusError, Message: err.Error()}, nil
	}

	if dsType == datasource.TypeFolder || dsType == datasource.TypeExcel {
		return &datasource.ValidateResponse{Status: datasource.StatusSuccess, Message: "skip validation for folder/excel datasource"}, nil
	}

	cfg, err := decodeConfig(cfgRaw)
	if err != nil {
		return &datasource.ValidateResponse{Status: datasource.StatusError, Message: "invalid configuration: " + err.Error()}, nil
	}

	host, port := parseHostPort(cfg)
	if host == "" || port <= 0 {
		return &datasource.ValidateResponse{Status: datasource.StatusError, Message: "missing host/port in datasource configuration"}, nil
	}

	err = pingTCP(host, port, 3*time.Second)
	if err != nil {
		return &datasource.ValidateResponse{Status: datasource.StatusError, Message: err.Error()}, nil
	}

	return &datasource.ValidateResponse{Status: datasource.StatusSuccess, Message: "connection check passed"}, nil
}

func (s *DatasourceService) Tree(req *datasource.ListRequest) ([]*datasource.CoreDatasource, error) {
	return s.repo.ListAll(req.Keyword)
}

func (s *DatasourceService) GetTables(req *datasource.TableRequest) ([]datasource.TableInfo, error) {
	if req.DatasourceID <= 0 {
		return []datasource.TableInfo{}, nil
	}
	return s.repo.ListTables(req.DatasourceID)
}

func (s *DatasourceService) GetTableStatus(req *datasource.TableRequest) ([]datasource.TableInfo, error) {
	list, err := s.GetTables(req)
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].Status = datasource.StatusSuccess
		list[i].LastUpdate = 0
	}
	return list, nil
}

func (s *DatasourceService) GetSchema() ([]string, error) {
	return s.repo.ListSchemas()
}

func (s *DatasourceService) GetTableField(req *datasource.TableRequest) ([]datasource.TableField, error) {
	if strings.TrimSpace(req.TableName) == "" {
		return []datasource.TableField{}, nil
	}
	return s.repo.ListTableFields(strings.TrimSpace(req.TableName))
}

func (s *DatasourceService) PreviewData(req *datasource.TableRequest) (*datasource.PreviewDataResponse, error) {
	tableName := strings.TrimSpace(req.TableName)
	if tableName == "" {
		return &datasource.PreviewDataResponse{Fields: []datasource.TableField{}, Data: []map[string]interface{}{}, Total: 0}, nil
	}

	fields, err := s.repo.ListTableFields(tableName)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.PreviewRows(tableName, req.Limit)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.CountRows(tableName)
	if err != nil {
		return nil, err
	}

	return &datasource.PreviewDataResponse{Fields: fields, Data: rows, Total: total}, nil
}

func (s *DatasourceService) ValidateByID(id int64) (*datasource.ValidateResponse, error) {
	if fixedID, err := s.compatDatasourceID(id); err == nil {
		id = fixedID
	}
	return s.Validate(&datasource.ValidateRequest{DatasourceID: &id})
}

func (s *DatasourceService) GetByID(id int64) (*datasource.CoreDatasource, error) {
	fixedID, err := s.compatDatasourceID(id)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(fixedID)
}

func (s *DatasourceService) compatDatasourceID(id int64) (int64, error) {
	if id <= 0 {
		return id, nil
	}
	if _, err := s.repo.GetByID(id); err == nil {
		return id, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return id, err
	}

	if id%100 != 0 {
		return id, gorm.ErrRecordNotFound
	}

	nearestID, err := s.repo.FindNearestIDInWindow(id, 100)
	if err != nil {
		return id, err
	}
	if nearestID == nil {
		return id, gorm.ErrRecordNotFound
	}
	return *nearestID, nil
}

func (s *DatasourceService) CheckRepeat(req *datasource.WriteRequest) (bool, error) {
	if req == nil {
		return false, nil
	}

	dsType := strings.TrimSpace(req.Type)
	if dsType == "" {
		dsType = strings.TrimSpace(req.NodeType)
	}
	if dsType == "" || shouldSkipRepeatCheck(dsType) {
		return false, nil
	}

	if req.Configuration == nil || strings.TrimSpace(*req.Configuration) == "" {
		return false, nil
	}
	currentCfg, err := decodeConfig(*req.Configuration)
	if err != nil {
		return false, nil
	}

	var excludeID *int64
	if req.ID > 0 {
		excludeID = &req.ID
	}

	candidates, err := s.repo.ListByType(dsType, excludeID)
	if err != nil {
		return false, err
	}

	for _, item := range candidates {
		if item == nil || shouldSkipRepeatCheck(item.Type) {
			continue
		}
		if item.Configuration == nil || strings.TrimSpace(*item.Configuration) == "" {
			continue
		}
		compareCfg, cfgErr := decodeConfig(*item.Configuration)
		if cfgErr != nil {
			continue
		}
		if isSameDatasourceConnection(dsType, currentCfg, compareCfg) {
			return true, nil
		}
	}

	return false, nil
}

func (s *DatasourceService) Save(req *datasource.WriteRequest) (*datasource.CoreDatasource, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("datasource name is required")
	}

	pid := normalizedPID(req.PID)
	dsType := strings.TrimSpace(req.Type)
	if dsType == "" {
		dsType = strings.TrimSpace(req.NodeType)
	}
	if dsType == "" {
		dsType = datasource.TypeFolder
	}
	if req.Configuration == nil || strings.TrimSpace(*req.Configuration) == "" {
		if dsType == datasource.TypeFolder {
			emptyConfig := "{}"
			req.Configuration = &emptyConfig
		}
	}

	count, err := s.repo.CountByNameAndPID(name, pid, nil)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("datasource name already exists")
	}

	now := time.Now().UnixMilli()
	status := datasource.StatusSuccess
	ds := &datasource.CoreDatasource{
		PID:            &pid,
		Name:           name,
		Description:    req.Description,
		Type:           dsType,
		EditType:       req.EditType,
		Configuration:  req.Configuration,
		Status:         &status,
		EnableDataFill: req.EnableDataFill,
		CreateTime:     &now,
		UpdateTime:     &now,
	}

	if err = s.repo.Create(ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *DatasourceService) Update(req *datasource.WriteRequest) (*datasource.CoreDatasource, error) {
	if req.ID <= 0 {
		return nil, fmt.Errorf("datasource id is required")
	}

	existing, err := s.repo.GetByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("datasource not found")
		}
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = existing.Name
	}
	pid := normalizedPID(req.PID)
	if req.PID == nil && existing.PID != nil {
		pid = *existing.PID
	}

	count, err := s.repo.CountByNameAndPID(name, pid, &req.ID)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("datasource name already exists")
	}

	existing.Name = name
	existing.PID = &pid
	if req.Description != nil {
		existing.Description = req.Description
	}
	if strings.TrimSpace(req.Type) != "" {
		existing.Type = req.Type
	}
	if req.EditType != nil {
		existing.EditType = req.EditType
	}
	if req.Configuration != nil {
		existing.Configuration = req.Configuration
	}
	if req.EnableDataFill != nil {
		existing.EnableDataFill = req.EnableDataFill
	}
	now := time.Now().UnixMilli()
	existing.UpdateTime = &now

	if err = s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DatasourceService) CreateFolder(name string, pid int64) (*datasource.CoreDatasource, error) {
	return s.Save(&datasource.WriteRequest{
		Name:     name,
		PID:      &pid,
		Type:     datasource.TypeFolder,
		NodeType: datasource.TypeFolder,
	})
}

func (s *DatasourceService) Rename(id int64, name string) (*datasource.CoreDatasource, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("datasource not found")
		}
		return nil, err
	}
	newName := strings.TrimSpace(name)
	if newName == "" {
		return nil, fmt.Errorf("datasource name is required")
	}
	pid := int64(0)
	if existing.PID != nil {
		pid = *existing.PID
	}
	count, err := s.repo.CountByNameAndPID(newName, pid, &id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("datasource name already exists")
	}

	existing.Name = newName
	now := time.Now().UnixMilli()
	existing.UpdateTime = &now
	if err = s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DatasourceService) Move(id int64, pid int64) (*datasource.CoreDatasource, error) {
	if id <= 0 {
		return nil, fmt.Errorf("datasource id is required")
	}
	if id == pid {
		return nil, fmt.Errorf("destination folder cannot be itself")
	}
	existing, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("datasource not found")
		}
		return nil, err
	}

	if pid > 0 {
		isDescendant, checkErr := s.isDescendant(id, pid)
		if checkErr != nil {
			return nil, checkErr
		}
		if isDescendant {
			return nil, fmt.Errorf("destination folder cannot be child of current datasource")
		}
	}

	count, err := s.repo.CountByNameAndPID(existing.Name, pid, &id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("datasource name already exists")
	}

	existing.PID = &pid
	now := time.Now().UnixMilli()
	existing.UpdateTime = &now
	if err = s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DatasourceService) Delete(id int64) error {
	if id <= 0 {
		return fmt.Errorf("datasource id is required")
	}
	return s.deleteRecursive(id)
}

func (s *DatasourceService) PerDelete(id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("datasource id is required")
	}
	count, err := s.repo.CountDatasourceRelations(id)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *DatasourceService) LatestTypes(createBy string) ([]string, error) {
	if createBy == "" {
		return []string{}, nil
	}
	return s.repo.ListLatestTypesByCreator(createBy, 5)
}

func (s *DatasourceService) ShowFinishPage(userID int64) (bool, error) {
	if userID <= 0 {
		return false, nil
	}
	exists, err := s.repo.ExistsFinishPageRecord(userID)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func (s *DatasourceService) SetShowFinishPage(userID int64) error {
	if userID <= 0 {
		return nil
	}
	return s.repo.CreateFinishPageRecord(userID)
}

func (s *DatasourceService) deleteRecursive(id int64) error {
	children, err := s.repo.ListChildren(id)
	if err != nil {
		return err
	}
	for _, child := range children {
		if err = s.deleteRecursive(child.ID); err != nil {
			return err
		}
	}
	return s.repo.SoftDelete(id)
}

func (s *DatasourceService) isDescendant(rootID int64, targetID int64) (bool, error) {
	children, err := s.repo.ListChildren(rootID)
	if err != nil {
		return false, err
	}
	for _, child := range children {
		if child.ID == targetID {
			return true, nil
		}
		descendant, innerErr := s.isDescendant(child.ID, targetID)
		if innerErr != nil {
			return false, innerErr
		}
		if descendant {
			return true, nil
		}
	}
	return false, nil
}

func normalizedPID(pid *int64) int64 {
	if pid == nil {
		return 0
	}
	if *pid < 0 {
		return 0
	}
	return *pid
}

func shouldSkipRepeatCheck(dsType string) bool {
	lowerType := strings.ToLower(strings.TrimSpace(dsType))
	if lowerType == "folder" || lowerType == "es" {
		return true
	}
	if strings.Contains(lowerType, "api") || strings.Contains(lowerType, "excel") {
		return true
	}
	return false
}

func requiresSchemaMatch(dsType string) bool {
	_, ok := repeatCheckSchemaTypes[strings.ToLower(strings.TrimSpace(dsType))]
	return ok
}

func isSameDatasourceConnection(dsType string, current *datasource.ConnectionConfig, compare *datasource.ConnectionConfig) bool {
	if current == nil || compare == nil {
		return false
	}

	currentHost, currentPort := parseHostPort(current)
	compareHost, comparePort := parseHostPort(compare)
	if currentPort <= 0 || comparePort <= 0 {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(currentHost), strings.TrimSpace(compareHost)) {
		return false
	}
	if currentPort != comparePort {
		return false
	}

	currentDB := strings.TrimSpace(current.Database)
	compareDB := strings.TrimSpace(compare.Database)
	if currentDB == "" || compareDB == "" {
		return false
	}
	if !strings.EqualFold(currentDB, compareDB) {
		return false
	}

	if requiresSchemaMatch(dsType) {
		return strings.EqualFold(strings.TrimSpace(current.Schema), strings.TrimSpace(compare.Schema))
	}

	return true
}

func (s *DatasourceService) resolveConfig(req *datasource.ValidateRequest) (string, string, error) {
	if req.DatasourceID != nil {
		ds, err := s.repo.GetByID(*req.DatasourceID)
		if err != nil {
			return "", "", fmt.Errorf("datasource not found")
		}
		if ds.Configuration == nil {
			return ds.Type, "", fmt.Errorf("datasource configuration is empty")
		}
		return ds.Type, *ds.Configuration, nil
	}

	if req.Type == nil || *req.Type == "" {
		return "", "", fmt.Errorf("datasource type is required")
	}
	if req.Configuration == nil || *req.Configuration == "" {
		return *req.Type, "", fmt.Errorf("datasource configuration is required")
	}

	return *req.Type, *req.Configuration, nil
}

func decodeConfig(raw string) (*datasource.ConnectionConfig, error) {
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		decoded = []byte(raw)
	}

	var cfg datasource.ConnectionConfig
	if err = json.Unmarshal(decoded, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func parseHostPort(cfg *datasource.ConnectionConfig) (string, int) {
	if cfg.Host != "" && cfg.Port > 0 {
		return cfg.Host, cfg.Port
	}

	jdbc := strings.TrimSpace(cfg.JDBCUrl)
	if jdbc == "" {
		return "", 0
	}

	matches := jdbcHostPortPattern.FindStringSubmatch(jdbc)
	if len(matches) != 3 {
		return "", 0
	}
	port, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", 0
	}
	return matches[1], port
}

func pingTCP(host string, port int, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return fmt.Errorf("failed to connect %s:%d: %w", host, port, err)
	}
	_ = conn.Close()
	return nil
}

func (s *DatasourceService) UploadFile(file multipart.File, header *multipart.FileHeader, datasourceID int64, editType int) (*datasource.ExcelFileData, error) {
	return s.excelService.UploadFile(file, header, datasourceID, editType)
}

func (s *DatasourceService) LoadRemoteFile(url, userName, password string, datasourceID int64) (*datasource.ExcelFileData, error) {
	return s.excelService.LoadRemoteFile(&RemoteExcelRequest{
		URL:          url,
		UserName:     userName,
		Password:     password,
		DatasourceID: datasourceID,
	})
}

func (s *DatasourceService) CheckAPIDatasource(req map[string]string) (map[string]interface{}, error) {
	if len(req) == 0 {
		return nil, fmt.Errorf("request is required")
	}
	dataRaw := strings.TrimSpace(req["data"])
	if dataRaw == "" {
		return nil, fmt.Errorf("data is required")
	}

	apiDefinition, err := decodeMaybeBase64JSONMap(dataRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid api definition: %w", err)
	}

	apiDefinition["type"] = "table"
	apiDefinition["showApiStructure"] = strings.EqualFold(strings.TrimSpace(req["type"]), "apiStructure")
	if _, ok := apiDefinition["name"]; !ok {
		apiDefinition["name"] = "api_table"
	}

	return apiDefinition, nil
}

func (s *DatasourceService) SyncAPITable(req map[string]string) (map[string]interface{}, error) {
	return s.submitSyncTask(req, "table")
}

func (s *DatasourceService) SyncAPIDs(req map[string]string) (map[string]interface{}, error) {
	return s.submitSyncTask(req, "datasource")
}

func (s *DatasourceService) ListSyncRecord(dsID int64, page int, limit int) (*datasource.SyncRecordPage, error) {
	if dsID <= 0 {
		return nil, fmt.Errorf("invalid datasource id")
	}
	if s.repo == nil {
		return nil, fmt.Errorf("datasource repository is unavailable")
	}

	records, total, err := s.repo.ListSyncTaskLogs(dsID, page, limit)
	if err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	return &datasource.SyncRecordPage{
		Records:      records,
		Total:        total,
		Current:      page,
		Size:         limit,
		DatasourceID: dsID,
	}, nil
}

func (s *DatasourceService) submitSyncTask(req map[string]string, syncType string) (map[string]interface{}, error) {
	dsID, err := parseDatasourceID(req)
	if err != nil {
		return nil, err
	}

	client, err := s.ensureSeatunnelClient()
	if err != nil {
		return nil, err
	}

	taskName := strings.TrimSpace(req["name"])
	if taskName == "" {
		taskName = fmt.Sprintf("%s-sync-%d", syncType, dsID)
	}
	tableName := strings.TrimSpace(req["tableName"])
	now := time.Now().UnixMilli()

	task := &seatunnel.SyncTask{
		Name:     taskName,
		Source:   strings.TrimSpace(req["source"]),
		Target:   strings.TrimSpace(req["target"]),
		Status:   seatunnel.StatusPending,
		Progress: 0,
	}

	taskID, submitErr := client.SubmitTask(context.Background(), task)

	if s.repo != nil {
		logRecord := &datasource.SyncRecord{
			DsID:        dsID,
			TaskID:      parseTaskID(taskID),
			StartTime:   now,
			CreateTime:  now,
			TaskStatus:  seatunnel.StatusRunning,
			TableName:   tableName,
			Name:        taskName,
			TriggerType: syncType,
		}
		if submitErr != nil {
			logRecord.TaskStatus = seatunnel.StatusFailed
			logRecord.EndTime = now
			logRecord.Info = submitErr.Error()
		}
		_ = s.repo.CreateSyncTaskLog(logRecord)
	}

	if submitErr != nil {
		return nil, submitErr
	}

	return map[string]interface{}{
		"taskId":       taskID,
		"status":       seatunnel.StatusRunning,
		"datasourceId": dsID,
		"syncType":     syncType,
	}, nil
}

func (s *DatasourceService) ensureSeatunnelClient() (*seatunnel.Client, error) {
	s.seatunnelMu.Lock()
	defer s.seatunnelMu.Unlock()

	if s.seatunnelClient != nil {
		return s.seatunnelClient, nil
	}
	if s.seatunnelAddress == "" {
		return nil, fmt.Errorf("seatunnel grpc address is not configured")
	}

	timeout := s.seatunnelTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	retries := s.seatunnelRetries
	if retries < 0 {
		retries = 0
	}

	client, err := seatunnel.NewClient(&seatunnel.Config{Address: s.seatunnelAddress, Timeout: timeout, MaxRetries: retries})
	if err != nil {
		return nil, err
	}
	s.seatunnelClient = client
	return s.seatunnelClient, nil
}

func decodeMaybeBase64JSONMap(raw string) (map[string]interface{}, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("json payload is empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err == nil {
		trimmed = strings.TrimSpace(string(decoded))
	}

	result := make(map[string]interface{})
	if err = json.Unmarshal([]byte(trimmed), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func parseDatasourceID(req map[string]string) (int64, error) {
	if len(req) == 0 {
		return 0, fmt.Errorf("request is required")
	}
	for _, key := range []string{"dsId", "datasourceId", "id"} {
		value := strings.TrimSpace(req[key])
		if value == "" {
			continue
		}
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid datasource id")
		}
		if id <= 0 {
			return 0, fmt.Errorf("invalid datasource id")
		}
		return id, nil
	}
	return 0, fmt.Errorf("datasource id is required")
}

func parseTaskID(raw string) int64 {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0
	}
	id, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func (s *DatasourceService) CancelSyncTask(taskID string) error {
	client, err := s.ensureSeatunnelClient()
	if err != nil {
		return err
	}
	return client.CancelTask(context.Background(), taskID)
}

func (s *DatasourceService) CancelSyncTask(taskID string) error {
	client, err := s.ensureSeatunnelClient()
	if err != nil {
		return err
	}
	return client.CancelTask(context.Background(), taskID)
}
