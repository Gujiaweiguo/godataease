package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/syncmodule"
	"dataease/backend/internal/integration/seatunnel"
	"dataease/backend/internal/repository"
)

type SyncService struct {
	syncRepo          *repository.SyncRepository
	datasourceRepo    *repository.DatasourceRepository
	datasourceService *DatasourceService
}

func NewSyncService(syncRepo *repository.SyncRepository, datasourceRepo *repository.DatasourceRepository, datasourceService *DatasourceService) *SyncService {
	return &SyncService{syncRepo: syncRepo, datasourceRepo: datasourceRepo, datasourceService: datasourceService}
}

func (s *SyncService) SourcePager(page int, size int, req *datasource.ListRequest) (*syncmodule.PageResult[syncmodule.SyncDatasourceDTO], error) {
	return s.datasourcePager(page, size, req)
}

func (s *SyncService) TargetPager(page int, size int, req *datasource.ListRequest) (*syncmodule.PageResult[syncmodule.SyncDatasourceDTO], error) {
	return s.datasourcePager(page, size, req)
}

func (s *SyncService) datasourcePager(page int, size int, req *datasource.ListRequest) (*syncmodule.PageResult[syncmodule.SyncDatasourceDTO], error) {
	if req == nil {
		req = &datasource.ListRequest{}
	}
	req.Current = page
	req.Size = size
	result, err := s.datasourceService.List(req)
	if err != nil {
		return nil, err
	}
	items := make([]syncmodule.SyncDatasourceDTO, 0, len(result.List))
	for _, item := range result.List {
		items = append(items, toSyncDatasourceDTO(item))
	}
	return &syncmodule.PageResult[syncmodule.SyncDatasourceDTO]{Records: items, Total: result.Total, Current: result.Current, Size: result.Size}, nil
}

func (s *SyncService) SaveDatasource(req *datasource.WriteRequest) (*syncmodule.SyncDatasourceDTO, error) {
	item, err := s.datasourceService.Save(req)
	if err != nil {
		return nil, err
	}
	result := toSyncDatasourceDTO(item)
	return &result, nil
}

func (s *SyncService) UpdateDatasource(req *datasource.WriteRequest) (*syncmodule.SyncDatasourceDTO, error) {
	item, err := s.datasourceService.Update(req)
	if err != nil {
		return nil, err
	}
	result := toSyncDatasourceDTO(item)
	return &result, nil
}

func (s *SyncService) GetDatasource(id int64) (*syncmodule.SyncDatasourceDTO, error) {
	item, err := s.datasourceService.GetByID(id)
	if err != nil {
		return nil, err
	}
	result := toSyncDatasourceDTO(item)
	return &result, nil
}

func (s *SyncService) DeleteDatasource(id int64) error {
	return s.datasourceService.Delete(id)
}

func (s *SyncService) BatchDeleteDatasource(ids []int64) error {
	for _, id := range ids {
		if err := s.datasourceService.Delete(id); err != nil {
			return err
		}
	}
	return nil
}

func (s *SyncService) ValidateDatasource(req *datasource.ValidateRequest) (*datasource.ValidateResponse, error) {
	return s.datasourceService.Validate(req)
}

func (s *SyncService) ValidateDatasourceByID(id int64) (*datasource.ValidateResponse, error) {
	return s.datasourceService.ValidateByID(id)
}

func (s *SyncService) GetSchemas() ([]string, error) {
	return s.datasourceService.GetSchema()
}

func (s *SyncService) GetDatasourceFields(req *syncmodule.SyncDatasourceFieldRequest) (map[string]interface{}, error) {
	if req == nil {
		return map[string]interface{}{"fieldList": []syncmodule.TableField{}, "targetFieldTypeList": []string{}}, nil
	}
	tableReq := &datasource.TableRequest{TableName: strings.TrimSpace(req.Table)}
	if id, err := parseStringID(req.ID); err == nil {
		tableReq.DatasourceID = id
	}
	fields, err := s.datasourceService.GetTableField(tableReq)
	if err != nil {
		return nil, err
	}
	result := make([]syncmodule.TableField, 0, len(fields))
	typeSet := make(map[string]struct{})
	types := make([]string, 0)
	for idx, field := range fields {
		mapped := syncmodule.TableField{
			ID:             strconv.Itoa(idx + 1),
			FieldSource:    field.OriginName,
			FieldName:      field.Name,
			Remarks:        field.OriginName,
			FieldType:      field.Type,
			FieldSize:      0,
			FieldPrecision: 0,
		}
		result = append(result, mapped)
		if field.Type != "" {
			if _, ok := typeSet[field.Type]; !ok {
				typeSet[field.Type] = struct{}{}
				types = append(types, field.Type)
			}
		}
	}
	return map[string]interface{}{"fieldList": result, "targetFieldTypeList": types}, nil
}

