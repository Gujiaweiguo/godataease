package repository

import (
	"dataease/backend/internal/domain/auto"

	"gorm.io/gorm"
)

type VisualizationBackgroundRepository struct {
	db *gorm.DB
}

func NewVisualizationBackgroundRepository(db *gorm.DB) *VisualizationBackgroundRepository {
	return &VisualizationBackgroundRepository{db: db}
}

func (r *VisualizationBackgroundRepository) FindAll() ([]auto.VisualizationBackground, error) {
	var backgrounds []auto.VisualizationBackground
	err := r.db.Order("sort ASC").Find(&backgrounds).Error
	return backgrounds, err
}
