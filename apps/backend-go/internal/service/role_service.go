package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/governance"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrLastRoleRemovalBlocked is returned when attempting to remove a user's last role.
// This is an intentional policy deviation: the system blocks removal rather than cascading user deletion.
var ErrLastRoleRemovalBlocked = errors.New("cannot remove user's last role")

type RoleService struct {
	repo                *repository.RoleRepository
	userRepo            *repository.UserRepository
	userRoleRepo        *repository.UserRoleRepository
	resourcePermRepo    *repository.ResourcePermissionRepository
	governancePolicySvc *GovernancePolicyService
}

func NewRoleService(repo *repository.RoleRepository, userRepo *repository.UserRepository, userRoleRepo *repository.UserRoleRepository, governancePolicySvc *GovernancePolicyService) *RoleService {
	return &RoleService{repo: repo, userRepo: userRepo, userRoleRepo: userRoleRepo, governancePolicySvc: governancePolicySvc}
}

func (s *RoleService) SetResourcePermissionRepository(repo *repository.ResourcePermissionRepository) {
	s.resourcePermRepo = repo
}

func (s *RoleService) CreateRole(req *role.RoleCreator, createBy string, callerOrgID int64) (int64, error) {
	if err := requireGovernedOrgContext(callerOrgID); err != nil {
		return 0, err
	}

	logger.Info("Role creation with org context", zap.Int64("orgId", callerOrgID), zap.String("createBy", createBy))
	if req.ParentID != nil && *req.ParentID > 0 {
		if err := s.validateInheritance(*req.ParentID); err != nil {
			return 0, err
		}
	}

	roleName := req.RoleName
	if roleName == "" {
		roleName = req.Name
	}
	if roleName == "" {
		return 0, fmt.Errorf("role name is required")
	}

	roleDesc := req.RoleDesc
	if roleDesc == nil {
		roleDesc = req.Desc
	}

	roleCode := fmt.Sprintf("role_%d", time.Now().UnixNano())
	if req.RoleKey != "" {
		roleCode = req.RoleKey
	}

	rle := &role.SysRole{
		RoleName:  roleName,
		RoleCode:  roleCode,
		RoleDesc:  roleDesc,
		Status:    role.StatusEnabled,
		CreateBy:  &createBy,
		DataScope: strPtr(role.DataScopeSelf),
		ParentID:  req.ParentID,
	}
	if req.ParentID != nil && *req.ParentID > 0 {
		roleType := role.RoleTypeCustom
		rle.RoleType = &roleType
	}

	if req.Status != nil {
		rle.Status = *req.Status
	}

	if err := s.repo.Create(rle); err != nil {
		logger.Error("Failed to create role", zap.Error(err))
		return 0, fmt.Errorf("failed to create role: %w", err)
	}

	logger.Info("Role created", zap.Int64("roleId", rle.RoleID), zap.String("name", rle.RoleName))
	return rle.RoleID, nil
}

func (s *RoleService) EditRole(req *role.RoleEditor, updateBy string, callerOrgID int64) error {
	if err := requireGovernedOrgContext(callerOrgID); err != nil {
		return err
	}

	logger.Info("Role edit with org context", zap.Int64("orgId", callerOrgID), zap.String("updateBy", updateBy))

	// Normalize: RoleID takes precedence over ID
	roleID := req.RoleID
	if roleID == 0 {
		roleID = req.ID
	}
	if roleID == 0 {
		return fmt.Errorf("role id is required")
	}

	rle, err := s.repo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	if rle.RoleType != nil && *rle.RoleType == role.RoleTypeSystem {
		return fmt.Errorf("cannot edit built-in system role")
	}

	roleName := req.RoleName
	if roleName == "" {
		roleName = req.Name
	}
	if roleName != "" {
		rle.RoleName = roleName
	}
	roleDesc := req.RoleDesc
	if roleDesc == nil {
		roleDesc = req.Desc
	}
	if roleDesc != nil {
		rle.RoleDesc = roleDesc
	}
	if req.Status != nil {
		rle.Status = *req.Status
	}
	if req.ParentID != nil {
		if *req.ParentID > 0 {
			if err := s.validateInheritance(*req.ParentID); err != nil {
				return err
			}
			rle.ParentID = req.ParentID
			roleType := role.RoleTypeCustom
			rle.RoleType = &roleType
		} else {
			rle.ParentID = req.ParentID
		}
	}

	now := time.Now()
	rle.UpdateBy = &updateBy
	rle.UpdateTime = &now

	if err := s.repo.Update(rle); err != nil {
		logger.Error("Failed to update role", zap.Error(err))
		return fmt.Errorf("failed to update role: %w", err)
	}

	logger.Info("Role updated", zap.Int64("roleId", req.ID))
	return nil
}

