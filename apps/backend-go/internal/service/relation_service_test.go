package service

import (
	"context"
	"errors"
	"testing"

	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRelationRepository struct {
	datasourceRows []repository.RelationQueryRow
	datasetRows    []repository.RelationQueryRow
	panelRows      []repository.RelationQueryRow
	err            error
}

func (m *mockRelationRepository) GetDatasourceRelations(id int64) ([]repository.RelationQueryRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.datasourceRows, nil
}

func (m *mockRelationRepository) GetDatasetRelations(id int64) ([]repository.RelationQueryRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.datasetRows, nil
}

func (m *mockRelationRepository) GetPanelRelations(id int64) ([]repository.RelationQueryRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.panelRows, nil
}

func TestRelationService_GetDatasourceRelationshipBuildsTree(t *testing.T) {
	service := NewRelationService(&mockRelationRepository{datasourceRows: []repository.RelationQueryRow{
		{
			DatasetID:        int64PtrRelation(10),
			DatasetName:      stringPtrRelation("dataset-a"),
			DatasetCreator:   stringPtrRelation("alice"),
			DatasetUpdate:    int64PtrRelation(1710000000000),
			ChartID:          int64PtrRelation(20),
			ChartName:        stringPtrRelation("chart-a"),
			ChartCreator:     stringPtrRelation("bob"),
			ChartUpdate:      int64PtrRelation(1710000001000),
			DashboardID:      int64PtrRelation(30),
			DashboardName:    stringPtrRelation("dashboard-a"),
			DashboardCreator: stringPtrRelation("carol"),
			DashboardUpdate:  int64PtrRelation(1710000002000),
		},
		{
			DatasetID:        int64PtrRelation(10),
			DatasetName:      stringPtrRelation("dataset-a"),
			DatasetCreator:   stringPtrRelation("alice"),
			DatasetUpdate:    int64PtrRelation(1710000000000),
			ChartID:          int64PtrRelation(20),
			ChartName:        stringPtrRelation("chart-a"),
			ChartCreator:     stringPtrRelation("bob"),
			ChartUpdate:      int64PtrRelation(1710000001000),
			DashboardID:      int64PtrRelation(30),
			DashboardName:    stringPtrRelation("dashboard-a"),
			DashboardCreator: stringPtrRelation("carol"),
			DashboardUpdate:  int64PtrRelation(1710000002000),
		},
		{
			DatasetID:      int64PtrRelation(10),
			DatasetName:    stringPtrRelation("dataset-a"),
			DatasetCreator: stringPtrRelation("alice"),
			DatasetUpdate:  int64PtrRelation(1710000000000),
			ChartID:        int64PtrRelation(21),
			ChartName:      stringPtrRelation("chart-b"),
			ChartCreator:   stringPtrRelation("dora"),
			ChartUpdate:    int64PtrRelation(1710000003000),
		},
		{
			DatasetID:      int64PtrRelation(11),
			DatasetName:    stringPtrRelation("dataset-b"),
			DatasetCreator: stringPtrRelation("eric"),
			DatasetUpdate:  int64PtrRelation(1710000004000),
		},
	}})

	result, err := service.GetDatasourceRelationship(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "datasource", result.BusiFlag)
	require.Len(t, result.RelationList, 2)

	firstDataset := result.RelationList[0]
	assert.Equal(t, int64(10), firstDataset.ID)
	require.Len(t, firstDataset.SubRelation, 2)
	assert.Equal(t, int64(20), firstDataset.SubRelation[0].ID)
	require.Len(t, firstDataset.SubRelation[0].SubRelation, 1)
	assert.Equal(t, int64(30), firstDataset.SubRelation[0].SubRelation[0].ID)

	secondDataset := result.RelationList[1]
	assert.Equal(t, int64(11), secondDataset.ID)
	assert.Empty(t, secondDataset.SubRelation)
}

func TestRelationService_GetPanelRelationshipBuildsReverseTree(t *testing.T) {
	service := NewRelationService(&mockRelationRepository{panelRows: []repository.RelationQueryRow{
		{
			ChartID:           int64PtrRelation(20),
			ChartName:         stringPtrRelation("chart-a"),
			ChartCreator:      stringPtrRelation("alice"),
			ChartUpdate:       int64PtrRelation(1710000000),
			DatasetID:         int64PtrRelation(10),
			DatasetName:       stringPtrRelation("dataset-a"),
			DatasetCreator:    stringPtrRelation("bob"),
			DatasetUpdate:     int64PtrRelation(1710000001000),
			DatasourceID:      int64PtrRelation(1),
			DatasourceName:    stringPtrRelation("datasource-a"),
			DatasourceCreator: stringPtrRelation("carol"),
			DatasourceUpdate:  int64PtrRelation(1710000002000),
		},
		{
			ChartID:           int64PtrRelation(20),
			ChartName:         stringPtrRelation("chart-a"),
			ChartCreator:      stringPtrRelation("alice"),
			ChartUpdate:       int64PtrRelation(1710000000),
			DatasetID:         int64PtrRelation(10),
			DatasetName:       stringPtrRelation("dataset-a"),
			DatasetCreator:    stringPtrRelation("bob"),
			DatasetUpdate:     int64PtrRelation(1710000001000),
			DatasourceID:      int64PtrRelation(1),
			DatasourceName:    stringPtrRelation("datasource-a"),
			DatasourceCreator: stringPtrRelation("carol"),
			DatasourceUpdate:  int64PtrRelation(1710000002000),
		},
	}})

	result, err := service.GetPanelRelationship(context.Background(), 30)
	require.NoError(t, err)
	assert.Equal(t, "dashboard", result.BusiFlag)
	require.Len(t, result.RelationList, 1)
	require.Len(t, result.RelationList[0].SubRelation, 1)
	require.Len(t, result.RelationList[0].SubRelation[0].SubRelation, 1)
	assert.Equal(t, int64(1710000000000), result.RelationList[0].UpdateTime)
}

func TestRelationService_GetDatasetRelationshipEmpty(t *testing.T) {
	service := NewRelationService(&mockRelationRepository{})

	result, err := service.GetDatasetRelationship(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.ID)
	assert.Equal(t, "dataset", result.BusiFlag)
	assert.Empty(t, result.RelationList)
}

func TestRelationService_CheckPermission(t *testing.T) {
	service := NewRelationService(&mockRelationRepository{})

	result, err := service.CheckPermission(context.Background(), 9)
	require.NoError(t, err)
	assert.Equal(t, int64(9), result.ID)
	assert.True(t, result.Editable)
	assert.True(t, result.Creatable)
}

func TestRelationService_PropagatesRepositoryErrors(t *testing.T) {
	service := NewRelationService(&mockRelationRepository{err: errors.New("boom")})

	result, err := service.GetDatasourceRelationship(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, result)
}

func int64PtrRelation(v int64) *int64 { return &v }

func stringPtrRelation(v string) *string { return &v }
