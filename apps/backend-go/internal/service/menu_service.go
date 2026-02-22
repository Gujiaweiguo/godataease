package service

import (
	"strings"

	"dataease/backend/internal/domain/menu"
	"errors"
	"dataease/backend/internal/repository"
	)


type MenuService struct {
	repo         *repository.MenuRepository
	roleMenuRepo *repository.RoleMenuRepository
}

func NewMenuService(repo *repository.MenuRepository) *MenuService {
	return &MenuService{repo: repo}
}

func NewMenuServiceWithRoleFilter(repo *repository.MenuRepository, roleMenuRepo *repository.RoleMenuRepository) *MenuService {
	return &MenuService{
		repo:         repo,
		roleMenuRepo: roleMenuRepo,
	}
}

func (s *MenuService) Query() ([]*menu.MenuVO, error) {
	menus, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return s.buildMenuTree(menus), nil
}

func (s *MenuService) QueryByRoleIDs(roleIDs []int64) ([]*menu.MenuVO, error) {
	if s.roleMenuRepo == nil || len(roleIDs) == 0 {
		if s.isAdminRole(roleIDs) {
			return s.Query()
		}
		return []*menu.MenuVO{}, nil
	}

	menuIDs, err := s.roleMenuRepo.GetMenuIDsByRoleIDs(roleIDs)
	if err != nil {
		return nil, err
	}

	if len(menuIDs) == 0 {
		return []*menu.MenuVO{}, nil
	}

	authorizedMenus, err := s.repo.GetByIDs(menuIDs)
	if err != nil {
		return nil, err
	}

	return s.buildMenuTree(authorizedMenus), nil
}

func (s *MenuService) isAdminRole(roleIDs []int64) bool {
	for _, id := range roleIDs {
		if id == 1 {
			return true
		}
	}
	return false
}

func (s *MenuService) GetAuthorizedMenuIDs(roleIDs []int64) ([]int64, error) {
	if s.roleMenuRepo == nil {
		return []int64{}, nil
	}
	if s.isAdminRole(roleIDs) {
		menus, err := s.repo.GetAll()
		if err != nil {
			return nil, err
		}
		ids := make([]int64, len(menus))
		for i, m := range menus {
			ids[i] = m.ID
		}
		return ids, nil
	}
	return s.roleMenuRepo.GetMenuIDsByRoleIDs(roleIDs)
}

func (s *MenuService) buildMenuTree(menus []*menu.CoreMenu) []*menu.MenuVO {
	childMap := make(map[int64][]*menu.CoreMenu)
	for _, m := range menus {
		childMap[m.Pid] = append(childMap[m.Pid], m)
	}

	var roots []*menu.MenuVO
	for _, m := range menus {
		if m.Pid == 0 {
			vo := s.convertToVO(m, childMap)
			if len(vo.Children) > 0 || m.Type != 1 {
				roots = append(roots, vo)
			}
		}
	}
	return roots
}

func (s *MenuService) convertToVO(m *menu.CoreMenu, childMap map[int64][]*menu.CoreMenu) *menu.MenuVO {
	path := m.Path
	if m.Pid != 0 && strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	vo := &menu.MenuVO{
		ID:        m.ID,
		Path:      path,
		Component: m.Component,
		Hidden:    m.Hidden,
		IsPlugin:  false,
		Name:      m.Name,
		InLayout:  m.InLayout,
		Meta: &menu.MenuMeta{
			Title: m.Name,
			Icon:  m.Icon,
		},
	}

	children := childMap[m.ID]
	if len(children) > 0 {
		for _, child := range children {
			childVO := s.convertToVO(child, childMap)
			if len(childVO.Children) > 0 || child.Type != 1 {
				vo.Children = append(vo.Children, childVO)
			}
		}
	}

	return vo
}

func (s *MenuService) ShouldUseDynamicMenu() bool {
	return true
}

func (s *MenuService) GetByID(id int64) (*menu.CoreMenu, error) {
	return s.repo.GetByID(id)
}

func (s *MenuService) Create(m *menu.CoreMenu) error {
	return s.repo.Create(m)
}

func (s *MenuService) Update(m *menu.CoreMenu) error {
	return s.repo.Update(m)
}

func (s *MenuService) Delete(id int64) error {
	hasChildren, err := s.repo.HasChildren(id)
	if err != nil {
		return err
	}
	if hasChildren {
		return ErrMenuHasChildren
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	if s.roleMenuRepo != nil {
		return s.roleMenuRepo.DeleteByMenuID(id)
	}
	return nil
}

func (s *MenuService) UpdateSort(id int64, sort int) error {
	return s.repo.UpdateSort(id, sort)
}

func (s *MenuService) UpdateHidden(id int64, hidden bool) error {
	return s.repo.UpdateHidden(id, hidden)
}

var ErrMenuHasChildren = errors.New("menu has children, cannot delete")
