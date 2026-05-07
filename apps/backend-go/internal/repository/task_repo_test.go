package repository

import (
	"context"
	"testing"
	"time"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTaskRepositoryTest(t *testing.T) (*TaskRepository, *SubTaskRepository, *SubInstanceRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&datafillingdomain.DataFillingTask{},
		&datafillingdomain.DataFillingSubTask{},
		&datafillingdomain.DataFillingSubInstance{},
	))

	return NewTaskRepository(db), NewSubTaskRepository(db), NewSubInstanceRepository(db), db
}

func TestTaskRepository_CRUDAndQueries(t *testing.T) {
	ctx := context.Background()
	repo, _, _, _ := setupTaskRepositoryTest(t)

	taskA := &datafillingdomain.DataFillingTask{FormID: 10, Name: "Alpha", Status: datafillingdomain.TaskStatusStarted, StartTime: 100, EndTime: 200, FillType: 1}
	taskB := &datafillingdomain.DataFillingTask{FormID: 10, Name: "Beta", Status: datafillingdomain.TaskStatusStopped, StartTime: 200, EndTime: 300, FillType: 2}
	taskC := &datafillingdomain.DataFillingTask{FormID: 11, Name: "Gamma", Status: datafillingdomain.TaskStatusStarted, StartTime: 300, EndTime: 400, FillType: 3}

	require.NoError(t, repo.CreateTask(ctx, taskA))
	require.NoError(t, repo.CreateTask(ctx, taskB))
	require.NoError(t, repo.CreateTask(ctx, taskC))
	require.Positive(t, taskA.ID)

	found, err := repo.GetTaskByID(ctx, taskA.ID)
	require.NoError(t, err)
	assert.Equal(t, "Alpha", found.Name)

	taskA.Name = "Alpha Updated"
	taskA.Status = datafillingdomain.TaskStatusStopped
	require.NoError(t, repo.UpdateTask(ctx, taskA))

	updated, err := repo.GetTaskByID(ctx, taskA.ID)
	require.NoError(t, err)
	assert.Equal(t, "Alpha Updated", updated.Name)
	assert.Equal(t, datafillingdomain.TaskStatusStopped, updated.Status)

	rows, total, err := repo.ListTasksByFormID(ctx, 10, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, rows, 2)
	assert.Equal(t, taskB.ID, rows[0].ID)
	assert.Equal(t, taskA.ID, rows[1].ID)

	paged, total, err := repo.ListTasksByFormID(ctx, 10, 2, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, paged, 1)
	assert.Equal(t, taskA.ID, paged[0].ID)

	started, err := repo.GetStartedTasks(ctx)
	require.NoError(t, err)
	require.Len(t, started, 1)
	assert.Equal(t, taskC.ID, started[0].ID)

	require.NoError(t, repo.DeleteTasksByIDs(ctx, nil))
	require.NoError(t, repo.DeleteTasksByIDs(ctx, []int64{taskB.ID}))
	_, err = repo.GetTaskByID(ctx, taskB.ID)
	require.Error(t, err)

	_, err = repo.GetTaskByID(ctx, 999999)
	require.Error(t, err)
}

