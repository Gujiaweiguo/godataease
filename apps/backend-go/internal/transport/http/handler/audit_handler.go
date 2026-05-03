package handler

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type AuditHandler struct {
	auditService       *service.AuditService
	systemParamService *service.SystemParamService
	auditAlertService  *service.AuditAlertService
}

const auditMenuPath = "/system/audit"

const auditExportRateLimitWindow = time.Minute

const auditExportRateLimitRequests = 10

func NewAuditHandler(auditService *service.AuditService, systemParamService *service.SystemParamService) *AuditHandler {
	return &AuditHandler{
		auditService:       auditService,
		systemParamService: systemParamService,
		auditAlertService:  service.NewAuditAlertService(systemParamService, auditService),
	}
}

func (h *AuditHandler) GetAuditAlertSettings(c *gin.Context) {
	defer recoverServicePanic(c)
	settings, err := h.systemParamService.QueryAuditAlertSettings()
	if err != nil {
		response.InternalError(c, "Failed to get audit alert settings")
		return
	}

	response.Success(c, settings)
}

func (h *AuditHandler) SaveAuditAlertSettings(c *gin.Context) {
	defer recoverServicePanic(c)
	var req audit.AuditAlertSettings
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.systemParamService.SaveAuditAlertSettings(&req); err != nil {
		response.InternalError(c, "Failed to save audit alert settings")
		return
	}

	response.Success(c, nil)
}

func (h *AuditHandler) CleanupNow(c *gin.Context) {
	defer recoverServicePanic(c)
	settings, err := h.systemParamService.QueryAuditAlertSettings()
	if err != nil {
		response.InternalError(c, "Failed to get audit alert settings")
		return
	}

	affected, err := h.auditService.DeleteAuditLogsBeforeDate(settings.RetentionDays)
	if err != nil {
		response.InternalError(c, "Failed to cleanup audit logs")
		return
	}

	response.Success(c, gin.H{
		"deleted":       affected,
		"retentionDays": settings.RetentionDays,
	})
}

func (h *AuditHandler) TestNotification(c *gin.Context) {
	defer recoverServicePanic(c)
	event := service.AlertEvent{
		Type:       service.AlertTypeBatchOperation,
		Username:   "system",
		Details:    "测试审计告警通知",
		DetectedAt: time.Now(),
	}
	if err := h.auditAlertService.Notify(event); err != nil {
		response.InternalError(c, "Failed to send test notification")
		return
	}

	response.Success(c, event)
}

func (h *AuditHandler) CreateAuditLog(c *gin.Context) {
	defer recoverServicePanic(c)
	var req audit.AuditLogCreateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	log, err := h.auditService.CreateAuditLog(&req)
	if err != nil {
		response.InternalError(c, "Failed to create audit log")
		return
	}

	response.Success(c, log)
}

func (h *AuditHandler) GetAuditLogByID(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsgBadRequest(c, "id", "Invalid audit log ID")
	if !ok {
		return
	}

	log, err := h.auditService.GetAuditLogByID(id)
	if err != nil {
		response.NotFound(c, "Audit log not found")
		return
	}

	response.Success(c, log)
}

func (h *AuditHandler) GetAuditLogsByUserID(c *gin.Context) {
	defer recoverServicePanic(c)
	userID, ok := parseIDParamMsgBadRequest(c, "userId", "Invalid user ID")
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := h.auditService.GetAuditLogsByUserID(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to get audit logs")
		return
	}

	response.Success(c, result)
}

func (h *AuditHandler) QueryAuditLogs(c *gin.Context) {
	defer recoverServicePanic(c)
	var query audit.AuditLogQuery

	if userIDStr := c.Query("userId"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			query.UserID = &userID
		}
	}

	if username := c.Query("username"); username != "" {
		query.Username = &username
	}

	if actionType := c.Query("actionType"); actionType != "" {
		at := audit.ActionType(actionType)
		query.ActionType = &at
	}

	if resourceType := c.Query("resourceType"); resourceType != "" {
		rt := audit.ResourceType(resourceType)
		query.ResourceType = &rt
	}

	if orgIDStr := c.Query("organizationId"); orgIDStr != "" {
		if orgID, err := strconv.ParseInt(orgIDStr, 10, 64); err == nil {
			query.OrganizationID = &orgID
		}
	}

	if status := c.Query("status"); status != "" {
		s := audit.Status(status)
		query.Status = &s
	}

	if startTimeStr := c.Query("startTime"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			query.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("endTime"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			query.EndTime = &endTime
		}
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := h.auditService.QueryAuditLogs(&query)
	if err != nil {
		response.InternalError(c, "Failed to query audit logs")
		return
	}

	response.Success(c, result)
}

type ExportRequest struct {
	IDs    []int64 `json:"ids" binding:"required"`
	Format string  `json:"format"`
}

