package repository

import (
	"dataease/backend/internal/domain/driver"

	"gorm.io/gorm"
)

type DriverRepository struct {
	db *gorm.DB
}

func NewDriverRepository(db *gorm.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

func (r *DriverRepository) List() ([]driver.Driver, error) {
	var list []driver.Driver
	err := r.db.Order("id ASC").Find(&list).Error
	return list, err
}

func (r *DriverRepository) ListByType(dsType string) ([]driver.Driver, error) {
	var list []driver.Driver
	err := r.db.Where("type = ?", dsType).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *DriverRepository) GetByID(id int64) (*driver.Driver, error) {
	var d driver.Driver
	err := r.db.Where("id = ?", id).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DriverRepository) ListDriverJars(driverID int64) ([]driver.DriverJar, error) {
	var list []driver.DriverJar
	err := r.db.Where("driver_id = ?", driverID).Order("id ASC").Find(&list).Error
	return list, err
}
