package service

import (
	"encoding/json"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/syncmodule"
	"dataease/backend/internal/integration/seatunnel"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSyncServiceRepoTest(t *testing.T) (*SyncService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.CoreDatasourceTask{}, &auto.CoreDatasourceTaskLog{}))

	syncRepo := repository.NewSyncRepository(db)
	return NewSyncService(syncRepo, nil, nil), db
}

func setupSyncServiceWithDatasourceTest(t *testing.T) (*SyncService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&auto.CoreDatasourceTask{},
		&auto.CoreDatasourceTaskLog{},
		&auto.CoreDatasetTable{},
		&auto.CoreDsFinishPage{},
		&datasource.CoreDatasource{},
	))

	syncRepo := repository.NewSyncRepository(db)
	datasourceRepo := repository.NewDatasourceRepository(db)
	datasourceService := NewDatasourceService(datasourceRepo)

	return NewSyncService(syncRepo, datasourceRepo, datasourceService), db
}

func setupClosedSyncServiceRepoTest(t *testing.T) *SyncService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.CoreDatasourceTask{}, &auto.CoreDatasourceTaskLog{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return NewSyncService(repository.NewSyncRepository(db), nil, nil)
}

func closeSyncDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func TestSyncServiceHelpers_ParseStringID(t *testing.T) {
	id, err := parseStringID("42")
	require.NoError(t, err)
	assert.Equal(t, int64(42), id)

	_, err = parseStringID("")
	assert.EqualError(t, err, "id is required")

	_, err = parseStringID("abc")
	assert.EqualError(t, err, "invalid id")
}

func TestSyncServiceHelpers_ParseMillisAndSchedulerRate(t *testing.T) {
	assert.Equal(t, int64(123456), parseMillis("123456"))
	assert.Equal(t, int64(0), parseMillis("bad"))
	assert.Equal(t, "1", schedulerRate("CRON"))
	assert.Equal(t, "1", schedulerRate("fix_rate"))
	assert.Equal(t, "0", schedulerRate("none"))
}

func TestSyncServiceHelpers_SchedulerTypeAndFirstNonEmpty(t *testing.T) {
	assert.Equal(t, "CRON", schedulerType(auto.CoreDatasourceTask{Cron: "0 0 * * * ?"}))
	assert.Equal(t, "FIX_RATE", schedulerType(auto.CoreDatasourceTask{SyncRate: "1"}))
	assert.Equal(t, "NONE", schedulerType(auto.CoreDatasourceTask{}))
	assert.Equal(t, "b", firstNonEmpty("", "b", "c"))
	assert.Equal(t, "", firstNonEmpty("", " "))
}

func TestSyncServiceHelpers_ValueOrEmptyAndToSyncDatasourceDTO(t *testing.T) {
	assert.Equal(t, "", valueOrEmpty(nil))
	value := "ok"
	assert.Equal(t, "ok", valueOrEmpty(&value))

	desc := "desc"
	config := "{}"
	status := datasource.StatusSuccess
	createTime := int64(1)
	updateTime := int64(2)
	item := &datasource.CoreDatasource{
		ID:            7,
		Name:          "mysql-ds",
		Description:   &desc,
		Type:          "mysql",
		Configuration: &config,
		Status:        &status,
		CreateTime:    &createTime,
		UpdateTime:    &updateTime,
	}
	dto := toSyncDatasourceDTO(item)
	assert.Equal(t, "7", dto.ID)
	assert.Equal(t, "mysql-ds", dto.Name)
	assert.Equal(t, "desc", dto.Desc)
	assert.Equal(t, datasource.StatusSuccess, dto.Status)
}

