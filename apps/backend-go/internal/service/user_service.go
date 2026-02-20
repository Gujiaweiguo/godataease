package service

import (
	"fmt"
	"time"

	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultBcryptCost = 10 // 与 Java 版本一致
)

type UserService struct {
	userRepo     *repository.UserRepository
	userRoleRepo *repository.UserRoleRepository
	userPermRepo *repository.UserPermRepository
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
	}
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

	logger.Info("User updated", zap.Int64("userId", req.ID))
	return nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(userID int64) error {
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
	existing, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), DefaultBcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	existing.Password = string(hashedPassword)
	now := time.Now()
	existing.UpdateTime = &now

	if err := s.userRepo.Update(existing); err != nil {
		logger.Error("Failed to reset password", zap.Error(err))
		return fmt.Errorf("failed to reset password: %w", err)
	}

	logger.Info("Password reset", zap.Int64("userId", userID))
	return nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(userID int64, status int) error {
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