func (s *RoleService) DeleteRole(roleID int64, callerOrgID int64) error {
	if err := requireGovernedOrgContext(callerOrgID); err != nil {
		return err
	}

	rle, err := s.repo.GetByID(roleID)
	if err != nil {
		return nil // role not found, nothing to delete
	}
	if rle.RoleType != nil && *rle.RoleType == role.RoleTypeSystem {
		return fmt.Errorf("cannot delete built-in role")
	}
	if err := s.repo.Delete(roleID); err != nil {
		logger.Error("Failed to delete role", zap.Error(err))
		return fmt.Errorf("failed to delete role: %w", err)
	}
	logger.Info("Role deleted", zap.Int64("roleId", roleID))
	return nil
}

func (s *RoleService) GetRoleByID(roleID int64) (*role.RoleDetailVO, error) {
	rle, err := s.repo.GetByID(roleID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	return &role.RoleDetailVO{
		ID:        rle.RoleID,
		Name:      rle.RoleName,
		Code:      rle.RoleCode,
		Desc:      rle.RoleDesc,
		ParentID:  rle.ParentID,
		Level:     rle.Level,
		DataScope: rle.DataScope,
		Status:    rle.Status,
	}, nil
}

func (s *RoleService) QueryRoles(req *role.RoleQueryRequest) ([]*role.RoleVO, error) {
	keyword := ""
	if req.Keyword != nil {
		keyword = *req.Keyword
	}

	roles, err := s.repo.Query(keyword)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}

	result := make([]*role.RoleVO, 0, len(roles))
	for _, rle := range roles {
		result = append(result, &role.RoleVO{
			ID:       rle.RoleID,
			Name:     rle.RoleName,
			Code:     rle.RoleCode,
			RoleType: rle.RoleType,
			Desc:     rle.RoleDesc,
			Status:   rle.Status,
			ReadOnly: rle.Readonly != nil && *rle.Readonly,
			Root:     rle.ParentID == nil || *rle.ParentID == 0,
		})
	}
	return result, nil
}