func (s *SyncService) ListDatasourceByType(dsType string) ([]syncmodule.SyncDatasourceDTO, error) {
	items, err := s.datasourceRepo.ListByType(dsType, nil)
	if err != nil {
		return nil, err
	}
	result := make([]syncmodule.SyncDatasourceDTO, 0, len(items))
	for _, item := range items {
		result = append(result, toSyncDatasourceDTO(item))
	}
	return result, nil
}

func (s *SyncService) ListDatasourceTables(id int64) ([]syncmodule.DBTableDTO, error) {
	tables, err := s.datasourceService.GetTables(&datasource.TableRequest{DatasourceID: id})
	if err != nil {
		return nil, err
	}
	result := make([]syncmodule.DBTableDTO, 0, len(tables))
	for _, item := range tables {
		remark := item.TableName
		if strings.TrimSpace(item.Name) != "" {
			remark = item.Name
		}
		result = append(result, syncmodule.DBTableDTO{
			DatasourceID: strconv.FormatInt(item.DatasourceID, 10),
			Name:         item.TableName,
			Remark:       remark,
			EnableCheck:  true,
			DatasetPath:  "",
		})
	}
	return result, nil
}

func (s *SyncService) LatestUse(_ string, creator string) ([]string, error) {
	return s.datasourceService.LatestTypes(creator)
}

func (s *SyncService) TaskPager(page int, size int, req *syncmodule.TaskGridRequest) (*syncmodule.PageResult[syncmodule.TaskInfo], error) {
	rows, total, err := s.syncRepo.ListTasks(page, size, req)
	if err != nil {
		return nil, err
	}
	result := make([]syncmodule.TaskInfo, 0, len(rows))
	for _, row := range rows {
		result = append(result, toTaskInfo(row))
	}
	return &syncmodule.PageResult[syncmodule.TaskInfo]{Records: result, Total: total, Current: page, Size: size}, nil
}

func (s *SyncService) GetTask(id int64) (*syncmodule.TaskInfo, error) {
	row, err := s.syncRepo.GetTask(id)
	if err != nil {
		return nil, err
	}
	result := toTaskInfo(*row)
	return &result, nil
}

func (s *SyncService) AddTask(req *syncmodule.TaskInfo) error {
	row, err := toTaskRow(req, nil)
	if err != nil {
		return err
	}
	return s.syncRepo.CreateTask(row)
}

func (s *SyncService) UpdateTask(req *syncmodule.TaskInfo) error {
	if req == nil {
		return fmt.Errorf("task is required")
	}
	id, err := parseStringID(req.ID)
	if err != nil {
		return err
	}
	existing, err := s.syncRepo.GetTask(id)
	if err != nil {
		return err
	}
	row, err := toTaskRow(req, existing)
	if err != nil {
		return err
	}
	row.ID = existing.ID
	return s.syncRepo.UpdateTask(row)
}

func (s *SyncService) RemoveTask(id int64) error {
	if err := s.syncRepo.DeleteTaskLogsByTaskID(id); err != nil {
		return err
	}
	return s.syncRepo.DeleteTask(id)
}

func (s *SyncService) BatchDeleteTasks(ids []int64) error {
	for _, id := range ids {
		if err := s.syncRepo.DeleteTaskLogsByTaskID(id); err != nil {
			return err
		}
	}
	return s.syncRepo.BatchDeleteTasks(ids)
}

