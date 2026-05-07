package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	datafillingdomain "dataease/backend/internal/domain/datafilling"
	datasourcedomain "dataease/backend/internal/domain/datasource"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const nilStringValue = "<nil>"
const maxDataFillingExcelUploadSize int64 = 10 * 1024 * 1024

type ExcelUploadSession struct {
	mu       sync.RWMutex
	sessions map[string]*datafillingdomain.DfExcelData
}

func NewExcelUploadSession() *ExcelUploadSession {
	return &ExcelUploadSession{sessions: make(map[string]*datafillingdomain.DfExcelData)}
}

func (s *ExcelUploadSession) Store(data *datafillingdomain.DfExcelData) string {
	if s == nil || data == nil {
		return ""
	}
	key := uuid.NewString()
	s.mu.Lock()
	s.sessions[key] = data
	s.mu.Unlock()
	return key
}

func (s *ExcelUploadSession) Load(key string) (*datafillingdomain.DfExcelData, bool) {
	if s == nil || strings.TrimSpace(key) == "" {
		return nil, false
	}
	s.mu.RLock()
	data, ok := s.sessions[key]
	s.mu.RUnlock()
	return data, ok
}

func (s *ExcelUploadSession) Delete(key string) {
	if s == nil || strings.TrimSpace(key) == "" {
		return
	}
	s.mu.Lock()
	delete(s.sessions, key)
	s.mu.Unlock()
}

type DataFillingRepo interface {
	Create(ctx context.Context, form *datafillingdomain.DataFillingForm) error
	GetByID(ctx context.Context, id int64) (*datafillingdomain.DataFillingForm, error)
	Update(ctx context.Context, form *datafillingdomain.DataFillingForm) error
	DeleteByID(ctx context.Context, id int64) error
	Rename(ctx context.Context, id int64, name string) error
	Move(ctx context.Context, id int64, pid int64) error
	GetTree(ctx context.Context) ([]*datafillingdomain.DataFillingForm, error)
	GetByPID(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error)
	GetChildren(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error)
}

type DataFillingDatasourceService interface {
	GetByID(id int64) (*datasourcedomain.CoreDatasource, error)
	Tree(req *datasourcedomain.ListRequest) ([]*datasourcedomain.CoreDatasource, error)
	GetTables(req *datasourcedomain.TableRequest) ([]datasourcedomain.TableInfo, error)
}

type CommitLogRepo interface {
	Create(ctx context.Context, log *datafillingdomain.DfCommitLog) error
	ListByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DfCommitLog, int64, error)
	DeleteByFormID(ctx context.Context, formID int64) error
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task *datafillingdomain.DataFillingTask) error
	UpdateTask(ctx context.Context, task *datafillingdomain.DataFillingTask) error
	GetTaskByID(ctx context.Context, taskID int64) (*datafillingdomain.DataFillingTask, error)
	ListTasksByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DataFillingTask, int64, error)
	DeleteTasksByIDs(ctx context.Context, taskIDs []int64) error
	GetStartedTasks(ctx context.Context) ([]*datafillingdomain.DataFillingTask, error)
}

type SubTaskRepository interface {
	CreateSubTask(ctx context.Context, subTask *datafillingdomain.DataFillingSubTask) error
	GetSubTaskByID(ctx context.Context, subTaskID int64) (*datafillingdomain.DataFillingSubTask, error)
	UpdateSubTaskCounts(ctx context.Context, subTaskID int64, totalCount, unfinishedCount, totalUserCount, unfinishedUserCount int) error
	ListSubTasksByTaskID(ctx context.Context, taskID int64, page, pageSize int) ([]*datafillingdomain.DataFillingSubTask, int64, error)
	DeleteSubTasksByIDs(ctx context.Context, subTaskIDs []int64) error
	ListSubTaskIDsByTaskIDs(ctx context.Context, taskIDs []int64) ([]int64, error)
	DecrementSubTaskUnfinishedCount(ctx context.Context, subTaskID int64) error
}

type SubInstanceRepository interface {
	BatchCreateSubInstances(ctx context.Context, instances []*datafillingdomain.DataFillingSubInstance) error
	DeleteSubInstancesByPID(ctx context.Context, pid int64) error
	DeleteSubInstancesByPIDs(ctx context.Context, pids []int64) error
	DeleteSubInstancesByTaskIDs(ctx context.Context, taskIDs []int64) error
	ListSubInstancesByPID(ctx context.Context, pid int64, statusFilter *int) ([]*datafillingdomain.DataFillingSubInstance, error)
	ListSubInstancesByUID(ctx context.Context, uid int64, page, pageSize int, req *datafillingdomain.UserTaskPageRequest) ([]*datafillingdomain.UserTaskVO, int64, error)
	CountOpenSubInstancesByUID(ctx context.Context, uid int64) (int64, error)
	GetSubInstanceByID(ctx context.Context, id int64) (*datafillingdomain.DataFillingSubInstance, error)
	GetSubInstanceByPIDAndUID(ctx context.Context, pid, uid int64) ([]*datafillingdomain.DataFillingSubInstance, error)
	UpdateSubInstanceStatus(ctx context.Context, id int64, status int, finishTime int64) error
}

type DataFillingService struct {
	repo                   DataFillingRepo
	datasourceService      DataFillingDatasourceService
	ddlProvider            DDLProvider
	commitLogRepo          CommitLogRepo
	datasourceConnProvider DatasourceConnectionProvider
	taskRepo               TaskRepository
	subTaskRepo            SubTaskRepository
	subInstanceRepo        SubInstanceRepository
	scheduler              *DataFillingScheduler
	excelUploadSession     *ExcelUploadSession
}

func NewDataFillingService(repo DataFillingRepo, datasourceService DataFillingDatasourceService, ddlProvider DDLProvider, commitLogRepo CommitLogRepo, taskRepo TaskRepository, subTaskRepo SubTaskRepository, subInstanceRepo SubInstanceRepository, scheduler *DataFillingScheduler) *DataFillingService {
	return &DataFillingService{repo: repo, datasourceService: datasourceService, ddlProvider: ddlProvider, commitLogRepo: commitLogRepo, taskRepo: taskRepo, subTaskRepo: subTaskRepo, subInstanceRepo: subInstanceRepo, scheduler: scheduler, excelUploadSession: NewExcelUploadSession()}
}

func (s *DataFillingService) SetDatasourceConnectionProvider(provider DatasourceConnectionProvider) {
	s.datasourceConnProvider = provider
}

func (s *DataFillingService) ExcelTemplateDownload(ctx context.Context, formID int64, writer io.Writer) error {
	if writer == nil {
		return gorm.ErrInvalidData
	}
	form, err := s.repo.GetByID(ctx, formID)
	if err != nil {
		return err
	}
	fields, err := parseExtTableFields(form.Forms)
	if err != nil {
		return err
	}
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	if err := writeExcelHeaders(file, activeExcelFieldMetas(fields)); err != nil {
		return err
	}
	return file.Write(writer)
}

func (s *DataFillingService) ExcelUpload(ctx context.Context, formID int64, fileHeader *multipart.FileHeader) (*datafillingdomain.DfExcelData, error) {
	if fileHeader == nil || fileHeader.Size <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if fileHeader.Size > maxDataFillingExcelUploadSize {
		return nil, fmt.Errorf("file size exceeds 10MB limit")
	}
	form, err := s.repo.GetByID(ctx, formID)
	if err != nil {
		return nil, err
	}
	formFields, err := parseFormFieldMaps(form.Forms)
	if err != nil {
		return nil, err
	}
	fields, err := parseExtTableFields(form.Forms)
	if err != nil {
		return nil, err
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	workbook, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = workbook.Close() }()
	dataList, err := parseExcelUploadRows(workbook, activeExcelFieldMetas(fields))
	if err != nil {
		return nil, err
	}
	result := &datafillingdomain.DfExcelData{
		FormFields: formFields,
		DataList:   dataList,
		ExcelName:  fileHeader.Filename,
		Path:       "",
		Suffix:     filepath.Ext(fileHeader.Filename),
	}
	result.ID = s.excelUploadSession.Store(result)
	return result, nil
}

