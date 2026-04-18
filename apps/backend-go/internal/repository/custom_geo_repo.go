package repository

import (
	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/domain/auto"

	"gorm.io/gorm"
)

type CustomGeoRepository struct {
	db *gorm.DB
}

func NewCustomGeoRepository(db *gorm.DB) *CustomGeoRepository {
	return &CustomGeoRepository{db: db}
}

// ListGeoAreas returns all custom geo areas
func (r *CustomGeoRepository) ListGeoAreas() ([]*auto.CoreCustomGeoArea, error) {
	var areas []*auto.CoreCustomGeoArea
	err := r.db.Find(&areas).Error
	return areas, err
}

// GetGeoArea returns all sub-areas for a given geo area
func (r *CustomGeoRepository) GetGeoArea(areaID string) ([]*auto.CoreCustomGeoSubArea, error) {
	var subAreas []*auto.CoreCustomGeoSubArea
	err := r.db.Where("geo_area_id = ?", areaID).Find(&subAreas).Error
	return subAreas, err
}

// DeleteGeoArea deletes a geo area and all its sub-areas (transactional)
func (r *CustomGeoRepository) DeleteGeoArea(areaID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("geo_area_id = ?", areaID).Delete(&auto.CoreCustomGeoSubArea{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", areaID).Delete(&auto.CoreCustomGeoArea{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// SaveGeoArea creates or updates a geo area (upsert by ID)
func (r *CustomGeoRepository) SaveGeoArea(area *auto.CoreCustomGeoArea) error {
	var existing auto.CoreCustomGeoArea
	err := r.db.Where("id = ?", area.ID).First(&existing).Error
	if err != nil {
		// Not found — create
		return r.db.Create(area).Error
	}
	// Found — update name
	return r.db.Model(&existing).Update("name", area.Name).Error
}

// CheckGeoAreaName checks if a geo area with the given name exists (excluding the given ID)
func (r *CustomGeoRepository) CheckGeoAreaName(name, excludeID string) (bool, error) {
	var count int64
	q := r.db.Model(&auto.CoreCustomGeoArea{}).Where("name = ?", name)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}

// DeleteGeoSubArea deletes a single sub-area
func (r *CustomGeoRepository) DeleteGeoSubArea(id int64) error {
	return r.db.Where("id = ?", id).Delete(&auto.CoreCustomGeoSubArea{}).Error
}

// SaveGeoSubArea creates or updates a sub-area (upsert by ID)
func (r *CustomGeoRepository) SaveGeoSubArea(subArea *auto.CoreCustomGeoSubArea) error {
	var existing auto.CoreCustomGeoSubArea
	err := r.db.Where("id = ?", subArea.ID).First(&existing).Error
	if err != nil {
		// Not found — create
		return r.db.Create(subArea).Error
	}
	// Found — update
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"name":        subArea.Name,
		"scope":       subArea.Scope,
		"geo_area_id": subArea.GeoAreaID,
	}).Error
}

// CheckGeoSubAreaName checks if a sub-area with the given name exists in the same geo area
func (r *CustomGeoRepository) CheckGeoSubAreaName(name, geoAreaID string, excludeID int64) (bool, error) {
	var count int64
	q := r.db.Model(&auto.CoreCustomGeoSubArea{}).Where("name = ? AND geo_area_id = ?", name, geoAreaID)
	if excludeID != 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}

// ListAreaOptions returns areas where pid='156' (China children) for sub-area options
func (r *CustomGeoRepository) ListAreaOptions() ([]*areamap.Area, error) {
	var areas []*areamap.Area
	err := r.db.Where("pid = ?", "156").Find(&areas).Error
	return areas, err
}
