package service

import (
	"fmt"
	"strings"

	"dataease/backend/internal/domain/permission"
)

type datasourceGovernanceBackfillRunner interface {
	BackfillGovernedResourcesWithOptions(options *GovernanceBackfillOptions) (*DatasourceGovernanceBackfillReport, error)
}

type datasetGovernanceBackfillRunner interface {
	BackfillGovernedResourcesWithOptions(options *GovernanceBackfillOptions) (*DatasetGovernanceBackfillReport, error)
}

type visualizationGovernanceBackfillRunner interface {
	BackfillGovernedVisualizationResourcesWithOptions(options *GovernanceBackfillOptions) (*VisualizationGovernanceBackfillReport, error)
}

type ResourceGovernanceBackfillRequest struct {
	ResourceType string `json:"resourceType"`
	AfterID      int64  `json:"afterId"`
	Limit        int    `json:"limit"`
	OrgID        *int64 `json:"orgId,omitempty"`
}

type ResourceGovernanceAdminService struct {
	datasourceSvc    datasourceGovernanceBackfillRunner
	datasetSvc       datasetGovernanceBackfillRunner
	visualizationSvc visualizationGovernanceBackfillRunner
}

func NewResourceGovernanceAdminService(
	datasourceSvc datasourceGovernanceBackfillRunner,
	datasetSvc datasetGovernanceBackfillRunner,
	visualizationSvc visualizationGovernanceBackfillRunner,
) *ResourceGovernanceAdminService {
	return &ResourceGovernanceAdminService{
		datasourceSvc:    datasourceSvc,
		datasetSvc:       datasetSvc,
		visualizationSvc: visualizationSvc,
	}
}

func (s *ResourceGovernanceAdminService) BackfillResources(req *ResourceGovernanceBackfillRequest) (*GovernanceBackfillReport, error) {
	if req == nil {
		return nil, fmt.Errorf("backfill request is required")
	}
	resourceType := strings.TrimSpace(req.ResourceType)
	options := &GovernanceBackfillOptions{AfterID: req.AfterID, Limit: req.Limit, OrgID: req.OrgID}

	switch resourceType {
	case permission.ResourceTypeDatasource:
		if req.OrgID != nil && *req.OrgID > 0 {
			return nil, fmt.Errorf("org-scoped backfill is unsupported for datasource: current resource model does not expose a safe org boundary")
		}
		if s.datasourceSvc == nil {
			return nil, fmt.Errorf("datasource governance service not initialized")
		}
		return s.datasourceSvc.BackfillGovernedResourcesWithOptions(options)
	case permission.ResourceTypeDataset:
		if req.OrgID != nil && *req.OrgID > 0 {
			return nil, fmt.Errorf("org-scoped backfill is unsupported for dataset: current resource model does not expose a safe org boundary")
		}
		if s.datasetSvc == nil {
			return nil, fmt.Errorf("dataset governance service not initialized")
		}
		return s.datasetSvc.BackfillGovernedResourcesWithOptions(options)
	case "visualization":
		if s.visualizationSvc == nil {
			return nil, fmt.Errorf("visualization governance service not initialized")
		}
		return s.visualizationSvc.BackfillGovernedVisualizationResourcesWithOptions(options)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}
