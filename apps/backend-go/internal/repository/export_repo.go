package repository

import (
	"dataease/backend/internal/domain/export"

	"gorm.io/gorm"
)

type coreExportTask struct {
	ID                string `gorm:"column:id;primaryKey;size:50"`
	UserID            int64  `gorm:"column:user_id"`
	FileName          string `gorm:"column:file_name;size:255"`
	FileSize          float64
	FileSizeUnit      string `gorm:"column:file_size_unit;size:20"`
	ExportFrom        int64  `gorm:"column:export_from"`
	ExportStatus      string `gorm:"column:export_status;size:50"`
	Msg               string `gorm:"column:msg;size:500"`
	ExportFromType    string `gorm:"column:export_from_type;size:50"`
	ExportTime        int64  `gorm:"column:export_time"`
	ExportProgress    string `gorm:"column:export_progress;size:20"`
	ExportMachineName string `gorm:"column:export_machine_name;size:100"`
	ExportFromName    string `gorm:"column:export_from_name;size:255"`
	OrgName           string `gorm:"column:org_name;size:255"`
}

func (coreExportTask) TableName() string {
	return "core_export_task"
}

type ExportRepository struct {
	db *gorm.DB
}

func NewExportRepository(db *gorm.DB) *ExportRepository {
	return &ExportRepository{db: db}
}

func (r *ExportRepository) Create(task *export.ExportTask) error {
	record := r.toRecord(task)
	return r.db.Create(record).Error
}

func (r *ExportRepository) GetByID(id string) (*export.ExportTask, error) {
	var record coreExportTask
	err := r.db.Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&record), nil
}

func (r *ExportRepository) List(page, pageSize int, status string) ([]export.ExportTask, int64, error) {
	var records []coreExportTask
	var total int64

	query := r.db.Model(&coreExportTask{})
	if status != "" && status != "all" {
		query = query.Where("export_status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("export_time DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]export.ExportTask, len(records))
	for i, record := range records {
		tasks[i] = *r.toDomain(&record)
	}

	return tasks, total, nil
}

func (r *ExportRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&coreExportTask{}).Where("id = ?", id).Update("export_status", status).Error
}

func (r *ExportRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&coreExportTask{}).Error
}

func (r *ExportRepository) DeleteBatch(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&coreExportTask{}).Error
}

func (r *ExportRepository) DeleteAllByType(exportFromType string) error {
	query := r.db
	if exportFromType != "" && exportFromType != "all" {
		query = query.Where("export_from_type = ?", exportFromType)
	}
	return query.Delete(&coreExportTask{}).Error
}

func (r *ExportRepository) CountByStatus() (map[string]int64, error) {
	type statusCount struct {
		Status string
		Count  int64
	}
	var counts []statusCount

	err := r.db.Model(&coreExportTask{}).
		Select("export_status as status, count(*) as count").
		Group("export_status").
		Scan(&counts).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, c := range counts {
		result[c.Status] = c.Count
	}
	return result, nil
}

func (r *ExportRepository) toRecord(task *export.ExportTask) *coreExportTask {
	return &coreExportTask{
		ID:                task.ID,
		UserID:            task.UserID,
		FileName:          task.FileName,
		FileSize:          task.FileSize,
		FileSizeUnit:      task.FileSizeUnit,
		ExportFrom:        task.ExportFrom,
		ExportStatus:      task.ExportStatus,
		Msg:               task.Msg,
		ExportFromType:    task.ExportFromType,
		ExportTime:        task.ExportTime,
		ExportProgress:    task.ExportProgress,
		ExportMachineName: task.ExportMachineName,
		ExportFromName:    task.ExportFromName,
		OrgName:           task.OrgName,
	}
}

func (r *ExportRepository) toDomain(record *coreExportTask) *export.ExportTask {
	return &export.ExportTask{
		ID:                record.ID,
		UserID:            record.UserID,
		FileName:          record.FileName,
		FileSize:          record.FileSize,
		FileSizeUnit:      record.FileSizeUnit,
		ExportFrom:        record.ExportFrom,
		ExportStatus:      record.ExportStatus,
		Msg:               record.Msg,
		ExportFromType:    record.ExportFromType,
		ExportTime:        record.ExportTime,
		ExportProgress:    record.ExportProgress,
		ExportMachineName: record.ExportMachineName,
		ExportFromName:    record.ExportFromName,
		OrgName:           record.OrgName,
	}
}