func (s *RoleService) QueryRolesPage(req *role.RolePageRequest) (*role.RolePageResult, error) {
	keyword := ""
	if req.Keyword != nil {
		keyword = *req.Keyword
	}
	roleType := ""
	if req.RoleType != nil {
		roleType = *req.RoleType
	}
	current := 1
	if req.Current > 0 {
		current = req.Current
	}
	size := 10
	if req.Size > 0 {
		size = req.Size
	}

	roles, total, err := s.repo.QueryWithPage(keyword, roleType, current, size)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}

	list := make([]*role.RoleVO, 0, len(roles))
	for _, rle := range roles {
		list = append(list, &role.RoleVO{
			ID:       rle.RoleID,
			Name:     rle.RoleName,
			Code:     rle.RoleCode,
			RoleType: rle.RoleType,
			Desc:     rle.RoleDesc,
			Status:   rle.Status,
			ReadOnly: rle.Readonly != nil && *rle.Readonly,
			Root:     rle.ParentID == nil || *rle.ParentID == 0,
		})
	}

	return &role.RolePageResult{
		List:    list,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func strPtr(s string) *string {
	return &s
}

// MountUsers 绑定组织用户到角色
func (s *RoleService) MountUsers(req *role.MountUserRequest) error {
	if s.userRoleRepo == nil {
		return fmt.Errorf("userRoleRepo not initialized")
	}
	if err := requireGovernedOrgContext(req.OrgId); err != nil {
		return err
	}

	for _, uid := range req.Uids {
		userRole := &user.SysUserRole{
			UserID: uid,
			RoleID: req.Rid,
			OrgID:  req.OrgId,
		}
		if _, err := s.userRoleRepo.CreateIfMissing(userRole); err != nil {
			logger.Error("Failed to bind user to role", zap.Int64("uid", uid), zap.Int64("rid", req.Rid), zap.Error(err))
			return fmt.Errorf("failed to bind user %d to role: %w", uid, err)
		}
	}

	logger.Info("Users mounted to role", zap.Int64("rid", req.Rid), zap.Int64s("uids", req.Uids))
	return nil
}

// MountExternalUser 绑定组织外用户到角色
func (s *RoleService) MountExternalUser(req *role.MountExternalUserRequest, orgID int64) error {
	if s.userRoleRepo == nil {
		return fmt.Errorf("userRoleRepo not initialized")
	}
	if orgID <= 0 {
		return fmt.Errorf("org id is required")
	}

	inOrg, err := s.userRoleRepo.IsUserInOrg(req.Uid, orgID)
	if err != nil {
		return fmt.Errorf("failed to validate external user organization: %w", err)
	}
	if inOrg {
		return fmt.Errorf("user already belongs to target organization")
	}

	userRole := &user.SysUserRole{
		UserID: req.Uid,
		RoleID: req.Rid,
		OrgID:  orgID,
	}
	if _, err := s.userRoleRepo.CreateIfMissing(userRole); err != nil {
		logger.Error("Failed to bind external user to role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid), zap.Error(err))
		return fmt.Errorf("failed to bind external user to role: %w", err)
	}

	logger.Info("External user mounted to role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid))
	return nil
}

// UnmountUser 解绑用户与角色（含唯一角色安全约束）
//
// Intentional Deviation: The system REJECTS removal of a user's last role instead of
// cascading to user deletion, even though some documentation implies cascade behavior.
// This is a documented policy decision for C1; cascade delete is deferred to C2 work.
//
// 如果用户只有这一个角色，将拒绝移除并返回 ErrLastRoleRemovalBlocked
func (s *RoleService) UnmountUser(req *role.UnmountUserRequest) error {
	if err := requireGovernedOrgContext(req.OrgId); err != nil {
		return err
	}

	var count int64
	var err error
	if req.OrgId > 0 {
		count, err = s.repo.CountUserRolesByOrg(req.Uid, req.OrgId)
	} else {
		count, err = s.repo.CountUserRoles(req.Uid)
	}
	if err != nil {
		logger.Error("Failed to count user roles", zap.Int64("uid", req.Uid), zap.Error(err))
		return fmt.Errorf("failed to check user role count: %w", err)
	}

	// 唯一角色安全约束：禁止移除用户的最后一个角色
	if count <= 1 {
		policy, err := s.getEffectiveLastRolePolicy(req.OrgId)
		if err != nil {
			return fmt.Errorf("failed to resolve last-role policy: %w", err)
		}

		switch policy {
		case governance.LastRolePolicyWarnAllow:
			if err := s.repo.UnbindUserRole(req.Uid, req.Rid, req.OrgId); err != nil {
				logger.Error("Failed to unbind user from role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid), zap.Error(err))
				return fmt.Errorf("failed to unbind user from role: %w", err)
			}
			s.recordLastRoleAudit(req, policy, audit.StatusSuccess, "warn_allow: removed user's last role without disabling account")
			logger.Warn("User last role removed under WARN_ALLOW policy", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid), zap.Int64("orgId", req.OrgId))
			return nil
		case governance.LastRolePolicyCascade:
			if s.userRepo == nil {
				return fmt.Errorf("userRepo not initialized")
			}
			if s.repo == nil || s.repo.DB() == nil {
				return fmt.Errorf("role repository is not configured")
			}
			if err := s.repo.DB().Transaction(func(txDB *gorm.DB) error {
				if req.Uid == 1 {
					return ErrBuiltInUserProtected
				}
				roleRepo := repository.NewRoleRepository(txDB)
				if err := roleRepo.UnbindUserRole(req.Uid, req.Rid, req.OrgId); err != nil {
					return fmt.Errorf("failed to unbind user from role: %w", err)
				}
				userRepo := repository.NewUserRepository(txDB)
				existing, err := userRepo.GetByID(req.Uid)
				if err != nil {
					return fmt.Errorf("user not found: %w", err)
				}
				now := time.Now()
				existing.Status = user.StatusDisabled
				existing.UpdateTime = &now
				if err := userRepo.Update(existing); err != nil {
					return fmt.Errorf("failed to disable user: %w", err)
				}
				return nil
			}); err != nil {
				s.recordLastRoleAudit(req, policy, audit.StatusFailed, err.Error())
				return err
			}
			s.recordLastRoleAudit(req, policy, audit.StatusSuccess, "cascade: removed user's last role and disabled account")
			logger.Warn("User last role removed under CASCADE policy", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid), zap.Int64("orgId", req.OrgId))
			return nil
		default:
			logger.Warn("Reject unmount last role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid), zap.Int64("count", count), zap.String("policy", string(policy)))
			s.recordLastRoleAudit(req, policy, audit.StatusFailed, ErrLastRoleRemovalBlocked.Error())
			return fmt.Errorf("%w", ErrLastRoleRemovalBlocked)
		}
	}

	if err := s.repo.UnbindUserRole(req.Uid, req.Rid, req.OrgId); err != nil {
		logger.Error("Failed to unbind user from role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid), zap.Error(err))
		return fmt.Errorf("failed to unbind user from role: %w", err)
	}

	logger.Info("User unmounted from role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid), zap.Int64("orgId", req.OrgId))
	return nil
}

func (s *RoleService) getEffectiveLastRolePolicy(orgID int64) (governance.LastRolePolicy, error) {
	if s.governancePolicySvc == nil {
		return governance.DefaultLastRolePolicy, nil
	}
	return s.governancePolicySvc.GetLastRolePolicy(orgID)
}

func (s *RoleService) recordLastRoleAudit(req *role.UnmountUserRequest, policy governance.LastRolePolicy, status audit.Status, detail string) {
	if s.governancePolicySvc == nil || s.governancePolicySvc.auditSvc == nil {
		return
	}
	beforeValue, _ := json.Marshal(map[string]interface{}{
		"uid":            req.Uid,
		"rid":            req.Rid,
		"orgId":          req.OrgId,
		"lastRolePolicy": policy,
	})
	afterValue, _ := json.Marshal(map[string]interface{}{
		"detail": detail,
	})
	resourceType := string(audit.ResourceTypeUser)
	userID := req.Uid
	username := systemActor
	_, _ = s.governancePolicySvc.auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
		UserID:         &userID,
		Username:       &username,
		ActionType:     audit.ActionTypeUserAction,
		ActionName:     "移除用户最后角色",
		ResourceType:   &resourceType,
		ResourceID:     &req.Uid,
		OrganizationID: &req.OrgId,
		Operation:      audit.OperationDelete,
		Status:         &status,
		FailureReason:  stringPtrIfNotEmpty(detail),
		BeforeValue:    stringPtrIfNotEmpty(string(beforeValue)),
		AfterValue:     stringPtrIfNotEmpty(string(afterValue)),
	})
}

// BeforeUnmountInfo 检查解绑前用户的角色数量（用于安全提示）
func (s *RoleService) BeforeUnmountInfo(req *role.UnmountUserRequest) (int, error) {
	if err := requireGovernedOrgContext(req.OrgId); err != nil {
		return 0, err
	}

	var count int64
	var err error
	if req.OrgId > 0 {
		count, err = s.repo.CountUserRolesByOrg(req.Uid, req.OrgId)
	} else {
		count, err = s.repo.CountUserRoles(req.Uid)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to count user roles: %w", err)
	}
	return int(count), nil
}

// SearchExternalUser 搜索组织外用户
func (s *RoleService) SearchExternalUser(keyword string, excludeOrgID int64) ([]*role.ExternalUserVO, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []*role.ExternalUserVO{}, nil
	}
	if excludeOrgID <= 0 {
		return nil, fmt.Errorf("org id is required")
	}
	if s.userRepo == nil {
		return nil, fmt.Errorf("userRepo not initialized")
	}

	users, err := s.userRepo.SearchExternalUser(keyword, excludeOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to search external users: %w", err)
	}

	result := make([]*role.ExternalUserVO, 0, len(users))
	for _, u := range users {
		result = append(result, &role.ExternalUserVO{
			Uid:     u.UserID,
			Account: u.Username,
			Name:    u.NickName,
			Email:   u.Email,
			Phone:   u.Phone,
		})
	}
	return result, nil
}

// OptionForUser 获取用户可选角色（组织内所有角色）
func (s *RoleService) OptionForUser(req *role.RoleRequest, orgID int64) ([]*role.RoleVO, error) {
	keyword := ""
	if req.Keyword != nil {
		keyword = *req.Keyword
	}
	return s.QueryRolesByOrgID(orgID, keyword)
}

func (s *RoleService) QueryRolesByOrgID(orgID int64, keyword string) ([]*role.RoleVO, error) {

	roles, err := s.repo.QueryByOrgID(orgID, keyword)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}

	result := make([]*role.RoleVO, 0, len(roles))
	for _, rle := range roles {
		result = append(result, &role.RoleVO{
			ID:       rle.RoleID,
			Name:     rle.RoleName,
			Code:     rle.RoleCode,
			RoleType: rle.RoleType,
			Desc:     rle.RoleDesc,
			Status:   rle.Status,
			ReadOnly: rle.Readonly != nil && *rle.Readonly,
			Root:     rle.ParentID == nil || *rle.ParentID == 0,
		})
	}
	return result, nil
}

// SelectedForUser 获取用户已选角色
func (s *RoleService) SelectedForUser(req *role.RoleRequest) ([]*role.RoleVO, error) {
	if req.Uid == nil {
		return nil, fmt.Errorf("uid is required")
	}

	roleIDs, err := s.repo.GetUserRoleIDs(*req.Uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role IDs: %w", err)
	}

	if len(roleIDs) == 0 {
		return []*role.RoleVO{}, nil
	}

	roles, err := s.repo.GetRolesByIDs(roleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles by IDs: %w", err)
	}

	result := make([]*role.RoleVO, 0, len(roles))
	for _, rle := range roles {
		result = append(result, &role.RoleVO{
			ID:       rle.RoleID,
			Name:     rle.RoleName,
			Code:     rle.RoleCode,
			RoleType: rle.RoleType,
			Desc:     rle.RoleDesc,
			Status:   rle.Status,
			ReadOnly: rle.Readonly != nil && *rle.Readonly,
			Root:     rle.ParentID == nil || *rle.ParentID == 0,
		})
	}
	return result, nil
}

// CreateRoleWithInheritance 创建带继承约束的自定义角色
// parentRoleID 指定父角色，新角色的权限不能超过父角色
func (s *RoleService) CreateRoleWithInheritance(req *role.RoleCreator, parentRoleID *int64, createBy string) (int64, error) {
	// 如果指定了父角色，验证继承约束
	if parentRoleID != nil && *parentRoleID > 0 {
		if err := s.validateInheritance(*parentRoleID); err != nil {
			return 0, err
		}
	}

	roleName := req.RoleName
	if roleName == "" {
		roleName = req.Name
	}
	if roleName == "" {
		return 0, fmt.Errorf("role name is required")
	}

	roleDesc := req.RoleDesc
	if roleDesc == nil {
		roleDesc = req.Desc
	}

	roleCode := fmt.Sprintf("role_%d", time.Now().UnixNano())
	if req.RoleKey != "" {
		roleCode = req.RoleKey
	}
	rle := &role.SysRole{
		RoleName:  roleName,
		RoleCode:  roleCode,
		RoleDesc:  roleDesc,
		Status:    role.StatusEnabled,
		CreateBy:  &createBy,
		DataScope: strPtr(role.DataScopeSelf),
		ParentID:  parentRoleID,
	}

	if err := s.repo.Create(rle); err != nil {
		logger.Error("Failed to create role with inheritance", zap.Error(err))
		return 0, fmt.Errorf("failed to create role: %w", err)
	}

	logger.Info("Role created with inheritance", zap.Int64("roleId", rle.RoleID), zap.String("name", rle.RoleName), zap.Int64p("parentId", parentRoleID))
	return rle.RoleID, nil
}

// validateInheritance 验证角色继承约束
// 确保父角色存在且为内置组织角色
func (s *RoleService) validateInheritance(parentRoleID int64) error {
	parent, err := s.repo.GetByID(parentRoleID)
	if err != nil {
		return fmt.Errorf("parent role not found: %w", err)
	}

	// 检查父角色状态
	if parent.Status != role.StatusEnabled {
		return fmt.Errorf("parent role is disabled")
	}

	if parent.ParentID != nil && *parent.ParentID != 0 {
		return fmt.Errorf("parent role must be a built-in root role")
	}

	if parent.RoleType != nil && *parent.RoleType == role.RoleTypeCustom {
		return fmt.Errorf("custom role cannot be used as parent role")
	}

	return nil
}

// ValidatePermissionInheritance 验证权限分配不超过继承边界
// 当角色有父角色时，分配的权限不能超过父角色的权限范围
func (s *RoleService) ValidatePermissionInheritance(roleID int64, permIDs []int64) error {
	rle, err := s.repo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 如果没有父角色，则无继承约束
	if rle.ParentID == nil || *rle.ParentID == 0 {
		return nil
	}
	if len(permIDs) == 0 {
		return nil
	}
	if s.resourcePermRepo == nil {
		return fmt.Errorf("resource permission repository is not configured")
	}

	parentPermIDs, err := s.resourcePermRepo.GetRolePerms(*rle.ParentID)
	if err != nil {
		return fmt.Errorf("failed to load parent role permissions: %w", err)
	}

	parentPermSet := make(map[int64]struct{}, len(parentPermIDs))
	for _, permID := range parentPermIDs {
		parentPermSet[permID] = struct{}{}
	}

	for _, permID := range permIDs {
		if _, ok := parentPermSet[permID]; ok {
			continue
		}
		return fmt.Errorf("permission inheritance violation: permission %d exceeds parent role scope", permID)
	}

	logger.Info("Permission inheritance validation",
		zap.Int64("roleID", roleID),
		zap.Int64("parentID", *rle.ParentID),
		zap.Int("permCount", len(permIDs)),
		zap.Int64s("validatedPermIDs", slices.Clone(permIDs)))

	return nil
}
