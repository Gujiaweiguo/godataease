package handler

import (
	"strconv"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditService *service.AuditService
}

func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

func (h *AuditHandler) CreateAuditLog(c *gin.Context) {
	var req audit.AuditLogCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid audit log ID")
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
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
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
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	var req RetentionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	var req audit.LoginFailureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	filePath := c.Query("path")
	if filePath == "" {
		response.BadRequest(c, "File path is required")
		return
	}

	c.FileAttachment(filePath, "audit_logs."+c.Query("format"))
}

func RegisterAuditRoutes(r *gin.RouterGroup, h *AuditHandler) {
	auditGroup := r.Group("/audit")
	{
		auditGroup.POST("/log", h.CreateAuditLog)
		auditGroup.GET("/list", h.QueryAuditLogs)
		auditGroup.GET("/:id", h.GetAuditLogByID)
		auditGroup.GET("/user/:userId", h.GetAuditLogsByUserID)
		auditGroup.POST("/export", h.ExportAuditLogs)
		auditGroup.DELETE("/retention", h.DeleteAuditLogsRetention)
		auditGroup.POST("/login-failure", h.RecordLoginFailure)
		auditGroup.GET("/download", h.DownloadExportFile)
	}
}
