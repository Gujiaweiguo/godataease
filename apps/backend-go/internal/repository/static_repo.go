package repository

import (
	"dataease/backend/internal/domain/static"

	"gorm.io/gorm"
)

type StaticRepository struct {
	db *gorm.DB
}

func NewStaticRepository(db *gorm.DB) *StaticRepository {
	return &StaticRepository{db: db}
}

func (r *StaticRepository) ListResources() ([]*static.StaticResource, error) {
	var resources []*static.StaticResource
	err := r.db.Find(&resources).Error
	return resources, err
}

func (r *StaticRepository) GetResourceByID(id string) (*static.StaticResource, error) {
	var resource static.StaticResource
	err := r.db.Where("id = ?", id).First(&resource).Error
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

type StoreRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) *StoreRepository {
	return &StoreRepository{db: db}
}

func (r *StoreRepository) ListStores() ([]*static.Store, error) {
	var stores []*static.Store
	err := r.db.Find(&stores).Error
	return stores, err
}

type TypefaceRepository struct {
	db *gorm.DB
}

func NewTypefaceRepository(db *gorm.DB) *TypefaceRepository {
	return &TypefaceRepository{db: db}
}

func (r *TypefaceRepository) ListTypefaces() ([]*static.Typeface, error) {
	var typefaces []*static.Typeface
	err := r.db.Find(&typefaces).Error
	return typefaces, err
}
