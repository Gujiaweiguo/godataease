package service

import (
	"context"
	"fmt"

	"dataease/backend/internal/domain/relation"
	"dataease/backend/internal/repository"
)

type RelationReader interface {
	GetDatasourceRelations(id int64) ([]repository.RelationQueryRow, error)
	GetDatasetRelations(id int64) ([]repository.RelationQueryRow, error)
	GetPanelRelations(id int64) ([]repository.RelationQueryRow, error)
}

type RelationService struct {
	repo RelationReader
}

func NewRelationService(repo RelationReader) *RelationService {
	return &RelationService{repo: repo}
}

func (s *RelationService) GetDatasourceRelationship(_ context.Context, id int64) (*relation.RelationResponse, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("relation service is unavailable")
	}
	rows, err := s.repo.GetDatasourceRelations(id)
	if err != nil {
		return nil, err
	}
	return &relation.RelationResponse{
		ID:           id,
		BusiFlag:     "datasource",
		RelationList: buildThreeLevelTree(rows, forwardLevels[0], forwardLevels[1], forwardLevels[2]),
	}, nil
}

func (s *RelationService) GetDatasetRelationship(_ context.Context, id int64) (*relation.RelationResponse, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("relation service is unavailable")
	}
	rows, err := s.repo.GetDatasetRelations(id)
	if err != nil {
		return nil, err
	}
	return &relation.RelationResponse{
		ID:           id,
		BusiFlag:     "dataset",
		RelationList: buildThreeLevelTree(rows, forwardLevels[0], forwardLevels[1], forwardLevels[2]),
	}, nil
}

func (s *RelationService) GetPanelRelationship(_ context.Context, id int64) (*relation.RelationResponse, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("relation service is unavailable")
	}
	rows, err := s.repo.GetPanelRelations(id)
	if err != nil {
		return nil, err
	}
	return &relation.RelationResponse{
		ID:           id,
		BusiFlag:     "dashboard",
		RelationList: buildThreeLevelTree(rows, reverseLevels[0], reverseLevels[1], reverseLevels[2]),
	}, nil
}

func (s *RelationService) CheckPermission(_ context.Context, id int64) (*relation.CheckPermissionResponse, error) {
	return &relation.CheckPermissionResponse{ID: id, Editable: true, Creatable: true}, nil
}

type relationLevel struct {
	idFunc      func(row repository.RelationQueryRow) *int64
	nameFunc    func(row repository.RelationQueryRow) *string
	creatorFunc func(row repository.RelationQueryRow) *string
	updateFunc  func(row repository.RelationQueryRow) *int64
	typeLabel   string
}

func buildThreeLevelTree(rows []repository.RelationQueryRow, l1, l2, l3 relationLevel) []*relation.RelationDTO {
	root := make([]*relation.RelationDTO, 0)
	level1Map := make(map[int64]*relation.RelationDTO)
	level2Map := make(map[int64]map[int64]*relation.RelationDTO)
	level3Map := make(map[int64]map[int64]*relation.RelationDTO)

	for _, row := range rows {
		if l1.idFunc(row) == nil {
			continue
		}

		level1ID := *l1.idFunc(row)
		level1Node, ok := level1Map[level1ID]
		if !ok {
			level1Node = newRelationDTO(level1ID, l1.nameFunc(row), l1.creatorFunc(row), l1.typeLabel, l1.updateFunc(row))
			level1Map[level1ID] = level1Node
			root = append(root, level1Node)
		}

		if l2.idFunc(row) == nil {
			continue
		}

		level2ID := *l2.idFunc(row)
		if level2Map[level1ID] == nil {
			level2Map[level1ID] = make(map[int64]*relation.RelationDTO)
		}
		level2Node, ok := level2Map[level1ID][level2ID]
		if !ok {
			level2Node = newRelationDTO(level2ID, l2.nameFunc(row), l2.creatorFunc(row), l2.typeLabel, l2.updateFunc(row))
			level2Map[level1ID][level2ID] = level2Node
			level1Node.SubRelation = append(level1Node.SubRelation, level2Node)
		}

		if l3.idFunc(row) == nil {
			continue
		}

		level3ID := *l3.idFunc(row)
		if level3Map[level2ID] == nil {
			level3Map[level2ID] = make(map[int64]*relation.RelationDTO)
		}
		if _, ok := level3Map[level2ID][level3ID]; ok {
			continue
		}

		level3Node := newRelationDTO(level3ID, l3.nameFunc(row), l3.creatorFunc(row), l3.typeLabel, l3.updateFunc(row))
		level3Map[level2ID][level3ID] = level3Node
		level2Node.SubRelation = append(level2Node.SubRelation, level3Node)
	}

	return root
}