func TestSyncServiceHelpers_TaskRowAndTaskInfoRoundTrip(t *testing.T) {
	row, err := toTaskRow(&syncmodule.TaskInfo{
		Name:            "sync-job",
		TaskKey:         "sync",
		SchedulerType:   "CRON",
		SchedulerConf:   "0 0 * * * ?",
		SchedulerOption: syncmodule.SchedulerOption{Interval: 5, Unit: "MINUTE"},
		Source:          syncmodule.Source{DatasourceID: "9", Type: "mysql"},
		Target:          syncmodule.Target{DatasourceID: "9", Type: "mysql", TableName: "orders"},
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(9), row.DsID)
	assert.Equal(t, seatunnel.StatusPending, row.TaskStatus)

	info := toTaskInfo(*row)
	assert.Equal(t, "sync-job", info.Name)
	assert.Equal(t, "9", info.Source.DatasourceID)
	assert.Equal(t, "orders", info.Target.TableName)

	_, err = toTaskRow(&syncmodule.TaskInfo{Name: "", Source: syncmodule.Source{DatasourceID: ""}}, nil)
	assert.Error(t, err)
}

func TestSyncServiceHelpers_TaskLogAndMarkTaskExecuted(t *testing.T) {
	logDTO := toTaskLog(auto.CoreDatasourceTaskLog{ID: 1, TaskID: 2, PhysicalTableName: "orders", Info: "done"})
	assert.Equal(t, "1", logDTO.ID)
	assert.Equal(t, "2", logDTO.JobID)
	assert.Equal(t, seatunnel.StatusPending, logDTO.Status)

	svc := &SyncService{}
	require.NoError(t, svc.markTaskExecuted(nil, nil, nil))
}

func TestSyncService_TaskLogDetailAndClear(t *testing.T) {
	t.Run("task log detail computes line ranges", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 1, TaskID: 11, Info: "line1\nline2\nline3", TaskStatus: seatunnel.StatusRunning}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 2, TaskID: 12, Info: "   ", TaskStatus: seatunnel.StatusPending}).Error)

		result, err := svc.TaskLogDetail(1, 7)
		require.NoError(t, err)
		assert.Equal(t, 7, result.FromLineNum)
		assert.Equal(t, 9, result.ToLineNum)
		assert.Equal(t, "line1\nline2\nline3", result.LogContent)

		result, err = svc.TaskLogDetail(2, 3)
		require.NoError(t, err)
		assert.Equal(t, 3, result.ToLineNum)
		assert.Empty(t, result.LogContent)
	})

	t.Run("task log detail propagates repo error", func(t *testing.T) {
		svc := setupClosedSyncServiceRepoTest(t)

		result, err := svc.TaskLogDetail(1, 1)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("clear task log handles nil blank and valid job id", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 10, TaskID: 101, TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 11, TaskID: 102, TaskStatus: seatunnel.StatusPending}).Error)

		err := svc.ClearTaskLog(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "WHERE conditions required")
		var count int64
		require.NoError(t, db.Model(&auto.CoreDatasourceTaskLog{}).Count(&count).Error)
		assert.Equal(t, int64(2), count)

		err = svc.ClearTaskLog(&syncmodule.TaskLog{JobID: "   "})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "WHERE conditions required")
		require.NoError(t, db.Model(&auto.CoreDatasourceTaskLog{}).Count(&count).Error)
		assert.Equal(t, int64(2), count)

		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 14, TaskID: 301, TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 15, TaskID: 302, TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, svc.ClearTaskLog(&syncmodule.TaskLog{JobID: "301"}))
		require.NoError(t, db.Model(&auto.CoreDatasourceTaskLog{}).Count(&count).Error)
		assert.Equal(t, int64(3), count)
	})
}

func TestSyncService_MarkTaskExecuted(t *testing.T) {
	t.Run("success path stores last task id and running status", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		task := &auto.CoreDatasourceTask{ID: 41, DsID: 9, Name: "job", ExtraData: "{bad json", TaskStatus: seatunnel.StatusPending}
		require.NoError(t, db.Create(task).Error)

		require.NoError(t, svc.markTaskExecuted(task, map[string]any{"taskId": 12345}, nil))

		var persisted auto.CoreDatasourceTask
		require.NoError(t, db.Where("id = ?", 41).First(&persisted).Error)
		assert.Equal(t, seatunnel.StatusRunning, persisted.TaskStatus)
		assert.Equal(t, seatunnel.StatusRunning, persisted.LastExecStatus)
		assert.NotZero(t, persisted.LastExecTime)
		var extra syncmodule.TaskPersistedData
		require.NoError(t, json.Unmarshal([]byte(persisted.ExtraData), &extra))
		assert.Equal(t, "12345", extra.LastTaskID)
	})

	t.Run("failure path stores failed status", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		task := &auto.CoreDatasourceTask{ID: 42, DsID: 9, Name: "job2", TaskStatus: seatunnel.StatusPending}
		require.NoError(t, db.Create(task).Error)

		require.NoError(t, svc.markTaskExecuted(task, nil, assert.AnError))

		var persisted auto.CoreDatasourceTask
		require.NoError(t, db.Where("id = ?", 42).First(&persisted).Error)
		assert.Equal(t, seatunnel.StatusFailed, persisted.TaskStatus)
		assert.Equal(t, seatunnel.StatusFailed, persisted.LastExecStatus)
	})
}