func (s *SyncService) ExecuteTask(id int64) (map[string]interface{}, error) {
	task, err := s.syncRepo.GetTask(id)
	if err != nil {
		return nil, err
	}
	info := toTaskInfo(*task)
	syncType := "datasource"
	req := map[string]string{"datasourceId": info.Source.DatasourceID, "name": info.Name}
	if strings.TrimSpace(info.Target.TableName) != "" {
		syncType = "table"
		req["tableName"] = info.Target.TableName
	}
	if syncType == "table" {
		result, execErr := s.datasourceService.SyncAPITable(req)
		_ = s.markTaskExecuted(task, result, execErr)
		return result, execErr
	}
	result, execErr := s.datasourceService.SyncAPIDs(req)
	_ = s.markTaskExecuted(task, result, execErr)
	return result, execErr
}

func (s *SyncService) StartTask(id int64) error {
	task, err := s.syncRepo.GetTask(id)
	if err != nil {
		return err
	}
	task.TaskStatus = seatunnel.StatusRunning
	return s.syncRepo.UpdateTask(task)
}

func (s *SyncService) StopTask(id int64) error {
	task, err := s.syncRepo.GetTask(id)
	if err != nil {
		return err
	}
	persisted := syncmodule.TaskPersistedData{}
	if strings.TrimSpace(task.ExtraData) != "" {
		_ = json.Unmarshal([]byte(task.ExtraData), &persisted)
	}
	if strings.TrimSpace(persisted.LastTaskID) != "" {
		if cancelErr := s.datasourceService.CancelSyncTask(persisted.LastTaskID); cancelErr != nil {
			return cancelErr
		}
	}
	task.TaskStatus = seatunnel.StatusCancelled
	task.LastExecStatus = seatunnel.StatusCancelled
	return s.syncRepo.UpdateTask(task)
}

func (s *SyncService) TaskLogPager(page int, size int, req *syncmodule.TaskLogGridRequest) (*syncmodule.PageResult[syncmodule.TaskLog], error) {
	rows, total, err := s.syncRepo.ListTaskLogs(page, size, req)
	if err != nil {
		return nil, err
	}
	result := make([]syncmodule.TaskLog, 0, len(rows))
	for _, row := range rows {
		result = append(result, toTaskLog(row))
	}
	return &syncmodule.PageResult[syncmodule.TaskLog]{Records: result, Total: total, Current: page, Size: size}, nil
}

func (s *SyncService) TaskLogDetail(id int64, fromLineNum int) (*syncmodule.LogResult, error) {
	logRow, err := s.syncRepo.GetTaskLog(id)
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(logRow.Info)
	toLineNum := fromLineNum
	if content != "" {
		toLineNum = fromLineNum + len(strings.Split(content, "\n")) - 1
	}
	return &syncmodule.LogResult{FromLineNum: fromLineNum, ToLineNum: toLineNum, LogContent: content, IsEnd: true}, nil
}

func (s *SyncService) DeleteTaskLog(id int64) error {
	return s.syncRepo.DeleteTaskLog(id)
}

func (s *SyncService) ClearTaskLog(req *syncmodule.TaskLog) error {
	if req == nil || strings.TrimSpace(req.JobID) == "" {
		return s.syncRepo.ClearTaskLogs(nil)
	}
	jobID, err := parseStringID(req.JobID)
	if err != nil {
		return err
	}
	return s.syncRepo.ClearTaskLogs(&jobID)
}

func (s *SyncService) TerminateTaskByLogID(id int64) error {
	logRow, err := s.syncRepo.GetTaskLog(id)
	if err != nil {
		return err
	}
	if logRow.TaskID > 0 {
		return s.datasourceService.CancelSyncTask(strconv.FormatInt(logRow.TaskID, 10))
	}
	return nil
}

func (s *SyncService) ResourceCount() (*syncmodule.ResourceCount, error) {
	jobCount, err := s.syncRepo.CountTasks()
	if err != nil {
		return nil, err
	}
	datasourceCount, err := s.syncRepo.CountDatasources()
	if err != nil {
		return nil, err
	}
	jobLogCount, err := s.syncRepo.CountTaskLogs()
	if err != nil {
		return nil, err
	}
	return &syncmodule.ResourceCount{JobCount: jobCount, DatasourceCount: datasourceCount, JobLogCount: jobLogCount}, nil
}

