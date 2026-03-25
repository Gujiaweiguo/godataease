package service

import (
	"fmt"
	"testing"

	"dataease/backend/internal/domain/permission"
)

type mockDatasourceGovernanceRunner struct {
	report *DatasourceGovernanceBackfillReport
	err    error
	last   *GovernanceBackfillOptions
}

func (m *mockDatasourceGovernanceRunner) BackfillGovernedResourcesWithOptions(options *GovernanceBackfillOptions) (*DatasourceGovernanceBackfillReport, error) {
	m.last = options
	return m.report, m.err
}

type mockDatasetGovernanceRunner struct {
	report *DatasetGovernanceBackfillReport
	err    error
	last   *GovernanceBackfillOptions
}

func (m *mockDatasetGovernanceRunner) BackfillGovernedResourcesWithOptions(options *GovernanceBackfillOptions) (*DatasetGovernanceBackfillReport, error) {
	m.last = options
	return m.report, m.err
}

type mockVisualizationGovernanceRunner struct {
	report *VisualizationGovernanceBackfillReport
	err    error
	last   *GovernanceBackfillOptions
}

func (m *mockVisualizationGovernanceRunner) BackfillGovernedVisualizationResourcesWithOptions(options *GovernanceBackfillOptions) (*VisualizationGovernanceBackfillReport, error) {
	m.last = options
	return m.report, m.err
}

func TestResourceGovernanceAdminService_BackfillResources_DispatchesDatasource(t *testing.T) {
	datasourceRunner := &mockDatasourceGovernanceRunner{report: &DatasourceGovernanceBackfillReport{ResourceType: permission.ResourceTypeDatasource, Governed: 1}}
	svc := NewResourceGovernanceAdminService(datasourceRunner, nil, nil)

	report, err := svc.BackfillResources(&ResourceGovernanceBackfillRequest{ResourceType: permission.ResourceTypeDatasource, AfterID: 10, Limit: 5})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report.ResourceType != permission.ResourceTypeDatasource || report.Governed != 1 {
		t.Fatalf("unexpected datasource report: %+v", report)
	}
	if datasourceRunner.last == nil || datasourceRunner.last.AfterID != 10 || datasourceRunner.last.Limit != 5 {
		t.Fatalf("expected datasource options to be forwarded, got %+v", datasourceRunner.last)
	}
}

func TestResourceGovernanceAdminService_BackfillResources_DispatchesVisualization(t *testing.T) {
	visualizationRunner := &mockVisualizationGovernanceRunner{report: &VisualizationGovernanceBackfillReport{ResourceType: "visualization", Governed: 2}}
	svc := NewResourceGovernanceAdminService(nil, nil, visualizationRunner)

	report, err := svc.BackfillResources(&ResourceGovernanceBackfillRequest{ResourceType: "visualization", Limit: 3})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report.ResourceType != "visualization" || report.Governed != 2 {
		t.Fatalf("unexpected visualization report: %+v", report)
	}
	if visualizationRunner.last == nil || visualizationRunner.last.Limit != 3 {
		t.Fatalf("expected visualization options to be forwarded, got %+v", visualizationRunner.last)
	}
}

func TestResourceGovernanceAdminService_BackfillResources_ForwardsVisualizationOrgScope(t *testing.T) {
	orgID := int64(9)
	visualizationRunner := &mockVisualizationGovernanceRunner{report: &VisualizationGovernanceBackfillReport{ResourceType: "visualization", Governed: 1}}
	svc := NewResourceGovernanceAdminService(nil, nil, visualizationRunner)

	_, err := svc.BackfillResources(&ResourceGovernanceBackfillRequest{ResourceType: "visualization", AfterID: 2, Limit: 4, OrgID: &orgID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if visualizationRunner.last == nil || visualizationRunner.last.OrgID == nil || *visualizationRunner.last.OrgID != orgID {
		t.Fatalf("expected visualization org scope to be forwarded, got %+v", visualizationRunner.last)
	}
}

func TestResourceGovernanceAdminService_BackfillResources_RejectsDatasourceOrgScope(t *testing.T) {
	orgID := int64(7)
	svc := NewResourceGovernanceAdminService(&mockDatasourceGovernanceRunner{}, nil, nil)

	_, err := svc.BackfillResources(&ResourceGovernanceBackfillRequest{ResourceType: permission.ResourceTypeDatasource, OrgID: &orgID})
	if err == nil {
		t.Fatalf("expected org-scoped datasource request to be rejected")
	}
	if err.Error() != "org-scoped backfill is unsupported for datasource: current resource model does not expose a safe org boundary" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResourceGovernanceAdminService_BackfillResources_RejectsDatasetOrgScope(t *testing.T) {
	orgID := int64(8)
	svc := NewResourceGovernanceAdminService(nil, &mockDatasetGovernanceRunner{}, nil)

	_, err := svc.BackfillResources(&ResourceGovernanceBackfillRequest{ResourceType: permission.ResourceTypeDataset, OrgID: &orgID})
	if err == nil {
		t.Fatalf("expected org-scoped dataset request to be rejected")
	}
	if err.Error() != "org-scoped backfill is unsupported for dataset: current resource model does not expose a safe org boundary" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResourceGovernanceAdminService_BackfillResources_RejectsUnsupportedType(t *testing.T) {
	svc := NewResourceGovernanceAdminService(nil, nil, nil)

	_, err := svc.BackfillResources(&ResourceGovernanceBackfillRequest{ResourceType: permission.ResourceTypeDashboard})
	if err == nil {
		t.Fatalf("expected unsupported resource type error")
	}
	if err.Error() != fmt.Sprintf("unsupported resource type: %s", permission.ResourceTypeDashboard) {
		t.Fatalf("unexpected error: %v", err)
	}
}
