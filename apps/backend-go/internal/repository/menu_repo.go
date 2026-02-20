package repository

import (
	"dataease/backend/internal/domain/menu"
	"gorm.io/gorm"
)

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) GetAll() ([]*menu.CoreMenu, error) {
	var menus []*menu.CoreMenu
	err := r.db.Order("menu_sort ASC").Find(&menus).Error
	return menus, err
}
