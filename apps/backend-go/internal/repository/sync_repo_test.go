package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/syncmodule"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var registerSyncSQLiteDriverOnce sync.Once

func setupSyncRepositoryTest(t *testing.T) *SyncRepository {
	t.Helper()

	registerSyncSQLiteDriverOnce.Do(func() {
			sql.Register("sqlite3_sync_repo", &sqlite3.SQLiteDriver{
				ConnectHook: func(conn *sqlite3.SQLiteConn) error {
					return conn.RegisterFunc("FROM_UNIXTIME", func(value any) string {
						var unixSeconds int64
						switch v := value.(type) {
						case int64:
							unixSeconds = v
						case float64:
							unixSeconds = int64(v)
						case []byte:
							parsed, _ := strconv.ParseFloat(string(v), 64)
							unixSeconds = int64(parsed)
						case string:
							parsed, _ := strconv.ParseFloat(v, 64)
							unixSeconds = int64(parsed)
						}
						return time.Unix(unixSeconds, 0).UTC().Format("2006-01-02 15:04:05")
					}, true)
				},
			})
		})

	db, err := gorm.Open(sqlite.Dialector{DriverName: "sqlite3_sync_repo", DSN: ":memory:"}, &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.CoreDatasourceTask{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDatasource{}))
	require.NoError(t, db.Exec("ALTER TABLE core_datasource ADD COLUMN del_flag INTEGER DEFAULT 0").Error)

	return NewSyncRepository(db)
}

func newUnitSyncTask(name string, dsID int64) *auto.CoreDatasourceTask {
	now := time.Now().UnixMilli()
	return &auto.CoreDatasourceTask{
		DsID:            dsID,
		Name:            name,
		UpdateType:      "full",
		StartTime:       now,
		SyncRate:        "0",
		SimpleCronType:  "minute",
		EndLimit:        "0",
		CreateTime:      now,
		LastExecStatus:  "PENDING",
		ExtraData:       "{}",
		TaskStatus:      "pending",
	}
}

