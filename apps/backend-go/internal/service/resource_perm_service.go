package service

import (
	"fmt"
	"strings"

	"dataease/backend/internal/domain/permission"
)

type ResourcePermRepo interface {
	GetPermByID(permID int64) (*permission.SysPerm, error)
	GetPermByKey(permKey string) (*permission.SysPerm, error)
	ListPerms(permType string, page, size int) ([]*permission.SysPerm, int64, error)
	CreatePerm(perm *permission.SysPerm) error
	UpdatePerm(perm *permission.SysPerm) error
	DeletePerm(permID int64) error
	GetUserPerms(userID int64) ([]int64, error)
	GetRolePerms(roleID int64) ([]int64, error)
	GetUserRoleIDs(userID int64) ([]int64, error)
	CheckUserPermission(userID, permID int64) (bool, error)
	CheckRolePermission(roleID, permID int64) (bool, error)
	GrantPermToUser(userID, permID int64, createBy string) error
	RevokePermFromUser(userID, permID int64) error
	GrantPermToRole(roleID, permID int64) error
	RevokePermFromRole(roleID, permID int64) error
	// 双视角接口
	GetUserResources(userID int64, resourceType string) ([]*permission.UserResourcePermVO, error)
	GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error)
	ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error
	RegisterResource(resourceID int64, resourceName, resourceType string, parentID *int64) error
	ReplaceResourcePermissions(resourceID int64, resourceType string, permIDs []int64) error
	GetResourcePermissionIDs(resourceID int64, resourceType string) ([]int64, bool, error)
	CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error)
}

type AdminChecker interface {
	IsAdmin(userID int64) bool
}
type ResourcePermissionService struct {
	repo         ResourcePermRepo
	adminChecker AdminChecker
}

func NewResourcePermissionService(repo ResourcePermRepo, adminChecker AdminChecker) *ResourcePermissionService {
	return &ResourcePermissionService{repo: repo, adminChecker: adminChecker}
}

func (s *ResourcePermissionService) CheckPermission(userID int64, resourceType string, resourceID int64, permKey string) *permission.PermissionCheckResult {
	if s.adminChecker != nil && s.adminChecker.IsAdmin(userID) {
		return &permission.PermissionCheckResult{HasPermission: true, Reason: "admin"}
	}

	perm, err := s.lookupPermission(resourceType, permKey)
	if err != nil {
		return &permission.PermissionCheckResult{HasPermission: false, Reason: "permission_not_found"}
	}

	if resourceID > 0 {
		resourcePermIDs, exists, resourceErr := s.repo.GetResourcePermissionIDs(resourceID, resourceType)
		if resourceErr != nil {
			return &permission.PermissionCheckResult{HasPermission: false, Reason: "resource_permission_lookup_failed"}
		}
		if exists {
			if len(resourcePermIDs) == 0 {
				return &permission.PermissionCheckResult{HasPermission: false, Reason: "resource_permission_denied"}
			}
			if !containsInt64(resourcePermIDs, perm.PermID) {
				return &permission.PermissionCheckResult{HasPermission: false, Reason: "resource_permission_denied"}
			}
		}
	}

	hasUserPerm, err := s.repo.CheckUserPermission(userID, perm.PermID)
	if err == nil && hasUserPerm {
		return &permission.PermissionCheckResult{HasPermission: true, Reason: "user_permission"}
	}

	roleIDs, err := s.repo.GetUserRoleIDs(userID)
	if err != nil || len(roleIDs) == 0 {
		return &permission.PermissionCheckResult{HasPermission: false, Reason: "no_roles"}
	}

	for _, roleID := range roleIDs {
		hasRolePerm, err := s.repo.CheckRolePermission(roleID, perm.PermID)
		if err == nil && hasRolePerm {
			return &permission.PermissionCheckResult{HasPermission: true, Reason: "role_permission"}
		}
	}

	return &permission.PermissionCheckResult{HasPermission: false, Reason: "permission_denied"}
}

func (s *ResourcePermissionService) CheckViewPermission(userID int64, resourceType string, resourceID int64) bool {
	result := s.CheckPermission(userID, resourceType, resourceID, permission.PermKeyView)
	return result.HasPermission
}

func (s *ResourcePermissionService) CheckEditPermission(userID int64, resourceType string, resourceID int64) bool {
	result := s.CheckPermission(userID, resourceType, resourceID, permission.PermKeyEdit)
	return result.HasPermission
}

func (s *ResourcePermissionService) CheckExportPermission(userID int64, resourceType string, resourceID int64) bool {
	result := s.CheckPermission(userID, resourceType, resourceID, permission.PermKeyExport)
	return result.HasPermission
}

