package repository

import (
	"dataease/backend/internal/domain/areamap"
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

// SaveCustomArea inserts a new custom area record
func (r *GeoRepository) SaveCustomArea(area *areamap.CoreAreaCustom) error {
	return r.db.Create(area).Error
}

// GetCustomAreaByID gets a custom area by ID
func (r *GeoRepository) GetCustomAreaByID(id string) (*areamap.CoreAreaCustom, error) {
	var area areamap.CoreAreaCustom
	err := r.db.Where("id = ?", id).First(&area).Error
	if err != nil {
		return nil, err
	}
	return &area, nil
}

// DeleteCustomArea deletes a custom area by ID
func (r *GeoRepository) DeleteCustomArea(id string) error {
	return r.db.Where("id = ?", id).Delete(&areamap.CoreAreaCustom{}).Error
}

// GetCustomAreaChildren gets all children of a custom area (recursive helper)
func (r *GeoRepository) GetCustomAreaChildren(pid string) ([]*areamap.CoreAreaCustom, error) {
	var areas []*areamap.CoreAreaCustom
	err := r.db.Where("pid = ?", pid).Find(&areas).Error
	return areas, err
}

// DeleteCustomAreasBatch deletes multiple custom areas by IDs
func (r *GeoRepository) DeleteCustomAreasBatch(ids []string) error {
	return r.db.Where("id IN ?", ids).Delete(&areamap.CoreAreaCustom{}).Error
}

// CheckAreaExists checks if an area with the given ID exists in the built-in area table
func (r *GeoRepository) CheckAreaExists(id string) (bool, error) {
	var count int64
	err := r.db.Table("area").Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