func (s *DataFillingService) ConfirmUpload(ctx context.Context, formID int64, uploadID string, userID int64, userName string) error {
	data, err := s.mustLoadExcelUpload(uploadID)
	if err != nil {
		return err
	}
	defer s.excelUploadSession.Delete(uploadID)
	return s.persistUploadedRows(ctx, formID, data.DataList, userID, userName)
}

func (s *DataFillingService) UserTaskConfirmUpload(ctx context.Context, userID, subTaskID, formID int64, uploadID string) error {
	if userID <= 0 || formID <= 0 {
		return gorm.ErrInvalidData
	}
	instanceSet, form, err := s.loadUserTaskForm(ctx, userID, subTaskID)
	if err != nil {
		return err
	}
	if form.ID != formID {
		return gorm.ErrRecordNotFound
	}
	data, err := s.mustLoadExcelUpload(uploadID)
	if err != nil {
		return err
	}
	defer s.excelUploadSession.Delete(uploadID)
	if err := s.persistUploadedRows(ctx, formID, data.DataList, userID, ""); err != nil {
		return err
	}
	return s.finishUserTaskIfOpen(ctx, instanceSet)
}

func (s *DataFillingService) ExtraDetails(ctx context.Context, req *datafillingdomain.ExtraDetailsRequest) ([]*datafillingdomain.ExtraDetails, error) {
	datasourceID, tableName, optionColumn, extraColumns, value, err := validateExtraDetailsRequest(req)
	if err != nil {
		return nil, err
	}
	db, err := s.GetDatasourceConnection(ctx, datasourceID)
	if err != nil {
		return nil, err
	}
	selectColumns := append([]string{quoteIdentifier(optionColumn)}, quoteIdentifiers(extraColumns)...)
	rows := make([]map[string]interface{}, 0)
	query := db.WithContext(ctx).Table(quoteIdentifier(tableName)).Select(strings.Join(selectColumns, ", ")).Where(fmt.Sprintf("%s = ?", quoteIdentifier(optionColumn)), value).Limit(1000)
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return flattenExtraDetails(rows, extraColumns), nil
}

func (s *DataFillingService) ListDatasourceOptions(ctx context.Context, datasourceID int64, req *datafillingdomain.DatasourceOptionsRequest) ([]*datafillingdomain.ColumnOption, error) {
	tableName, optionColumn, orderColumn, err := validateDatasourceOptionsRequest(req)
	if err != nil {
		return nil, err
	}
	db, err := s.GetDatasourceConnection(ctx, datasourceID)
	if err != nil {
		return nil, err
	}
	values := make([]string, 0)
	if err := db.WithContext(ctx).Table(quoteIdentifier(tableName)).Distinct().Order(quoteIdentifier(orderColumn)).Limit(1000).Pluck(optionColumn, &values).Error; err != nil {
		return nil, err
	}
	result := make([]*datafillingdomain.ColumnOption, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		result = append(result, &datafillingdomain.ColumnOption{Name: trimmed, Value: trimmed})
	}
	return result, nil
}

func (s *DataFillingService) GetTemplateByUserTaskItem(ctx context.Context, itemID int64) (string, error) {
	if s.subInstanceRepo == nil || s.taskRepo == nil || itemID <= 0 {
		return "", gorm.ErrInvalidData
	}
	instance, err := s.subInstanceRepo.GetSubInstanceByID(ctx, itemID)
	if err != nil {
		return "", err
	}
	task, err := s.taskRepo.GetTaskByID(ctx, instance.TaskID)
	if err != nil {
		return "", err
	}
	formID := task.FormID
	if formID <= 0 {
		formID = instance.FormID
	}
	form, err := s.repo.GetByID(ctx, formID)
	if err != nil {
		return "", err
	}
	return form.Forms, nil
}

func (s *DataFillingService) ExportFormData(ctx context.Context, formID int64, writer io.Writer) error {
	if writer == nil {
		return gorm.ErrInvalidData
	}
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return err
	}
	fields, err := parseExtTableFields(form.Forms)
	if err != nil {
		return err
	}
	metas := activeExcelFieldMetas(fields)
	rows, err := s.loadAllFormRows(ctx, db, form.PhysicalTableName)
	if err != nil {
		return err
	}
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	if err := writeExcelHeaders(file, metas); err != nil {
		return err
	}
	if err := writeExcelDataRows(file, metas, rows); err != nil {
		return err
	}
	return file.Write(writer)
}

func (s *DataFillingService) Save(ctx context.Context, req *datafillingdomain.CreateFormRequest, userID int64) (*datafillingdomain.DataFillingForm, error) {
	if err := validateDataFillingCreateRequest(req); err != nil {
		return nil, err
	}
	nodeType := normalizeDataFillingNodeType(req.NodeType)
	level, err := s.resolveLevel(ctx, req.PID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	form := &datafillingdomain.DataFillingForm{
		Name:              strings.TrimSpace(req.Name),
		PID:               req.PID,
		Level:             level,
		NodeType:          nodeType,
		PhysicalTableName: strings.TrimSpace(req.TableName),
		DatasourceID:      req.DatasourceID,
		Forms:             strings.TrimSpace(req.Forms),
		CreateIndex:       req.CreateIndex,
		TableIndexes:      strings.TrimSpace(req.TableIndexes),
		CreateBy:          userID,
		CreateTime:        now,
		UpdateBy:          userID,
		UpdateTime:        now,
		UseExistsTable:    req.UseExistsTable,
	}
	if err := s.repo.Create(ctx, form); err != nil {
		return nil, err
	}
	if nodeType == datafillingdomain.NodeTypeForm && !req.UseExistsTable {
		if err := s.createPhysicalTable(ctx, form); err != nil {
			_ = s.repo.DeleteByID(ctx, form.ID)
			return nil, err
		}
	}
	return form, nil
}

func (s *DataFillingService) Get(ctx context.Context, id int64) (*datafillingdomain.DataFillingForm, error) {
	if id <= 0 {
		return nil, gorm.ErrInvalidData
	}
	return s.repo.GetByID(ctx, id)
}

func (s *DataFillingService) Update(ctx context.Context, req *datafillingdomain.UpdateFormRequest, userID int64) (*datafillingdomain.DataFillingForm, error) {
	if req == nil || req.ID <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if err := validateDataFillingCreateRequest(&req.CreateFormRequest); err != nil {
		return nil, err
	}
	current, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	originalFields := strings.TrimSpace(current.Forms)
	level, err := s.resolveLevel(ctx, req.PID)
	if err != nil {
		return nil, err
	}
	current.Name = strings.TrimSpace(req.Name)
	current.PID = req.PID
	current.Level = level
	current.NodeType = normalizeDataFillingNodeType(req.NodeType)
	current.PhysicalTableName = strings.TrimSpace(req.TableName)
	current.DatasourceID = req.DatasourceID
	current.Forms = strings.TrimSpace(req.Forms)
	current.CreateIndex = req.CreateIndex
	current.TableIndexes = strings.TrimSpace(req.TableIndexes)
	current.UpdateBy = userID
	current.UpdateTime = time.Now().UnixMilli()
	if normalizeDataFillingNodeType(current.NodeType) == datafillingdomain.NodeTypeForm && originalFields != current.Forms {
		if err := s.alterPhysicalTableForFieldChanges(ctx, current, originalFields, current.Forms); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *DataFillingService) SearchTableData(ctx context.Context, formID int64, req *datafillingdomain.TableDataRequest) (*datafillingdomain.TableDataResponse, error) {
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return nil, err
	}
	currentPage := int64(1)
	pageSize := int64(20)
	searchParams := []datafillingdomain.SearchParam{}
	if req != nil {
		if req.CurrentPage > 0 {
			currentPage = req.CurrentPage
		}
		if req.PageSize > 0 {
			pageSize = req.PageSize
		}
		searchParams = req.SearchParams
	}
	whereClause, args, err := buildWhereClause(searchParams)
	if err != nil {
		return nil, err
	}
	offset := (currentPage - 1) * pageSize
	rows, err := s.ddlProvider.SearchRows(ctx, db, form.PhysicalTableName, whereClause, args, pageSize, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.ddlProvider.CountRows(ctx, db, form.PhysicalTableName, whereClause, args)
	if err != nil {
		return nil, err
	}
	return &datafillingdomain.TableDataResponse{Data: rows, Fields: form.Forms, Total: total, CurrentPage: currentPage, PageSize: pageSize, Key: "id"}, nil
}

func (s *DataFillingService) SaveRowData(ctx context.Context, formID int64, rowData map[string]interface{}, userID int64, userName string) (*datafillingdomain.TableDataResponse, error) {
	if len(rowData) == 0 {
		return nil, gorm.ErrInvalidData
	}
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return nil, err
	}
	rowID := strings.TrimSpace(fmt.Sprint(rowData["id"]))
	operate := 1
	if rowID == "" || rowID == nilStringValue {
		delete(rowData, "id")
		if err := s.ddlProvider.InsertRow(ctx, db, form.PhysicalTableName, rowData); err != nil {
			return nil, err
		}
		rowID = strings.TrimSpace(fmt.Sprint(rowData["id"]))
	} else {
		operate = 2
		if err := s.ddlProvider.UpdateRow(ctx, db, form.PhysicalTableName, rowID, rowData); err != nil {
			return nil, err
		}
	}
	if err := s.writeCommitLog(ctx, formID, rowID, operate, userID, userName, 1); err != nil {
		return nil, err
	}
	return &datafillingdomain.TableDataResponse{Data: []map[string]interface{}{rowData}, Fields: form.Forms, Total: 1, CurrentPage: 1, PageSize: 1, Key: "id"}, nil
}

func (s *DataFillingService) DeleteRowData(ctx context.Context, formID int64, rowID string, userID int64, userName string) error {
	if strings.TrimSpace(rowID) == "" {
		return gorm.ErrInvalidData
	}
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return err
	}
	if err := s.ddlProvider.DeleteRows(ctx, db, form.PhysicalTableName, []string{rowID}); err != nil {
		return err
	}
	return s.writeCommitLog(ctx, formID, rowID, 0, userID, userName, 1)
}

