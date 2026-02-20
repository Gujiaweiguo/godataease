package repository

import (
	"dataease/backend/internal/domain/permission"
	"gorm.io/gorm"
)

type PermRepository struct {
	db *gorm.DB
}

func NewPermRepository(db *gorm.DB) *PermRepository {
	return &PermRepository{db: db}
}

func (r *PermRepository) Create(p *permission.SysPerm) error {
	return r.db.Create(p).Error
}

func (r *PermRepository) Update(p *permission.SysPerm) error {
	return r.db.Save(p).Error
}

func (r *PermRepository) Delete(permID int64) error {
	return r.db.Model(&permission.SysPerm{}).
		Where("perm_id = ?", permID).
		Update("del_flag", permission.DelFlagDeleted).Error
}

func (r *PermRepository) GetByID(permID int64) (*permission.SysPerm, error) {
	var p permission.SysPerm
	err := r.db.Where("perm_id = ? AND del_flag = ?", permID, permission.DelFlagNormal).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PermRepository) GetByKey(permKey string) (*permission.SysPerm, error) {
	var p permission.SysPerm
	err := r.db.Where("perm_key = ? AND del_flag = ?", permKey, permission.DelFlagNormal).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PermRepository) List() ([]*permission.SysPerm, error) {
	var perms []*permission.SysPerm
	err := r.db.Where("del_flag = ?", permission.DelFlagNormal).
		Order("create_time DESC").
		Find(&perms).Error
	return perms, err
}

func (r *PermRepository) CheckKeyExists(permKey string, excludePermID int64) (int64, error) {
	var count int64
	query := r.db.Model(&permission.SysPerm{}).
		Where("perm_key = ? AND del_flag = ?", permKey, permission.DelFlagNormal)
	if excludePermID > 0 {
		query = query.Where("perm_id != ?", excludePermID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *PermRepository) GetByType(permType string) ([]*permission.SysPerm, error) {
	var perms []*permission.SysPerm
	err := r.db.Where("perm_type = ? AND del_flag = ?", permType, permission.DelFlagNormal).
		Order("create_time DESC").
		Find(&perms).Error
	return perms, err
}
