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

func (r *MenuRepository) GetByIDs(ids []int64) ([]*menu.CoreMenu, error) {
	if len(ids) == 0 {
		return []*menu.CoreMenu{}, nil
	}
	var menus []*menu.CoreMenu
	err := r.db.Where("id IN ?", ids).Order("menu_sort ASC").Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) GetByID(id int64) (*menu.CoreMenu, error) {
	var m menu.CoreMenu
	err := r.db.Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MenuRepository) Create(m *menu.CoreMenu) error {
	return r.db.Create(m).Error
}

func (r *MenuRepository) Update(m *menu.CoreMenu) error {
	return r.db.Save(m).Error
}

func (r *MenuRepository) Delete(id int64) error {
	return r.db.Delete(&menu.CoreMenu{}, id).Error
}

func (r *MenuRepository) HasChildren(id int64) (bool, error) {
	var count int64
	err := r.db.Model(&menu.CoreMenu{}).Where("pid = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *MenuRepository) UpdateSort(id int64, sort int) error {
	return r.db.Model(&menu.CoreMenu{}).Where("id = ?", id).Update("menu_sort", sort).Error
}

func (r *MenuRepository) UpdateHidden(id int64, hidden bool) error {
	return r.db.Model(&menu.CoreMenu{}).Where("id = ?", id).Update("hidden", hidden).Error
}