func TestSubTaskRepository_CRUDAndCounters(t *testing.T) {
	ctx := context.Background()
	_, repo, _, _ := setupTaskRepositoryTest(t)

	subTaskA := &datafillingdomain.DataFillingSubTask{TaskID: 10, StartTime: 100, EndTime: 200, TotalCount: 9, UnfinishedCount: 3, TotalUserCount: 5, UnfinishedUserCount: 2, FillType: 1}
	subTaskB := &datafillingdomain.DataFillingSubTask{TaskID: 10, StartTime: 200, EndTime: 300, TotalCount: 8, UnfinishedCount: 0, TotalUserCount: 4, UnfinishedUserCount: 0, FillType: 2}
	subTaskC := &datafillingdomain.DataFillingSubTask{TaskID: 11, StartTime: 300, EndTime: 400, TotalCount: 7, UnfinishedCount: 2, TotalUserCount: 3, UnfinishedUserCount: 1, FillType: 3}

	require.NoError(t, repo.CreateSubTask(ctx, subTaskA))
	require.NoError(t, repo.CreateSubTask(ctx, subTaskB))
	require.NoError(t, repo.CreateSubTask(ctx, subTaskC))

	found, err := repo.GetSubTaskByID(ctx, subTaskA.ID)
	require.NoError(t, err)
	assert.Equal(t, subTaskA.TaskID, found.TaskID)

	require.NoError(t, repo.UpdateSubTaskCounts(ctx, subTaskA.ID, 10, 4, 6, 3))
	updated, err := repo.GetSubTaskByID(ctx, subTaskA.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, updated.TotalCount)
	assert.Equal(t, 4, updated.UnfinishedCount)
	assert.Equal(t, 6, updated.TotalUserCount)
	assert.Equal(t, 3, updated.UnfinishedUserCount)

	rows, total, err := repo.ListSubTasksByTaskID(ctx, 10, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, rows, 2)
	assert.Equal(t, subTaskB.ID, rows[0].ID)
	assert.Equal(t, subTaskA.ID, rows[1].ID)

	paged, total, err := repo.ListSubTasksByTaskID(ctx, 10, 2, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, paged, 1)
	assert.Equal(t, subTaskA.ID, paged[0].ID)

	ids, err := repo.ListSubTaskIDsByTaskIDs(ctx, []int64{10, 11})
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{subTaskA.ID, subTaskB.ID, subTaskC.ID}, ids)

	emptyIDs, err := repo.ListSubTaskIDsByTaskIDs(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, emptyIDs)

	require.NoError(t, repo.DecrementSubTaskUnfinishedCount(ctx, subTaskA.ID))
	decremented, err := repo.GetSubTaskByID(ctx, subTaskA.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, decremented.UnfinishedCount)
	assert.Equal(t, 2, decremented.UnfinishedUserCount)

	require.NoError(t, repo.DecrementSubTaskUnfinishedCount(ctx, subTaskB.ID))
	unchanged, err := repo.GetSubTaskByID(ctx, subTaskB.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, unchanged.UnfinishedCount)
	assert.Equal(t, 0, unchanged.UnfinishedUserCount)

	require.NoError(t, repo.DeleteSubTasksByIDs(ctx, nil))
	require.NoError(t, repo.DeleteSubTasksByIDs(ctx, []int64{subTaskC.ID}))
	_, err = repo.GetSubTaskByID(ctx, subTaskC.ID)
	require.Error(t, err)
}

