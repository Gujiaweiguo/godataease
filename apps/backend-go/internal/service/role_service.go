package service

import (
	"fmt"
	"time"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
)

type RoleService struct {
	repo *repository.RoleRepository
}

func NewRoleService(repo *repository.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func (s *RoleService) CreateRole(req *role.RoleCreator, createBy string) (int64, error) {
	roleCode := fmt.Sprintf("role_%d", time.Now().UnixNano())

	rle := &role.SysRole{
		RoleName:  req.Name,
		RoleCode:  roleCode,
		RoleDesc:  req.Desc,
		Status:    role.StatusEnabled,
		CreateBy:  &createBy,
		DataScope: strPtr(role.DataScopeSelf),
	}

	if err := s.repo.Create(rle); err != nil {
		logger.Error("Failed to create role", zap.Error(err))
		return 0, fmt.Errorf("failed to create role: %w", err)
	}

	logger.Info("Role created", zap.Int64("roleId", rle.RoleID), zap.String("name", rle.RoleName))
	return rle.RoleID, nil
}

func (s *RoleService) EditRole(req *role.RoleEditor, updateBy string) error {
	rle, err := s.repo.GetByID(req.ID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	rle.RoleName = req.Name
	rle.RoleDesc = req.Desc
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
			ReadOnly: false,
			Root:     rle.ParentID == nil || *rle.ParentID == 0,
		})
	}
	return result, nil
}

func strPtr(s string) *string {
	return &s
}
