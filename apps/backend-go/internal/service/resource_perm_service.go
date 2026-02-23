package service

import (
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"
)

type ResourcePermissionService struct {
	repo         *repository.ResourcePermissionRepository
	adminChecker AdminChecker
}

type AdminChecker interface {
	IsAdmin(userID int64) bool
}

func NewResourcePermissionService(repo *repository.ResourcePermissionRepository, adminChecker AdminChecker) *ResourcePermissionService {
	return &ResourcePermissionService{repo: repo, adminChecker: adminChecker}
}

func (s *ResourcePermissionService) CheckPermission(userID int64, resourceType, resourceID int64, permKey string) *permission.PermissionCheckResult {
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

func (s *ResourcePermissionService) CheckViewPermission(userID, resourceType, resourceID int64) bool {
	result := s.CheckPermission(userID, resourceType, resourceID, permission.PermKeyView)
	return result.HasPermission
}

func (s *ResourcePermissionService) CheckEditPermission(userID, resourceType, resourceID int64) bool {
	result := s.CheckPermission(userID, resourceType, resourceID, permission.PermKeyEdit)
	return result.HasPermission
}

func (s *ResourcePermissionService) CheckExportPermission(userID, resourceType, resourceID int64) bool {
	result := s.CheckPermission(userID, resourceType, resourceID, permission.PermKeyExport)
	return result.HasPermission
}

func (s *ResourcePermissionService) CheckManagePermission(userID, resourceType, resourceID int64) bool {
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
