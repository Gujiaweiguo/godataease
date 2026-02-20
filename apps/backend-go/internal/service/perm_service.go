package service

import (
	"fmt"
	"time"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
)

type PermService struct {
	permRepo repository.PermRepositoryInterface
}

func NewPermService(permRepo repository.PermRepositoryInterface) *PermService {
	return &PermService{
		permRepo: permRepo,
	}
}

func (s *PermService) CreatePerm(req *permission.PermCreateRequest) (int64, error) {
	count, err := s.permRepo.CheckKeyExists(req.PermKey, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to check perm key: %w", err)
	}
	if count > 0 {
		return 0, fmt.Errorf("permission key already exists")
	}

	permType := req.PermType
	if permType == "" {
		permType = permission.PermTypeMenu
	}

	p := &permission.SysPerm{
		PermName: req.PermName,
		PermKey:  req.PermKey,
		PermType: permType,
		PermDesc: req.PermDesc,
		Status:   permission.StatusEnabled,
		DelFlag:  permission.DelFlagNormal,
	}

	if req.Status != nil {
		p.Status = *req.Status
	}

	if err := s.permRepo.Create(p); err != nil {
		logger.Error("Failed to create permission", zap.Error(err))
		return 0, fmt.Errorf("failed to create permission: %w", err)
	}

	logger.Info("Permission created", zap.Int64("permId", p.PermID), zap.String("permKey", p.PermKey))
	return p.PermID, nil
}

func (s *PermService) UpdatePerm(req *permission.PermUpdateRequest) error {
	existing, err := s.permRepo.GetByID(req.PermID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	if req.PermKey != "" && req.PermKey != existing.PermKey {
		count, err := s.permRepo.CheckKeyExists(req.PermKey, req.PermID)
		if err != nil {
			return fmt.Errorf("failed to check perm key: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("permission key already exists")
		}
		existing.PermKey = req.PermKey
	}

	if req.PermName != "" {
		existing.PermName = req.PermName
	}
	if req.PermType != "" {
		existing.PermType = req.PermType
	}
	if req.PermDesc != nil {
		existing.PermDesc = req.PermDesc
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}

	now := time.Now()
	existing.UpdateTime = &now

	if err := s.permRepo.Update(existing); err != nil {
		logger.Error("Failed to update permission", zap.Error(err))
		return fmt.Errorf("failed to update permission: %w", err)
	}

	logger.Info("Permission updated", zap.Int64("permId", req.PermID))
	return nil
}

func (s *PermService) DeletePerm(permID int64) error {
	if err := s.permRepo.Delete(permID); err != nil {
		logger.Error("Failed to delete permission", zap.Error(err))
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	logger.Info("Permission deleted", zap.Int64("permId", permID))
	return nil
}

func (s *PermService) GetPermByID(permID int64) (*permission.SysPerm, error) {
	return s.permRepo.GetByID(permID)
}

func (s *PermService) ListPerms(req *permission.PermQueryRequest) (*permission.PermListResponse, error) {
	perms, err := s.permRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	page := req.Current
	pageSize := req.Size
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	total := int64(len(perms))
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > int(len(perms)) {
		end = len(perms)
	}
	if start > int(len(perms)) {
		start = len(perms)
	}

	return &permission.PermListResponse{
		List:    perms[start:end],
		Total:   total,
		Current: page,
		Size:    pageSize,
	}, nil
}

func (s *PermService) CheckPermKeyExists(permKey string) (bool, error) {
	count, err := s.permRepo.CheckKeyExists(permKey, 0)
	if err != nil {
		return false, fmt.Errorf("failed to check perm key: %w", err)
	}
	return count > 0, nil
}
