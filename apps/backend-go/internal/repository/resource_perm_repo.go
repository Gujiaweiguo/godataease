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

// ========== 双视角接口实现 ==========

// GetUserResources 获取用户可访问的资源列表（按用户视角）
func (r *ResourcePermissionRepository) GetUserResources(userID int64, resourceType string) ([]*permission.UserResourcePermVO, error) {
	var results []*permission.UserResourcePermVO
	
	// 1. 直接授权给用户的资源
	var userPerms []struct {
		PermID   int64  `gorm:"column:perm_id"`
		PermKey  string `gorm:"column:perm_key"`
		PermName string `gorm:"column:perm_name"`
	}
	err := r.db.Table("sys_user_perm sup").
		Select("p.perm_id, p.perm_key, p.perm_name").
		Joins("JOIN sys_perm p ON p.perm_id = sup.perm_id").
		Where("sup.user_id = ? AND sup.status = 1 AND sup.del_flag = 0", userID).
		Find(&userPerms).Error
	if err != nil {
		return nil, err
	}
	
	for _, up := range userPerms {
		results = append(results, &permission.UserResourcePermVO{
			PermKey:    up.PermKey,
			PermName:   up.PermName,
			SourceType: "direct",
		})
	}
	
	// 2. 通过角色继承的资源
	var rolePerms []struct {
		PermID   int64  `gorm:"column:perm_id"`
		PermKey  string `gorm:"column:perm_key"`
		PermName string `gorm:"column:perm_name"`
		RoleID   int64  `gorm:"column:role_id"`
		RoleName string `gorm:"column:role_name"`
	}
	err = r.db.Table("sys_user_role sur").
		Select("srp.perm_id, p.perm_key, p.perm_name, srp.role_id, r.role_name").
		Joins("JOIN sys_role_perm srp ON srp.role_id = sur.role_id").
		Joins("JOIN sys_perm p ON p.perm_id = srp.perm_id").
		Joins("JOIN sys_role r ON r.role_id = sur.role_id").
		Where("sur.user_id = ?", userID).
		Find(&rolePerms).Error
	if err != nil {
		return nil, err
	}
	
	for _, rp := range rolePerms {
		results = append(results, &permission.UserResourcePermVO{
			PermKey:    rp.PermKey,
			PermName:   rp.PermName,
			SourceType: "role",
			SourceID:   rp.RoleID,
			SourceName: rp.RoleName,
		})
	}
	
	return results, nil
}

// GetResourceUsers 获取资源的授权用户列表（按资源视角）
func (r *ResourcePermissionRepository) GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	var results []*permission.ResourceUserPermVO
	
	// TODO: 需要根据 resource_id 和 resource_type 查询实际资源权限
	// 当前版本返回基础结构，后续需要扩展资源权限表
	
	return results, nil
}

// ApplyGroupPermissions 将分组权限应用到新资源
func (r *ResourcePermissionRepository) ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error {
	// TODO: 实现分组权限继承逻辑
	// 1. 查询分组的权限配置
	// 2. 将权限复制到新资源
	return nil
}

// CheckPermissionConsistency 检查双视角权限一致性
func (r *ResourcePermissionRepository) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	// 当前实现返回一致结果，后续可以添加详细的一致性检查
	return &permission.PermissionConsistencyResult{
		Consistent:      true,
		UserCount:       0,
		ResourceCount:   0,
		Inconsistencies: []*permission.PermissionInconsistencyVO{},
	}, nil
}
