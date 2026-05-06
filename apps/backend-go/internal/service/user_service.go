package service

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"dataease/backend/internal/domain/audit"
	domainorg "dataease/backend/internal/domain/org"
	domainrole "dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	DefaultBcryptCost      = 10 // 与 Java 版本一致
	DefaultPasswordEnvName = "USER_DEFAULT_PASSWORD"
	FallbackDefaultPwd     = "DataEase123456"
)

var (
	ErrBuiltInUserProtected = fmt.Errorf("cannot modify built-in admin user")
	ErrInvalidStatus        = fmt.Errorf("invalid status value: must be 0 (disabled) or 1 (enabled)")
	ErrUserNotInCurrentOrg  = fmt.Errorf("user not in current organization")
)

type UserService struct {
	userRepo     *repository.UserRepository
	userRoleRepo *repository.UserRoleRepository
	userPermRepo *repository.UserPermRepository
	roleRepo     *repository.RoleRepository
	orgRepo      *repository.OrgRepository
	auditSvc     *AuditService // 审计服务（可选）
}

func NewUserService(
	userRepo *repository.UserRepository,
	userRoleRepo *repository.UserRoleRepository,
	userPermRepo *repository.UserPermRepository,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		userPermRepo: userPermRepo,
		auditSvc:     nil, // 默认为空，可通过 SetAuditService 注入
	}
}

func (s *UserService) SetRoleRepository(repo *repository.RoleRepository) {
	s.roleRepo = repo
}

func (s *UserService) SetOrgRepository(repo *repository.OrgRepository) {
	s.orgRepo = repo
}

// SetAuditService 设置审计服务
func (s *UserService) SetAuditService(svc *AuditService) {
	s.auditSvc = svc
}

func (s *UserService) isBuiltInAdminUser(userID int64) bool {
	return userID == 1
}

func (s *UserService) EnsureUserInOrg(userID int64, orgID int64) error {
	if orgID <= 0 {
		return fmt.Errorf("org id is required")
	}
	if s.userRoleRepo == nil {
		return fmt.Errorf("user-role repository is not configured")
	}

	inOrg, err := s.userRoleRepo.IsUserInOrg(userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to validate user organization: %w", err)
	}
	if !inOrg {
		return ErrUserNotInCurrentOrg
	}

	return nil
}

