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
		RelationList: buildForwardTree(rows),
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
		RelationList: buildForwardTree(rows),
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
		RelationList: buildReverseTree(rows),
	}, nil
}

func (s *RelationService) CheckPermission(_ context.Context, id int64) (*relation.CheckPermissionResponse, error) {
	return &relation.CheckPermissionResponse{ID: id, Editable: true, Creatable: true}, nil
}

func buildForwardTree(rows []repository.RelationQueryRow) []*relation.RelationDTO {
	root := make([]*relation.RelationDTO, 0)
	datasets := make(map[int64]*relation.RelationDTO)
	chartNodes := make(map[int64]map[int64]*relation.RelationDTO)
	dashboardNodes := make(map[int64]map[int64]*relation.RelationDTO)

	for _, row := range rows {
		if row.DatasetID == nil {
			continue
		}

		datasetNode, ok := datasets[*row.DatasetID]
		if !ok {
			datasetNode = newRelationDTO(*row.DatasetID, row.DatasetName, row.DatasetCreator, "dataset", row.DatasetUpdate)
			datasets[*row.DatasetID] = datasetNode
			root = append(root, datasetNode)
		}

		if row.ChartID == nil {
			continue
		}

		if chartNodes[*row.DatasetID] == nil {
			chartNodes[*row.DatasetID] = make(map[int64]*relation.RelationDTO)
		}
		chartNode, ok := chartNodes[*row.DatasetID][*row.ChartID]
		if !ok {
			chartNode = newRelationDTO(*row.ChartID, row.ChartName, row.ChartCreator, "chart", row.ChartUpdate)
			chartNodes[*row.DatasetID][*row.ChartID] = chartNode
			datasetNode.SubRelation = append(datasetNode.SubRelation, chartNode)
		}

		if row.DashboardID == nil {
			continue
		}

		if dashboardNodes[*row.ChartID] == nil {
			dashboardNodes[*row.ChartID] = make(map[int64]*relation.RelationDTO)
		}
		if _, ok := dashboardNodes[*row.ChartID][*row.DashboardID]; ok {
			continue
		}

		dashboardNode := newRelationDTO(*row.DashboardID, row.DashboardName, row.DashboardCreator, "dashboard", row.DashboardUpdate)
		dashboardNodes[*row.ChartID][*row.DashboardID] = dashboardNode
		chartNode.SubRelation = append(chartNode.SubRelation, dashboardNode)
	}

	return root
}

func buildReverseTree(rows []repository.RelationQueryRow) []*relation.RelationDTO {
	root := make([]*relation.RelationDTO, 0)
	charts := make(map[int64]*relation.RelationDTO)
	datasets := make(map[int64]map[int64]*relation.RelationDTO)
	datasources := make(map[int64]map[int64]*relation.RelationDTO)

	for _, row := range rows {
		if row.ChartID == nil {
			continue
		}

		chartNode, ok := charts[*row.ChartID]
		if !ok {
			chartNode = newRelationDTO(*row.ChartID, row.ChartName, row.ChartCreator, "chart", row.ChartUpdate)
			charts[*row.ChartID] = chartNode
			root = append(root, chartNode)
		}

		if row.DatasetID == nil {
			continue
		}

		if datasets[*row.ChartID] == nil {
			datasets[*row.ChartID] = make(map[int64]*relation.RelationDTO)
		}
		datasetNode, ok := datasets[*row.ChartID][*row.DatasetID]
		if !ok {
			datasetNode = newRelationDTO(*row.DatasetID, row.DatasetName, row.DatasetCreator, "dataset", row.DatasetUpdate)
			datasets[*row.ChartID][*row.DatasetID] = datasetNode
			chartNode.SubRelation = append(chartNode.SubRelation, datasetNode)
		}

		if row.DatasourceID == nil {
			continue
		}

		if datasources[*row.DatasetID] == nil {
			datasources[*row.DatasetID] = make(map[int64]*relation.RelationDTO)
		}
		if _, ok := datasources[*row.DatasetID][*row.DatasourceID]; ok {
			continue
		}

		datasourceNode := newRelationDTO(*row.DatasourceID, row.DatasourceName, row.DatasourceCreator, "datasource", row.DatasourceUpdate)
		datasources[*row.DatasetID][*row.DatasourceID] = datasourceNode
		datasetNode.SubRelation = append(datasetNode.SubRelation, datasourceNode)
	}

	return root
}

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