func (s *SyncService) LogChartData() (map[string]interface{}, error) {
	points, err := s.syncRepo.ListLogChartData(7)
	if err != nil {
		return nil, err
	}
	xAxis := make([]string, 0, len(points))
	values := make([]int64, 0, len(points))
	for _, point := range points {
		xAxis = append(xAxis, point.Day)
		values = append(values, point.Count)
	}
	return map[string]interface{}{"x": xAxis, "y": values, "values": points}, nil
}

func (s *SyncService) markTaskExecuted(task *auto.CoreDatasourceTask, result map[string]interface{}, execErr error) error {
	if task == nil {
		return nil
	}
	persisted := syncmodule.TaskPersistedData{}
	if strings.TrimSpace(task.ExtraData) != "" {
		_ = json.Unmarshal([]byte(task.ExtraData), &persisted)
	}
	if result != nil {
		if taskID, ok := result["taskId"]; ok && taskID != nil {
			persisted.LastTaskID = fmt.Sprint(taskID)
		}
	}
	if encoded, err := json.Marshal(persisted); err == nil {
		task.ExtraData = string(encoded)
	}
	now := time.Now().UnixMilli()
	task.LastExecTime = now
	if execErr != nil {
		task.LastExecStatus = seatunnel.StatusFailed
		task.TaskStatus = seatunnel.StatusFailed
	} else {
		task.LastExecStatus = seatunnel.StatusRunning
		task.TaskStatus = seatunnel.StatusRunning
	}
	return s.syncRepo.UpdateTask(task)
}

func toSyncDatasourceDTO(item *datasource.CoreDatasource) syncmodule.SyncDatasourceDTO {
	if item == nil {
		return syncmodule.SyncDatasourceDTO{}
	}
	createTime := int64(0)
	if item.CreateTime != nil {
		createTime = *item.CreateTime
	}
	updateTime := int64(0)
	if item.UpdateTime != nil {
		updateTime = *item.UpdateTime
	}
	status := ""
	if item.Status != nil {
		status = *item.Status
	}
	configuration := ""
	if item.Configuration != nil {
		configuration = *item.Configuration
	}
	return syncmodule.SyncDatasourceDTO{
		ID:            strconv.FormatInt(item.ID, 10),
		Name:          item.Name,
		Desc:          valueOrEmpty(item.Description),
		Type:          item.Type,
		Configuration: configuration,
		CreateTime:    createTime,
		UpdateTime:    updateTime,
		Status:        status,
	}
}