func (s *DataFillingService) BatchDeleteRowData(ctx context.Context, formID int64, ids []string, userID int64, userName string) error {
	cleaned := cleanRowIDs(ids)
	if len(cleaned) == 0 {
		return gorm.ErrInvalidData
	}
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return err
	}
	for start := 0; start < len(cleaned); start += 500 {
		end := start + 500
		if end > len(cleaned) {
			end = len(cleaned)
		}
		if err := s.ddlProvider.DeleteRows(ctx, db, form.PhysicalTableName, cleaned[start:end]); err != nil {
			return err
		}
	}
	return s.writeCommitLog(ctx, formID, "", 0, userID, userName, len(cleaned))
}

func (s *DataFillingService) TruncateTableData(ctx context.Context, formID int64) error {
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return err
	}
	return s.ddlProvider.TruncateTable(ctx, db, form.PhysicalTableName)
}

func (s *DataFillingService) ListColumnData(ctx context.Context, formID int64, columnName string) ([]string, error) {
	if !isValidDDLIdentifier(columnName) {
		return nil, gorm.ErrInvalidData
	}
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return nil, err
	}
	return s.ddlProvider.ListColumnData(ctx, db, form.PhysicalTableName, columnName)
}

func (s *DataFillingService) ListCommitLogs(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DfCommitLog, int64, error) {
	if s.commitLogRepo == nil {
		return nil, 0, fmt.Errorf("commit log repository not configured")
	}
	if formID <= 0 || page <= 0 || pageSize <= 0 {
		return nil, 0, gorm.ErrInvalidData
	}
	return s.commitLogRepo.ListByFormID(ctx, formID, page, pageSize)
}

func (s *DataFillingService) ClearCommitLogs(ctx context.Context, formID int64) error {
	if s.commitLogRepo == nil {
		return fmt.Errorf("commit log repository not configured")
	}
	if formID <= 0 {
		return gorm.ErrInvalidData
	}
	return s.commitLogRepo.DeleteByFormID(ctx, formID)
}

func marshalTaskLists(req *datafillingdomain.TaskSaveRequest) (reciFlagList, uidList, ridList string, err error) {
	rf, err := json.Marshal(req.ReciFlagList)
	if err != nil {
		return "", "", "", err
	}
	ul, err := json.Marshal(req.UIDList)
	if err != nil {
		return "", "", "", err
	}
	rl, err := json.Marshal(req.RIDList)
	if err != nil {
		return "", "", "", err
	}
	return string(rf), string(ul), string(rl), nil
}

func (s *DataFillingService) updateExistingTask(ctx context.Context, req *datafillingdomain.TaskSaveRequest, userID int64, reciFlagList, uidList, ridList string) (int64, error) {
	current, err := s.taskRepo.GetTaskByID(ctx, *req.ID)
	if err != nil {
		return 0, err
	}
	wasStarted := current.Status == datafillingdomain.TaskStatusStarted
	current.FormID = req.FormID
	current.Name = strings.TrimSpace(req.Name)
	current.ReciFlagList = reciFlagList
	current.UIDList = uidList
	current.RIDList = ridList
	current.FillType = req.FillType
	current.FitType = req.FitType
	current.FitColumn = strings.TrimSpace(req.FitColumn)
	current.RateType = req.RateType
	current.RateVal = strings.TrimSpace(req.RateVal)
	current.OneTimeType = req.OneTimeType
	current.StartTime = req.StartTime
	current.EndTime = req.EndTime
	current.PublishRangeTime = req.PublishRangeTime
	current.PublishRangeTimeType = req.PublishRangeTimeType
	current.FormExtSetting = strings.TrimSpace(req.FormExtSetting)
	current.FormFilterSetting = strings.TrimSpace(req.FormFilterSetting)
	current.UpdateBy = userID
	current.UpdateTime = time.Now().UnixMilli()
	if wasStarted && s.scheduler != nil {
		s.scheduler.UnregisterTask(current.ID)
	}
	if err := s.taskRepo.UpdateTask(ctx, current); err != nil {
		return 0, err
	}
	if wasStarted && s.scheduler != nil {
		nextExecTime, err := s.scheduler.computeNextExecTime(current)
		if err != nil {
			return 0, err
		}
		current.NextExecTime = nextExecTime
		if err := s.taskRepo.UpdateTask(ctx, current); err != nil {
			return 0, err
		}
		if err := s.scheduler.RegisterTask(ctx, current.ID); err != nil {
			return 0, err
		}
	}
	return current.ID, nil
}

