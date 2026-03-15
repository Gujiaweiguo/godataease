package repository

import (
	"errors"

	"dataease/backend/internal/domain/visualization"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WatermarkRepository struct {
	db *gorm.DB
}

func NewWatermarkRepository(db *gorm.DB) *WatermarkRepository {
	return &WatermarkRepository{db: db}
}

func (r *WatermarkRepository) FindLatest() (*visualization.Watermark, error) {
	var record visualization.Watermark
	result := r.db.Order("create_time desc").Limit(1).Find(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	if record.ID == "" {
		return nil, nil
	}
	return &record, nil
}

func (r *WatermarkRepository) SaveDefault(settingContent string, createBy string, createTime int64) (*visualization.Watermark, error) {
	record := &visualization.Watermark{
		ID:             "default",
		Version:        "v1",
		SettingContent: settingContent,
		CreateBy:       createBy,
		CreateTime:     createTime,
	}
	err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"setting_content", "version", "create_by", "create_time"}),
	}).Create(record).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return record, nil
}
