package service

import (
	"fmt"
	"strings"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/repository"
)

type scopedResourcePermRepo interface {
	GetUserResourcesByOrg(userID int64, resourceType string, orgID int64) ([]*permission.UserResourcePermVO, error)
	GetResourceUsersByOrg(resourceID int64, resourceType string, orgID int64) ([]*permission.ResourceUserPermVO, error)
	CheckPermissionConsistencyByOrg(orgID int64) (*permission.PermissionConsistencyResult, error)
	GrantPermToUserInOrg(userID, permID int64, createBy string, orgID int64) error
	RevokePermFromUserInOrg(userID, permID, orgID int64) error
}

type roleScopeResolver interface {
	GetByID(roleID int64) (*role.SysRole, error)
}

type userOrgMembershipResolver interface {
	IsUserInOrg(userID, orgID int64) (bool, error)
}

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
	auditor      permissionMutationAuditor
	roleResolver roleScopeResolver
	userOrgScope userOrgMembershipResolver
}

type ResourcePermissionServiceOption func(*ResourcePermissionService)

func WithResourcePermissionAuditor(auditor permissionMutationAuditor) ResourcePermissionServiceOption {
	return func(s *ResourcePermissionService) {
		s.auditor = auditor
	}
}

func WithResourcePermissionRoleResolver(resolver roleScopeResolver) ResourcePermissionServiceOption {
	return func(s *ResourcePermissionService) {
		s.roleResolver = resolver
	}
}

func WithResourcePermissionUserOrgResolver(resolver userOrgMembershipResolver) ResourcePermissionServiceOption {
	return func(s *ResourcePermissionService) {
		s.userOrgScope = resolver
	}
}

func NewResourcePermissionService(repo ResourcePermRepo, adminChecker AdminChecker, opts ...ResourcePermissionServiceOption) *ResourcePermissionService {
	svc := &ResourcePermissionService{repo: repo, adminChecker: adminChecker}
	for _, opt := range opts {
		if opt != nil {
			opt(svc)
		}
	}
	if repoImpl, ok := repo.(*repository.ResourcePermissionRepository); ok {
		if svc.auditor == nil {
			svc.auditor = newPermAuditHelperFromDB(repoImpl.DB())
		}
		if svc.roleResolver == nil {
			svc.roleResolver = repository.NewRoleRepository(repoImpl.DB())
		}
		if svc.userOrgScope == nil {
			svc.userOrgScope = repository.NewUserRoleRepository(repoImpl.DB())
		}
	}
	return svc
}