func TestSubInstanceRepository_CRUDAndUserViews(t *testing.T) {
	ctx := context.Background()
	taskRepo, subTaskRepo, repo, _ := setupTaskRepositoryTest(t)

	now := time.Now().UnixMilli()
	taskOpen := &datafillingdomain.DataFillingTask{FormID: 100, Name: "Quarterly Report", FillType: 2, Status: datafillingdomain.TaskStatusStarted}
	taskClosed := &datafillingdomain.DataFillingTask{FormID: 101, Name: "Monthly Report", FillType: 3, Status: datafillingdomain.TaskStatusStopped}
	require.NoError(t, taskRepo.CreateTask(ctx, taskOpen))
	require.NoError(t, taskRepo.CreateTask(ctx, taskClosed))

	subTaskOpen := &datafillingdomain.DataFillingSubTask{TaskID: taskOpen.ID, StartTime: now - 2000, EndTime: now + 2000, TotalCount: 6, UnfinishedCount: 2}
	subTaskExpired := &datafillingdomain.DataFillingSubTask{TaskID: taskClosed.ID, StartTime: now - 5000, EndTime: now - 1000, TotalCount: 5, UnfinishedCount: 1}
	require.NoError(t, subTaskRepo.CreateSubTask(ctx, subTaskOpen))
	require.NoError(t, subTaskRepo.CreateSubTask(ctx, subTaskExpired))

	instances := []*datafillingdomain.DataFillingSubInstance{
		{TaskID: taskOpen.ID, PID: subTaskOpen.ID, UID: 7, FormID: taskOpen.FormID, DataID: "A", Status: datafillingdomain.SubInstanceStatusOpen},
		{TaskID: taskOpen.ID, PID: subTaskOpen.ID, UID: 7, FormID: taskOpen.FormID, DataID: "B", Status: datafillingdomain.SubInstanceStatusFinished, FinishTime: now - 10},
		{TaskID: taskClosed.ID, PID: subTaskExpired.ID, UID: 7, FormID: taskClosed.FormID, DataID: "C", Status: datafillingdomain.SubInstanceStatusOpen},
		{TaskID: taskOpen.ID, PID: subTaskOpen.ID, UID: 8, FormID: taskOpen.FormID, DataID: "D", Status: datafillingdomain.SubInstanceStatusOpen},
	}
	require.NoError(t, repo.BatchCreateSubInstances(ctx, instances))
	require.NoError(t, repo.BatchCreateSubInstances(ctx, nil))

	byPID, err := repo.ListSubInstancesByPID(ctx, subTaskOpen.ID, nil)
	require.NoError(t, err)
	require.Len(t, byPID, 3)
	assert.Equal(t, instances[0].ID, byPID[0].ID)

	statusOpen := datafillingdomain.SubInstanceStatusOpen
	byPIDAndStatus, err := repo.ListSubInstancesByPID(ctx, subTaskOpen.ID, &statusOpen)
	require.NoError(t, err)
	require.Len(t, byPIDAndStatus, 2)

	userReq := &datafillingdomain.UserTaskPageRequest{Type: &statusOpen, TaskName: "Report"}
	userRows, total, err := repo.ListSubInstancesByUID(ctx, 7, 1, 10, userReq)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, userRows, 2)
	assert.Equal(t, taskOpen.ID, userRows[0].TaskID)
	assert.False(t, userRows[0].Expired)
	assert.Equal(t, taskClosed.ID, userRows[1].TaskID)
	assert.True(t, userRows[1].Expired)

	pagedRows, total, err := repo.ListSubInstancesByUID(ctx, 7, 2, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, pagedRows, 1)

	openCount, err := repo.CountOpenSubInstancesByUID(ctx, 7)
	require.NoError(t, err)
	assert.Equal(t, int64(2), openCount)

	found, err := repo.GetSubInstanceByID(ctx, instances[1].ID)
	require.NoError(t, err)
	assert.Equal(t, datafillingdomain.SubInstanceStatusFinished, found.Status)

	byPIDAndUID, err := repo.GetSubInstanceByPIDAndUID(ctx, subTaskOpen.ID, 7)
	require.NoError(t, err)
	require.Len(t, byPIDAndUID, 2)

	require.NoError(t, repo.UpdateSubInstanceStatus(ctx, instances[0].ID, datafillingdomain.SubInstanceStatusFinished, now))
	updated, err := repo.GetSubInstanceByID(ctx, instances[0].ID)
	require.NoError(t, err)
	assert.Equal(t, datafillingdomain.SubInstanceStatusFinished, updated.Status)
	assert.Equal(t, now, updated.FinishTime)

	require.NoError(t, repo.DeleteSubInstancesByPID(ctx, subTaskExpired.ID))
	afterDeleteByPID, err := repo.ListSubInstancesByPID(ctx, subTaskExpired.ID, nil)
	require.NoError(t, err)
	assert.Empty(t, afterDeleteByPID)

	require.NoError(t, repo.DeleteSubInstancesByPIDs(ctx, nil))
	require.NoError(t, repo.DeleteSubInstancesByPIDs(ctx, []int64{subTaskOpen.ID}))
	afterDeleteByPIDs, err := repo.ListSubInstancesByPID(ctx, subTaskOpen.ID, nil)
	require.NoError(t, err)
	assert.Empty(t, afterDeleteByPIDs)

	moreInstances := []*datafillingdomain.DataFillingSubInstance{
		{TaskID: taskOpen.ID, PID: subTaskOpen.ID, UID: 7, FormID: taskOpen.FormID, DataID: "E", Status: datafillingdomain.SubInstanceStatusOpen},
		{TaskID: taskClosed.ID, PID: subTaskExpired.ID, UID: 9, FormID: taskClosed.FormID, DataID: "F", Status: datafillingdomain.SubInstanceStatusOpen},
	}
	require.NoError(t, repo.BatchCreateSubInstances(ctx, moreInstances))
	require.NoError(t, repo.DeleteSubInstancesByTaskIDs(ctx, nil))
	require.NoError(t, repo.DeleteSubInstancesByTaskIDs(ctx, []int64{taskOpen.ID}))
	remaining, err := repo.GetSubInstanceByPIDAndUID(ctx, subTaskExpired.ID, 9)
	require.NoError(t, err)
	require.Len(t, remaining, 1)

	_, err = repo.GetSubInstanceByID(ctx, 999999)
	require.Error(t, err)
}
