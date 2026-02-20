package service

import (
	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/repository"
)

type MapService struct {
	repo *repository.AreaRepository
}

func NewMapService(repo *repository.AreaRepository) *MapService {
	return &MapService{repo: repo}
}

func (s *MapService) GetWorldTree() (*areamap.AreaNode, error) {
	areas, err := s.repo.GetAllAreas()
	if err != nil {
		return nil, err
	}

	customAreas, err := s.repo.GetAllCustomAreas()
	if err != nil {
		return nil, err
	}

	world := &areamap.AreaNode{
		ID:     "000",
		Level:  "world",
		Name:   "世界村",
		Custom: false,
	}

	nodeMap := make(map[string]*areamap.AreaNode)
	nodeMap[world.ID] = world

	for _, area := range areas {
		node := &areamap.AreaNode{
			ID:     area.ID,
			Level:  area.Level,
			Name:   area.Name,
			Pid:    area.Pid,
			Custom: false,
		}
		nodeMap[area.ID] = node
	}

	for _, area := range customAreas {
		node := &areamap.AreaNode{
			ID:     area.ID,
			Level:  area.Level,
			Name:   area.Name,
			Pid:    area.Pid,
			Custom: true,
		}
		nodeMap[area.ID] = node
	}

	for _, area := range areas {
		node := nodeMap[area.ID]
		parent, exists := nodeMap[area.Pid]
		if exists && parent != nil {
			if parent.Children == nil {
				parent.Children = []*areamap.AreaNode{}
			}
			parent.Children = append(parent.Children, node)
		}
	}

	for _, area := range customAreas {
		node := nodeMap[area.ID]
		parent, exists := nodeMap[area.Pid]
		if exists && parent != nil {
			if parent.Children == nil {
				parent.Children = []*areamap.AreaNode{}
			}
			parent.Children = append(parent.Children, node)
		}
	}

	return world, nil
}