func TestSyncServiceHelpers_TaskShapeBranches(t *testing.T) {
	t.Run("apply task identity keeps key or defaults to sync", func(t *testing.T) {
		row := &auto.CoreDatasourceTask{}
		applyTaskIdentity(row, &syncmodule.TaskInfo{TaskKey: "manual"}, "job-a")
		assert.Equal(t, "job-a", row.Name)
		assert.Equal(t, "manual", row.UpdateType)

		row = &auto.CoreDatasourceTask{}
		applyTaskIdentity(row, &syncmodule.TaskInfo{}, "job-b")
		assert.Equal(t, "sync", row.UpdateType)
	})

	t.Run("apply task schedule persists fields and runtime defaults", func(t *testing.T) {
		row := &auto.CoreDatasourceTask{}
		req := &syncmodule.TaskInfo{
			TaskKey:         "cron-sync",
			Desc:            "desc",
			SchedulerType:   "CRON",
			SchedulerConf:   "0 0 * * * ?",
			SchedulerOption: syncmodule.SchedulerOption{Interval: 5, Unit: "MINUTE"},
			Source:          syncmodule.Source{DatasourceID: "8", Type: "mysql"},
			Target:          syncmodule.Target{DatasourceID: "9", TableName: "orders"},
			StartTime:       "1000",
			StopTime:        "2000",
		}
		require.NoError(t, applyTaskSchedule(row, req))
		applyTaskRuntime(row, req)
		assert.Equal(t, "1", row.SyncRate)
		assert.Equal(t, "0 0 * * * ?", row.Cron)
		assert.Equal(t, int64(5), row.SimpleCronValue)
		assert.Equal(t, "MINUTE", row.SimpleCronType)
		assert.Equal(t, int64(1000), row.StartTime)
		assert.Equal(t, int64(2000), row.EndTime)
		assert.Equal(t, seatunnel.StatusPending, row.TaskStatus)

		var persisted syncmodule.TaskPersistedData
		require.NoError(t, json.Unmarshal([]byte(row.ExtraData), &persisted))
		assert.Equal(t, "cron-sync", persisted.TaskKey)
		assert.Equal(t, "orders", persisted.Target.TableName)
		assert.Equal(t, "1000", persisted.StartTime)
	})

	t.Run("apply task runtime respects explicit status and toTaskInfo falls back", func(t *testing.T) {
		row := auto.CoreDatasourceTask{ID: 99, Name: "job", UpdateType: "", TaskStatus: "", StartTime: 123, EndTime: 456, LastExecStatus: seatunnel.StatusRunning}
		applyTaskRuntime(&row, &syncmodule.TaskInfo{Status: seatunnel.StatusCancelled, StartTime: "bad", StopTime: "bad"})
		assert.Equal(t, seatunnel.StatusCancelled, row.TaskStatus)

		info := toTaskInfo(auto.CoreDatasourceTask{ID: 100, Name: "job2", UpdateType: "sync", Cron: "0 0 * * * ?", StartTime: 111, EndTime: 222})
		assert.Equal(t, "100", info.ID)
		assert.Equal(t, "CRON", info.SchedulerType)
		assert.Equal(t, "sync", info.TaskKey)
		assert.Equal(t, "111", info.StartTime)
		assert.Equal(t, "222", info.StopTime)
		assert.Equal(t, seatunnel.StatusPending, info.Status)
		assert.Equal(t, seatunnel.StatusPending, firstNonEmpty("", seatunnel.StatusPending))
	})
}