func (s *DataFillingService) createNewTask(ctx context.Context, req *datafillingdomain.TaskSaveRequest, userID int64, reciFlagList, uidList, ridList string) (int64, error) {
	now := time.Now().UnixMilli()
	task := &datafillingdomain.DataFillingTask{
		FormID:               req.FormID,
		Name:                 strings.TrimSpace(req.Name),
		ReciFlagList:         reciFlagList,
		UIDList:              uidList,
		RIDList:              ridList,
		FillType:             req.FillType,
		FitType:              req.FitType,
		FitColumn:            strings.TrimSpace(req.FitColumn),
		RateType:             req.RateType,
		RateVal:              strings.TrimSpace(req.RateVal),
		OneTimeType:          req.OneTimeType,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		PublishRangeTime:     req.PublishRangeTime,
		PublishRangeTimeType: req.PublishRangeTimeType,
		Status:               datafillingdomain.TaskStatusStopped,
		CreateBy:             userID,
		CreateTime:           now,
		UpdateBy:             userID,
		UpdateTime:           now,
		FormExtSetting:       strings.TrimSpace(req.FormExtSetting),
		FormFilterSetting:    strings.TrimSpace(req.FormFilterSetting),
	}
	if err := s.taskRepo.CreateTask(ctx, task); err != nil {
		return 0, err
	}
	return task.ID, nil
}

func (s *DataFillingService) SaveTask(ctx context.Context, req *datafillingdomain.TaskSaveRequest, userID int64) (int64, error) {
	if s.taskRepo == nil {
		return 0, fmt.Errorf("task repository not configured")
	}
	if req == nil || req.FormID <= 0 || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.RateVal) == "" {
		return 0, gorm.ErrInvalidData
	}
	if _, err := s.repo.GetByID(ctx, req.FormID); err != nil {
		return 0, err
	}
	reciFlagList, uidList, ridList, err := marshalTaskLists(req)
	if err != nil {
		return 0, err
	}
	if req.ID != nil && *req.ID > 0 {
		return s.updateExistingTask(ctx, req, userID, reciFlagList, uidList, ridList)
	}
	return s.createNewTask(ctx, req, userID, reciFlagList, uidList, ridList)
}

func (s *DataFillingService) GetTaskInfo(ctx context.Context, taskID int64) (*datafillingdomain.TaskInfoVO, error) {
	if s.taskRepo == nil || taskID <= 0 {
		return nil, gorm.ErrInvalidData
	}
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return buildTaskInfoVO(task)
}

func (s *DataFillingService) StartTask(ctx context.Context, formID, taskID int64) error {
	if s.taskRepo == nil || s.scheduler == nil || formID <= 0 || taskID <= 0 {
		return gorm.ErrInvalidData
	}
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task.FormID != formID {
		return gorm.ErrRecordNotFound
	}
	if task.Status == datafillingdomain.TaskStatusStarted {
		return fmt.Errorf("task already started")
	}
	nextExecTime, err := s.scheduler.computeNextExecTime(task)
	if err != nil {
		return err
	}
	if err := s.scheduler.RegisterTask(ctx, task.ID); err != nil {
		return err
	}
	task.Status = datafillingdomain.TaskStatusStarted
	task.NextExecTime = nextExecTime
	task.UpdateTime = time.Now().UnixMilli()
	return s.taskRepo.UpdateTask(ctx, task)
}

func (s *DataFillingService) StopTask(ctx context.Context, formID, taskID int64) error {
	if s.taskRepo == nil || formID <= 0 || taskID <= 0 {
		return gorm.ErrInvalidData
	}
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task.FormID != formID {
		return gorm.ErrRecordNotFound
	}
	if s.scheduler != nil {
		s.scheduler.UnregisterTask(task.ID)
	}
	task.Status = datafillingdomain.TaskStatusStopped
	task.NextExecTime = 0
	task.UpdateTime = time.Now().UnixMilli()
	return s.taskRepo.UpdateTask(ctx, task)
}

func (s *DataFillingService) ExecuteNowTask(ctx context.Context, taskID int64) error {
	if s.scheduler == nil || taskID <= 0 {
		return gorm.ErrInvalidData
	}
	return s.scheduler.FireTask(ctx, taskID)
}

func (s *DataFillingService) TaskPageList(ctx context.Context, formID int64, page, pageSize int) (*datafillingdomain.TaskPageResponse, error) {
	if s.taskRepo == nil || formID <= 0 || page <= 0 || pageSize <= 0 {
		return nil, gorm.ErrInvalidData
	}
	rows, total, err := s.taskRepo.ListTasksByFormID(ctx, formID, page, pageSize)
	if err != nil {
		return nil, err
	}
	records := make([]*datafillingdomain.TaskInfoVO, 0, len(rows))
	for _, row := range rows {
		item, err := buildTaskInfoVO(row)
		if err != nil {
			return nil, err
		}
		records = append(records, item)
	}
	return &datafillingdomain.TaskPageResponse{Records: records, Total: total, Current: page, Size: pageSize}, nil
}

func (s *DataFillingService) DeleteTasks(ctx context.Context, formID int64, taskIDs []int64) error {
	if s.taskRepo == nil || s.subTaskRepo == nil || s.subInstanceRepo == nil || formID <= 0 || len(taskIDs) == 0 {
		return gorm.ErrInvalidData
	}
	for _, taskID := range taskIDs {
		task, err := s.taskRepo.GetTaskByID(ctx, taskID)
		if err != nil {
			return err
		}
		if task.FormID != formID {
			return gorm.ErrRecordNotFound
		}
		if s.scheduler != nil {
			s.scheduler.UnregisterTask(taskID)
		}
	}
	subTaskIDs, err := s.subTaskRepo.ListSubTaskIDsByTaskIDs(ctx, taskIDs)
	if err != nil {
		return err
	}
	if err := s.subInstanceRepo.DeleteSubInstancesByTaskIDs(ctx, taskIDs); err != nil {
		return err
	}
	if err := s.subTaskRepo.DeleteSubTasksByIDs(ctx, subTaskIDs); err != nil {
		return err
	}
	return s.taskRepo.DeleteTasksByIDs(ctx, taskIDs)
}

func (s *DataFillingService) SubTaskPageList(ctx context.Context, taskID int64, page, pageSize int) (*datafillingdomain.SubTaskPageResponse, error) {
	if s.subTaskRepo == nil || taskID <= 0 || page <= 0 || pageSize <= 0 {
		return nil, gorm.ErrInvalidData
	}
	rows, total, err := s.subTaskRepo.ListSubTasksByTaskID(ctx, taskID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &datafillingdomain.SubTaskPageResponse{Records: rows, Total: total, Current: page, Size: pageSize}, nil
}

func (s *DataFillingService) DeleteSubTasks(ctx context.Context, formID int64, subTaskIDs []int64) error {
	if s.subTaskRepo == nil || s.subInstanceRepo == nil || formID <= 0 || len(subTaskIDs) == 0 {
		return gorm.ErrInvalidData
	}
	if _, err := s.repo.GetByID(ctx, formID); err != nil {
		return err
	}
	if err := s.subInstanceRepo.DeleteSubInstancesByPIDs(ctx, subTaskIDs); err != nil {
		return err
	}
	return s.subTaskRepo.DeleteSubTasksByIDs(ctx, subTaskIDs)
}

func (s *DataFillingService) SubTaskUsersList(ctx context.Context, subTaskID int64, listType string) ([]*datafillingdomain.SubTaskUserItem, error) {
	if s.subInstanceRepo == nil || subTaskID <= 0 {
		return nil, gorm.ErrInvalidData
	}
	statusFilter, err := mapSubTaskUserStatusFilter(listType)
	if err != nil {
		return nil, err
	}
	rows, err := s.subInstanceRepo.ListSubInstancesByPID(ctx, subTaskID, statusFilter)
	if err != nil {
		return nil, err
	}
	result := make([]*datafillingdomain.SubTaskUserItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, &datafillingdomain.SubTaskUserItem{ID: row.ID, TaskID: row.TaskID, PID: row.PID, UID: row.UID, FormID: row.FormID, DataID: row.DataID, FinishTime: row.FinishTime, Status: row.Status})
	}
	return result, nil
}

