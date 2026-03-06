package service

import (
	"fmt"
	"time"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
)

type RoleService struct {
	repo         *repository.RoleRepository
	userRepo     *repository.UserRepository
	userRoleRepo *repository.UserRoleRepository
}

func NewRoleService(repo *repository.RoleRepository, userRepo *repository.UserRepository, userRoleRepo *repository.UserRoleRepository) *RoleService {
	return &RoleService{repo: repo, userRepo: userRepo, userRoleRepo: userRoleRepo}
}

func (s *RoleService) CreateRole(req *role.RoleCreator, createBy string) (int64, error) {
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

func (s *RoleService) EditRole(req *role.RoleEditor, updateBy string) error {
	if req.ID == 0 && req.RoleID > 0 {
		req.ID = req.RoleID
	}
	if req.ID == 0 {
		return fmt.Errorf("role id is required")
	}

	rle, err := s.repo.GetByID(req.ID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
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

func (s *RoleService) DeleteRole(roleID int64) error {
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
			Desc:     rle.RoleDesc,
			Status:   rle.Status,
			ReadOnly: false,
			Root:     rle.ParentID == nil || *rle.ParentID == 0,
		})
	}
	return result, nil
}

func strPtr(s string) *string {
	return &s
}

// MountUsers 绑定组织用户到角色
func (s *RoleService) MountUsers(req *role.MountUserRequest) error {
	if s.userRoleRepo == nil {
		return fmt.Errorf("userRoleRepo not initialized")
	}

	for _, uid := range req.Uids {
		userRole := &user.SysUserRole{
			UserID: uid,
			RoleID: req.Rid,
			OrgID:  req.OrgId,
		}
		if err := s.userRoleRepo.Create(userRole); err != nil {
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

	userRole := &user.SysUserRole{
		UserID: req.Uid,
		RoleID: req.Rid,
		OrgID:  orgID,
	}
	if err := s.userRoleRepo.Create(userRole); err != nil {
		logger.Error("Failed to bind external user to role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid), zap.Error(err))
		return fmt.Errorf("failed to bind external user to role: %w", err)
	}

	logger.Info("External user mounted to role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid))
	return nil
}

// UnmountUser 解绑用户与角色（含唯一角色安全约束）
// 如果用户只有这一个角色，将拒绝移除并返回错误
func (s *RoleService) UnmountUser(req *role.UnmountUserRequest) error {
	// 检查用户的角色数量
	count, err := s.repo.CountUserRoles(req.Uid)
	if err != nil {
		logger.Error("Failed to count user roles", zap.Int64("uid", req.Uid), zap.Error(err))
		return fmt.Errorf("failed to check user role count: %w", err)
	}

	// 唯一角色安全约束：禁止移除用户的最后一个角色
	if count <= 1 {
		logger.Warn("Reject unmount last role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid), zap.Int64("count", count))
		return fmt.Errorf("cannot remove user's last role: user must have at least one role")
	}

	if err := s.repo.UnbindUserRole(req.Uid, req.Rid); err != nil {
		logger.Error("Failed to unbind user from role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid), zap.Error(err))
		return fmt.Errorf("failed to unbind user from role: %w", err)
	}

	logger.Info("User unmounted from role", zap.Int64("uid", req.Uid), zap.Int64("rid", req.Rid))
	return nil
}

// BeforeUnmountInfo 检查解绑前用户的角色数量（用于安全提示）
func (s *RoleService) BeforeUnmountInfo(req *role.UnmountUserRequest) (int, error) {
	count, err := s.repo.CountUserRoles(req.Uid)
	if err != nil {
		return 0, fmt.Errorf("failed to count user roles: %w", err)
	}
	return int(count), nil
}

// SearchExternalUser 搜索组织外用户
func (s *RoleService) SearchExternalUser(keyword string, excludeOrgID int64) ([]*role.ExternalUserVO, error) {
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
			Desc:     rle.RoleDesc,
			Status:   rle.Status,
			ReadOnly: false,
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
			Desc:     rle.RoleDesc,
			Status:   rle.Status,
			ReadOnly: false,
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

	// 内置角色（root 角色）才能作为父角色
	// parent_id 为 null 或 0 表示根角色/内置角色
	if parent.ParentID != nil && *parent.ParentID != 0 {
		// 父角色本身也是自定义角色，需要递归验证
		if err := s.validateInheritance(*parent.ParentID); err != nil {
			return fmt.Errorf("grandparent role inheritance invalid: %w", err)
		}
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

	// TODO: 实现权限边界验证
	// 这需要查询父角色的权限列表，并验证新分配的权限是否在父角色权限范围内
	// 当前版本先记录日志，后续迭代完善
	logger.Info("Permission inheritance validation",
		zap.Int64("roleID", roleID),
		zap.Int64("parentID", *rle.ParentID),
		zap.Int("permCount", len(permIDs)))

	return nil
}
