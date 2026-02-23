package service

import (
	"context"

	"dataease/backend/internal/domain/permission"
)

type ExportPermissionService struct {
	resourcePermSvc *ResourcePermissionService
	cache           *permission.PermissionCacheService
}

func NewExportPermissionService(resourcePermSvc *ResourcePermissionService, cache *permission.PermissionCacheService) *ExportPermissionService {
	return &ExportPermissionService{
		resourcePermSvc: resourcePermSvc,
		cache:           cache,
	}
}

func (s *ExportPermissionService) CheckDashboardExport(ctx context.Context, userID int64, dashboardID int64) *permission.PermissionCheckResult {
	result := s.resourcePermSvc.CheckPermission(userID, permission.ResourceTypeDashboard, dashboardID, permission.PermKeyExport)
	return result
}

func (s *ExportPermissionService) CheckDatasetExport(ctx context.Context, userID int64, datasetID int64) *permission.PermissionCheckResult {
	result := s.resourcePermSvc.CheckPermission(userID, permission.ResourceTypeDataset, datasetID, permission.PermKeyExport)
	return result
}

func (s *ExportPermissionService) CheckScreenExport(ctx context.Context, userID int64, screenID int64) *permission.PermissionCheckResult {
	result := s.resourcePermSvc.CheckPermission(userID, permission.ResourceTypeScreen, screenID, permission.PermKeyExport)
	return result
}

func (s *ExportPermissionService) CheckDatasourceExport(ctx context.Context, userID int64, datasourceID int64) *permission.PermissionCheckResult {
	result := s.resourcePermSvc.CheckPermission(userID, permission.ResourceTypeDatasource, datasourceID, permission.PermKeyExport)
	return result
}

func (s *ExportPermissionService) CanExportImage(ctx context.Context, userID int64, resourceType string, resourceID int64) bool {
	result := s.resourcePermSvc.CheckPermission(userID, resourceType, resourceID, permission.PermKeyExport)
	return result.HasPermission
}

func (s *ExportPermissionService) CanExportPDF(ctx context.Context, userID int64, resourceType string, resourceID int64) bool {
	result := s.resourcePermSvc.CheckPermission(userID, resourceType, resourceID, permission.PermKeyExport)
	return result.HasPermission
}

func (s *ExportPermissionService) CanExportExcel(ctx context.Context, userID int64, resourceType string, resourceID int64) bool {
	result := s.resourcePermSvc.CheckPermission(userID, resourceType, resourceID, permission.PermKeyExport)
	return result.HasPermission
}

func (s *ExportPermissionService) CanExportTemplate(ctx context.Context, userID int64, resourceType string, resourceID int64) bool {
	result := s.resourcePermSvc.CheckPermission(userID, resourceType, resourceID, permission.PermKeyManage)
	return result.HasPermission
}

type ExportType string

const (
	ExportTypeImage    ExportType = "image"
	ExportTypePDF      ExportType = "pdf"
	ExportTypeExcel    ExportType = "excel"
	ExportTypeTemplate ExportType = "template"
)

type ExportCheckRequest struct {
	UserID       int64      `json:"userId"`
	ResourceType string     `json:"resourceType"`
	ResourceID   int64      `json:"resourceId"`
	ExportType   ExportType `json:"exportType"`
}

type ExportCheckResult struct {
	CanExport    bool   `json:"canExport"`
	Reason       string `json:"reason,omitempty"`
	RequiredPerm string `json:"requiredPerm,omitempty"`
}

func (s *ExportPermissionService) CheckExportPermission(ctx context.Context, req *ExportCheckRequest) *ExportCheckResult {
	var requiredPerm string
	switch req.ExportType {
	case ExportTypeImage, ExportTypePDF, ExportTypeExcel:
		requiredPerm = permission.PermKeyExport
	case ExportTypeTemplate:
		requiredPerm = permission.PermKeyManage
	default:
		requiredPerm = permission.PermKeyExport
	}

	result := s.resourcePermSvc.CheckPermission(req.UserID, req.ResourceType, req.ResourceID, requiredPerm)

	return &ExportCheckResult{
		CanExport:    result.HasPermission,
		Reason:       result.Reason,
		RequiredPerm: requiredPerm,
	}
}

func (s *ExportPermissionService) CheckBatchExportPermission(ctx context.Context, userID int64, resourceType string, resourceIDs []int64) map[int64]bool {
	results := make(map[int64]bool)
	for _, id := range resourceIDs {
		result := s.resourcePermSvc.CheckPermission(userID, resourceType, id, permission.PermKeyExport)
		results[id] = result.HasPermission
	}
	return results
}

func (s *ExportPermissionService) FilterExportableResources(ctx context.Context, userID int64, resourceType string, resourceIDs []int64) []int64 {
	var exportable []int64
	for _, id := range resourceIDs {
		result := s.resourcePermSvc.CheckPermission(userID, resourceType, id, permission.PermKeyExport)
		if result.HasPermission {
			exportable = append(exportable, id)
		}
	}
	return exportable
}
