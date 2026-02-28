package service

import (
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