func TestSyncServiceHelpers_NilAndValidationBranches(t *testing.T) {
	svc := &SyncService{}
	fields, err := svc.GetDatasourceFields(nil)
	require.NoError(t, err)
	assert.NotNil(t, fields)

	err = svc.AddTask(nil)
	assert.EqualError(t, err, "task is required")

	err = svc.UpdateTask(nil)
	assert.EqualError(t, err, "task is required")

	err = svc.ClearTaskLog(&syncmodule.TaskLog{JobID: "bad"})
	assert.EqualError(t, err, "invalid id")
}

func TestSyncService_AddTask(t *testing.T) {
	t.Run("creates task on valid request", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)

		err := svc.AddTask(&syncmodule.TaskInfo{
			Name:          "daily-sync",
			TaskKey:       "sync",
			SchedulerType: "NONE",
			Source:        syncmodule.Source{DatasourceID: "9", Type: "mysql"},
		})
		require.NoError(t, err)

		var row auto.CoreDatasourceTask
		require.NoError(t, db.First(&row).Error)
		assert.Equal(t, "daily-sync", row.Name)
		assert.Equal(t, int64(9), row.DsID)
		assert.Equal(t, seatunnel.StatusPending, row.TaskStatus)
		assert.Equal(t, "sync", row.UpdateType)
	})

	t.Run("rejects missing datasource id for new task", func(t *testing.T) {
		svc, _ := setupSyncServiceRepoTest(t)

		err := svc.AddTask(&syncmodule.TaskInfo{
			Name:   "bad-sync",
			Source: syncmodule.Source{DatasourceID: "bad"},
		})
		assert.EqualError(t, err, "source datasourceId is required")
	})
}

func TestSyncService_UpdateTask(t *testing.T) {
	t.Run("rejects invalid task id", func(t *testing.T) {
		svc, _ := setupSyncServiceRepoTest(t)

		err := svc.UpdateTask(&syncmodule.TaskInfo{ID: "bad", Name: "ignored"})
		assert.EqualError(t, err, "invalid id")
	})

	t.Run("propagates repo error when task does not exist", func(t *testing.T) {
		svc, _ := setupSyncServiceRepoTest(t)

		err := svc.UpdateTask(&syncmodule.TaskInfo{ID: "999", Name: "missing"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "record not found")
	})

	t.Run("preserves existing name and datasource when update omits them", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		existing := auto.CoreDatasourceTask{
			ID:         5,
			DsID:       77,
			Name:       "existing-job",
			UpdateType: "manual",
			TaskStatus: seatunnel.StatusRunning,
		}
		require.NoError(t, db.Create(&existing).Error)

		err := svc.UpdateTask(&syncmodule.TaskInfo{
			ID:            "5",
			Name:          "   ",
			TaskKey:       "",
			SchedulerType: "NONE",
			Source:        syncmodule.Source{DatasourceID: "bad"},
			Status:        seatunnel.StatusCancelled,
		})
		require.NoError(t, err)

		var updated auto.CoreDatasourceTask
		require.NoError(t, db.First(&updated, 5).Error)
		assert.Equal(t, int64(77), updated.DsID)
		assert.Equal(t, "existing-job", updated.Name)
		assert.Equal(t, "manual", updated.UpdateType)
		assert.Equal(t, seatunnel.StatusCancelled, updated.TaskStatus)
	})
}