// CreateUser 创建用户（含密码加密）
func (s *UserService) CreateUser(req *user.UserCreateRequest) (int64, error) {
	// 检查用户名是否存在
	count, err := s.userRepo.CountByUsername(req.Username)
	if err != nil {
		return 0, fmt.Errorf("failed to check username: %w", err)
	}
	if count > 0 {
		return 0, fmt.Errorf("username already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), DefaultBcryptCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	u := &user.SysUser{
		Username: req.Username,
		Password: string(hashedPassword),
		NickName: req.RealName,
		Email:    req.Email,
		Phone:    req.Phone,
		From:     user.FromLocal,
		Status:   user.StatusEnabled,
		DelFlag:  user.DelFlagNormal,
	}

	if req.Status != nil {
		u.Status = *req.Status
	}

	if err := s.userRepo.Create(u); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	if orgID, ok := requestedOrgID(req.OrgID, req.OrganizationID); ok {
		if err := s.bindUserToOrgBaseline(u.UserID, orgID); err != nil {
			_ = s.userRepo.Delete(u.UserID)
			return 0, err
		}
	}

	logger.Info("User created", zap.Int64("userId", u.UserID), zap.String("username", u.Username))
	return u.UserID, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(req *user.UserUpdateRequest) error {
	existing, err := s.userRepo.GetByID(req.ID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if req.Username != "" {
		existing.Username = req.Username
	}
	if req.RealName != "" {
		existing.NickName = req.RealName
	}
	if req.Email != nil {
		existing.Email = req.Email
	}
	if req.Phone != nil {
		existing.Phone = req.Phone
	}
	if req.Password != nil && *req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), DefaultBcryptCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		existing.Password = string(hashedPassword)
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}

	now := time.Now()
	existing.UpdateTime = &now

	if err := s.userRepo.Update(existing); err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	newOrgID := resolveOrgID(req.OrgID, req.OrganizationID)
	if newOrgID != nil && *newOrgID > 0 {
		if err := s.switchUserOrg(req.ID, *newOrgID); err != nil {
			logger.Error("Failed to switch user org", zap.Int64("userId", req.ID), zap.Error(err))
			return fmt.Errorf("failed to update user organization: %w", err)
		}
	}

	logger.Info("User updated", zap.Int64("userId", req.ID))
	return nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(userID int64) error {
	if s.isBuiltInAdminUser(userID) {
		return ErrBuiltInUserProtected
	}

	if err := s.userRepo.Delete(userID); err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// 删除关联的角色和权限
	_ = s.userRoleRepo.DeleteByUserID(userID)
	_ = s.userPermRepo.DeleteByUserID(userID)

	logger.Info("User deleted", zap.Int64("userId", userID))
	return nil
}

// GetUserByID 根据ID查询用户
func (s *UserService) GetUserByID(userID int64) (*user.SysUser, error) {
	return s.userRepo.GetByID(userID)
}

// GetUserByUsername 根据用户名查询用户
func (s *UserService) GetUserByUsername(username string) (*user.SysUser, error) {
	return s.userRepo.GetByUsername(username)
}

// SearchUsers 搜索用户（多条件查询 + 分页）
func (s *UserService) SearchUsers(req *user.UserQueryRequest) (*user.UserListResponse, error) {
	users, total, err := s.userRepo.Query(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	page := req.Current
	pageSize := req.Size
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	return &user.UserListResponse{
		List:    users,
		Total:   total,
		Current: page,
		Size:    pageSize,
	}, nil
}

// ResetPassword 重置密码
func (s *UserService) ResetPassword(userID int64, newPassword string) error {
	return s.ResetPasswordWithAudit(userID, newPassword, 0, systemActor, "127.0.0.1")
}

// ResetPasswordWithAudit 重置密码（含审计日志）
func (s *UserService) ResetPasswordWithAudit(userID int64, newPassword string, operatorID int64, operatorName string, ipAddress string) error {
	if s.isBuiltInAdminUser(userID) {
		return ErrBuiltInUserProtected
	}

	existing, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.recordPasswordResetAudit(nil, userID, operatorID, operatorName, ipAddress, audit.StatusFailed, "user not found")
		return fmt.Errorf("user not found: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), DefaultBcryptCost)
	if err != nil {
		s.recordPasswordResetAudit(existing, userID, operatorID, operatorName, ipAddress, audit.StatusFailed, "failed to hash password")
		return fmt.Errorf("failed to hash password: %w", err)
	}

	existing.Password = string(hashedPassword)
	now := time.Now()
	existing.UpdateTime = &now

	if err := s.userRepo.Update(existing); err != nil {
		logger.Error("Failed to reset password", zap.Error(err))
		s.recordPasswordResetAudit(existing, userID, operatorID, operatorName, ipAddress, audit.StatusFailed, err.Error())
		return fmt.Errorf("failed to reset password: %w", err)
	}

	logger.Info("Password reset", zap.Int64("userId", userID))
	s.recordPasswordResetAudit(existing, userID, operatorID, operatorName, ipAddress, audit.StatusSuccess, "")
	return nil
}

// recordPasswordResetAudit 记录密码重置审计日志
func (s *UserService) recordPasswordResetAudit(user *user.SysUser, userID int64, operatorID int64, operatorName string, ipAddress string, status audit.Status, failureReason string) {
	if s.auditSvc == nil {
		return
	}

	resourceType := string(audit.ResourceTypeUser)
	req := &audit.AuditLogCreateRequest{
		UserID:       &operatorID,
		Username:     &operatorName,
		ActionType:   audit.ActionTypeUserAction,
		ActionName:   "重置密码",
		ResourceType: &resourceType,
		ResourceID:   &userID,
		Operation:    audit.OperationUpdate,
		IPAddress:    &ipAddress,
		Status:       &status,
	}

	if user != nil {
		req.ResourceName = &user.Username
		if status == audit.StatusSuccess {
			req.BeforeValue = ptrUserStr("[REDACTED]")
			req.AfterValue = ptrUserStr("password reset to default policy")
		}
	}

	if failureReason != "" {
		req.FailureReason = &failureReason
	}

	// 异步记录审计日志
	go func() {
		_, _ = s.auditSvc.CreateAuditLog(req)
	}()
}

func ptrUserStr(v string) *string {
	return &v
}

func (s *UserService) ResolveDefaultPassword() string {
	if pwd := os.Getenv(DefaultPasswordEnvName); pwd != "" {
		return pwd
	}
	return FallbackDefaultPwd
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(userID int64, status int) error {
	if status != user.StatusEnabled && status != user.StatusDisabled {
		return ErrInvalidStatus
	}
	if s.isBuiltInAdminUser(userID) && status == user.StatusDisabled {
		return ErrBuiltInUserProtected
	}

	existing, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	existing.Status = status
	now := time.Now()
	existing.UpdateTime = &now

	if err := s.userRepo.Update(existing); err != nil {
		logger.Error("Failed to update user status", zap.Error(err))
		return fmt.Errorf("failed to update user status: %w", err)
	}

	logger.Info("User status updated", zap.Int64("userId", userID), zap.Int("status", status))
	return nil
}

func (s *UserService) SwitchLanguage(userID int64, lang string) error {
	existing, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	normalized := normalizeLanguage(lang)
	existing.Language = &normalized
	now := time.Now()
	existing.UpdateTime = &now

	if err := s.userRepo.Update(existing); err != nil {
		logger.Error("Failed to switch user language", zap.Error(err))
		return fmt.Errorf("failed to switch user language: %w", err)
	}

	logger.Info("User language updated", zap.Int64("userId", userID), zap.String("language", normalized))
	return nil
}

func normalizeLanguage(lang string) string {
	switch strings.TrimSpace(strings.ToLower(strings.ReplaceAll(lang, "_", "-"))) {
	case "zh-cn", "zh":
		return defaultLanguageZhCN
	case "zh-tw", "tw":
		return "tw"
	case "en", "en-us":
		return "en"
	default:
		return defaultLanguageZhCN
	}
}

func requestedOrgID(orgID *int64, organizationID *int64) (int64, bool) {
	if organizationID != nil && *organizationID > 0 {
		return *organizationID, true
	}
	if orgID != nil && *orgID > 0 {
		return *orgID, true
	}
	return 0, false
}

// bindUserToOrgBaseline establishes the user's organization-scoped membership baseline.
// This is the authoritative entry point for org-scoped user binding that downstream
// role workflows (MountUsers, UnmountUser, etc.) must reuse through the same org context.
func (s *UserService) bindUserToOrgBaseline(userID int64, orgID int64) error {
	if s.orgRepo == nil {
		return fmt.Errorf("org repository is not configured")
	}
	if s.userRoleRepo == nil {
		return fmt.Errorf("user-role repository is not configured")
	}

	orgEntity, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}
	if orgEntity.Status != domainorg.StatusEnabled {
		return fmt.Errorf("organization is disabled")
	}

	defaultRoleID, err := s.ensureDefaultOrgUserRole()
	if err != nil {
		return err
	}

	if err := s.userRoleRepo.Create(&user.SysUserRole{UserID: userID, RoleID: defaultRoleID, OrgID: orgID}); err != nil {
		return fmt.Errorf("failed to persist organization membership baseline: %w", err)
	}

	return nil
}

func (s *UserService) ensureDefaultOrgUserRole() (int64, error) {
	if s.roleRepo == nil {
		return 0, fmt.Errorf("role repository is not configured")
	}

	existing, err := s.roleRepo.GetByRoleCode(domainrole.BuiltInOrgUserRoleCode)
	if err == nil {
		if existing.Status != domainrole.StatusEnabled {
			existing.Status = domainrole.StatusEnabled
			if updateErr := s.roleRepo.Update(existing); updateErr != nil {
				return 0, fmt.Errorf("failed to enable default organization role: %w", updateErr)
			}
		}
		return existing.RoleID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to load default organization role: %w", err)
	}

	roleType := domainrole.RoleTypeOrganization
	dataScope := domainrole.DataScopeSelf
	createBy := systemActor
	defaultRole := &domainrole.SysRole{
		RoleName:  domainrole.BuiltInOrgUserRoleName,
		RoleCode:  domainrole.BuiltInOrgUserRoleCode,
		RoleType:  &roleType,
		DataScope: &dataScope,
		Status:    domainrole.StatusEnabled,
		CreateBy:  &createBy,
	}

	if err := s.roleRepo.Create(defaultRole); err != nil {
		return 0, fmt.Errorf("failed to create default organization role: %w", err)
	}

	return defaultRole.RoleID, nil
}

func resolveOrgID(orgID *int64, organizationID *int64) *int64 {
	if orgID != nil && *orgID > 0 {
		return orgID
	}
	if organizationID != nil && *organizationID > 0 {
		return organizationID
	}
	return nil
}

func (s *UserService) switchUserOrg(userID int64, newOrgID int64) error {
	if s.orgRepo == nil {
		return fmt.Errorf("org repository is not configured")
	}
	orgEntity, err := s.orgRepo.GetByID(newOrgID)
	if err != nil {
		return fmt.Errorf("target organization not found: %w", err)
	}
	if orgEntity.Status != domainorg.StatusEnabled {
		return fmt.Errorf("target organization is disabled")
	}
	return s.userRoleRepo.SwitchOrgForUser(userID, newOrgID)
}
