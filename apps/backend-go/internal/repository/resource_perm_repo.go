package repository

import (
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/user"
	"errors"
	"fmt"
	"strings"

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
	permKeyPrefix := resourcePermKeyPrefix(resourceType)

	// 1. 直接授权给用户的资源
	var userPerms []struct {
		PermID   int64  `gorm:"column:perm_id"`
		PermKey  string `gorm:"column:perm_key"`
		PermName string `gorm:"column:perm_name"`
	}
	err := r.db.Table("sys_user_perm sup").
		Select("p.perm_id, p.perm_key, p.perm_name").
		Joins("JOIN sys_perm p ON p.perm_id = sup.perm_id AND p.del_flag = 0").
		Where("sup.user_id = ? AND sup.status = 1 AND sup.del_flag = 0", userID).
		Find(&userPerms).Error
	if err != nil {
		return nil, err
	}

	for _, up := range userPerms {
		if permKeyPrefix != "" && !strings.HasPrefix(up.PermKey, permKeyPrefix) {
			continue
		}
		results = append(results, &permission.UserResourcePermVO{
			PermKey:      up.PermKey,
			PermName:     up.PermName,
			SourceType:   "direct",
			ResourceType: resourceType,
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
		Joins("JOIN sys_perm p ON p.perm_id = srp.perm_id AND p.del_flag = 0").
		Joins("JOIN sys_role r ON r.role_id = sur.role_id").
		Where("sur.user_id = ?", userID).
		Find(&rolePerms).Error
	if err != nil {
		return nil, err
	}

	for _, rp := range rolePerms {
		if permKeyPrefix != "" && !strings.HasPrefix(rp.PermKey, permKeyPrefix) {
			continue
		}
		results = append(results, &permission.UserResourcePermVO{
			PermKey:      rp.PermKey,
			PermName:     rp.PermName,
			SourceType:   "role",
			SourceID:     rp.RoleID,
			SourceName:   rp.RoleName,
			ResourceType: resourceType,
		})
	}

	return results, nil
}

// GetResourceUsers 获取资源的授权用户列表（按资源视角）
func (r *ResourcePermissionRepository) GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	var results []*permission.ResourceUserPermVO
	resourcePermIDs, exists, err := r.GetResourcePermissionIDs(resourceID, resourceType)
	if err != nil {
		return nil, err
	}
	if exists {
		if len(resourcePermIDs) == 0 {
			return results, nil
		}
		var directPerms []*permission.ResourceUserPermVO
		err = r.db.Table("sys_user_perm sup").
			Select("DISTINCT u.user_id, u.username, COALESCE(u.nick_name, u.username) AS nick_name, p.perm_key, p.perm_name, ? AS source_type, u.user_id AS source_id, u.username AS source_name", "direct").
			Joins("JOIN sys_perm p ON p.perm_id = sup.perm_id AND p.del_flag = 0").
			Joins("JOIN sys_user u ON u.user_id = sup.user_id AND u.del_flag = 0").
			Where("sup.status = 1 AND sup.del_flag = 0 AND p.perm_id IN ?", resourcePermIDs).
			Order("u.user_id ASC, p.perm_key ASC").
			Find(&directPerms).Error
		if err != nil {
			return nil, err
		}
		results = append(results, directPerms...)

		var rolePerms []*permission.ResourceUserPermVO
		err = r.db.Table("sys_role_perm srp").
			Select("DISTINCT u.user_id, u.username, COALESCE(u.nick_name, u.username) AS nick_name, p.perm_key, p.perm_name, ? AS source_type, r.role_id AS source_id, r.role_name AS source_name", "role").
			Joins("JOIN sys_perm p ON p.perm_id = srp.perm_id AND p.del_flag = 0").
			Joins("JOIN sys_user_role sur ON sur.role_id = srp.role_id").
			Joins("JOIN sys_user u ON u.user_id = sur.user_id AND u.del_flag = 0").
			Joins("JOIN sys_role r ON r.role_id = srp.role_id").
			Where("p.perm_id IN ?", resourcePermIDs).
			Order("u.user_id ASC, p.perm_key ASC, r.role_id ASC").
			Find(&rolePerms).Error
		if err != nil {
			return nil, err
		}
		results = append(results, rolePerms...)
		return results, nil
	}

	permKeyPrefix := resourcePermKeyPrefix(resourceType)
	if permKeyPrefix == "" {
		return results, nil
	}

	var directPerms []*permission.ResourceUserPermVO
	err = r.db.Table("sys_user_perm sup").
		Select("DISTINCT u.user_id, u.username, COALESCE(u.nick_name, u.username) AS nick_name, p.perm_key, p.perm_name, ? AS source_type, u.user_id AS source_id, u.username AS source_name", "direct").
		Joins("JOIN sys_perm p ON p.perm_id = sup.perm_id AND p.del_flag = 0").
		Joins("JOIN sys_user u ON u.user_id = sup.user_id AND u.del_flag = 0").
		Where("sup.status = 1 AND sup.del_flag = 0 AND p.perm_key LIKE ?", permKeyPrefix+"%").
		Order("u.user_id ASC, p.perm_key ASC").
		Find(&directPerms).Error
	if err != nil {
		return nil, err
	}
	results = append(results, directPerms...)

	var rolePerms []*permission.ResourceUserPermVO
	err = r.db.Table("sys_role_perm srp").
		Select("DISTINCT u.user_id, u.username, COALESCE(u.nick_name, u.username) AS nick_name, p.perm_key, p.perm_name, ? AS source_type, r.role_id AS source_id, r.role_name AS source_name", "role").
		Joins("JOIN sys_perm p ON p.perm_id = srp.perm_id AND p.del_flag = 0").
		Joins("JOIN sys_user_role sur ON sur.role_id = srp.role_id").
		Joins("JOIN sys_user u ON u.user_id = sur.user_id AND u.del_flag = 0").
		Joins("JOIN sys_role r ON r.role_id = srp.role_id").
		Where("p.perm_key LIKE ?", permKeyPrefix+"%").
		Order("u.user_id ASC, p.perm_key ASC, r.role_id ASC").
		Find(&rolePerms).Error
	if err != nil {
		return nil, err
	}
	results = append(results, rolePerms...)

	return results, nil
}

// ApplyGroupPermissions 将分组权限应用到新资源
func (r *ResourcePermissionRepository) ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error {
	if groupID <= 0 || resourceID <= 0 || strings.TrimSpace(resourceType) == "" {
		return nil
	}

	inheritedPermIDs, parentExists, err := r.GetResourcePermissionIDs(groupID, resourceType)
	if err != nil {
		return err
	}
	if !parentExists {
		return nil
	}

	return r.ReplaceResourcePermissions(resourceID, resourceType, inheritedPermIDs)
}

func (r *ResourcePermissionRepository) RegisterResource(resourceID int64, resourceName, resourceType string, parentID *int64) error {
	if resourceID <= 0 || strings.TrimSpace(resourceType) == "" {
		return fmt.Errorf("resource id and type are required")
	}
	scopedResourceID := scopedResourceID(resourceType, resourceID)

	name := strings.TrimSpace(resourceName)
	if name == "" {
		name = fmt.Sprintf("%s-%d", resourceType, resourceID)
	}

	var existing permission.SysResource
	err := r.db.Where("resource_id = ? AND resource_type = ?", scopedResourceID, resourceType).First(&existing).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		entry := &permission.SysResource{
			ResourceID:   scopedResourceID,
			ResourceName: name,
			ResourceType: resourceType,
			ParentID:     normalizeParentID(resourceType, parentID),
		}
		return r.db.Create(entry).Error
	}

	if strings.TrimSpace(resourceName) != "" {
		existing.ResourceName = name
	}
	if parentID != nil {
		existing.ParentID = normalizeParentID(resourceType, parentID)
	}
	return r.db.Save(&existing).Error
}

func (r *ResourcePermissionRepository) ReplaceResourcePermissions(resourceID int64, resourceType string, permIDs []int64) error {
	if resourceID <= 0 || strings.TrimSpace(resourceType) == "" {
		return fmt.Errorf("resource id and type are required")
	}
	scopedResourceID := scopedResourceID(resourceType, resourceID)
	if err := r.RegisterResource(resourceID, "", resourceType, nil); err != nil {
		return err
	}

	uniquePermIDs := dedupeInt64(permIDs)
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("resource_id = ?", scopedResourceID).Delete(&permission.SysResourcePerm{}).Error; err != nil {
			return err
		}
		if len(uniquePermIDs) == 0 {
			return nil
		}
		rows := make([]*permission.SysResourcePerm, 0, len(uniquePermIDs))
		for _, permID := range uniquePermIDs {
			rows = append(rows, &permission.SysResourcePerm{ResourceID: scopedResourceID, PermID: permID})
		}
		return tx.Create(&rows).Error
	})
}