func (s *DataFillingService) UserTaskPageList(ctx context.Context, userID int64, page, pageSize int, req *datafillingdomain.UserTaskPageRequest) ([]*datafillingdomain.UserTaskVO, int64, error) {
	if s.subInstanceRepo == nil || userID <= 0 || page <= 0 || pageSize <= 0 {
		return nil, 0, gorm.ErrInvalidData
	}
	rows, total, err := s.subInstanceRepo.ListSubInstancesByUID(ctx, userID, page, pageSize, req)
	if err != nil {
		return nil, 0, err
	}
	now := time.Now().UnixMilli()
	for _, row := range rows {
		if row != nil {
			row.Expired = row.EndTime > 0 && row.EndTime < now
		}
	}
	return rows, total, nil
}

func (s *DataFillingService) UserTaskTodoCount(ctx context.Context, userID int64) (int64, error) {
	if s.subInstanceRepo == nil || userID <= 0 {
		return 0, gorm.ErrInvalidData
	}
	return s.subInstanceRepo.CountOpenSubInstancesByUID(ctx, userID)
}

func (s *DataFillingService) GetUserTaskData(ctx context.Context, userID, subTaskID int64) (*datafillingdomain.UserTaskData, error) {
	if s.subInstanceRepo == nil || s.subTaskRepo == nil || s.taskRepo == nil || userID <= 0 || subTaskID <= 0 {
		return nil, gorm.ErrInvalidData
	}
	instances, err := s.subInstanceRepo.GetSubInstanceByPIDAndUID(ctx, subTaskID, userID)
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	subTask, err := s.getSubTaskByID(ctx, subTaskID)
	if err != nil {
		return nil, err
	}
	task, err := s.taskRepo.GetTaskByID(ctx, subTask.TaskID)
	if err != nil {
		return nil, err
	}
	form, err := s.repo.GetByID(ctx, task.FormID)
	if err != nil {
		return nil, err
	}
	dataIDs := make([]string, 0, len(instances))
	subInstances := make([]datafillingdomain.SubInstanceItem, 0, len(instances))
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		if trimmed := strings.TrimSpace(instance.DataID); trimmed != "" {
			dataIDs = append(dataIDs, trimmed)
		}
		subInstances = append(subInstances, buildSubInstanceItem(instance))
	}
	return &datafillingdomain.UserTaskData{
		FormID:         form.ID,
		FormTitle:      form.Name,
		DataIDs:        dataIDs,
		SubInstances:   subInstances,
		Form:           form.Forms,
		FormExtSetting: task.FormExtSetting,
		FillType:       task.FillType,
	}, nil
}

func (s *DataFillingService) SaveUserTaskData(ctx context.Context, userID, subTaskID int64, data []map[string]interface{}) error {
	if len(data) == 0 {
		return gorm.ErrInvalidData
	}
	instanceSet, form, err := s.loadUserTaskForm(ctx, userID, subTaskID)
	if err != nil {
		return err
	}
	for _, row := range data {
		if len(row) == 0 {
			return gorm.ErrInvalidData
		}
		if _, err := s.SaveRowData(ctx, form.ID, row, userID, ""); err != nil {
			return err
		}
	}
	return s.finishUserTaskIfOpen(ctx, instanceSet)
}

func (s *DataFillingService) AppendUserTaskData(ctx context.Context, userID, subTaskID int64, data []map[string]interface{}) error {
	if len(data) == 0 {
		return gorm.ErrInvalidData
	}
	instanceSet, form, err := s.loadUserTaskForm(ctx, userID, subTaskID)
	if err != nil {
		return err
	}
	_, db, err := s.loadFormAndDatasource(ctx, form.ID)
	if err != nil {
		return err
	}
	for _, row := range data {
		if len(row) == 0 {
			return gorm.ErrInvalidData
		}
		delete(row, "id")
		if err := s.ddlProvider.InsertRow(ctx, db, form.PhysicalTableName, row); err != nil {
			return err
		}
	}
	return s.finishUserTaskIfOpen(ctx, instanceSet)
}

func (s *DataFillingService) DeleteUserTaskData(ctx context.Context, userID, subTaskID int64, dataIDs []string) error {
	cleaned := cleanRowIDs(dataIDs)
	if len(cleaned) == 0 {
		return gorm.ErrInvalidData
	}
	_, form, err := s.loadUserTaskForm(ctx, userID, subTaskID)
	if err != nil {
		return err
	}
	for _, dataID := range cleaned {
		if err := s.DeleteRowData(ctx, form.ID, dataID, userID, ""); err != nil {
			return err
		}
	}
	return nil
}

func (s *DataFillingService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return gorm.ErrInvalidData
	}
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	children, err := s.repo.GetChildren(ctx, item.ID)
	if err != nil {
		return err
	}
	for _, child := range children {
		if err := s.Delete(ctx, child.ID); err != nil {
			return err
		}
	}
	if normalizeDataFillingNodeType(item.NodeType) == datafillingdomain.NodeTypeForm && item.DatasourceID > 0 && strings.TrimSpace(item.PhysicalTableName) != "" && s.ddlProvider != nil {
		db, err := s.GetDatasourceConnection(ctx, item.DatasourceID)
		if err != nil {
			return err
		}
		if err := s.ddlProvider.DropTable(ctx, db, item.PhysicalTableName); err != nil {
			return err
		}
	}
	return s.repo.DeleteByID(ctx, id)
}

func (s *DataFillingService) Rename(ctx context.Context, id int64, name string) (*datafillingdomain.DataFillingForm, error) {
	if id <= 0 || strings.TrimSpace(name) == "" {
		return nil, gorm.ErrInvalidData
	}
	if err := s.repo.Rename(ctx, id, strings.TrimSpace(name)); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *DataFillingService) Move(ctx context.Context, id int64, pid int64) (*datafillingdomain.DataFillingForm, error) {
	if id <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if id == pid {
		return nil, fmt.Errorf("cannot move node to itself")
	}
	if err := s.repo.Move(ctx, id, pid); err != nil {
		return nil, err
	}
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.recalculateChildrenLevels(ctx, item.ID, item.Level); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *DataFillingService) Tree(ctx context.Context, req *datafillingdomain.TreeRequest) (datafillingdomain.TreeResponse, error) {
	rows, err := s.repo.GetTree(ctx)
	if err != nil {
		return nil, err
	}
	root := buildDataFillingTree(rows)
	if req == nil || req.Keyword == nil || strings.TrimSpace(*req.Keyword) == "" {
		return root, nil
	}
	keyword := strings.ToLower(strings.TrimSpace(*req.Keyword))
	return filterDataFillingTree(root, keyword), nil
}

func (s *DataFillingService) ListDatasourceList(ctx context.Context) ([]datafillingdomain.DatasourceSummary, error) {
	_ = ctx
	return s.listDatasourceSummaries(true)
}

func (s *DataFillingService) ListDatasourceListAll(ctx context.Context) ([]datafillingdomain.DatasourceSummary, error) {
	_ = ctx
	return s.listDatasourceSummaries(false)
}

func (s *DataFillingService) GetBuiltInTables(ctx context.Context) ([]datafillingdomain.BuiltInTable, error) {
	_ = ctx
	return []datafillingdomain.BuiltInTable{}, nil
}

