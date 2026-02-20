package repository

import (
	"dataease/backend/internal/domain/geo"

	"gorm.io/gorm"
)

type GeoRepository struct {
	db *gorm.DB
}

func NewGeoRepository(db *gorm.DB) *GeoRepository {
	return &GeoRepository{db: db}
}

func (r *GeoRepository) ListAreas() ([]*geo.GeometryArea, error) {
	var areas []*geo.GeometryArea
	err := r.db.Find(&areas).Error
	return areas, err
}

func (r *GeoRepository) GetAreaByID(id string) (*geo.GeometryArea, error) {
	var area geo.GeometryArea
	err := r.db.Where("id = ?", id).First(&area).Error
	if err != nil {
		return nil, err
	}
	return &area, nil
}
