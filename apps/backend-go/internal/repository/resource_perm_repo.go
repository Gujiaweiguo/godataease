package repository

import (
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/user"

	"gorm.io/gorm"
)

type ResourcePermissionRepository struct {
	db *gorm.DB
}

func NewResourcePermissionRepository(db *gorm.DB) *ResourcePermissionRepository {
	return &ResourcePermissionRepository{db: db}
}

func (r *ResourcePermissionRepository) GetPermByID(permID int64) (*permission.SysPerm, error) {
	var perm permission.SysPerm
	err := r.db.Where("perm_id = ? AND del_flag = 0", permID).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *ResourcePermissionRepository) GetPermByKey(permKey string) (*permission.SysPerm, error) {
	var perm permission.SysPerm
	err := r.db.Where("perm_key = ? AND del_flag = 0", permKey).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *ResourcePermissionRepository) ListPerms(permType string, page, size int) ([]*permission.SysPerm, int64, error) {
	var perms []*permission.SysPerm
	var total int64

	query := r.db.Model(&permission.SysPerm{}).Where("del_flag = 0")
	if permType != "" {
		query = query.Where("perm_type = ?", permType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Offset(offset).Limit(size).Find(&perms).Error
	return perms, total, err
}

func (r *ResourcePermissionRepository) CreatePerm(perm *permission.SysPerm) error {
	return r.db.Create(perm).Error
}

func (r *ResourcePermissionRepository) UpdatePerm(perm *permission.SysPerm) error {
	return r.db.Save(perm).Error
}

func (r *ResourcePermissionRepository) DeletePerm(permID int64) error {
	return r.db.Model(&permission.SysPerm{}).
		Where("perm_id = ?", permID).
		Update("del_flag", 1).Error
}

func (r *ResourcePermissionRepository) GetUserPerms(userID int64) ([]int64, error) {
	var permIDs []int64
	err := r.db.Model(&permission.SysUserPerm{}).
		Where("user_id = ? AND status = 1 AND del_flag = 0", userID).
		Pluck("perm_id", &permIDs).Error
	return permIDs, err
}

func (r *ResourcePermissionRepository) GetRolePerms(roleID int64) ([]int64, error) {
	var permIDs []int64
	err := r.db.Model(&permission.SysRolePerm{}).
		Where("role_id = ?", roleID).
		Pluck("perm_id", &permIDs).Error
	return permIDs, err
}

func (r *ResourcePermissionRepository) GetUserRoleIDs(userID int64) ([]int64, error) {
	var roleIDs []int64
	err := r.db.Model(&user.SysUserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *ResourcePermissionRepository) CheckUserPermission(userID, permID int64) (bool, error) {
	var count int64
	err := r.db.Model(&permission.SysUserPerm{}).
		Where("user_id = ? AND perm_id = ? AND status = 1 AND del_flag = 0", userID, permID).
		Count(&count).Error
	return count > 0, err
}

func (r *ResourcePermissionRepository) CheckRolePermission(roleID, permID int64) (bool, error) {
	var count int64
	err := r.db.Model(&permission.SysRolePerm{}).
		Where("role_id = ? AND perm_id = ?", roleID, permID).
		Count(&count).Error
	return count > 0, err
}

func (r *ResourcePermissionRepository) GrantPermToUser(userID, permID int64, createBy string) error {
	perm := &permission.SysUserPerm{
		UserID:   userID,
		PermID:   permID,
		CreateBy: &createBy,
		Status:   1,
		DelFlag:  0,
	}
	return r.db.Create(perm).Error
}

func (r *ResourcePermissionRepository) RevokePermFromUser(userID, permID int64) error {
	return r.db.Model(&permission.SysUserPerm{}).
		Where("user_id = ? AND perm_id = ?", userID, permID).
		Updates(map[string]interface{}{"status": 0, "del_flag": 1}).Error
}

func (r *ResourcePermissionRepository) GrantPermToRole(roleID, permID int64) error {
	perm := &permission.SysRolePerm{
		RoleID: roleID,
		PermID: permID,
	}
	return r.db.Create(perm).Error
}

func (r *ResourcePermissionRepository) RevokePermFromRole(roleID, permID int64) error {
	return r.db.Where("role_id = ? AND perm_id = ?", roleID, permID).
		Delete(&permission.SysRolePerm{}).Error
}
