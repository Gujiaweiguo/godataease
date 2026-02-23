package middleware

import (
	"context"
	"strconv"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	UserIDKey        = "user_id"
	ResourceTypeKey  = "resource_type"
	ResourceIDKey    = "resource_id"
	PermissionKeyKey = "permission_key"
	DatasetIDKey     = "dataset_id"
)

type PermissionMiddleware struct {
	resourcePermSvc *service.ResourcePermissionService
	exportPermSvc   *service.ExportPermissionService
	adminChecker    AdminChecker
}

type AdminChecker interface {
	IsAdmin(userID int64) bool
}

func NewPermissionMiddleware(
	resourcePermSvc *service.ResourcePermissionService,
	exportPermSvc *service.ExportPermissionService,
	adminChecker AdminChecker,
) *PermissionMiddleware {
	return &PermissionMiddleware{
		resourcePermSvc: resourcePermSvc,
		exportPermSvc:   exportPermSvc,
		adminChecker:    adminChecker,
	}
}

func (m *PermissionMiddleware) CheckResourcePermission(resourceType, permKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		if m.adminChecker != nil && m.adminChecker.IsAdmin(int64(userID)) {
			c.Next()
			return
		}

		resourceIDStr := c.Param("id")
		if resourceIDStr == "" {
			resourceIDStr = c.Query("id")
		}
		if resourceIDStr == "" {
			response.BadRequest(c, "resource id is required")
			c.Abort()
			return
		}

		resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid resource id")
			c.Abort()
			return
		}

		result := m.resourcePermSvc.CheckPermission(int64(userID), resourceType, resourceID, permKey)
		if !result.HasPermission {
			logger.Warn("Permission denied",
				zap.Uint64("user_id", userID),
				zap.String("resource_type", resourceType),
				zap.Int64("resource_id", resourceID),
				zap.String("perm_key", permKey),
				zap.String("reason", result.Reason),
			)
			response.Forbidden(c, "insufficient permissions")
			c.Abort()
			return
		}

		c.Set(ResourceTypeKey, resourceType)
		c.Set(ResourceIDKey, resourceID)
		c.Set(PermissionKeyKey, permKey)
		c.Next()
	}
}

func (m *PermissionMiddleware) CheckDatasetView() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeDataset, permission.PermKeyView)
}

func (m *PermissionMiddleware) CheckDatasetEdit() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeDataset, permission.PermKeyEdit)
}

func (m *PermissionMiddleware) CheckDatasetExport() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeDataset, permission.PermKeyExport)
}

func (m *PermissionMiddleware) CheckDashboardView() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeDashboard, permission.PermKeyView)
}

func (m *PermissionMiddleware) CheckDashboardEdit() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeDashboard, permission.PermKeyEdit)
}

func (m *PermissionMiddleware) CheckDashboardExport() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeDashboard, permission.PermKeyExport)
}

func (m *PermissionMiddleware) CheckScreenView() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeScreen, permission.PermKeyView)
}

func (m *PermissionMiddleware) CheckScreenEdit() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeScreen, permission.PermKeyEdit)
}

func (m *PermissionMiddleware) CheckScreenExport() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeScreen, permission.PermKeyExport)
}

func (m *PermissionMiddleware) CheckDatasourceView() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeDatasource, permission.PermKeyView)
}

func (m *PermissionMiddleware) CheckDatasourceEdit() gin.HandlerFunc {
	return m.CheckResourcePermission(permission.ResourceTypeDatasource, permission.PermKeyEdit)
}

func (m *PermissionMiddleware) CheckExportPermission(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		if m.adminChecker != nil && m.adminChecker.IsAdmin(int64(userID)) {
			c.Next()
			return
		}

		resourceIDStr := c.Param("id")
		if resourceIDStr == "" {
			resourceIDStr = c.Query("resourceId")
		}
		if resourceIDStr == "" {
			resourceIDStr = c.Query("id")
		}
		if resourceIDStr == "" {
			response.BadRequest(c, "resource id is required for export permission check")
			c.Abort()
			return
		}

		resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid resource id")
			c.Abort()
			return
		}

		exportType := service.ExportType(c.Query("exportType"))
		if exportType == "" {
			exportType = service.ExportTypeImage
		}

		req := &service.ExportCheckRequest{
			UserID:       int64(userID),
			ResourceType: resourceType,
			ResourceID:   resourceID,
			ExportType:   exportType,
		}

		result := m.exportPermSvc.CheckExportPermission(context.Background(), req)
		if !result.CanExport {
			logger.Warn("Export permission denied",
				zap.Uint64("user_id", userID),
				zap.String("resource_type", resourceType),
				zap.Int64("resource_id", resourceID),
				zap.String("export_type", string(exportType)),
				zap.String("required_perm", result.RequiredPerm),
				zap.String("reason", result.Reason),
			)
			response.Forbidden(c, "no export permission")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *PermissionMiddleware) CheckBatchExportPermission(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		if m.adminChecker != nil && m.adminChecker.IsAdmin(int64(userID)) {
			c.Next()
			return
		}

		var req struct {
			ResourceIDs []int64 `json:"resourceIds"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request")
			c.Abort()
			return
		}

		if len(req.ResourceIDs) == 0 {
			c.Next()
			return
		}

		exportable := m.exportPermSvc.FilterExportableResources(context.Background(), int64(userID), resourceType, req.ResourceIDs)
		c.Set("exportable_resource_ids", exportable)
		c.Next()
	}
}

func (m *PermissionMiddleware) CheckDatasetDataPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		datasetIDStr := c.Param("datasetId")
		if datasetIDStr == "" {
			datasetIDStr = c.Query("datasetId")
		}
		if datasetIDStr == "" {
			var req struct {
				DatasetGroupID int64 `json:"datasetGroupId"`
			}
			if err := c.ShouldBindJSON(&req); err == nil && req.DatasetGroupID > 0 {
				datasetIDStr = strconv.FormatInt(req.DatasetGroupID, 10)
			}
		}

		if datasetIDStr != "" {
			datasetID, err := strconv.ParseInt(datasetIDStr, 10, 64)
			if err == nil {
				c.Set(DatasetIDKey, datasetID)
			}
		}

		c.Next()
	}
}

func GetDatasetID(c *gin.Context) int64 {
	if datasetID, exists := c.Get(DatasetIDKey); exists {
		return datasetID.(int64)
	}
	return 0
}

func GetResourceID(c *gin.Context) int64 {
	if resourceID, exists := c.Get(ResourceIDKey); exists {
		return resourceID.(int64)
	}
	return 0
}

func GetExportableResourceIDs(c *gin.Context) []int64 {
	if ids, exists := c.Get("exportable_resource_ids"); exists {
		return ids.([]int64)
	}
	return nil
}

type DefaultAdminChecker struct {
	adminUserIDs map[int64]bool
}

func NewDefaultAdminChecker(adminUserIDs []int64) *DefaultAdminChecker {
	ids := make(map[int64]bool)
	for _, id := range adminUserIDs {
		ids[id] = true
	}
	return &DefaultAdminChecker{adminUserIDs: ids}
}

func (c *DefaultAdminChecker) IsAdmin(userID int64) bool {
	return c.adminUserIDs[userID]
}
