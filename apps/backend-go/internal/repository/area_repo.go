package repository

import (
	"dataease/backend/internal/domain/areamap"
	"gorm.io/gorm"
)

type AreaRepository struct {
	db *gorm.DB
}

func NewAreaRepository(db *gorm.DB) *AreaRepository {
	return &AreaRepository{db: db}
}

func (r *AreaRepository) GetAllAreas() ([]*areamap.Area, error) {
	var areas []*areamap.Area
	err := r.db.Find(&areas).Error
	return areas, err
}

func (r *AreaRepository) GetAllCustomAreas() ([]*areamap.CoreAreaCustom, error) {
	var areas []*areamap.CoreAreaCustom
	err := r.db.Find(&areas).Error
	return areas, err
}
