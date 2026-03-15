//go:build integration

package service

import (
	"strconv"
	"testing"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/syncmodule"
	"dataease/backend/internal/integration/seatunnel"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanupSyncTables() {
	testDB.Exec("DELETE FROM core_datasource_task_log")
	testDB.Exec("DELETE FROM core_datasource_task")
	testDB.Exec("DELETE FROM core_datasource")
	testDB.Exec("DELETE FROM core_ds_finish_page")
}

func TestSyncService_TaskLifecycleAndSummary(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	require.NoError(t, testDB.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasourceTask{}, &auto.CoreDatasourceTaskLog{}))
	cleanupSyncTables()

	dsRepo := repository.NewDatasourceRepository(testDB)
	syncRepo := repository.NewSyncRepository(testDB)
	dsSvc := NewDatasourceService(dsRepo)
	svc := NewSyncService(syncRepo, dsRepo, dsSvc)
	addr, cleanup := startSeatunnelServerForIntegration(t, false)
	defer cleanup()
	dsSvc.SetSeatunnelConfig(addr, 3*time.Second, 0)

	now := time.Now().UnixMilli()
	status := datasource.StatusSuccess
	ds := &datasource.CoreDatasource{
		Name:       "sync-source-ds",
		Type:       "mysql",
		Status:     &status,
		CreateTime: &now,
		UpdateTime: &now,
	}
	require.NoError(t, dsRepo.Create(ds))

	addReq := &syncmodule.TaskInfo{
		Name:          "nightly-sync",
		TaskKey:       "sync",
		SchedulerType: "NONE",
		Source: syncmodule.Source{
			DatasourceID: strconv.FormatInt(ds.ID, 10),
			Type:         "mysql",
		},
		Target: syncmodule.Target{
			DatasourceID: strconv.FormatInt(ds.ID, 10),
			Type:         "mysql",
			TableName:    "orders",
		},
	}
	require.NoError(t, svc.AddTask(addReq))

	page, err := svc.TaskPager(1, 10, &syncmodule.TaskGridRequest{Name: "nightly"})
	require.NoError(t, err)
	require.NotNil(t, page)
	require.Len(t, page.Records, 1)
	assert.Equal(t, "nightly-sync", page.Records[0].Name)

	id, err := strconv.ParseInt(page.Records[0].ID, 10, 64)
	require.NoError(t, err)

	detail, err := svc.GetTask(id)
	require.NoError(t, err)
	assert.Equal(t, "nightly-sync", detail.Name)
	assert.Equal(t, strconv.FormatInt(ds.ID, 10), detail.Source.DatasourceID)

	result, err := svc.ExecuteTask(id)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "99001", result["taskId"])

	logs, err := svc.TaskLogPager(1, 10, &syncmodule.TaskLogGridRequest{})
	require.NoError(t, err)
	require.NotNil(t, logs)
	assert.NotEmpty(t, logs.Records)

	counts, err := svc.ResourceCount()
	require.NoError(t, err)
	require.NotNil(t, counts)
	assert.Equal(t, int64(1), counts.JobCount)
	assert.Equal(t, int64(1), counts.DatasourceCount)
	assert.GreaterOrEqual(t, counts.JobLogCount, int64(1))

	chart, err := svc.LogChartData()
	require.NoError(t, err)
	assert.NotNil(t, chart)
	assert.Contains(t, chart, "values")
}