var (
	forwardLevels = [3]relationLevel{
		{
			idFunc:      func(r repository.RelationQueryRow) *int64 { return r.DatasetID },
			nameFunc:    func(r repository.RelationQueryRow) *string { return r.DatasetName },
			creatorFunc: func(r repository.RelationQueryRow) *string { return r.DatasetCreator },
			updateFunc:  func(r repository.RelationQueryRow) *int64 { return r.DatasetUpdate },
			typeLabel:   "dataset",
		},
		{
			idFunc:      func(r repository.RelationQueryRow) *int64 { return r.ChartID },
			nameFunc:    func(r repository.RelationQueryRow) *string { return r.ChartName },
			creatorFunc: func(r repository.RelationQueryRow) *string { return r.ChartCreator },
			updateFunc:  func(r repository.RelationQueryRow) *int64 { return r.ChartUpdate },
			typeLabel:   "chart",
		},
		{
			idFunc:      func(r repository.RelationQueryRow) *int64 { return r.DashboardID },
			nameFunc:    func(r repository.RelationQueryRow) *string { return r.DashboardName },
			creatorFunc: func(r repository.RelationQueryRow) *string { return r.DashboardCreator },
			updateFunc:  func(r repository.RelationQueryRow) *int64 { return r.DashboardUpdate },
			typeLabel:   "dashboard",
		},
	}
	reverseLevels = [3]relationLevel{
		{
			idFunc:      func(r repository.RelationQueryRow) *int64 { return r.ChartID },
			nameFunc:    func(r repository.RelationQueryRow) *string { return r.ChartName },
			creatorFunc: func(r repository.RelationQueryRow) *string { return r.ChartCreator },
			updateFunc:  func(r repository.RelationQueryRow) *int64 { return r.ChartUpdate },
			typeLabel:   "chart",
		},
		{
			idFunc:      func(r repository.RelationQueryRow) *int64 { return r.DatasetID },
			nameFunc:    func(r repository.RelationQueryRow) *string { return r.DatasetName },
			creatorFunc: func(r repository.RelationQueryRow) *string { return r.DatasetCreator },
			updateFunc:  func(r repository.RelationQueryRow) *int64 { return r.DatasetUpdate },
			typeLabel:   "dataset",
		},
		{
			idFunc:      func(r repository.RelationQueryRow) *int64 { return r.DatasourceID },
			nameFunc:    func(r repository.RelationQueryRow) *string { return r.DatasourceName },
			creatorFunc: func(r repository.RelationQueryRow) *string { return r.DatasourceCreator },
			updateFunc:  func(r repository.RelationQueryRow) *int64 { return r.DatasourceUpdate },
			typeLabel:   "datasource",
		},
	}
)

func newRelationDTO(id int64, name, creator *string, relationType string, update *int64) *relation.RelationDTO {
	return &relation.RelationDTO{
		ID:          id,
		Name:        relationValueOrEmpty(name),
		Auths:       "",
		Type:        relationType,
		Creator:     relationValueOrEmpty(creator),
		UpdateTime:  normalizeRelationTime(update),
		SubRelation: make([]*relation.RelationDTO, 0),
	}
}

func relationValueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func normalizeRelationTime(v *int64) int64 {
	if v == nil || *v <= 0 {
		return 0
	}
	if *v < 1_000_000_000_000 {
		return *v * 1000
	}
	return *v
}
