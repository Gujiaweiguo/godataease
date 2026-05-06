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

func (r *RoleRepository) DB() *gorm.DB {
	return r.db
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
	err := r.db.Where("role_id = ?", roleID).First(&rle).Error
	if err != nil {
		return nil, err
	}
	return &rle, nil
}

func (r *RoleRepository) GetByRoleCode(roleCode string) (*role.SysRole, error) {
	var rle role.SysRole
	err := r.db.Where("role_code = ?", roleCode).First(&rle).Error
	if err != nil {
		return nil, err
	}
	return &rle, nil
}

func (r *RoleRepository) Query(keyword string) ([]*role.SysRole, error) {
	var roles []*role.SysRole
	db := r.db.Model(&role.SysRole{})
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

// CountUserRoles 统计用户的角色数量
func (r *RoleRepository) CountUserRoles(userID int64) (int64, error) {
	var count int64
	err := r.db.Table("sys_user_role").
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetUserRoleIDs 获取用户的角色ID列表
func (r *RoleRepository) GetUserRoleIDs(userID int64) ([]int64, error) {
	var roleIDs []int64
	err := r.db.Table("sys_user_role").
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

// BindUserRole 绑定用户到角色
func (r *RoleRepository) BindUserRole(userID, roleID, orgID int64) error {
	userRole := map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
		"org_id":  orgID,
	}
	return r.db.Table("sys_user_role").Create(userRole).Error
}

// UnbindUserRole 解绑用户与角色
func (r *RoleRepository) UnbindUserRole(userID, roleID, orgID int64) error {
	query := r.db.Table("sys_user_role").
		Where("user_id = ? AND role_id = ?", userID, roleID)
	if orgID > 0 {
		query = query.Where("org_id = ?", orgID)
	}
	return query.Delete(nil).Error
}

// CountUserRolesByOrg 统计用户在指定组织内的角色数量
func (r *RoleRepository) CountUserRolesByOrg(userID, orgID int64) (int64, error) {
	var count int64
	query := r.db.Table("sys_user_role").Where("user_id = ?", userID)
	if orgID > 0 {
		query = query.Where("org_id = ?", orgID)
	}
	err := query.Count(&count).Error
	return count, err
}

// GetRolesByIDs 根据角色ID列表查询角色
func (r *RoleRepository) GetRolesByIDs(roleIDs []int64) ([]*role.SysRole, error) {
	var roles []*role.SysRole
	err := r.db.Where("role_id IN ? AND status = ?", roleIDs, role.StatusEnabled).
		Find(&roles).Error
	return roles, err
}

// QueryByOrgID 查询组织下的角色
func (r *RoleRepository) QueryByOrgID(orgID int64, keyword string) ([]*role.SysRole, error) {
	var roles []*role.SysRole
	assignedRoles := r.db.Table("sys_user_role").Select("role_id").Where("org_id = ?", orgID)
	db := r.db.Model(&role.SysRole{}).
		Where("status = ?", role.StatusEnabled).
		Where("((role_type = ? AND (org_id = ? OR org_id IS NULL OR org_id = 0)) OR role_id IN (?))", role.RoleTypeOrganization, orgID, assignedRoles)
	if keyword != "" {
		db = db.Where("role_name LIKE ?", "%"+keyword+"%")
	}
	err := db.Distinct().Order("create_time DESC").Find(&roles).Error
	return roles, err
}

// QueryWithPage 分页查询角色
func (r *RoleRepository) QueryWithPage(keyword, roleType string, current, size int) ([]*role.SysRole, int64, error) {
	var roles []*role.SysRole
	var total int64

	db := r.db.Model(&role.SysRole{})

	if keyword != "" {
		db = db.Where("role_name LIKE ? OR role_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if roleType != "" {
		db = db.Where("role_type = ?", roleType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (current - 1) * size
	if offset < 0 {
		offset = 0
	}

	if err := db.Order("role_type ASC, create_time DESC").Offset(offset).Limit(size).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}