func (s *ResourcePermissionService) CheckPermission(userID int64, resourceType string, resourceID int64, permKey string) *permission.PermissionCheckResult {
	if s.adminChecker != nil && s.adminChecker.IsAdmin(userID) {
		return &permission.PermissionCheckResult{HasPermission: true, Reason: "admin"}
	}

	perm, err := s.lookupPermission(resourceType, permKey)
	if err != nil {
		return &permission.PermissionCheckResult{HasPermission: false, Reason: "permission_not_found"}
	}

	if result := s.checkResourcePermission(resourceID, resourceType, perm.PermID); result != nil {
		return result
	}

	if result := s.checkDirectUserPermission(userID, perm.PermID); result != nil {
		return result
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

func (s *ResourcePermissionService) checkResourcePermission(resourceID int64, resourceType string, permID int64) *permission.PermissionCheckResult {
	if resourceID <= 0 {
		return nil
	}
	resourcePermIDs, exists, err := s.repo.GetResourcePermissionIDs(resourceID, resourceType)
	if err != nil {
		return &permission.PermissionCheckResult{HasPermission: false, Reason: "resource_permission_lookup_failed"}
	}
	if !exists {
		return nil
	}
	if len(resourcePermIDs) == 0 || !containsInt64(resourcePermIDs, permID) {
		return &permission.PermissionCheckResult{HasPermission: false, Reason: "resource_permission_denied"}
	}
	return nil
}

func (s *ResourcePermissionService) checkDirectUserPermission(userID int64, permID int64) *permission.PermissionCheckResult {
	hasUserPerm, err := s.repo.CheckUserPermission(userID, permID)
	if err == nil && hasUserPerm {
		return &permission.PermissionCheckResult{HasPermission: true, Reason: "user_permission"}
	}
	return nil
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

func (s *ResourcePermissionService) GrantPermissionToUser(userID, permID int64, createBy string, scopes ...PermissionMutationScope) error {
	scope := resolvePermissionScope(scopes)
	if err := s.requireUserMutationScope(userID, scope); err != nil {
		return err
	}
	if scopedRepo, ok := s.repo.(scopedResourcePermRepo); ok && scope.OrgID > 0 && !s.isAdminActor(scope) {
		if err := scopedRepo.GrantPermToUserInOrg(userID, permID, createBy, scope.OrgID); err != nil {
			return err
		}
	} else if err := s.repo.GrantPermToUser(userID, permID, createBy); err != nil {
		return err
	}
	return s.recordMutationAudit("PERM_GRANT", scope, "user", userID, permID, 0, map[string]interface{}{"createBy": createBy})
}

func (s *ResourcePermissionService) RevokePermissionFromUser(userID, permID int64, scopes ...PermissionMutationScope) error {
	scope := resolvePermissionScope(scopes)
	if err := s.requireUserMutationScope(userID, scope); err != nil {
		return err
	}
	if scopedRepo, ok := s.repo.(scopedResourcePermRepo); ok && scope.OrgID > 0 && !s.isAdminActor(scope) {
		if err := scopedRepo.RevokePermFromUserInOrg(userID, permID, scope.OrgID); err != nil {
			return err
		}
	} else if err := s.repo.RevokePermFromUser(userID, permID); err != nil {
		return err
	}
	return s.recordMutationAudit("PERM_REVOKE", scope, "user", userID, permID, 0, nil)
}

func (s *ResourcePermissionService) GrantPermissionToRole(roleID, permID int64, scopes ...PermissionMutationScope) error {
	scope := resolvePermissionScope(scopes)
	if err := s.requireRoleMutationScope(roleID, scope); err != nil {
		return err
	}
	if err := s.repo.GrantPermToRole(roleID, permID); err != nil {
		return err
	}
	return s.recordMutationAudit("PERM_GRANT", scope, "role", roleID, permID, 0, nil)
}

func (s *ResourcePermissionService) RevokePermissionFromRole(roleID, permID int64, scopes ...PermissionMutationScope) error {
	scope := resolvePermissionScope(scopes)
	if err := s.requireRoleMutationScope(roleID, scope); err != nil {
		return err
	}
	if err := s.repo.RevokePermFromRole(roleID, permID); err != nil {
		return err
	}
	return s.recordMutationAudit("PERM_REVOKE", scope, "role", roleID, permID, 0, nil)
}

// ========== 双视角接口实现 ==========

// GetUserPerspective 获取用户视角的权限列表
func (s *ResourcePermissionService) GetUserPerspective(userID int64, resourceType string, scopes ...PermissionMutationScope) ([]*permission.UserResourcePermVO, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	scope := resolvePermissionScope(scopes)

	// 管理员返回所有权限标识
	if s.isAdminActor(scope) || (scope.isZero() && s.adminChecker != nil && s.adminChecker.IsAdmin(userID)) {
		return []*permission.UserResourcePermVO{
			{PermKey: "*", PermName: "全部权限", SourceType: "admin"},
		}, nil
	}
	if scope.OrgID > 0 {
		if err := requireOrgScope(scope); err != nil {
			return nil, err
		}
		if scopedRepo, ok := s.repo.(scopedResourcePermRepo); ok {
			return scopedRepo.GetUserResourcesByOrg(userID, resourceType, scope.OrgID)
		}
	}

	return s.repo.GetUserResources(userID, resourceType)
}

// GetResourcePerspective 获取资源视角的授权列表
func (s *ResourcePermissionService) GetResourcePerspective(resourceID int64, resourceType string, scopes ...PermissionMutationScope) ([]*permission.ResourceUserPermVO, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	scope := resolvePermissionScope(scopes)
	if scope.OrgID > 0 && !s.isAdminActor(scope) {
		if err := requireOrgScope(scope); err != nil {
			return nil, err
		}
		if scopedRepo, ok := s.repo.(scopedResourcePermRepo); ok {
			return scopedRepo.GetResourceUsersByOrg(resourceID, resourceType, scope.OrgID)
		}
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
func (s *ResourcePermissionService) CheckPermissionConsistency(scopes ...PermissionMutationScope) (*permission.PermissionConsistencyResult, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	scope := resolvePermissionScope(scopes)
	if scope.OrgID > 0 && !s.isAdminActor(scope) {
		if err := requireOrgScope(scope); err != nil {
			return nil, err
		}
		if scopedRepo, ok := s.repo.(scopedResourcePermRepo); ok {
			return scopedRepo.CheckPermissionConsistencyByOrg(scope.OrgID)
		}
	}

	return s.repo.CheckPermissionConsistency()
}

func (s *ResourcePermissionService) isAdminActor(scope PermissionMutationScope) bool {
	return scope.ActorID > 0 && s.adminChecker != nil && s.adminChecker.IsAdmin(scope.ActorID)
}

func (s *ResourcePermissionService) requireUserMutationScope(userID int64, scope PermissionMutationScope) error {
	if scope.OrgID <= 0 || s.isAdminActor(scope) {
		return nil
	}
	if err := requireOrgScope(scope); err != nil {
		return err
	}
	if s.userOrgScope == nil {
		return fmt.Errorf("user organization scope resolver not configured")
	}
	inOrg, err := s.userOrgScope.IsUserInOrg(userID, scope.OrgID)
	if err != nil {
		return fmt.Errorf("failed to validate user organization scope: %w", err)
	}
	if !inOrg {
		return fmt.Errorf("user does not belong to current organization")
	}
	return nil
}

func (s *ResourcePermissionService) requireRoleMutationScope(roleID int64, scope PermissionMutationScope) error {
	if scope.OrgID <= 0 || s.isAdminActor(scope) {
		return nil
	}
	if err := requireOrgScope(scope); err != nil {
		return err
	}
	if s.roleResolver == nil {
		return fmt.Errorf("role organization scope resolver not configured")
	}
	rle, err := s.roleResolver.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("failed to load role: %w", err)
	}
	return validateRoleOrgScope(rle, scope.OrgID)
}

func (s *ResourcePermissionService) recordMutationAudit(operation string, scope PermissionMutationScope, targetType string, targetID, permID, resourceID int64, details map[string]interface{}) error {
	if s.auditor == nil {
		return nil
	}
	if details == nil {
		details = map[string]interface{}{}
	}
	details["targetType"] = targetType
	details["targetId"] = targetID
	details["permId"] = permID
	return s.auditor.RecordPermissionMutationAudit(operation, scope, targetType, targetID, permID, resourceID, details)
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