func (s *DataFillingService) GetDatasourceConnection(ctx context.Context, datasourceID int64) (*gorm.DB, error) {
	if s.datasourceConnProvider != nil {
		return s.datasourceConnProvider.GetDatasourceConnection(ctx, datasourceID)
	}
	_ = ctx
	if s.datasourceService == nil {
		return nil, fmt.Errorf("datasource service not configured")
	}
	ds, err := s.datasourceService.GetByID(datasourceID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(ds.Type), "mysql") {
		return nil, fmt.Errorf("unsupported datasource type: %s", ds.Type)
	}
	if ds.Configuration == nil || strings.TrimSpace(*ds.Configuration) == "" {
		return nil, fmt.Errorf("datasource configuration is empty")
	}
	cfg, err := decodeConfig(*ds.Configuration)
	if err != nil {
		return nil, err
	}
	mysqlCfg, err := buildMySQLPreviewConfig(cfg)
	if err != nil {
		return nil, err
	}
	return gorm.Open(mysql.Open(mysqlCfg.FormatDSN()), &gorm.Config{})
}

func (s *DataFillingService) createPhysicalTable(ctx context.Context, form *datafillingdomain.DataFillingForm) error {
	if s.ddlProvider == nil {
		return fmt.Errorf("ddl provider not configured")
	}
	fields := make([]datafillingdomain.ExtTableField, 0)
	if strings.TrimSpace(form.Forms) != "" {
		if err := json.Unmarshal([]byte(form.Forms), &fields); err != nil {
			return fmt.Errorf("parse form fields: %w", err)
		}
	}
	db, err := s.GetDatasourceConnection(ctx, form.DatasourceID)
	if err != nil {
		return err
	}
	return s.ddlProvider.CreateTable(ctx, db, form.PhysicalTableName, fields)
}

func (s *DataFillingService) loadFormAndDatasource(ctx context.Context, formID int64) (*datafillingdomain.DataFillingForm, *gorm.DB, error) {
	if formID <= 0 {
		return nil, nil, gorm.ErrInvalidData
	}
	if s.ddlProvider == nil {
		return nil, nil, fmt.Errorf("ddl provider not configured")
	}
	form, err := s.repo.GetByID(ctx, formID)
	if err != nil {
		return nil, nil, err
	}
	db, err := s.GetDatasourceConnection(ctx, form.DatasourceID)
	if err != nil {
		return nil, nil, err
	}
	return form, db, nil
}

func (s *DataFillingService) alterPhysicalTableForFieldChanges(ctx context.Context, form *datafillingdomain.DataFillingForm, oldForms, newForms string) error {
	if s.ddlProvider == nil || form == nil || strings.TrimSpace(form.PhysicalTableName) == "" || form.DatasourceID <= 0 {
		return nil
	}
	oldFields, err := parseExtTableFields(oldForms)
	if err != nil {
		return err
	}
	newFields, err := parseExtTableFields(newForms)
	if err != nil {
		return err
	}
	toAdd, toDrop := diffExtTableFields(oldFields, newFields)
	if len(toAdd) == 0 && len(toDrop) == 0 {
		return nil
	}
	db, err := s.GetDatasourceConnection(ctx, form.DatasourceID)
	if err != nil {
		return err
	}
	if len(toAdd) > 0 {
		if err := s.ddlProvider.AddTableColumns(ctx, db, form.PhysicalTableName, toAdd); err != nil {
			return err
		}
	}
	if len(toDrop) > 0 {
		if err := s.ddlProvider.DropTableColumns(ctx, db, form.PhysicalTableName, toDrop); err != nil {
			return err
		}
	}
	return nil
}

func (s *DataFillingService) writeCommitLog(ctx context.Context, formID int64, dataID string, operate int, userID int64, userName string, count int) error {
	if s.commitLogRepo == nil {
		return nil
	}
	return s.commitLogRepo.Create(ctx, &datafillingdomain.DfCommitLog{FormID: formID, DataID: dataID, Operate: operate, CommitBy: userID, Committer: userName, CommitTime: time.Now().UnixMilli(), Count: count})
}

type excelFieldMeta struct {
	DisplayName string
	ColumnName  string
	ID          string
}

func activeExcelFieldMetas(fields []datafillingdomain.ExtTableField) []excelFieldMeta {
	metas := make([]excelFieldMeta, 0, len(fields))
	for _, field := range fields {
		if field.Removed {
			continue
		}
		columnName := strings.TrimSpace(field.Settings.Mapping.ColumnName)
		if columnName == "" {
			continue
		}
		displayName := strings.TrimSpace(field.Settings.Name)
		if displayName == "" {
			displayName = columnName
		}
		metas = append(metas, excelFieldMeta{DisplayName: displayName, ColumnName: columnName, ID: strings.TrimSpace(field.ID)})
	}
	return metas
}

func writeExcelHeaders(file *excelize.File, metas []excelFieldMeta) error {
	sheetName := file.GetSheetName(file.GetActiveSheetIndex())
	for idx, meta := range metas {
		cell, err := excelize.CoordinatesToCellName(idx+1, 1)
		if err != nil {
			return err
		}
		if err := file.SetCellValue(sheetName, cell, meta.DisplayName); err != nil {
			return err
		}
	}
	return nil
}

func writeExcelDataRows(file *excelize.File, metas []excelFieldMeta, rows []map[string]interface{}) error {
	sheetName := file.GetSheetName(file.GetActiveSheetIndex())
	for rowIndex, row := range rows {
		for colIndex, meta := range metas {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+2)
			if err != nil {
				return err
			}
			if err := file.SetCellValue(sheetName, cell, stringifyExcelValue(row[meta.ColumnName])); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseFormFieldMaps(raw string) ([]map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return []map[string]interface{}{}, nil
	}
	result := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parse form fields: %w", err)
	}
	return result, nil
}

func parseExcelUploadRows(workbook *excelize.File, metas []excelFieldMeta) ([]datafillingdomain.RowDataDatum, error) {
	if workbook == nil {
		return nil, gorm.ErrInvalidData
	}
	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		return nil, gorm.ErrInvalidData
	}
	rows, err := workbook.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []datafillingdomain.RowDataDatum{}, nil
	}
	headerMap := buildExcelHeaderMap(rows[0], metas)
	result := make([]datafillingdomain.RowDataDatum, 0)
	for _, row := range rows[1:] {
		item := buildExcelRowData(row, headerMap)
		if item != nil {
			result = append(result, *item)
		}
	}
	return result, nil
}

func buildExcelHeaderMap(headers []string, metas []excelFieldMeta) map[int]excelFieldMeta {
	metaMap := make(map[string]excelFieldMeta, len(metas)*2)
	for _, meta := range metas {
		metaMap[strings.ToLower(strings.TrimSpace(meta.DisplayName))] = meta
		metaMap[strings.ToLower(strings.TrimSpace(meta.ColumnName))] = meta
	}
	result := make(map[int]excelFieldMeta)
	for idx, header := range headers {
		if meta, ok := metaMap[strings.ToLower(strings.TrimSpace(header))]; ok {
			result[idx] = meta
		}
	}
	return result
}

func buildExcelRowData(row []string, headerMap map[int]excelFieldMeta) *datafillingdomain.RowDataDatum {
	data := make(map[string]interface{})
	rowID := ""
	hasValue := false
	for idx, meta := range headerMap {
		if idx >= len(row) {
			continue
		}
		value := strings.TrimSpace(row[idx])
		if value == "" {
			continue
		}
		hasValue = true
		data[meta.ColumnName] = value
		if meta.ColumnName == "id" {
			rowID = value
		}
	}
	if !hasValue {
		return nil
	}
	insert := strings.TrimSpace(rowID) == ""
	if rowID != "" {
		data["id"] = rowID
	}
	return &datafillingdomain.RowDataDatum{ID: rowID, Data: data, Insert: insert}
}

func (s *DataFillingService) mustLoadExcelUpload(uploadID string) (*datafillingdomain.DfExcelData, error) {
	if strings.TrimSpace(uploadID) == "" {
		return nil, gorm.ErrInvalidData
	}
	data, ok := s.excelUploadSession.Load(uploadID)
	if !ok || data == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return data, nil
}

