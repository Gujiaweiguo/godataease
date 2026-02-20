package repository

import (
	"dataease/backend/internal/domain/role"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(role *role.SysRole) error {
	return r.db.Create(role).Error
}

func (r *RoleRepository) Update(role *role.SysRole) error {
	return r.db.Save(role).Error
}

func (r *RoleRepository) Delete(roleID int64) error {
	return r.db.Delete(&role.SysRole{}, roleID).Error
}

func (r *RoleRepository) GetByID(roleID int64) (*role.SysRole, error) {
	var rle role.SysRole
	err := r.db.Where("role_id = ? AND status = ?", roleID, role.StatusEnabled).
		First(&rle).Error
	if err != nil {
		return nil, err
	}
	return &rle, nil
}

func (r *RoleRepository) Query(keyword string) ([]*role.SysRole, error) {
	var roles []*role.SysRole
	db := r.db.Model(&role.SysRole{}).Where("status = ?", role.StatusEnabled)
	if keyword != "" {
		db = db.Where("role_name LIKE ?", "%"+keyword+"%")
	}
	err := db.Order("create_time DESC").Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) CountByRoleCode(roleCode string) (int64, error) {
	var count int64
	err := r.db.Model(&role.SysRole{}).
		Where("role_code = ? AND status = ?", roleCode, role.StatusEnabled).
		Count(&count).Error
	return count, err
}