func TestSyncService_TaskLifecycleAndPaging(t *testing.T) {
	t.Run("remove task deletes logs then task", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 101, DsID: 9, Name: "remove-me", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 201, TaskID: 101, TaskStatus: seatunnel.StatusRunning}).Error)

		require.NoError(t, svc.RemoveTask(101))

		var taskCount, logCount int64
		require.NoError(t, db.Model(&auto.CoreDatasourceTask{}).Where("id = ?", 101).Count(&taskCount).Error)
		require.NoError(t, db.Model(&auto.CoreDatasourceTaskLog{}).Where("task_id = ?", 101).Count(&logCount).Error)
		assert.Zero(t, taskCount)
		assert.Zero(t, logCount)
	})

	t.Run("remove task stops when deleting logs fails", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 102, DsID: 9, Name: "remove-fail", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 202, TaskID: 102, TaskStatus: seatunnel.StatusRunning}).Error)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_task_log_delete BEFORE DELETE ON core_datasource_task_log BEGIN SELECT RAISE(FAIL, 'deny task log delete'); END;").Error)

		err := svc.RemoveTask(102)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deny task log delete")

		var taskCount int64
		require.NoError(t, db.Model(&auto.CoreDatasourceTask{}).Where("id = ?", 102).Count(&taskCount).Error)
		assert.Equal(t, int64(1), taskCount)
	})

	t.Run("batch delete tasks stops on log deletion error", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 103, DsID: 9, Name: "batch-a", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 104, DsID: 9, Name: "batch-b", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 203, TaskID: 103, TaskStatus: seatunnel.StatusRunning}).Error)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_batch_log_delete BEFORE DELETE ON core_datasource_task_log BEGIN SELECT RAISE(FAIL, 'deny batch log delete'); END;").Error)

		err := svc.BatchDeleteTasks([]int64{103, 104})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deny batch log delete")

		var taskCount int64
		require.NoError(t, db.Model(&auto.CoreDatasourceTask{}).Count(&taskCount).Error)
		assert.Equal(t, int64(2), taskCount)
	})

	t.Run("batch delete tasks success", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 105, DsID: 9, Name: "batch-ok-a", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 106, DsID: 9, Name: "batch-ok-b", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 204, TaskID: 105, TaskStatus: seatunnel.StatusRunning}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 205, TaskID: 106, TaskStatus: seatunnel.StatusRunning}).Error)

		require.NoError(t, svc.BatchDeleteTasks([]int64{105, 106}))

		var taskCount, logCount int64
		require.NoError(t, db.Model(&auto.CoreDatasourceTask{}).Count(&taskCount).Error)
		require.NoError(t, db.Model(&auto.CoreDatasourceTaskLog{}).Count(&logCount).Error)
		assert.Zero(t, taskCount)
		assert.Zero(t, logCount)
	})

	t.Run("start task sets running status", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 107, DsID: 9, Name: "start-me", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusPending}).Error)

		require.NoError(t, svc.StartTask(107))

		var task auto.CoreDatasourceTask
		require.NoError(t, db.First(&task, 107).Error)
		assert.Equal(t, seatunnel.StatusRunning, task.TaskStatus)
	})

	t.Run("stop task without last task id cancels nothing and persists cancelled", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		persisted, err := json.Marshal(syncmodule.TaskPersistedData{LastTaskID: ""})
		require.NoError(t, err)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 108, DsID: 9, Name: "stop-me", UpdateType: "sync", SyncRate: "0", ExtraData: string(persisted), TaskStatus: seatunnel.StatusRunning}).Error)

		require.NoError(t, svc.StopTask(108))

		var task auto.CoreDatasourceTask
		require.NoError(t, db.First(&task, 108).Error)
		assert.Equal(t, seatunnel.StatusCancelled, task.TaskStatus)
		assert.Equal(t, seatunnel.StatusCancelled, task.LastExecStatus)
	})

	t.Run("task pager maps rows and totals", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		persisted, err := json.Marshal(syncmodule.TaskPersistedData{TaskKey: "sync", Source: syncmodule.Source{DatasourceID: "9"}, Target: syncmodule.Target{TableName: "orders"}})
		require.NoError(t, err)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 109, DsID: 9, Name: "task-one", UpdateType: "sync", SyncRate: "0", ExtraData: string(persisted), TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 110, DsID: 10, Name: "task-two", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusRunning}).Error)

		page, err := svc.TaskPager(1, 10, &syncmodule.TaskGridRequest{Name: "task"})
		require.NoError(t, err)
		require.NotNil(t, page)
		assert.Equal(t, int64(2), page.Total)
		assert.Len(t, page.Records, 2)
		assert.Equal(t, "110", page.Records[0].ID)
		assert.Equal(t, seatunnel.StatusRunning, page.Records[0].Status)
		assert.Equal(t, "orders", page.Records[1].Target.TableName)
	})

	t.Run("task log pager defaults pending status and totals", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 301, TaskID: 201, PhysicalTableName: "orders", TaskStatus: "", Info: "pending", StartTime: 10, EndTime: 20}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 302, TaskID: 202, PhysicalTableName: "users", TaskStatus: seatunnel.StatusRunning, Info: "running", StartTime: 30, EndTime: 40}).Error)

		page, err := svc.TaskLogPager(1, 10, &syncmodule.TaskLogGridRequest{JobID: "201"})
		require.NoError(t, err)
		require.NotNil(t, page)
		assert.Equal(t, int64(1), page.Total)
		require.Len(t, page.Records, 1)
		assert.Equal(t, seatunnel.StatusPending, page.Records[0].Status)
		assert.Equal(t, "201", page.Records[0].JobID)
	})

	t.Run("resource count aggregates all counters and log chart data propagates repo error", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TABLE core_datasource (id INTEGER PRIMARY KEY AUTOINCREMENT, type TEXT, del_flag INTEGER)").Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 111, DsID: 9, Name: "count-task", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 401, TaskID: 111, TaskStatus: seatunnel.StatusRunning}).Error)
		require.NoError(t, db.Exec("INSERT INTO core_datasource (type, del_flag) VALUES ('mysql', 0), ('folder', 0), ('pg', 0), ('mysql', 1)").Error)

		count, err := svc.ResourceCount()
		require.NoError(t, err)
		require.NotNil(t, count)
		assert.Equal(t, int64(1), count.JobCount)
		assert.Equal(t, int64(2), count.DatasourceCount)
		assert.Equal(t, int64(1), count.JobLogCount)

		closeSyncDB(t, db)
		chart, err := svc.LogChartData()
		require.Error(t, err)
		assert.Nil(t, chart)
	})
}

