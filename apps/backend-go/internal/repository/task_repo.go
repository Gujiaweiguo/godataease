package repository

import (
	"context"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"gorm.io/gorm"
)

type TaskRepositoryInterface interface {
	CreateTask(ctx context.Context, task *datafillingdomain.DataFillingTask) error
	UpdateTask(ctx context.Context, task *datafillingdomain.DataFillingTask) error
	GetTaskByID(ctx context.Context, taskID int64) (*datafillingdomain.DataFillingTask, error)
	ListTasksByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DataFillingTask, int64, error)
	DeleteTasksByIDs(ctx context.Context, taskIDs []int64) error
	GetStartedTasks(ctx context.Context) ([]*datafillingdomain.DataFillingTask, error)
}

type SubTaskRepositoryInterface interface {
	CreateSubTask(ctx context.Context, subTask *datafillingdomain.DataFillingSubTask) error
	UpdateSubTaskCounts(ctx context.Context, subTaskID int64, totalCount, unfinishedCount, totalUserCount, unfinishedUserCount int) error
	ListSubTasksByTaskID(ctx context.Context, taskID int64, page, pageSize int) ([]*datafillingdomain.DataFillingSubTask, int64, error)
	DeleteSubTasksByIDs(ctx context.Context, subTaskIDs []int64) error
	ListSubTaskIDsByTaskIDs(ctx context.Context, taskIDs []int64) ([]int64, error)
}

type SubInstanceRepositoryInterface interface {
	BatchCreateSubInstances(ctx context.Context, instances []*datafillingdomain.DataFillingSubInstance) error
	DeleteSubInstancesByPID(ctx context.Context, pid int64) error
	DeleteSubInstancesByPIDs(ctx context.Context, pids []int64) error
	DeleteSubInstancesByTaskIDs(ctx context.Context, taskIDs []int64) error
	ListSubInstancesByPID(ctx context.Context, pid int64, statusFilter *int) ([]*datafillingdomain.DataFillingSubInstance, error)
}

var _ TaskRepositoryInterface = (*TaskRepository)(nil)
var _ SubTaskRepositoryInterface = (*SubTaskRepository)(nil)
var _ SubInstanceRepositoryInterface = (*SubInstanceRepository)(nil)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTask(ctx context.Context, task *datafillingdomain.DataFillingTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *TaskRepository) UpdateTask(ctx context.Context, task *datafillingdomain.DataFillingTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *TaskRepository) GetTaskByID(ctx context.Context, taskID int64) (*datafillingdomain.DataFillingTask, error) {
	var task datafillingdomain.DataFillingTask
	if err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) ListTasksByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DataFillingTask, int64, error) {
	query := r.db.WithContext(ctx).Model(&datafillingdomain.DataFillingTask{}).Where("form_id = ?", formID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := 0
	if page > 1 {
		offset = (page - 1) * pageSize
	}
	rows := make([]*datafillingdomain.DataFillingTask, 0)
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *TaskRepository) DeleteTasksByIDs(ctx context.Context, taskIDs []int64) error {
	if len(taskIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", taskIDs).Delete(&datafillingdomain.DataFillingTask{}).Error
}

func (r *TaskRepository) GetStartedTasks(ctx context.Context) ([]*datafillingdomain.DataFillingTask, error) {
	rows := make([]*datafillingdomain.DataFillingTask, 0)
	err := r.db.WithContext(ctx).Where("status = ?", datafillingdomain.TaskStatusStarted).Order("id ASC").Find(&rows).Error
	return rows, err
}

type SubTaskRepository struct {
	db *gorm.DB
}

func NewSubTaskRepository(db *gorm.DB) *SubTaskRepository {
	return &SubTaskRepository{db: db}
}

func (r *SubTaskRepository) CreateSubTask(ctx context.Context, subTask *datafillingdomain.DataFillingSubTask) error {
	return r.db.WithContext(ctx).Create(subTask).Error
}

func (r *SubTaskRepository) UpdateSubTaskCounts(ctx context.Context, subTaskID int64, totalCount, unfinishedCount, totalUserCount, unfinishedUserCount int) error {
	return r.db.WithContext(ctx).Model(&datafillingdomain.DataFillingSubTask{}).Where("id = ?", subTaskID).Updates(map[string]any{
		"total_count":           totalCount,
		"unfinished_count":      unfinishedCount,
		"total_user_count":      totalUserCount,
		"unfinished_user_count": unfinishedUserCount,
	}).Error
}

func (r *SubTaskRepository) ListSubTasksByTaskID(ctx context.Context, taskID int64, page, pageSize int) ([]*datafillingdomain.DataFillingSubTask, int64, error) {
	query := r.db.WithContext(ctx).Model(&datafillingdomain.DataFillingSubTask{}).Where("task_id = ?", taskID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := 0
	if page > 1 {
		offset = (page - 1) * pageSize
	}
	rows := make([]*datafillingdomain.DataFillingSubTask, 0)
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *SubTaskRepository) DeleteSubTasksByIDs(ctx context.Context, subTaskIDs []int64) error {
	if len(subTaskIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", subTaskIDs).Delete(&datafillingdomain.DataFillingSubTask{}).Error
}

func (r *SubTaskRepository) ListSubTaskIDsByTaskIDs(ctx context.Context, taskIDs []int64) ([]int64, error) {
	if len(taskIDs) == 0 {
		return []int64{}, nil
	}
	rows := make([]int64, 0)
	err := r.db.WithContext(ctx).Model(&datafillingdomain.DataFillingSubTask{}).Where("task_id IN ?", taskIDs).Pluck("id", &rows).Error
	return rows, err
}

type SubInstanceRepository struct {
	db *gorm.DB
}

func NewSubInstanceRepository(db *gorm.DB) *SubInstanceRepository {
	return &SubInstanceRepository{db: db}
}

func (r *SubInstanceRepository) BatchCreateSubInstances(ctx context.Context, instances []*datafillingdomain.DataFillingSubInstance) error {
	if len(instances) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(instances, 500).Error
}

func (r *SubInstanceRepository) DeleteSubInstancesByPID(ctx context.Context, pid int64) error {
	return r.db.WithContext(ctx).Where("pid = ?", pid).Delete(&datafillingdomain.DataFillingSubInstance{}).Error
}

func (r *SubInstanceRepository) DeleteSubInstancesByPIDs(ctx context.Context, pids []int64) error {
	if len(pids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("pid IN ?", pids).Delete(&datafillingdomain.DataFillingSubInstance{}).Error
}

func (r *SubInstanceRepository) DeleteSubInstancesByTaskIDs(ctx context.Context, taskIDs []int64) error {
	if len(taskIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("task_id IN ?", taskIDs).Delete(&datafillingdomain.DataFillingSubInstance{}).Error
}

func (r *SubInstanceRepository) ListSubInstancesByPID(ctx context.Context, pid int64, statusFilter *int) ([]*datafillingdomain.DataFillingSubInstance, error) {
	query := r.db.WithContext(ctx).Model(&datafillingdomain.DataFillingSubInstance{}).Where("pid = ?", pid)
	if statusFilter != nil {
		query = query.Where("status = ?", *statusFilter)
	}
	rows := make([]*datafillingdomain.DataFillingSubInstance, 0)
	err := query.Order("id ASC").Find(&rows).Error
	return rows, err
}
