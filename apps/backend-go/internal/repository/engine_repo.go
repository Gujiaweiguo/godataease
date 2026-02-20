package repository

import (
	"dataease/backend/internal/domain/engine"

	"gorm.io/gorm"
)

type EngineRepository struct {
	db *gorm.DB
}

func NewEngineRepository(db *gorm.DB) *EngineRepository {
	return &EngineRepository{db: db}
}

func (r *EngineRepository) Get() (*engine.Engine, error) {
	var e engine.Engine
	err := r.db.First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}
