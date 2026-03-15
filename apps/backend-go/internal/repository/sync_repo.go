package repository

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/syncmodule"

	"gorm.io/gorm"
)

type SyncRepository struct {
	db *gorm.DB
}

func NewSyncRepository(db *gorm.DB) *SyncRepository {
	return &SyncRepository{db: db}
}

func (r *SyncRepository) ListTasks(page int, size int, req *syncmodule.TaskGridRequest) ([]auto.CoreDatasourceTask, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 200 {
		size = 200
	}
	offset := (page - 1) * size

	query := r.db.Model(&auto.CoreDatasourceTask{})
	if req != nil {
		if strings.TrimSpace(req.Name) != "" {
			query = query.Where("name LIKE ?", "%"+strings.TrimSpace(req.Name)+"%")
		}
		if strings.TrimSpace(req.ID) != "" {
			if id, err := strconv.ParseInt(strings.TrimSpace(req.ID), 10, 64); err == nil && id > 0 {
				query = query.Where("id = ?", id)
			}
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]auto.CoreDatasourceTask, 0)
	if err := query.Order("id DESC").Offset(offset).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *SyncRepository) GetTask(id int64) (*auto.CoreDatasourceTask, error) {
	var task auto.CoreDatasourceTask
	if err := r.db.Where("id = ?", id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *SyncRepository) CreateTask(task *auto.CoreDatasourceTask) error {
	return r.db.Create(task).Error
}

func (r *SyncRepository) UpdateTask(task *auto.CoreDatasourceTask) error {
	return r.db.Save(task).Error
}

func (r *SyncRepository) DeleteTask(id int64) error {
	return r.db.Where("id = ?", id).Delete(&auto.CoreDatasourceTask{}).Error
}

func (r *SyncRepository) BatchDeleteTasks(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&auto.CoreDatasourceTask{}).Error
}

func (r *SyncRepository) CountTasks() (int64, error) {
	var count int64
	err := r.db.Model(&auto.CoreDatasourceTask{}).Count(&count).Error
	return count, err
}

func (r *SyncRepository) CountTaskLogs() (int64, error) {
	var count int64
	err := r.db.Model(&auto.CoreDatasourceTaskLog{}).Count(&count).Error
	return count, err
}

func (r *SyncRepository) CountDatasources() (int64, error) {
	var count int64
	err := r.db.Table("core_datasource").Where("COALESCE(del_flag, 0) = 0 AND type <> ?", "folder").Count(&count).Error
	return count, err
}

func (r *SyncRepository) ListTaskLogs(page int, size int, req *syncmodule.TaskLogGridRequest) ([]auto.CoreDatasourceTaskLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 200 {
		size = 200
	}
	offset := (page - 1) * size

	query := r.db.Model(&auto.CoreDatasourceTaskLog{})
	if req != nil && strings.TrimSpace(req.JobID) != "" {
		if jobID, err := strconv.ParseInt(strings.TrimSpace(req.JobID), 10, 64); err == nil && jobID > 0 {
			query = query.Where("task_id = ?", jobID)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]auto.CoreDatasourceTaskLog, 0)
	if err := query.Order("start_time DESC").Offset(offset).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *SyncRepository) GetTaskLog(id int64) (*auto.CoreDatasourceTaskLog, error) {
	var log auto.CoreDatasourceTaskLog
	if err := r.db.Where("id = ?", id).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *SyncRepository) CreateTaskLog(log *auto.CoreDatasourceTaskLog) error {
	if log == nil {
		return fmt.Errorf("task log is required")
	}
	return r.db.Create(log).Error
}

func (r *SyncRepository) DeleteTaskLog(id int64) error {
	return r.db.Where("id = ?", id).Delete(&auto.CoreDatasourceTaskLog{}).Error
}

func (r *SyncRepository) DeleteTaskLogsByTaskID(taskID int64) error {
	return r.db.Where("task_id = ?", taskID).Delete(&auto.CoreDatasourceTaskLog{}).Error
}

func (r *SyncRepository) ClearTaskLogs(taskID *int64) error {
	query := r.db.Model(&auto.CoreDatasourceTaskLog{})
	if taskID != nil && *taskID > 0 {
		query = query.Where("task_id = ?", *taskID)
	}
	return query.Delete(&auto.CoreDatasourceTaskLog{}).Error
}

func (r *SyncRepository) ListLogChartData(days int) ([]syncmodule.ChartPoint, error) {
	if days <= 0 {
		days = 7
	}
	type row struct {
		Day   string `gorm:"column:day"`
		Count int64  `gorm:"column:count"`
	}
	rows := make([]row, 0)
	since := time.Now().AddDate(0, 0, -days+1).UnixMilli()
	err := r.db.Model(&auto.CoreDatasourceTaskLog{}).
		Select("DATE(FROM_UNIXTIME(create_time / 1000)) as day, COUNT(1) as count").
		Where("create_time >= ?", since).
		Group("DATE(FROM_UNIXTIME(create_time / 1000))").
		Order("day ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]syncmodule.ChartPoint, 0, len(rows))
	for _, item := range rows {
		result = append(result, syncmodule.ChartPoint{Day: item.Day, Count: item.Count})
	}
	return result, nil
}