func TestSyncService_TaskManagementPaths(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	require.NoError(t, testDB.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasourceTask{}, &auto.CoreDatasourceTaskLog{}))
	cleanupSyncTables()

	dsRepo := repository.NewDatasourceRepository(testDB)
	syncRepo := repository.NewSyncRepository(testDB)
	dsSvc := NewDatasourceService(dsRepo)
	svc := NewSyncService(syncRepo, dsRepo, dsSvc)
	addr, cleanup := startSeatunnelServerForIntegration(t, false)
	defer cleanup()
	dsSvc.SetSeatunnelConfig(addr, 3*time.Second, 0)

	now := time.Now().UnixMilli()
	status := datasource.StatusSuccess
	creator := "sync-owner"
	emptyConfig := "{}"
	ds := &datasource.CoreDatasource{
		Name:          "sync-manage-ds",
		Type:          datasource.TypeFolder,
		Configuration: &emptyConfig,
		Status:        &status,
		CreateBy:      &creator,
		CreateTime:    &now,
		UpdateTime:    &now,
	}
	require.NoError(t, dsRepo.Create(ds))

	created, err := svc.SaveDatasource(&datasource.WriteRequest{Name: "sync-created-ds", Type: "mysql"})
	require.NoError(t, err)
	require.NotNil(t, created)

	_, err = svc.SaveDatasource(&datasource.WriteRequest{Name: "", Type: "mysql"})
	assert.Error(t, err)

	validateResp, err := svc.ValidateDatasource(&datasource.ValidateRequest{Type: &ds.Type, Configuration: &emptyConfig})
	require.NoError(t, err)
	assert.Equal(t, datasource.StatusSuccess, validateResp.Status)

	validateByIDResp, err := svc.ValidateDatasourceByID(ds.ID)
	require.NoError(t, err)
	assert.Equal(t, datasource.StatusSuccess, validateByIDResp.Status)

	schemas, err := svc.GetSchemas()
	require.NoError(t, err)
	assert.NotNil(t, schemas)

	tables, err := svc.ListDatasourceTables(ds.ID)
	require.NoError(t, err)
	assert.NotNil(t, tables)
	createdID, err := strconv.ParseInt(created.ID, 10, 64)
	require.NoError(t, err)

	fetchedCreated, err := svc.GetDatasource(createdID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetchedCreated.ID)

	updatedDatasource, err := svc.UpdateDatasource(&datasource.WriteRequest{ID: createdID, Name: "sync-created-ds-updated", Type: "mysql"})
	require.NoError(t, err)
	assert.Equal(t, "sync-created-ds-updated", updatedDatasource.Name)

	_, err = svc.GetDatasource(999999)
	assert.Error(t, err)

	_, err = svc.ValidateDatasourceByID(999999)
	assert.NoError(t, err)

	sourcePage, err := svc.SourcePager(1, 10, &datasource.ListRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(sourcePage.Records), 1)

	targetPage, err := svc.TargetPager(1, 10, &datasource.ListRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(targetPage.Records), 1)

	byType, err := svc.ListDatasourceByType("mysql")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(byType), 1)

	addReq := &syncmodule.TaskInfo{
		Name:          "manageable-sync",
		TaskKey:       "sync",
		SchedulerType: "NONE",
		Source:        syncmodule.Source{DatasourceID: strconv.FormatInt(ds.ID, 10), Type: "mysql"},
		Target:        syncmodule.Target{DatasourceID: strconv.FormatInt(ds.ID, 10), Type: "mysql", TableName: "orders"},
	}
	require.NoError(t, svc.AddTask(addReq))
	require.NoError(t, svc.AddTask(&syncmodule.TaskInfo{
		Name:          "datasource-sync",
		TaskKey:       "sync",
		SchedulerType: "NONE",
		Source:        syncmodule.Source{DatasourceID: strconv.FormatInt(ds.ID, 10), Type: "mysql"},
		Target:        syncmodule.Target{DatasourceID: strconv.FormatInt(ds.ID, 10), Type: "mysql"},
	}))

	page, err := svc.TaskPager(1, 10, &syncmodule.TaskGridRequest{Name: "manageable"})
	require.NoError(t, err)
	require.Len(t, page.Records, 1)
	taskID, err := strconv.ParseInt(page.Records[0].ID, 10, 64)
	require.NoError(t, err)

	require.NoError(t, svc.UpdateTask(&syncmodule.TaskInfo{
		ID:            strconv.FormatInt(taskID, 10),
		Name:          "manageable-sync-updated",
		SchedulerType: "CRON",
		SchedulerConf: "0 0 * * * ?",
		Source:        syncmodule.Source{DatasourceID: strconv.FormatInt(ds.ID, 10), Type: "mysql"},
		Target:        syncmodule.Target{DatasourceID: strconv.FormatInt(ds.ID, 10), Type: "mysql", TableName: "orders"},
	}))

	updated, err := svc.GetTask(taskID)
	require.NoError(t, err)
	assert.Equal(t, "manageable-sync-updated", updated.Name)

	require.NoError(t, svc.StartTask(taskID))
	started, err := svc.GetTask(taskID)
	require.NoError(t, err)
	assert.Equal(t, seatunnel.StatusRunning, started.Status)

	taskRow, err := syncRepo.GetTask(taskID)
	require.NoError(t, err)
	taskRow.ExtraData = `{"lastTaskId":"99001"}`
	require.NoError(t, syncRepo.UpdateTask(taskRow))

	require.NoError(t, svc.StopTask(taskID))
	stopped, err := svc.GetTask(taskID)
	require.NoError(t, err)
	assert.Equal(t, seatunnel.StatusCancelled, stopped.Status)

	logRow := &auto.CoreDatasourceTaskLog{
		TaskID:            taskID,
		PhysicalTableName: "orders",
		TaskStatus:        "SUCCESS",
		Info:              "line1\nline2",
		StartTime:         now,
		EndTime:           now + 1000,
		CreateTime:        now,
	}
	require.NoError(t, syncRepo.CreateTaskLog(logRow))

	logs, err := svc.TaskLogPager(1, 10, &syncmodule.TaskLogGridRequest{JobID: strconv.FormatInt(taskID, 10)})
	require.NoError(t, err)
	require.Len(t, logs.Records, 1)

	logDetail, err := svc.TaskLogDetail(logRow.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, logDetail.ToLineNum)
	assert.Contains(t, logDetail.LogContent, "line1")

	require.NoError(t, svc.TerminateTaskByLogID(logRow.ID))

	require.NoError(t, svc.DeleteTaskLog(logRow.ID))

	logs, err = svc.TaskLogPager(1, 10, &syncmodule.TaskLogGridRequest{JobID: strconv.FormatInt(taskID, 10)})
	require.NoError(t, err)
	assert.Empty(t, logs.Records)

	logRow2 := &auto.CoreDatasourceTaskLog{TaskID: taskID, PhysicalTableName: "orders", TaskStatus: "SUCCESS", Info: "line3", StartTime: now, EndTime: now + 1000, CreateTime: now}
	require.NoError(t, syncRepo.CreateTaskLog(logRow2))
	require.NoError(t, svc.ClearTaskLog(&syncmodule.TaskLog{JobID: strconv.FormatInt(taskID, 10)}))

	logs, err = svc.TaskLogPager(1, 10, &syncmodule.TaskLogGridRequest{JobID: strconv.FormatInt(taskID, 10)})
	require.NoError(t, err)
	assert.Empty(t, logs.Records)

	require.NoError(t, svc.BatchDeleteTasks([]int64{taskID}))
	_, err = svc.GetTask(taskID)
	assert.Error(t, err)

	page, err = svc.TaskPager(1, 10, &syncmodule.TaskGridRequest{Name: "datasource-sync"})
	require.NoError(t, err)
	require.Len(t, page.Records, 1)
	dsTaskID, err := strconv.ParseInt(page.Records[0].ID, 10, 64)
	require.NoError(t, err)
	_, err = svc.ExecuteTask(dsTaskID)
	require.NoError(t, err)
	require.NoError(t, svc.RemoveTask(dsTaskID))
	_, err = svc.GetTask(dsTaskID)
	assert.Error(t, err)

	emptyLog := &auto.CoreDatasourceTaskLog{TaskID: 0, PhysicalTableName: "orders", TaskStatus: "SUCCESS", Info: "", StartTime: now, EndTime: now, CreateTime: now}
	require.NoError(t, syncRepo.CreateTaskLog(emptyLog))
	detailZero, err := svc.TaskLogDetail(emptyLog.ID, 5)
	require.NoError(t, err)
	assert.Equal(t, 5, detailZero.ToLineNum)
	require.NoError(t, svc.TerminateTaskByLogID(emptyLog.ID))
	assert.Error(t, svc.ClearTaskLog(nil))

	batchCreated, err := svc.SaveDatasource(&datasource.WriteRequest{Name: "sync-batch-ds", Type: "mysql"})
	require.NoError(t, err)
	batchID, err := strconv.ParseInt(batchCreated.ID, 10, 64)
	require.NoError(t, err)
	require.NoError(t, svc.BatchDeleteDatasource([]int64{batchID}))
	_, err = svc.GetDatasource(batchID)
	assert.Error(t, err)

	require.NoError(t, svc.DeleteDatasource(createdID))
	_, err = svc.GetDatasource(createdID)
	assert.Error(t, err)
}
