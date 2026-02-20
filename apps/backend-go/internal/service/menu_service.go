package service

import (
	"strings"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/repository"
)

type MenuService struct {
	repo *repository.MenuRepository
}

func NewMenuService(repo *repository.MenuRepository) *MenuService {
	return &MenuService{repo: repo}
}

func (s *MenuService) Query() ([]*menu.MenuVO, error) {
	menus, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return s.buildMenuTree(menus), nil
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
