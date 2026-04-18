package repository

import (
	"dataease/backend/internal/domain/auto"
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

// Font CRUD methods use core_font table for full font management.

func (r *TypefaceRepository) ListFonts() ([]auto.CoreFont, error) {
	var fonts []auto.CoreFont
	err := r.db.Find(&fonts).Error
	return fonts, err
}

func (r *TypefaceRepository) GetFontByID(id int64) (*auto.CoreFont, error) {
	var font auto.CoreFont
	err := r.db.Where("id = ?", id).First(&font).Error
	return &font, err
}

func (r *TypefaceRepository) FindFontByName(name string) (*auto.CoreFont, error) {
	var font auto.CoreFont
	err := r.db.Where("name = ?", name).First(&font).Error
	return &font, err
}

func (r *TypefaceRepository) CreateFont(font *auto.CoreFont) error {
	return r.db.Create(font).Error
}

func (r *TypefaceRepository) UpdateFont(font *auto.CoreFont) error {
	return r.db.Save(font).Error
}

func (r *TypefaceRepository) DeleteFont(id int64) error {
	return r.db.Where("id = ?", id).Delete(&auto.CoreFont{}).Error
}

func (r *TypefaceRepository) SetDefaultFont(id int64, isDefault bool) error {
	return r.db.Model(&auto.CoreFont{}).Where("id = ?", id).Update("is_default", isDefault).Error
}

func (r *TypefaceRepository) ClearDefaultFonts(excludeID int64) error {
	return r.db.Model(&auto.CoreFont{}).Where("id != ?", excludeID).Update("is_default", false).Error
}

func (r *TypefaceRepository) ListDefaultFonts() ([]auto.CoreFont, error) {
	var fonts []auto.CoreFont
	err := r.db.Where("is_default = ?", true).Find(&fonts).Error
	return fonts, err
}
