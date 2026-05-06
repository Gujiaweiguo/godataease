package service

import (
	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"
	"errors"

	"go.uber.org/zap"
)

var (
	ErrRoleNotFound   = errors.New("role not found")
	ErrMenuNotFound   = errors.New("menu not found")
	ErrInvalidRoleID  = errors.New("invalid role id")
	ErrInvalidMenuIDs = errors.New("invalid menu ids")
)

type RoleMenuService struct {
	roleMenuRepo *repository.RoleMenuRepository
	roleRepo     *repository.RoleRepository
	menuRepo     *repository.MenuRepository
	menuService  *MenuService
}

func NewRoleMenuService(
	roleMenuRepo *repository.RoleMenuRepository,
	roleRepo *repository.RoleRepository,
	menuRepo *repository.MenuRepository,
	menuService *MenuService,
) *RoleMenuService {
	return &RoleMenuService{
		roleMenuRepo: roleMenuRepo,
		roleRepo:     roleRepo,
		menuRepo:     menuRepo,
		menuService:  menuService,
	}
}

type RoleMenuAuthVO struct {
	RoleID   int64          `json:"roleId"`
	MenuIDs  []int64        `json:"menuIds"`
	MenuTree []*menu.MenuVO `json:"menuTree"`
}

type SaveRoleMenuRequest struct {
	RoleID  int64   `json:"roleId" binding:"required"`
	MenuIDs []int64 `json:"menuIds"`
}

func (s *RoleMenuService) GetRoleMenuAuth(roleID int64, scopes ...PermissionMutationScope) (*RoleMenuAuthVO, error) {
	if roleID <= 0 {
		return nil, ErrInvalidRoleID
	}

	rle, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		logger.Error("Role not found", zap.Int64("roleId", roleID), zap.Error(err))
		return nil, ErrRoleNotFound
	}
	scope := resolvePermissionScope(scopes)
	if scope.OrgID > 0 {
		if err := requireOrgScope(scope); err != nil {
			return nil, err
		}
		if err := validateRoleOrgScope(rle, scope.OrgID); err != nil {
			return nil, err
		}
	}

	menuIDs, err := s.roleMenuRepo.GetMenuIDsByRoleID(roleID)
	if err != nil {
		logger.Error("Failed to get role menu auth", zap.Int64("roleId", roleID), zap.Error(err))
		return nil, err
	}

	var menuTree []*menu.MenuVO
	if s.menuService != nil {
		menuTree, _ = s.menuService.Query()
	}

	return &RoleMenuAuthVO{
		RoleID:   roleID,
		MenuIDs:  menuIDs,
		MenuTree: menuTree,
	}, nil
}

func (s *RoleMenuService) SaveRoleMenuAuth(req *SaveRoleMenuRequest) error {
	if req.RoleID <= 0 {
		return ErrInvalidRoleID
	}

	_, err := s.roleRepo.GetByID(req.RoleID)
	if err != nil {
		logger.Error("Role not found", zap.Int64("roleId", req.RoleID), zap.Error(err))
		return ErrRoleNotFound
	}

	if len(req.MenuIDs) > 0 {
		allMenus, err := s.menuRepo.GetAll()
		if err != nil {
			logger.Error("Failed to get all menus", zap.Error(err))
			return err
		}
		validMenuIDs := make(map[int64]bool)
		for _, m := range allMenus {
			validMenuIDs[m.ID] = true
		}
		for _, menuID := range req.MenuIDs {
			if !validMenuIDs[menuID] {
				logger.Warn("Invalid menu ID in request", zap.Int64("menuId", menuID))
				return ErrMenuNotFound
			}
		}
	}

	if err := s.roleMenuRepo.SaveRoleMenus(req.RoleID, req.MenuIDs); err != nil {
		logger.Error("Failed to save role menu auth",
			zap.Int64("roleId", req.RoleID),
			zap.Int("menuCount", len(req.MenuIDs)),
			zap.Error(err))
		return err
	}

	logger.Info("Role menu auth saved",
		zap.Int64("roleId", req.RoleID),
		zap.Int("menuCount", len(req.MenuIDs)))
	return nil
}

func (s *RoleMenuService) GetAuthorizedMenuIDs(roleIDs []int64) ([]int64, error) {
	if len(roleIDs) == 0 {
		return []int64{}, nil
	}
	return s.roleMenuRepo.GetMenuIDsByRoleIDs(roleIDs)
}

func (s *RoleMenuService) IsMenuAuthorized(roleIDs []int64, menuID int64) (bool, error) {
	if len(roleIDs) == 0 {
		return false, nil
	}
	return s.roleMenuRepo.IsMenuAuthorizedForRoles(roleIDs, menuID)
}

func (s *RoleMenuService) DeleteRoleMenuAuth(roleID int64) error {
	if roleID <= 0 {
		return ErrInvalidRoleID
	}

	if err := s.roleMenuRepo.DeleteByRoleID(roleID); err != nil {
		logger.Error("Failed to delete role menu auth", zap.Int64("roleId", roleID), zap.Error(err))
		return err
	}

	logger.Info("Role menu auth deleted", zap.Int64("roleId", roleID))
	return nil
}
