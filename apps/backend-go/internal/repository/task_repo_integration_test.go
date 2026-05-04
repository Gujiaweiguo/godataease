//go:build integration

package repository

import (
	"context"
	"testing"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskRepository_CRUD(t *testing.T) {
	cleanupTables("data_filling_sub_instance", "data_filling_sub_task", "data_filling_task", "data_filling_forms")
	ctx := context.Background()
	formRepo := NewDataFillingRepository(testDB)
	require.NoError(t, formRepo.Create(ctx, &datafillingdomain.DataFillingForm{Name: "form-a", NodeType: datafillingdomain.NodeTypeForm}))

	repo := NewTaskRepository(testDB)
	task := &datafillingdomain.DataFillingTask{FormID: 1, Name: "task-a", UIDList: "[1,2]", RateType: 1, RateVal: "09:00:00", Status: datafillingdomain.TaskStatusStarted}
	require.NoError(t, repo.CreateTask(ctx, task))
	require.NotZero(t, task.ID)

	loaded, err := repo.GetTaskByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "task-a", loaded.Name)

	loaded.Name = "task-b"
	require.NoError(t, repo.UpdateTask(ctx, loaded))

	rows, total, err := repo.ListTasksByFormID(ctx, 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	assert.Equal(t, "task-b", rows[0].Name)

	started, err := repo.GetStartedTasks(ctx)
	require.NoError(t, err)
	require.Len(t, started, 1)

	require.NoError(t, repo.DeleteTasksByIDs(ctx, []int64{task.ID}))
	_, err = repo.GetTaskByID(ctx, task.ID)
	assert.Error(t, err)
}

func TestSubTaskAndSubInstanceRepository_CRUD(t *testing.T) {
	cleanupTables("data_filling_sub_instance", "data_filling_sub_task", "data_filling_task")
	ctx := context.Background()
	taskRepo := NewTaskRepository(testDB)
	require.NoError(t, taskRepo.CreateTask(ctx, &datafillingdomain.DataFillingTask{FormID: 1, Name: "task", UIDList: "[1]", RateType: 1, RateVal: "09:00:00"}))

	subTaskRepo := NewSubTaskRepository(testDB)
	subTask := &datafillingdomain.DataFillingSubTask{TaskID: 1, Status: datafillingdomain.SubTaskStatusActive}
	require.NoError(t, subTaskRepo.CreateSubTask(ctx, subTask))
	require.NoError(t, subTaskRepo.UpdateSubTaskCounts(ctx, subTask.ID, 2, 2, 2, 2))

	rows, total, err := subTaskRepo.ListSubTasksByTaskID(ctx, 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	assert.Equal(t, 2, rows[0].TotalCount)

	instanceRepo := NewSubInstanceRepository(testDB)
	require.NoError(t, instanceRepo.BatchCreateSubInstances(ctx, []*datafillingdomain.DataFillingSubInstance{{TaskID: 1, PID: subTask.ID, UID: 10, FormID: 1, Status: datafillingdomain.SubInstanceStatusOpen}, {TaskID: 1, PID: subTask.ID, UID: 11, FormID: 1, Status: datafillingdomain.SubInstanceStatusFinished}}))

	openStatus := datafillingdomain.SubInstanceStatusOpen
	instances, err := instanceRepo.ListSubInstancesByPID(ctx, subTask.ID, &openStatus)
	require.NoError(t, err)
	require.Len(t, instances, 1)

	ids, err := subTaskRepo.ListSubTaskIDsByTaskIDs(ctx, []int64{1})
	require.NoError(t, err)
	assert.Equal(t, []int64{subTask.ID}, ids)

	require.NoError(t, instanceRepo.DeleteSubInstancesByPIDs(ctx, []int64{subTask.ID}))
	instances, err = instanceRepo.ListSubInstancesByPID(ctx, subTask.ID, nil)
	require.NoError(t, err)
	assert.Empty(t, instances)

	require.NoError(t, subTaskRepo.DeleteSubTasksByIDs(ctx, []int64{subTask.ID}))
	rows, total, err = subTaskRepo.ListSubTasksByTaskID(ctx, 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, rows)
}