func (s *DataFillingService) persistUploadedRows(ctx context.Context, formID int64, rows []datafillingdomain.RowDataDatum, userID int64, userName string) error {
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := s.persistUploadedRow(ctx, form, db, row, userID, userName); err != nil {
			return err
		}
	}
	return nil
}

func (s *DataFillingService) persistUploadedRow(ctx context.Context, form *datafillingdomain.DataFillingForm, db *gorm.DB, row datafillingdomain.RowDataDatum, userID int64, userName string) error {
	if form == nil || len(row.Data) == 0 {
		return nil
	}
	rowData := copyMap(row.Data)
	rowID := strings.TrimSpace(row.ID)
	if row.Insert || rowID == "" || rowID == nilStringValue {
		delete(rowData, "id")
		if err := s.ddlProvider.InsertRow(ctx, db, form.PhysicalTableName, rowData); err != nil {
			return err
		}
		return s.writeCommitLog(ctx, form.ID, strings.TrimSpace(fmt.Sprint(rowData["id"])), 1, userID, userName, 1)
	}
	if err := s.ddlProvider.UpdateRow(ctx, db, form.PhysicalTableName, rowID, rowData); err != nil {
		return err
	}
	return s.writeCommitLog(ctx, form.ID, rowID, 2, userID, userName, 1)
}

func validateExtraDetailsRequest(req *datafillingdomain.ExtraDetailsRequest) (int64, string, string, []string, string, error) {
	if req == nil {
		return 0, "", "", nil, "", gorm.ErrInvalidData
	}
	datasourceID, err := strconv.ParseInt(strings.TrimSpace(req.OptionDatasource), 10, 64)
	if err != nil || datasourceID <= 0 {
		return 0, "", "", nil, "", gorm.ErrInvalidData
	}
	tableName := strings.TrimSpace(req.OptionTable)
	optionColumn := strings.TrimSpace(req.OptionColumn)
	if !isValidDDLIdentifier(tableName) || !isValidDDLIdentifier(optionColumn) {
		return 0, "", "", nil, "", gorm.ErrInvalidData
	}
	extraColumns := make([]string, 0, len(req.ExtraColumns))
	for _, column := range req.ExtraColumns {
		trimmed := strings.TrimSpace(column)
		if trimmed == "" {
			continue
		}
		if !isValidDDLIdentifier(trimmed) {
			return 0, "", "", nil, "", gorm.ErrInvalidData
		}
		extraColumns = append(extraColumns, trimmed)
	}
	return datasourceID, tableName, optionColumn, extraColumns, strings.TrimSpace(req.Value), nil
}

func validateDatasourceOptionsRequest(req *datafillingdomain.DatasourceOptionsRequest) (string, string, string, error) {
	if req == nil {
		return "", "", "", gorm.ErrInvalidData
	}
	tableName := strings.TrimSpace(req.OptionTable)
	optionColumn := strings.TrimSpace(req.OptionColumn)
	orderColumn := strings.TrimSpace(req.OptionOrder)
	if orderColumn == "" {
		orderColumn = optionColumn
	}
	if !isValidDDLIdentifier(tableName) || !isValidDDLIdentifier(optionColumn) || !isValidDDLIdentifier(orderColumn) {
		return "", "", "", gorm.ErrInvalidData
	}
	return tableName, optionColumn, orderColumn, nil
}

func quoteIdentifier(identifier string) string {
	return "`" + strings.TrimSpace(identifier) + "`"
}

func quoteIdentifiers(columns []string) []string {
	result := make([]string, 0, len(columns))
	for _, column := range columns {
		result = append(result, quoteIdentifier(column))
	}
	return result
}

func flattenExtraDetails(rows []map[string]interface{}, extraColumns []string) []*datafillingdomain.ExtraDetails {
	result := make([]*datafillingdomain.ExtraDetails, 0, len(rows)*len(extraColumns))
	for _, row := range rows {
		for _, column := range extraColumns {
			result = append(result, &datafillingdomain.ExtraDetails{Name: column, Value: stringifyExcelValue(row[column])})
		}
	}
	return result
}

func stringifyExcelValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func copyMap(input map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func (s *DataFillingService) loadAllFormRows(ctx context.Context, db *gorm.DB, tableName string) ([]map[string]interface{}, error) {
	total, err := s.ddlProvider.CountRows(ctx, db, tableName, "", nil)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return []map[string]interface{}{}, nil
	}
	return s.ddlProvider.SearchRows(ctx, db, tableName, "", nil, total, 0)
}

func parseExtTableFields(raw string) ([]datafillingdomain.ExtTableField, error) {
	if strings.TrimSpace(raw) == "" {
		return []datafillingdomain.ExtTableField{}, nil
	}
	fields := make([]datafillingdomain.ExtTableField, 0)
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return nil, fmt.Errorf("parse form fields: %w", err)
	}
	return fields, nil
}

func diffExtTableFields(oldFields, newFields []datafillingdomain.ExtTableField) ([]datafillingdomain.ExtTableField, []string) {
	oldMap := make(map[string]datafillingdomain.ExtTableField)
	newMap := make(map[string]datafillingdomain.ExtTableField)
	for _, field := range oldFields {
		column := strings.TrimSpace(field.Settings.Mapping.ColumnName)
		if column != "" && !field.Removed {
			oldMap[column] = field
		}
	}
	for _, field := range newFields {
		column := strings.TrimSpace(field.Settings.Mapping.ColumnName)
		if column != "" && !field.Removed {
			newMap[column] = field
		}
	}
	toAdd := make([]datafillingdomain.ExtTableField, 0)
	toDrop := make([]string, 0)
	for column, field := range newMap {
		if _, ok := oldMap[column]; !ok {
			toAdd = append(toAdd, field)
		}
	}
	for column := range oldMap {
		if _, ok := newMap[column]; !ok {
			toDrop = append(toDrop, column)
		}
	}
	sort.Slice(toAdd, func(i, j int) bool {
		return toAdd[i].Settings.Mapping.ColumnName < toAdd[j].Settings.Mapping.ColumnName
	})
	sort.Strings(toDrop)
	return toAdd, toDrop
}