func TestSyncService_GetTaskDeleteLogAndErrorBranches(t *testing.T) {
	t.Run("get task returns mapped info", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		persisted, err := json.Marshal(syncmodule.TaskPersistedData{TaskKey: "sync", SchedulerType: "CRON", SchedulerConf: "0 0 * * * ?", Source: syncmodule.Source{DatasourceID: "9"}, Target: syncmodule.Target{TableName: "orders"}})
		require.NoError(t, err)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 120, DsID: 9, Name: "get-me", UpdateType: "sync", Cron: "0 0 * * * ?", ExtraData: string(persisted), TaskStatus: seatunnel.StatusRunning}).Error)

		info, err := svc.GetTask(120)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, "120", info.ID)
		assert.Equal(t, "get-me", info.Name)
		assert.Equal(t, "orders", info.Target.TableName)
		assert.Equal(t, seatunnel.StatusRunning, info.Status)
	})

	t.Run("get task repo error", func(t *testing.T) {
		svc := setupClosedSyncServiceRepoTest(t)

		info, err := svc.GetTask(1)
		require.Error(t, err)
		assert.Nil(t, info)
	})

	t.Run("delete task log success and repo error", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 501, TaskID: 301, TaskStatus: seatunnel.StatusRunning}).Error)

		require.NoError(t, svc.DeleteTaskLog(501))
		var count int64
		require.NoError(t, db.Model(&auto.CoreDatasourceTaskLog{}).Where("id = ?", 501).Count(&count).Error)
		assert.Zero(t, count)

		closeSyncDB(t, db)
		err := svc.DeleteTaskLog(999)
		require.Error(t, err)
	})

	t.Run("terminate task by log id missing log and zero task id", func(t *testing.T) {
		svc := setupClosedSyncServiceRepoTest(t)
		err := svc.TerminateTaskByLogID(1)
		require.Error(t, err)

		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 502, TaskID: 0, TaskStatus: seatunnel.StatusPending}).Error)
		require.NoError(t, svc.TerminateTaskByLogID(502))
	})

	t.Run("start and stop task get task errors", func(t *testing.T) {
		svc := setupClosedSyncServiceRepoTest(t)

		err := svc.StartTask(1)
		require.Error(t, err)

		err = svc.StopTask(1)
		require.Error(t, err)
	})

	t.Run("resource count datasource count error", func(t *testing.T) {
		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 121, DsID: 9, Name: "count-err", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusPending}).Error)

		count, err := svc.ResourceCount()
		require.Error(t, err)
		assert.Nil(t, count)
	})

	t.Run("resource count task count error and task log count error", func(t *testing.T) {
		svc := setupClosedSyncServiceRepoTest(t)

		count, err := svc.ResourceCount()
		require.Error(t, err)
		assert.Nil(t, count)

		svc, db := setupSyncServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TABLE core_datasource (id INTEGER PRIMARY KEY AUTOINCREMENT, type TEXT, del_flag INTEGER)").Error)
		require.NoError(t, db.Exec("INSERT INTO core_datasource (type, del_flag) VALUES ('mysql', 0)").Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 122, DsID: 9, Name: "count-log-err", UpdateType: "sync", SyncRate: "0", TaskStatus: seatunnel.StatusPending}).Error)

		require.NoError(t, db.Exec("DROP TABLE core_datasource_task_log").Error)
		count, err = svc.ResourceCount()
		require.Error(t, err)
		assert.Nil(t, count)
	})
}

