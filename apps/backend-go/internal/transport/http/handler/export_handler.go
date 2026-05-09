package handler

import (
	"dataease/backend/internal/domain/export"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type ExportHandler struct {
	service       *service.ExportService
	exportPermSvc *service.ExportPermissionService
	adminChecker  middleware.AdminChecker
}

func NewExportHandler(
	service *service.ExportService,
	exportPermSvc *service.ExportPermissionService,
	adminChecker middleware.AdminChecker,
) *ExportHandler {
	return &ExportHandler{
		service:       service,
		exportPermSvc: exportPermSvc,
		adminChecker:  adminChecker,
	}
}

func (h *ExportHandler) ExportTasks(c *gin.Context) {
	defer recoverServicePanic(c)
	result := h.service.ExportTasks()
	response.Success(c, result)
}

func (h *ExportHandler) Pager(c *gin.Context) {
	defer recoverServicePanic(c)
	status := c.Param("status")
	goPage := 1
	pageSize := 10

	req := &export.PagerRequest{
		GoPage:   goPage,
		PageSize: pageSize,
		Status:   status,
	}

	result := h.service.Pager(req)
	response.Success(c, result)
}

func (h *ExportHandler) Delete(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}

	if err := h.service.Delete(id); err != nil {
		response.InternalError(c, "failed to delete export task")
		return
	}

	response.Success(c, nil)
}

func (h *ExportHandler) DeleteBatch(c *gin.Context) {
	defer recoverServicePanic(c)
	var req export.DeleteRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	if err := h.service.DeleteBatch(req.IDs); err != nil {
		response.InternalError(c, "failed to delete export tasks")
		return
	}

	response.Success(c, nil)
}

func (h *ExportHandler) DeleteAll(c *gin.Context) {
	defer recoverServicePanic(c)
	exportFromType := c.Param("type")

	if err := h.service.DeleteAll(exportFromType); err != nil {
		response.InternalError(c, "failed to delete all export tasks")
		return
	}

	response.Success(c, nil)
}

func (h *ExportHandler) Download(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}

	task, err := h.service.GetByID(id)
	if err != nil {
		response.NotFoundExport(c, "导出任务不存在")
		return
	}

	userID := int64(middleware.GetUserID(c))
	role := middleware.GetRole(c)
	isAdmin := role == defaultAdminCredential

	if err := h.service.CheckAccess(task, userID, isAdmin); err != nil {
		if err == service.ErrUnauthorized {
			response.ForbiddenExport(c, "无权访问该导出任务")
			return
		}
		response.NotFoundExport(c, "导出任务不存在")
		return
	}

	if !h.checkExportPermission(c, task, userID, isAdmin) {
		return
	}

	response.Success(c, &export.DownloadResponse{URL: "/downloads/" + task.ID})
}

func (h *ExportHandler) GenerateDownloadURI(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}

	task, err := h.service.GetByID(id)
	if err != nil {
		response.NotFoundExport(c, "导出任务不存在")
		return
	}

	userID := int64(middleware.GetUserID(c))
	role := middleware.GetRole(c)
	isAdmin := role == defaultAdminCredential

	if err = h.service.CheckAccess(task, userID, isAdmin); err != nil {
		if err == service.ErrUnauthorized {
			response.ForbiddenExport(c, "无权访问该导出任务")
			return
		}
		response.NotFoundExport(c, "导出任务不存在")
		return
	}

	if !h.checkExportPermission(c, task, userID, isAdmin) {
		return
	}

	response.Success(c, "/downloads/"+id)
}

func (h *ExportHandler) checkExportPermission(c *gin.Context, task *export.ExportTask, userID int64, isAdmin bool) bool {
	if task == nil || h.exportPermSvc == nil {
		return true
	}

	if isAdmin || (h.adminChecker != nil && h.adminChecker.IsAdmin(userID)) {
		return true
	}

	if task.ExportFrom <= 0 || task.ExportFromType == "" {
		return true
	}

	resourceType := normalizeExportResourceType(task.ExportFromType)
	if resourceType == "" {
		return true
	}

	exportType := service.ExportType(c.Query("exportType"))
	if exportType == "" {
		exportType = service.ExportTypeImage
	}

	req := &service.ExportCheckRequest{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   task.ExportFrom,
		ExportType:   exportType,
	}

	result := h.exportPermSvc.CheckExportPermission(c.Request.Context(), req)
	if !result.CanExport {
		response.ForbiddenExport(c, "无权导出该资源")
		return false
	}

	return true
}

func normalizeExportResourceType(raw string) string {
	switch raw {
	case permission.ResourceTypeDataset:
		return permission.ResourceTypeDataset
	case permission.ResourceTypeDashboard:
		return permission.ResourceTypeDashboard
	case permission.ResourceTypeScreen:
		return permission.ResourceTypeScreen
	case permission.ResourceTypeDatasource:
		return permission.ResourceTypeDatasource
	default:
		return ""
	}
}

func (h *ExportHandler) Retry(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}

	if err := h.service.Retry(id); err != nil {
		response.InternalError(c, "failed to retry export task")
		return
	}

	response.Success(c, nil)
}

func (h *ExportHandler) ExportLimit(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.service == nil {
		response.Error(c, response.CodeInternalError, "Service unavailable")
		return
	}
	result := h.service.ExportLimit()
	response.Success(c, result.Limit)
}

func RegisterExportRoutes(r gin.IRouter, h *ExportHandler) {
	group := r.Group("/exportTasks")
	{
		group.POST("/records", h.ExportTasks)
		group.POST("/:status/:goPage/:pageSize", h.Pager)
		group.GET("/delete/:id", h.Delete)
		group.POST("/delete", h.DeleteBatch)
		group.POST("/deleteAll/:type", h.DeleteAll)
		group.GET("/download/:id", h.Download)
		group.GET("/generateDownloadUri/:id", h.GenerateDownloadURI)
		group.POST("/retry/:id", h.Retry)
		group.POST("/exportLimit", h.ExportLimit)
	}

	exportCenter := r.Group("/exportCenter")
	{
		exportCenter.POST("/exportLimit", h.ExportLimit)
		exportCenter.POST("/exportTasks/records", h.ExportTasks)
		exportCenter.POST("/exportTasks/:status/:goPage/:pageSize", h.Pager)
		exportCenter.GET("/exportTasks", h.ExportTasks)
		exportCenter.GET("/delete/:id", h.Delete)
		exportCenter.POST("/delete", h.DeleteBatch)
		exportCenter.POST("/deleteAll/:type", h.DeleteAll)
		exportCenter.GET("/download/:id", h.Download)
		exportCenter.GET("/generateDownloadUri/:id", h.GenerateDownloadURI)
		exportCenter.POST("/retry/:id", h.Retry)
	}
}