func (r *ResourcePermissionRepository) GetResourcePermissionIDs(resourceID int64, resourceType string) ([]int64, bool, error) {
	if resourceID <= 0 || strings.TrimSpace(resourceType) == "" {
		return nil, false, nil
	}
	scopedResourceID := scopedResourceID(resourceType, resourceID)

	var resource permission.SysResource
	err := r.db.Where("resource_id = ? AND resource_type = ?", scopedResourceID, resourceType).First(&resource).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var permIDs []int64
	err = r.db.Model(&permission.SysResourcePerm{}).
		Where("resource_id = ?", scopedResourceID).
		Pluck("perm_id", &permIDs).Error
	if err != nil {
		return nil, true, err
	}

	return permIDs, true, nil
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

func resourcePermKeyPrefix(resourceType string) string {
	switch resourceType {
	case permission.ResourceTypeDashboard:
		return "dashboard:"
	case permission.ResourceTypeScreen:
		return "screen:"
	case permission.ResourceTypeDataset:
		return "dataset:"
	case permission.ResourceTypeDatasource:
		return "datasource:"
	default:
		return fmt.Sprintf("%s:", resourceType)
	}
}

func normalizeParentID(resourceType string, parentID *int64) *int64 {
	if parentID == nil {
		return nil
	}
	if *parentID <= 0 {
		zero := int64(0)
		return &zero
	}
	value := scopedResourceID(resourceType, *parentID)
	return &value
}

func dedupeInt64(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func scopedResourceID(resourceType string, resourceID int64) int64 {
	const resourceNamespaceStride int64 = 1_000_000_000_000
	var namespace int64
	switch resourceType {
	case permission.ResourceTypeDatasource:
		namespace = 1
	case permission.ResourceTypeDataset:
		namespace = 2
	case permission.ResourceTypeDashboard:
		namespace = 3
	case permission.ResourceTypeScreen:
		namespace = 4
	default:
		namespace = 9
	}
	return namespace*resourceNamespaceStride + resourceID
}