func TestSyncService_DatasourceWrappers(t *testing.T) {
	t.Run("source target pager and datasource CRUD wrappers", func(t *testing.T) {
		svc, db := setupSyncServiceWithDatasourceTest(t)
		rootPID := int64(0)
		folderCfg := "{}"
		createTime1 := int64(10)
		createTime2 := int64(20)
		desc := "desc"

		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 1, PID: &rootPID, Name: "Folder", Type: datasource.TypeFolder, Configuration: &folderCfg, CreateTime: &createTime1}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 2, PID: &rootPID, Name: "MySQL", Type: "mysql", Configuration: &folderCfg, CreateTime: &createTime2}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 3, PID: &rootPID, Name: "PG", Type: "pg", Configuration: &folderCfg}).Error)

		sourcePage, err := svc.SourcePager(1, 10, &datasource.ListRequest{})
		require.NoError(t, err)
		require.NotNil(t, sourcePage)
		assert.Equal(t, int64(3), sourcePage.Total)
		assert.Len(t, sourcePage.Records, 3)
		assert.Equal(t, "3", sourcePage.Records[0].ID)

		targetPage, err := svc.TargetPager(1, 10, nil)
		require.NoError(t, err)
		require.NotNil(t, targetPage)
		assert.Equal(t, int64(3), targetPage.Total)

		saved, err := svc.SaveDatasource(&datasource.WriteRequest{Name: "Folder 2", PID: &rootPID, Type: datasource.TypeFolder})
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, datasource.TypeFolder, saved.Type)

		updated, err := svc.UpdateDatasource(&datasource.WriteRequest{ID: 2, Name: "MySQL Updated", Description: &desc})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "MySQL Updated", updated.Name)

		got, err := svc.GetDatasource(2)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "2", got.ID)
		assert.Equal(t, "MySQL Updated", got.Name)

		require.NoError(t, svc.DeleteDatasource(3))
		_, err = svc.GetDatasource(3)
		require.Error(t, err)

		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 5, PID: &rootPID, Name: "Batch A", Type: "mysql", Configuration: &folderCfg}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 6, PID: &rootPID, Name: "Batch B", Type: "pg", Configuration: &folderCfg}).Error)
		require.NoError(t, svc.BatchDeleteDatasource([]int64{5, 6}))
		_, err = svc.GetDatasource(5)
		require.Error(t, err)
		_, err = svc.GetDatasource(6)
		require.Error(t, err)

		require.NoError(t, svc.BatchDeleteDatasource(nil))

		latest, err := svc.LatestUse("", "")
		require.NoError(t, err)
		assert.Empty(t, latest)
	})

	t.Run("validate list type table schema and field wrappers", func(t *testing.T) {
		svc, db := setupSyncServiceWithDatasourceTest(t)
		rootPID := int64(0)
		folderCfg := "{}"
		tableType := "db"
		folderType := datasource.TypeFolder

		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 11, PID: &rootPID, Name: "Folder", Type: datasource.TypeFolder, Configuration: &folderCfg}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 12, PID: &rootPID, Name: "Mysql A", Type: "mysql", Configuration: &folderCfg}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 13, PID: &rootPID, Name: "Mysql B", Type: "mysql", Configuration: &folderCfg}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasetTable{ID: 101, Name: "Orders", PhysicalTableName: "orders", DatasourceID: 12, DatasetGroupID: 1, Type: tableType}).Error)

		resp, err := svc.ValidateDatasource(&datasource.ValidateRequest{Type: &folderType, Configuration: &folderCfg})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, datasource.StatusSuccess, resp.Status)

		resp, err = svc.ValidateDatasourceByID(11)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, datasource.StatusSuccess, resp.Status)

		typed, err := svc.ListDatasourceByType("mysql")
		require.NoError(t, err)
		assert.Len(t, typed, 2)

		tables, err := svc.ListDatasourceTables(12)
		require.NoError(t, err)
		require.Len(t, tables, 1)
		assert.Equal(t, "orders", tables[0].Name)
		assert.Equal(t, "Orders", tables[0].Remark)

		fields, err := svc.GetDatasourceFields(nil)
		require.NoError(t, err)
		require.NotNil(t, fields)
		assert.Contains(t, fields, "fieldList")
		assert.Contains(t, fields, "targetFieldTypeList")
		emptyFieldResult, err := svc.GetDatasourceFields(&syncmodule.SyncDatasourceFieldRequest{ID: "12", Table: "   "})
		require.NoError(t, err)
		require.NotNil(t, emptyFieldResult)
		assert.Contains(t, emptyFieldResult, "fieldList")
		assert.Contains(t, emptyFieldResult, "targetFieldTypeList")

		_, err = svc.GetDatasourceFields(&syncmodule.SyncDatasourceFieldRequest{ID: "12", Table: "bad-name;drop"})
		require.Error(t, err)

		_, err = svc.GetSchemas()
		require.Error(t, err)

		_, err = svc.SaveDatasource(&datasource.WriteRequest{Name: "   "})
		require.Error(t, err)

		_, err = svc.UpdateDatasource(&datasource.WriteRequest{Name: "missing-id"})
		require.Error(t, err)
	})
}

