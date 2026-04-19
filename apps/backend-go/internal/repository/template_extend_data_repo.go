package repository

import (
	"dataease/backend/internal/domain/auto"

	"gorm.io/gorm"
)

type TemplateExtendDataRepository struct {
	db *gorm.DB
}

func NewTemplateExtendDataRepository(db *gorm.DB) *TemplateExtendDataRepository {
	return &TemplateExtendDataRepository{db: db}
}

func (r *TemplateExtendDataRepository) BatchCreate(records []auto.VisualizationTemplateExtendDatum) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.Create(&records).Error
}