func (s *ResourcePermissionService) CheckManagePermission(userID int64, resourceType string, resourceID int64) bool {
	result := s.CheckPermission(userID, resourceType, resourceID, permission.PermKeyManage)
	return result.HasPermission
}

func (s *ResourcePermissionService) GetUserPermissionIDs(userID int64) ([]int64, error) {
	return s.repo.GetUserPerms(userID)
}

func (s *ResourcePermissionService) GetRolePermissionIDs(roleID int64) ([]int64, error) {
	return s.repo.GetRolePerms(roleID)
}

func (s *ResourcePermissionService) GetPermissionByID(permID int64) (*permission.SysPerm, error) {
	return s.repo.GetPermByID(permID)
}

func (s *ResourcePermissionService) GrantPermissionToUser(userID, permID int64, createBy string) error {
	return s.repo.GrantPermToUser(userID, permID, createBy)
}

func (s *ResourcePermissionService) RevokePermissionFromUser(userID, permID int64) error {
	return s.repo.RevokePermFromUser(userID, permID)
}

func (s *ResourcePermissionService) GrantPermissionToRole(roleID, permID int64) error {
	return s.repo.GrantPermToRole(roleID, permID)
}

func (s *ResourcePermissionService) RevokePermissionFromRole(roleID, permID int64) error {
	return s.repo.RevokePermFromRole(roleID, permID)
}

// ========== 双视角接口实现 ==========

// GetUserPerspective 获取用户视角的权限列表
func (s *ResourcePermissionService) GetUserPerspective(userID int64, resourceType string) ([]*permission.UserResourcePermVO, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	// 管理员返回所有权限标识
	if s.adminChecker != nil && s.adminChecker.IsAdmin(userID) {
		return []*permission.UserResourcePermVO{
			{PermKey: "*", PermName: "全部权限", SourceType: "admin"},
		}, nil
	}

	return s.repo.GetUserResources(userID, resourceType)
}

// GetResourcePerspective 获取资源视角的授权列表
func (s *ResourcePermissionService) GetResourcePerspective(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	return s.repo.GetResourceUsers(resourceID, resourceType)
}

// ApplyGroupPermissionsToResource 将分组权限应用到新资源
func (s *ResourcePermissionService) ApplyGroupPermissionsToResource(groupID, resourceID int64, resourceType string) error {
	if s.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	return s.repo.ApplyGroupPermissions(groupID, resourceID, resourceType)
}

func (s *ResourcePermissionService) RegisterResource(resourceID int64, resourceName, resourceType string, parentID *int64) error {
	if s.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	return s.repo.RegisterResource(resourceID, resourceName, resourceType, parentID)
}

func (s *ResourcePermissionService) InheritParentResourcePermissions(parentID, resourceID int64, resourceName, resourceType string) error {
	_, err := s.TryInheritParentResourcePermissions(parentID, resourceID, resourceName, resourceType)
	return err
}

func (s *ResourcePermissionService) TryInheritParentResourcePermissions(parentID, resourceID int64, resourceName, resourceType string) (bool, error) {
	if s.repo == nil {
		return false, fmt.Errorf("repository not initialized")
	}
	if parentID <= 0 || resourceID <= 0 || strings.TrimSpace(resourceType) == "" {
		return false, nil
	}
	permIDs, exists, err := s.repo.GetResourcePermissionIDs(parentID, resourceType)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	if err := s.repo.RegisterResource(resourceID, resourceName, resourceType, &parentID); err != nil {
		return false, err
	}
	if err := s.repo.ReplaceResourcePermissions(resourceID, resourceType, permIDs); err != nil {
		return false, err
	}
	return true, nil
}

func (s *ResourcePermissionService) ReplaceResourcePermissions(resourceID int64, resourceType string, permIDs []int64) error {
	if s.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	return s.repo.ReplaceResourcePermissions(resourceID, resourceType, permIDs)
}

func (s *ResourcePermissionService) ResolvePermission(resourceType, permKey string) (*permission.SysPerm, error) {
	return s.lookupPermission(resourceType, permKey)
}

// CheckPermissionConsistency 校验双视角权限一致性
func (s *ResourcePermissionService) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	return s.repo.CheckPermissionConsistency()
}

func (s *ResourcePermissionService) lookupPermission(resourceType, permKey string) (*permission.SysPerm, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	lookupKeys := make([]string, 0, 2)
	if strings.Contains(permKey, ":") {
		lookupKeys = append(lookupKeys, permKey)
	} else {
		lookupKeys = append(lookupKeys, fmt.Sprintf("%s:%s", resourceType, permKey), permKey)
	}

	for _, key := range lookupKeys {
		perm, err := s.repo.GetPermByKey(key)
		if err == nil && perm != nil {
			return perm, nil
		}
	}

	return nil, fmt.Errorf("permission %s not found", permKey)
}

func containsInt64(items []int64, target int64) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
