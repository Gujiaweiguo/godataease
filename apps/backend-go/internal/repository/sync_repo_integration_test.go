//go:build integration
// +build integration

package repository

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/syncmodule"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSyncRepository_TaskCRUDAndCounts(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_datasource_task", "core_datasource_task_log", "core_datasource")

	repo := NewSyncRepository(testDB)

	taskOne := newSyncTask("alpha-task", 101)
	require.NoError(t, repo.CreateTask(taskOne))
	assert.NotZero(t, taskOne.ID)

	taskTwo := newSyncTask("beta-task", 102)
	require.NoError(t, repo.CreateTask(taskTwo))

	taskThree := newSyncTask("gamma-task", 103)
	require.NoError(t, repo.CreateTask(taskThree))

	count, err := repo.CountTasks()
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	got, err := repo.GetTask(taskOne.ID)
	require.NoError(t, err)
	assert.Equal(t, "alpha-task", got.Name)

	_, err = repo.GetTask(999999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

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
	assert.Equal(t, taskTwo.ID, allRows[1].ID)
	assert.Equal(t, taskOne.ID, allRows[2].ID)

	byName, totalByName, err := repo.ListTasks(1, 10, &syncmodule.TaskGridRequest{Name: "beta"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalByName)
	require.Len(t, byName, 1)
	assert.Equal(t, taskTwo.ID, byName[0].ID)

	byID, totalByID, err := repo.ListTasks(1, 10, &syncmodule.TaskGridRequest{ID: fmt.Sprintf("%d", taskOne.ID)})
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalByID)
	require.Len(t, byID, 1)
	assert.Equal(t, taskOne.ID, byID[0].ID)

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

	remaining, err := repo.GetTask(taskTwo.ID)
	require.NoError(t, err)
	assert.Equal(t, taskTwo.ID, remaining.ID)

	_, err = repo.GetTask(taskOne.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	seedCountDatasources(t)
	datasourceCount, err := repo.CountDatasources()
	require.NoError(t, err)
	assert.Equal(t, int64(2), datasourceCount)

	taskCountAfterDelete, err := repo.CountTasks()
	require.NoError(t, err)
	assert.Equal(t, int64(1), taskCountAfterDelete)
}

func TestSyncRepository_TaskLogsAndChartData(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_datasource_task", "core_datasource_task_log")

	repo := NewSyncRepository(testDB)
	taskOne := newSyncTask("log-task-1", 201)
	require.NoError(t, repo.CreateTask(taskOne))
	taskTwo := newSyncTask("log-task-2", 202)
	require.NoError(t, repo.CreateTask(taskTwo))

	now := time.Now()
	logOne := &auto.CoreDatasourceTaskLog{DsID: taskOne.DsID, TaskID: taskOne.ID, StartTime: now.Add(-2 * time.Hour).UnixMilli(), EndTime: now.Add(-time.Hour).UnixMilli(), TaskStatus: "SUCCESS", PhysicalTableName: "table_one", Info: "ok", CreateTime: now.UnixMilli(), TriggerType: "manual"}
	require.NoError(t, repo.CreateTaskLog(logOne))
	assert.NotZero(t, logOne.ID)

	logTwo := &auto.CoreDatasourceTaskLog{DsID: taskTwo.DsID, TaskID: taskTwo.ID, StartTime: now.Add(-26 * time.Hour).UnixMilli(), EndTime: now.Add(-25 * time.Hour).UnixMilli(), TaskStatus: "FAILED", PhysicalTableName: "table_two", Info: "boom", CreateTime: now.AddDate(0, 0, -2).UnixMilli(), TriggerType: "schedule"}
	require.NoError(t, repo.CreateTaskLog(logTwo))

	logThree := &auto.CoreDatasourceTaskLog{DsID: taskTwo.DsID, TaskID: taskTwo.ID, StartTime: now.Add(-10 * time.Minute).UnixMilli(), EndTime: now.UnixMilli(), TaskStatus: "SUCCESS", PhysicalTableName: "table_three", Info: "fresh", CreateTime: now.UnixMilli(), TriggerType: "manual"}
	require.NoError(t, repo.CreateTaskLog(logThree))

	err := repo.CreateTaskLog(nil)
	assert.EqualError(t, err, "task log is required")

	countLogs, err := repo.CountTaskLogs()
	require.NoError(t, err)
	assert.Equal(t, int64(3), countLogs)

	got, err := repo.GetTaskLog(logOne.ID)
	require.NoError(t, err)
	assert.Equal(t, logOne.TaskID, got.TaskID)

	_, err = repo.GetTaskLog(999999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	allLogs, totalLogs, err := repo.ListTaskLogs(1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), totalLogs)
	require.Len(t, allLogs, 3)
	assert.True(t, allLogs[0].StartTime >= allLogs[1].StartTime)

	filteredLogs, filteredTotal, err := repo.ListTaskLogs(1, 10, &syncmodule.TaskLogGridRequest{JobID: fmt.Sprintf("%d", taskTwo.ID)})
	require.NoError(t, err)
	assert.Equal(t, int64(2), filteredTotal)
	require.Len(t, filteredLogs, 2)
	assert.Equal(t, taskTwo.ID, filteredLogs[0].TaskID)
	assert.Equal(t, taskTwo.ID, filteredLogs[1].TaskID)

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

	clearAllRepo := NewSyncRepository(testDB.Session(&gorm.Session{AllowGlobalUpdate: true}))
	require.NoError(t, clearAllRepo.ClearTaskLogs(nil))
	countAfterFullClear, err := repo.CountTaskLogs()
	require.NoError(t, err)
	assert.Equal(t, int64(0), countAfterFullClear)
}

func newSyncTask(name string, dsID int64) *auto.CoreDatasourceTask {
	now := time.Now().UnixMilli()
	return &auto.CoreDatasourceTask{
		DsID:            dsID,
		Name:            name,
		UpdateType:      "full",
		StartTime:       now,
		SyncRate:        "0",
		Cron:            "",
		SimpleCronValue: 0,
		SimpleCronType:  "minute",
		EndLimit:        "0",
		EndTime:         0,
		CreateTime:      now,
		LastExecTime:    0,
		LastExecStatus:  "PENDING",
		ExtraData:       "{}",
		TaskStatus:      "pending",
	}
}

func seedCountDatasources(t *testing.T) {
	t.Helper()
	success := datasource.StatusSuccess
	delFlagOne := 1
	now := time.Now().Unix()
	records := []datasource.CoreDatasource{
		{Name: "mysql-a", Type: "mysql", Status: &success, CreateTime: &now},
		{Name: "mysql-b", Type: "postgresql", Status: &success, CreateTime: &now},
		{Name: "folder-a", Type: datasource.TypeFolder, Status: &success, CreateTime: &now},
		{Name: "deleted-a", Type: "mysql", Status: &success, CreateTime: &now, DelFlag: &delFlagOne},
	}
	for i := range records {
		require.NoError(t, testDB.Create(&records[i]).Error)
	}
}