func TestSyncService_ExecuteAndCancelBranches(t *testing.T) {
	t.Run("execute task covers datasource and table error branches", func(t *testing.T) {
		svc, db := setupSyncServiceWithDatasourceTest(t)

		datasourcePayload, err := json.Marshal(syncmodule.TaskPersistedData{
			Source: syncmodule.Source{DatasourceID: "12"},
			Target: syncmodule.Target{},
		})
		require.NoError(t, err)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 201, DsID: 12, Name: "sync-datasource", UpdateType: "sync", SyncRate: "0", ExtraData: string(datasourcePayload), TaskStatus: seatunnel.StatusPending}).Error)

		result, err := svc.ExecuteTask(201)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "seatunnel grpc address is not configured")

		var task auto.CoreDatasourceTask
		require.NoError(t, db.First(&task, 201).Error)
		assert.Equal(t, seatunnel.StatusFailed, task.TaskStatus)
		assert.Equal(t, seatunnel.StatusFailed, task.LastExecStatus)

		tablePayload, err := json.Marshal(syncmodule.TaskPersistedData{
			Source: syncmodule.Source{DatasourceID: "12"},
			Target: syncmodule.Target{TableName: "orders"},
		})
		require.NoError(t, err)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 202, DsID: 12, Name: "sync-table", UpdateType: "sync", SyncRate: "0", ExtraData: string(tablePayload), TaskStatus: seatunnel.StatusPending}).Error)

		result, err = svc.ExecuteTask(202)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "seatunnel grpc address is not configured")

		var tableTask auto.CoreDatasourceTask
		require.NoError(t, db.First(&tableTask, 202).Error)
		assert.Equal(t, seatunnel.StatusFailed, tableTask.TaskStatus)
		assert.Equal(t, seatunnel.StatusFailed, tableTask.LastExecStatus)
	})

	t.Run("stop task and terminate task propagate cancel errors", func(t *testing.T) {
		svc, db := setupSyncServiceWithDatasourceTest(t)

		persisted, err := json.Marshal(syncmodule.TaskPersistedData{LastTaskID: "123"})
		require.NoError(t, err)
		require.NoError(t, db.Create(&auto.CoreDatasourceTask{ID: 203, DsID: 12, Name: "stop-error", UpdateType: "sync", SyncRate: "0", ExtraData: string(persisted), TaskStatus: seatunnel.StatusRunning}).Error)

		err = svc.StopTask(203)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "seatunnel grpc address is not configured")

		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{ID: 601, TaskID: 456, TaskStatus: seatunnel.StatusRunning}).Error)
		err = svc.TerminateTaskByLogID(601)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "seatunnel grpc address is not configured")
	})
}