func TestSyncRepository_TaskCRUDFiltersAndCounts(t *testing.T) {
	repo := setupSyncRepositoryTest(t)

	taskOne := newUnitSyncTask("alpha-task", 101)
	taskTwo := newUnitSyncTask("beta-task", 102)
	taskThree := newUnitSyncTask("gamma-task", 103)
	require.NoError(t, repo.CreateTask(taskOne))
	require.NoError(t, repo.CreateTask(taskTwo))
	require.NoError(t, repo.CreateTask(taskThree))

	count, err := repo.CountTasks()
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	got, err := repo.GetTask(taskOne.ID)
	require.NoError(t, err)
	assert.Equal(t, taskOne.Name, got.Name)

	taskOne.Name = "alpha-task-updated"
	taskOne.TaskStatus = "running"
	require.NoError(t, repo.UpdateTask(taskOne))

	updated, err := repo.GetTask(taskOne.ID)
	require.NoError(t, err)
	assert.Equal(t, "alpha-task-updated", updated.Name)
	assert.Equal(t, "running", updated.TaskStatus)

	allRows, total, err := repo.ListTasks(0, 0, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, allRows, 3)
	assert.Equal(t, taskThree.ID, allRows[0].ID)

	byName, totalByName, err := repo.ListTasks(1, 10, &syncmodule.TaskGridRequest{Name: " beta "})
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalByName)
	require.Len(t, byName, 1)
	assert.Equal(t, taskTwo.ID, byName[0].ID)

	byID, totalByID, err := repo.ListTasks(1, 10, &syncmodule.TaskGridRequest{ID: fmt.Sprintf(" %d ", taskOne.ID)})
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalByID)
	require.Len(t, byID, 1)
	assert.Equal(t, taskOne.ID, byID[0].ID)

	invalidID, invalidTotal, err := repo.ListTasks(1, 10, &syncmodule.TaskGridRequest{ID: "not-a-number"})
	require.NoError(t, err)
	assert.Equal(t, int64(3), invalidTotal)
	require.Len(t, invalidID, 3)

	paged, pagedTotal, err := repo.ListTasks(2, 2, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), pagedTotal)
	require.Len(t, paged, 1)
	assert.Equal(t, taskOne.ID, paged[0].ID)

	require.NoError(t, repo.DeleteTask(taskThree.ID))
	_, err = repo.GetTask(taskThree.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	require.NoError(t, repo.BatchDeleteTasks([]int64{taskOne.ID}))
	require.NoError(t, repo.BatchDeleteTasks(nil))
	remainingCount, err := repo.CountTasks()
	require.NoError(t, err)
	assert.Equal(t, int64(1), remainingCount)

	datasources := []*auto.CoreDatasource{
		{Name: "mysql-a", Type: "mysql", CreateTime: 1, UpdateTime: 1},
		{Name: "mysql-b", Type: "postgresql", CreateTime: 1, UpdateTime: 1},
		{Name: "folder-a", Type: "folder", CreateTime: 1, UpdateTime: 1},
		{Name: "deleted-a", Type: "mysql", CreateTime: 1, UpdateTime: 1},
	}
	for _, item := range datasources {
		require.NoError(t, repo.db.Create(item).Error)
	}
	require.NoError(t, repo.db.Exec("UPDATE core_datasource SET del_flag = 1 WHERE name = ?", "deleted-a").Error)

	datasourceCount, err := repo.CountDatasources()
	require.NoError(t, err)
	assert.Equal(t, int64(2), datasourceCount)

	_, err = repo.GetTask(999999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestSyncRepository_TaskLogsAndChartData(t *testing.T) {
	repo := setupSyncRepositoryTest(t)

	taskOne := newUnitSyncTask("log-task-1", 201)
	taskTwo := newUnitSyncTask("log-task-2", 202)
	require.NoError(t, repo.CreateTask(taskOne))
	require.NoError(t, repo.CreateTask(taskTwo))

	now := time.Now().UTC()
	logOne := &auto.CoreDatasourceTaskLog{DsID: taskOne.DsID, TaskID: taskOne.ID, StartTime: now.Add(-2 * time.Hour).UnixMilli(), EndTime: now.Add(-time.Hour).UnixMilli(), TaskStatus: "SUCCESS", PhysicalTableName: "table_one", Info: "ok", CreateTime: now.UnixMilli(), TriggerType: "manual"}
	logTwo := &auto.CoreDatasourceTaskLog{DsID: taskTwo.DsID, TaskID: taskTwo.ID, StartTime: now.Add(-26 * time.Hour).UnixMilli(), EndTime: now.Add(-25 * time.Hour).UnixMilli(), TaskStatus: "FAILED", PhysicalTableName: "table_two", Info: "boom", CreateTime: now.AddDate(0, 0, -2).UnixMilli(), TriggerType: "schedule"}
	logThree := &auto.CoreDatasourceTaskLog{DsID: taskTwo.DsID, TaskID: taskTwo.ID, StartTime: now.Add(-10 * time.Minute).UnixMilli(), EndTime: now.UnixMilli(), TaskStatus: "SUCCESS", PhysicalTableName: "table_three", Info: "fresh", CreateTime: now.UnixMilli(), TriggerType: "manual"}
	require.NoError(t, repo.CreateTaskLog(logOne))
	require.NoError(t, repo.CreateTaskLog(logTwo))
	require.NoError(t, repo.CreateTaskLog(logThree))

	err := repo.CreateTaskLog(nil)
	assert.EqualError(t, err, "task log is required")

	countLogs, err := repo.CountTaskLogs()
	require.NoError(t, err)
	assert.Equal(t, int64(3), countLogs)

	got, err := repo.GetTaskLog(logOne.ID)
	require.NoError(t, err)
	assert.Equal(t, logOne.TaskID, got.TaskID)

	allLogs, totalLogs, err := repo.ListTaskLogs(0, 0, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), totalLogs)
	require.Len(t, allLogs, 3)
	assert.True(t, allLogs[0].StartTime >= allLogs[1].StartTime)

	filteredLogs, filteredTotal, err := repo.ListTaskLogs(1, 10, &syncmodule.TaskLogGridRequest{JobID: fmt.Sprintf(" %d ", taskTwo.ID)})
	require.NoError(t, err)
	assert.Equal(t, int64(2), filteredTotal)
	require.Len(t, filteredLogs, 2)
	assert.Equal(t, taskTwo.ID, filteredLogs[0].TaskID)

	invalidFilter, invalidTotal, err := repo.ListTaskLogs(1, 10, &syncmodule.TaskLogGridRequest{JobID: "bad-id"})
	require.NoError(t, err)
	assert.Equal(t, int64(3), invalidTotal)
	require.Len(t, invalidFilter, 3)

	pagedLogs, pagedTotal, err := repo.ListTaskLogs(2, 2, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), pagedTotal)
	require.Len(t, pagedLogs, 1)

	points, err := repo.ListLogChartData(7)
	require.NoError(t, err)
	require.Len(t, points, 2)
	dayCounts := map[string]int64{}
	for _, point := range points {
		dayCounts[strings.Split(point.Day, "T")[0]] = point.Count
	}
	assert.Equal(t, int64(2), dayCounts[now.Format("2006-01-02")])
	assert.Equal(t, int64(1), dayCounts[now.AddDate(0, 0, -2).Format("2006-01-02")])

	defaultPoints, err := repo.ListLogChartData(0)
	require.NoError(t, err)
	assert.NotEmpty(t, defaultPoints)

	require.NoError(t, repo.DeleteTaskLog(logOne.ID))
	_, err = repo.GetTaskLog(logOne.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	require.NoError(t, repo.DeleteTaskLogsByTaskID(taskTwo.ID))
	remainingCount, err := repo.CountTaskLogs()
	require.NoError(t, err)
	assert.Equal(t, int64(0), remainingCount)

	logFour := &auto.CoreDatasourceTaskLog{DsID: taskOne.DsID, TaskID: taskOne.ID, StartTime: now.UnixMilli(), EndTime: now.UnixMilli(), TaskStatus: "SUCCESS", PhysicalTableName: "table_four", CreateTime: now.UnixMilli()}
	logFive := &auto.CoreDatasourceTaskLog{DsID: taskTwo.DsID, TaskID: taskTwo.ID, StartTime: now.UnixMilli(), EndTime: now.UnixMilli(), TaskStatus: "SUCCESS", PhysicalTableName: "table_five", CreateTime: now.UnixMilli()}
	require.NoError(t, repo.CreateTaskLog(logFour))
	require.NoError(t, repo.CreateTaskLog(logFive))

	require.NoError(t, repo.ClearTaskLogs(&taskOne.ID))
	countAfterScopedClear, err := repo.CountTaskLogs()
	require.NoError(t, err)
	assert.Equal(t, int64(1), countAfterScopedClear)

	clearAllRepo := NewSyncRepository(repo.db.Session(&gorm.Session{AllowGlobalUpdate: true}))
	require.NoError(t, clearAllRepo.ClearTaskLogs(nil))
	countAfterFullClear, err := repo.CountTaskLogs()
	require.NoError(t, err)
	assert.Equal(t, int64(0), countAfterFullClear)

	_, err = repo.GetTaskLog(999999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
