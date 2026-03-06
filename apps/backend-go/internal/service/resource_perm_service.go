package service

import (
	"fmt"

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

	perm, err := s.repo.GetPermByKey(permKey)
	if err != nil {
		return &permission.PermissionCheckResult{HasPermission: false, Reason: "permission_not_found"}
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

// CheckPermissionConsistency 校验双视角权限一致性
func (s *ResourcePermissionService) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	
	return s.repo.CheckPermissionConsistency()
}