func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	defer recoverServicePanic(c)
	var req ExportRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "No audit log IDs provided")
		return
	}

	format := req.Format
	if format == "" {
		format = "csv"
	}

	filePath, err := h.auditService.ExportAuditLogs(req.IDs, format)
	if err != nil {
		response.InternalError(c, "Failed to export audit logs")
		return
	}

	response.Success(c, gin.H{
		"filePath": filePath,
		"format":   format,
	})
}

type RetentionRequest struct {
	Days int `json:"days"`
}

func (h *AuditHandler) DeleteAuditLogsRetention(c *gin.Context) {
	defer recoverServicePanic(c)
	var req RetentionRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		req.Days = 90
	}

	affected, err := h.auditService.DeleteAuditLogsBeforeDate(req.Days)
	if err != nil {
		response.InternalError(c, "Failed to delete audit logs")
		return
	}

	response.Success(c, gin.H{
		"deleted": affected,
	})
}

func (h *AuditHandler) RecordLoginFailure(c *gin.Context) {
	defer recoverServicePanic(c)
	var req audit.LoginFailureRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if req.IPAddress == nil {
		req.IPAddress = &ipAddress
	}
	if req.UserAgent == nil {
		req.UserAgent = &userAgent
	}

	failure, err := h.auditService.RecordLoginFailure(&req)
	if err != nil {
		response.InternalError(c, "Failed to record login failure")
		return
	}

	response.Success(c, failure)
}

func (h *AuditHandler) DownloadExportFile(c *gin.Context) {
	defer recoverServicePanic(c)
	filePath := c.Query("path")
	if filePath == "" {
		response.BadRequest(c, "File path is required")
		return
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		response.BadRequest(c, "Invalid file path")
		return
	}
	tempDir, err := filepath.Abs(os.TempDir())
	if err != nil {
		response.InternalError(c, "Failed to resolve export directory")
		return
	}
	if absPath != tempDir && !strings.HasPrefix(absPath, tempDir+string(os.PathSeparator)) {
		response.BadRequest(c, "Invalid file path")
		return
	}

	baseName := filepath.Base(absPath)
	if !strings.HasPrefix(baseName, "audit_logs_") {
		response.BadRequest(c, "Invalid file path")
		return
	}

	ext := strings.ToLower(filepath.Ext(baseName))
	if ext != ".csv" && ext != ".json" {
		response.BadRequest(c, "Invalid export format")
		return
	}
	format := strings.TrimPrefix(ext, ".")
	if requested := strings.TrimSpace(c.Query("format")); requested != "" && requested != format {
		response.BadRequest(c, "Invalid export format")
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			response.NotFound(c, "Export file not found")
			return
		}
		response.InternalError(c, "Failed to access export file")
		return
	}
	if info.IsDir() {
		response.BadRequest(c, "Invalid file path")
		return
	}

	c.FileAttachment(absPath, "audit_logs."+format)
}

func RegisterAuditRoutes(r *gin.RouterGroup, h *AuditHandler, rateLimitOpts *middleware.RouteRateLimitOptions, menuAuthMiddlewares ...*middleware.MenuAuthMiddleware) {
	var menuAuthMiddleware *middleware.MenuAuthMiddleware
	if len(menuAuthMiddlewares) > 0 {
		menuAuthMiddleware = menuAuthMiddlewares[0]
	}
	auditGroup := r.Group("/audit")
	exportGroup := auditGroup.Group("")
	if menuAuthMiddleware != nil {
		exportGroup.Use(menuAuthMiddleware.RequireMenuAuth(auditMenuPath))
	}
	exportMiddleware := middleware.RateLimit(
		"audit-export",
		auditExportRateLimitRequests,
		auditExportRateLimitWindow,
		middleware.AuthenticatedUserKey,
	)
	if rateLimitOpts != nil {
		enabled, maxRequests, window := middleware.ResolveRouteLimit(rateLimitOpts.Config, "audit-export", auditExportRateLimitRequests, auditExportRateLimitWindow)
		if enabled {
			exportMiddleware = middleware.ConfigurableRateLimit("audit-export", maxRequests, window, rateLimitOpts.Backend, middleware.AuthenticatedUserKey)
		} else {
			exportMiddleware = nil
		}
	}
	if exportMiddleware != nil {
		exportGroup.Use(exportMiddleware)
	}
	{
		auditGroup.POST("/log", h.CreateAuditLog)
		auditGroup.GET("/list", h.QueryAuditLogs)
		auditGroup.GET("/settings", h.GetAuditAlertSettings)
		auditGroup.PUT("/settings", h.SaveAuditAlertSettings)
		auditGroup.POST("/cleanup", h.CleanupNow)
		auditGroup.POST("/test-notification", h.TestNotification)
		auditGroup.GET("/:id", h.GetAuditLogByID)
		auditGroup.GET("/user/:userId", h.GetAuditLogsByUserID)
		exportGroup.POST("/export", h.ExportAuditLogs)
		auditGroup.DELETE("/retention", h.DeleteAuditLogsRetention)
		auditGroup.POST("/login-failure", h.RecordLoginFailure)
		exportGroup.GET("/download", h.DownloadExportFile)
	}
}