func toTaskRow(req *syncmodule.TaskInfo, existing *auto.CoreDatasourceTask) (*auto.CoreDatasourceTask, error) {
	if req == nil {
		return nil, fmt.Errorf("task is required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" && existing == nil {
		return nil, fmt.Errorf("task name is required")
	}
	row := &auto.CoreDatasourceTask{}
	if existing != nil {
		*row = *existing
	}
	if err := applyTaskDatasource(row, req, existing); err != nil {
		return nil, err
	}
	applyTaskIdentity(row, req, name)
	if err := applyTaskSchedule(row, req); err != nil {
		return nil, err
	}
	applyTaskRuntime(row, req)
	now := time.Now().UnixMilli()
	if row.CreateTime == 0 {
		row.CreateTime = now
	}
	return row, nil
}

func applyTaskDatasource(row *auto.CoreDatasourceTask, req *syncmodule.TaskInfo, existing *auto.CoreDatasourceTask) error {
	dsID, err := parseStringID(req.Source.DatasourceID)
	if err != nil {
		if existing == nil {
			return fmt.Errorf("source datasourceId is required")
		}
		return nil
	}
	row.DsID = dsID
	return nil
}

func applyTaskIdentity(row *auto.CoreDatasourceTask, req *syncmodule.TaskInfo, name string) {
	if name != "" {
		row.Name = name
	}
	if strings.TrimSpace(req.TaskKey) != "" {
		row.UpdateType = req.TaskKey
		return
	}
	if row.UpdateType == "" {
		row.UpdateType = "sync"
	}
}

func applyTaskSchedule(row *auto.CoreDatasourceTask, req *syncmodule.TaskInfo) error {
	persisted := syncmodule.TaskPersistedData{
		TaskKey:                req.TaskKey,
		Desc:                   req.Desc,
		ExecutorTimeout:        req.ExecutorTimeout,
		ExecutorFailRetryCount: req.ExecutorFailRetryCount,
		SchedulerType:          req.SchedulerType,
		SchedulerConf:          req.SchedulerConf,
		SchedulerOption:        req.SchedulerOption,
		Source:                 req.Source,
		Target:                 req.Target,
		StartTime:              req.StartTime,
		StopTime:               req.StopTime,
	}
	extraData, err := json.Marshal(persisted)
	if err != nil {
		return err
	}
	row.SyncRate = schedulerRate(req.SchedulerType)
	row.Cron = req.SchedulerConf
	row.SimpleCronValue = int64(req.SchedulerOption.Interval)
	row.SimpleCronType = req.SchedulerOption.Unit
	row.ExtraData = string(extraData)
	return nil
}

func applyTaskRuntime(row *auto.CoreDatasourceTask, req *syncmodule.TaskInfo) {
	if ts := parseMillis(req.StartTime); ts > 0 {
		row.StartTime = ts
	}
	if ts := parseMillis(req.StopTime); ts > 0 {
		row.EndTime = ts
	}
	if strings.TrimSpace(req.Status) != "" {
		row.TaskStatus = req.Status
		return
	}
	if row.TaskStatus == "" {
		row.TaskStatus = seatunnel.StatusPending
	}
}

func toTaskInfo(row auto.CoreDatasourceTask) syncmodule.TaskInfo {
	persisted := syncmodule.TaskPersistedData{}
	if strings.TrimSpace(row.ExtraData) != "" {
		_ = json.Unmarshal([]byte(row.ExtraData), &persisted)
	}
	startTime := persisted.StartTime
	if startTime == "" && row.StartTime > 0 {
		startTime = strconv.FormatInt(row.StartTime, 10)
	}
	stopTime := persisted.StopTime
	if stopTime == "" && row.EndTime > 0 {
		stopTime = strconv.FormatInt(row.EndTime, 10)
	}
	return syncmodule.TaskInfo{
		ID:                     strconv.FormatInt(row.ID, 10),
		Name:                   row.Name,
		SchedulerType:          firstNonEmpty(persisted.SchedulerType, schedulerType(row)),
		SchedulerConf:          firstNonEmpty(persisted.SchedulerConf, row.Cron),
		SchedulerOption:        persisted.SchedulerOption,
		TaskKey:                firstNonEmpty(persisted.TaskKey, row.UpdateType),
		Desc:                   persisted.Desc,
		ExecutorTimeout:        persisted.ExecutorTimeout,
		ExecutorFailRetryCount: persisted.ExecutorFailRetryCount,
		Source:                 persisted.Source,
		Target:                 persisted.Target,
		Status:                 firstNonEmpty(row.TaskStatus, seatunnel.StatusPending),
		StartTime:              startTime,
		StopTime:               stopTime,
		LastExecuteStatus:      row.LastExecStatus,
	}
}

func toTaskLog(row auto.CoreDatasourceTaskLog) syncmodule.TaskLog {
	status := row.TaskStatus
	if status == "" {
		status = seatunnel.StatusPending
	}
	return syncmodule.TaskLog{
		ID:                strconv.FormatInt(row.ID, 10),
		JobID:             strconv.FormatInt(row.TaskID, 10),
		JobName:           row.PhysicalTableName,
		ExecutorStartTime: row.StartTime,
		ExecutorEndTime:   row.EndTime,
		Status:            status,
		ExecutorMsg:       row.Info,
	}
}

func parseStringID(raw string) (int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, fmt.Errorf("id is required")
	}
	id, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func parseMillis(raw string) int64 {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0
	}
	value, err := strconv.ParseInt(trimmed, 10, 64)
	if err == nil {
		return value
	}
	return 0
}

func schedulerRate(schedulerType string) string {
	switch strings.ToUpper(strings.TrimSpace(schedulerType)) {
	case "CRON", "FIX_RATE", "FIX_DELAY":
		return "1"
	default:
		return "0"
	}
}

func schedulerType(row auto.CoreDatasourceTask) string {
	if strings.TrimSpace(row.Cron) != "" {
		return "CRON"
	}
	if row.SyncRate == "1" {
		return "FIX_RATE"
	}
	return "NONE"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
