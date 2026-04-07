package middleware

import (
	"context"
	"fmt"
	"strconv"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

const (
	UserIDKey        = "user_id"
	ResourceTypeKey  = "resource_type"
	ResourceIDKey    = "resource_id"
	ResourceIDsKey   = "resource_ids"
	PermissionKeyKey = "permission_key"
	DatasetIDKey     = "dataset_id"
)

type PermissionMiddleware struct {
	resourcePermSvc      *service.ResourcePermissionService
	exportPermSvc        *service.ExportPermissionService
	adminChecker         AdminChecker
	chartDatasetResolver ChartDatasetGroupResolver
}

type AdminChecker interface {
	IsAdmin(userID int64) bool
}

type ChartDatasetGroupResolver interface {
	GetDatasetGroupIDByChartID(chartID int64) (int64, error)
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

func (m *PermissionMiddleware) SetChartDatasetResolver(resolver ChartDatasetGroupResolver) {
	m.chartDatasetResolver = resolver
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

		resourceID, err := extractResourceID(c)
		if err != nil {
			response.BadRequest(c, err.Error())
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

func (m *PermissionMiddleware) CheckChartDataView() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		chartID, err := extractChartID(c)
		if err != nil {
			response.BadRequest(c, err.Error())
			c.Abort()
			return
		}

		if m.chartDatasetResolver == nil {
			response.Error(c, "500000", "Failed: chart dataset resolver is unavailable")
			c.Abort()
			return
		}

		datasetID, err := m.chartDatasetResolver.GetDatasetGroupIDByChartID(chartID)
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			c.Abort()
			return
		}
		if datasetID <= 0 {
			response.Error(c, "500000", "Failed: resolved dataset group id is invalid")
			c.Abort()
			return
		}

		c.Set(DatasetIDKey, datasetID)
		c.Set(ResourceTypeKey, permission.ResourceTypeDataset)
		c.Set(ResourceIDKey, datasetID)
		c.Set(PermissionKeyKey, permission.PermKeyView)
		c.Set(RowPermissionDatasetIDKey, datasetID)
		c.Set(RowPermissionDatasetIDsKey, []int64{datasetID})

		if m.adminChecker != nil && m.adminChecker.IsAdmin(int64(userID)) {
			c.Next()
			return
		}
		if m.resourcePermSvc == nil {
			response.Error(c, "500000", "Failed: resource permission service is unavailable")
			c.Abort()
			return
		}

		result := m.resourcePermSvc.CheckPermission(int64(userID), permission.ResourceTypeDataset, datasetID, permission.PermKeyView)
		if !result.HasPermission {
			logger.Warn("Chart permission denied",
				zap.Uint64("user_id", userID),
				zap.Int64("chart_id", chartID),
				zap.Int64("dataset_id", datasetID),
				zap.String("perm_key", permission.PermKeyView),
				zap.String("reason", result.Reason),
			)
			response.Forbidden(c, "insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

func extractChartID(c *gin.Context) (int64, error) {
	for _, rawID := range []string{c.Param("id"), c.Query("id")} {
		if rawID == "" {
			continue
		}
		chartID, err := parseResourceID(rawID)
		if err != nil {
			return 0, err
		}
		return chartID, nil
	}

	var payload struct {
		ID int64 `json:"id"`
	}
	if err := c.ShouldBindBodyWith(&payload, binding.JSON); err == nil && payload.ID > 0 {
		return payload.ID, nil
	}

	return 0, fmt.Errorf("chart id is required")
}

func (m *PermissionMiddleware) CheckDatasetBatchView() gin.HandlerFunc {
	return m.CheckBatchResourcePermission(permission.ResourceTypeDataset, permission.PermKeyView)
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

func (m *PermissionMiddleware) CheckBatchResourcePermission(resourceType, permKey string) gin.HandlerFunc {
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

		resourceIDs, err := extractResourceIDs(c)
		if err != nil {
			response.BadRequest(c, err.Error())
			c.Abort()
			return
		}

		for _, resourceID := range resourceIDs {
			result := m.resourcePermSvc.CheckPermission(int64(userID), resourceType, resourceID, permKey)
			if !result.HasPermission {
				logger.Warn("Batch permission denied",
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
		}

		c.Set(ResourceTypeKey, resourceType)
		c.Set(ResourceIDKey, resourceIDs[0])
		c.Set(ResourceIDsKey, resourceIDs)
		c.Set(PermissionKeyKey, permKey)
		c.Next()
	}
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

		resourceID, err := extractResourceID(c)
		if err != nil {
			response.BadRequest(c, err.Error())
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

func extractResourceID(c *gin.Context) (int64, error) {
	resourceIDs, err := extractResourceIDs(c)
	if err != nil {
		return 0, err
	}
	return resourceIDs[0], nil
}

func extractResourceIDs(c *gin.Context) ([]int64, error) {
	if resourceIDs, ok, err := extractResourceIDsFromRequestLine(c); ok || err != nil {
		return resourceIDs, err
	}

	resourceIDs, err := extractResourceIDsFromBody(c)
	if err != nil {
		return nil, err
	}
	resourceIDs = uniqueInt64(resourceIDs)
	if len(resourceIDs) == 0 {
		return nil, fmt.Errorf("resource id is required")
	}

	return resourceIDs, nil
}

func extractResourceIDsFromRequestLine(c *gin.Context) ([]int64, bool, error) {
	for _, rawID := range []string{c.Param("id"), c.Query("resourceId"), c.Query("id")} {
		if rawID == "" {
			continue
		}
		resourceID, err := parseResourceID(rawID)
		if err != nil {
			return nil, true, err
		}
		return []int64{resourceID}, true, nil
	}
	return nil, false, nil
}

func extractResourceIDsFromBody(c *gin.Context) ([]int64, error) {
	resourceIDs := make([]int64, 0, 4)

	var payload map[string]interface{}
	if err := c.ShouldBindBodyWith(&payload, binding.JSON); err == nil {
		ids, parseErr := collectResourceIDsFromObject(payload)
		if parseErr != nil {
			return nil, parseErr
		}
		resourceIDs = append(resourceIDs, ids...)
	}

	var payloadList []interface{}
	if err := c.ShouldBindBodyWith(&payloadList, binding.JSON); err == nil {
		ids, parseErr := collectResourceIDsFromList(payloadList)
		if parseErr != nil {
			return nil, parseErr
		}
		resourceIDs = append(resourceIDs, ids...)
	}

	return resourceIDs, nil
}

func collectResourceIDsFromObject(payload map[string]interface{}) ([]int64, error) {
	resourceIDs := make([]int64, 0, 4)
	for _, key := range []string{"id", "resourceId", "datasetGroupId", "datasetId"} {
		if value, ok := payload[key]; ok {
			resourceID, err := parseResourceIDFromAny(value)
			if err != nil {
				return nil, err
			}
			resourceIDs = append(resourceIDs, resourceID)
		}
	}

	values, ok := payload["ids"].([]interface{})
	if !ok {
		return resourceIDs, nil
	}

	ids, err := collectResourceIDsFromList(values)
	if err != nil {
		return nil, err
	}
	return append(resourceIDs, ids...), nil
}

func collectResourceIDsFromList(values []interface{}) ([]int64, error) {
	resourceIDs := make([]int64, 0, len(values))
	for _, value := range values {
		resourceID, err := parseResourceIDFromAny(value)
		if err != nil {
			return nil, err
		}
		resourceIDs = append(resourceIDs, resourceID)
	}
	return resourceIDs, nil
}

func uniqueInt64(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func parseResourceIDFromAny(v interface{}) (int64, error) {
	switch value := v.(type) {
	case float64:
		return int64(value), nil
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	case string:
		return parseResourceID(value)
	default:
		return 0, fmt.Errorf("invalid resource id")
	}
}

func parseResourceID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid resource id")
	}
	return id, nil
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