func cleanRowIDs(ids []string) []string {
	cleaned := make([]string, 0, len(ids))
	for _, id := range ids {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

func (s *DataFillingService) getSubTaskByID(ctx context.Context, subTaskID int64) (*datafillingdomain.DataFillingSubTask, error) {
	return s.subTaskRepo.GetSubTaskByID(ctx, subTaskID)
}

type userTaskInstanceSet struct {
	current   *datafillingdomain.DataFillingSubInstance
	instances []*datafillingdomain.DataFillingSubInstance
	subTaskID int64
}

func (s *DataFillingService) loadUserTaskForm(ctx context.Context, userID, subTaskID int64) (*userTaskInstanceSet, *datafillingdomain.DataFillingForm, error) {
	if s.subInstanceRepo == nil || s.taskRepo == nil || userID <= 0 || subTaskID <= 0 {
		return nil, nil, gorm.ErrInvalidData
	}
	instances, err := s.subInstanceRepo.GetSubInstanceByPIDAndUID(ctx, subTaskID, userID)
	if err != nil {
		return nil, nil, err
	}
	if len(instances) == 0 {
		return nil, nil, gorm.ErrRecordNotFound
	}
	current := firstActiveSubInstance(instances)
	if current == nil {
		current = instances[0]
	}
	task, err := s.taskRepo.GetTaskByID(ctx, current.TaskID)
	if err != nil {
		return nil, nil, err
	}
	form, err := s.repo.GetByID(ctx, task.FormID)
	if err != nil {
		return nil, nil, err
	}
	return &userTaskInstanceSet{current: current, instances: instances, subTaskID: subTaskID}, form, nil
}

func firstActiveSubInstance(instances []*datafillingdomain.DataFillingSubInstance) *datafillingdomain.DataFillingSubInstance {
	for _, instance := range instances {
		if instance != nil && instance.Status == datafillingdomain.SubInstanceStatusOpen {
			return instance
		}
	}
	return nil
}

func (s *DataFillingService) finishUserTaskIfOpen(ctx context.Context, instanceSet *userTaskInstanceSet) error {
	if instanceSet == nil || instanceSet.current == nil {
		return gorm.ErrInvalidData
	}
	if instanceSet.current.Status != datafillingdomain.SubInstanceStatusOpen {
		return nil
	}
	finishTime := time.Now().UnixMilli()
	if err := s.subInstanceRepo.UpdateSubInstanceStatus(ctx, instanceSet.current.ID, datafillingdomain.SubInstanceStatusFinished, finishTime); err != nil {
		return err
	}
	if s.subTaskRepo != nil {
		if err := s.subTaskRepo.DecrementSubTaskUnfinishedCount(ctx, instanceSet.subTaskID); err != nil {
			return err
		}
	}
	instanceSet.current.Status = datafillingdomain.SubInstanceStatusFinished
	instanceSet.current.FinishTime = finishTime
	return nil
}

func buildSubInstanceItem(instance *datafillingdomain.DataFillingSubInstance) datafillingdomain.SubInstanceItem {
	item := datafillingdomain.SubInstanceItem{
		ID:     instance.ID,
		TaskID: instance.TaskID,
		PID:    instance.PID,
		UID:    instance.UID,
		FormID: instance.FormID,
		DataID: instance.DataID,
		Status: instance.Status,
	}
	if instance.FinishTime > 0 {
		finishTime := instance.FinishTime
		item.FinishTime = &finishTime
	}
	return item
}

func (s *DataFillingService) resolveLevel(ctx context.Context, pid int64) (int, error) {
	if pid <= 0 {
		return 0, nil
	}
	parent, err := s.repo.GetByID(ctx, pid)
	if err != nil {
		return 0, err
	}
	return parent.Level + 1, nil
}

func (s *DataFillingService) recalculateChildrenLevels(ctx context.Context, pid int64, parentLevel int) error {
	children, err := s.repo.GetChildren(ctx, pid)
	if err != nil {
		return err
	}
	for _, child := range children {
		child.Level = parentLevel + 1
		if err := s.repo.Update(ctx, child); err != nil {
			return err
		}
		if err := s.recalculateChildrenLevels(ctx, child.ID, child.Level); err != nil {
			return err
		}
	}
	return nil
}

func (s *DataFillingService) listDatasourceSummaries(enabledOnly bool) ([]datafillingdomain.DatasourceSummary, error) {
	if s.datasourceService == nil {
		return []datafillingdomain.DatasourceSummary{}, nil
	}
	list, err := s.datasourceService.Tree(&datasourcedomain.ListRequest{})
	if err != nil {
		return nil, err
	}
	result := make([]datafillingdomain.DatasourceSummary, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		enabled := item.EnableDataFill != nil && *item.EnableDataFill
		if enabledOnly && !enabled {
			continue
		}
		pid := int64(0)
		if item.PID != nil {
			pid = *item.PID
		}
		status := ""
		if item.Status != nil {
			status = *item.Status
		}
		result = append(result, datafillingdomain.DatasourceSummary{ID: item.ID, PID: pid, Name: item.Name, Type: item.Type, Status: status, EnableDataFill: enabled})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func validateDataFillingCreateRequest(req *datafillingdomain.CreateFormRequest) error {
	if req == nil || strings.TrimSpace(req.Name) == "" {
		return gorm.ErrInvalidData
	}
	if normalizeDataFillingNodeType(req.NodeType) == datafillingdomain.NodeTypeFolder {
		return nil
	}
	if strings.TrimSpace(req.TableName) == "" || req.DatasourceID <= 0 {
		return gorm.ErrInvalidData
	}
	return nil
}

func normalizeDataFillingNodeType(nodeType string) string {
	if strings.EqualFold(strings.TrimSpace(nodeType), datafillingdomain.NodeTypeFolder) {
		return datafillingdomain.NodeTypeFolder
	}
	return datafillingdomain.NodeTypeForm
}

func buildDataFillingTree(rows []*datafillingdomain.DataFillingForm) []*datafillingdomain.TreeNode {
	nodeMap := make(map[int64]*datafillingdomain.TreeNode, len(rows))
	result := make([]*datafillingdomain.TreeNode, 0)
	for _, row := range rows {
		if row == nil {
			continue
		}
		nodeMap[row.ID] = &datafillingdomain.TreeNode{ID: row.ID, Name: row.Name, PID: row.PID, NodeType: normalizeDataFillingNodeType(row.NodeType)}
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		node := nodeMap[row.ID]
		if row.PID <= 0 {
			result = append(result, node)
			continue
		}
		if parent, ok := nodeMap[row.PID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			result = append(result, node)
		}
	}
	return result
}

func filterDataFillingTree(nodes []*datafillingdomain.TreeNode, keyword string) []*datafillingdomain.TreeNode {
	filtered := make([]*datafillingdomain.TreeNode, 0)
	for _, node := range nodes {
		children := filterDataFillingTree(node.Children, keyword)
		if strings.Contains(strings.ToLower(node.Name), keyword) || len(children) > 0 {
			copied := *node
			copied.Children = children
			filtered = append(filtered, &copied)
		}
	}
	return filtered
}

func buildTaskInfoVO(task *datafillingdomain.DataFillingTask) (*datafillingdomain.TaskInfoVO, error) {
	if task == nil {
		return nil, gorm.ErrInvalidData
	}
	reciFlagList, err := parseJSONIntList(task.ReciFlagList)
	if err != nil {
		return nil, err
	}
	uidList, err := parseJSONInt64List(task.UIDList)
	if err != nil {
		return nil, err
	}
	ridList, err := parseJSONInt64List(task.RIDList)
	if err != nil {
		return nil, err
	}
	return &datafillingdomain.TaskInfoVO{
		ID:                   task.ID,
		FormID:               task.FormID,
		Name:                 task.Name,
		ReciFlagList:         reciFlagList,
		UIDList:              uidList,
		RIDList:              ridList,
		FillType:             task.FillType,
		FitType:              task.FitType,
		FitColumn:            task.FitColumn,
		RateType:             task.RateType,
		RateVal:              task.RateVal,
		OneTimeType:          task.OneTimeType,
		StartTime:            task.StartTime,
		EndTime:              task.EndTime,
		PublishRangeTime:     task.PublishRangeTime,
		PublishRangeTimeType: task.PublishRangeTimeType,
		Status:               task.Status,
		LastExecStatus:       task.LastExecStatus,
		LastExecTime:         task.LastExecTime,
		NextExecTime:         task.NextExecTime,
		CreateBy:             task.CreateBy,
		CreateTime:           task.CreateTime,
		UpdateBy:             task.UpdateBy,
		UpdateTime:           task.UpdateTime,
		FormExtSetting:       task.FormExtSetting,
		FormFilterSetting:    task.FormFilterSetting,
	}, nil
}

func parseJSONIntList(raw string) ([]int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []int{}, nil
	}
	values := make([]int, 0)
	if err := json.Unmarshal([]byte(trimmed), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func mapSubTaskUserStatusFilter(listType string) (*int, error) {
	switch strings.ToLower(strings.TrimSpace(listType)) {
	case "", "all":
		return nil, nil
	case "unfinished", "open", "todo":
		status := datafillingdomain.SubInstanceStatusOpen
		return &status, nil
	case "finished", "done":
		status := datafillingdomain.SubInstanceStatusFinished
		return &status, nil
	default:
		return nil, gorm.ErrInvalidData
	}
}
