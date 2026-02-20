package handler

import (
	"dataease/backend/internal/domain/export"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	service *service.ExportService
}

func NewExportHandler(service *service.ExportService) *ExportHandler {
	return &ExportHandler{service: service}
}

func (h *ExportHandler) ExportTasks(c *gin.Context) {
	result := h.service.ExportTasks()
	response.Success(c, result)
}

func (h *ExportHandler) Pager(c *gin.Context) {
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
	var req export.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	exportFromType := c.Param("type")

	if err := h.service.DeleteAll(exportFromType); err != nil {
		response.InternalError(c, "failed to delete all export tasks")
		return
	}

	response.Success(c, nil)
}

func (h *ExportHandler) Download(c *gin.Context) {
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
	isAdmin := role == "admin"

	if err := h.service.CheckAccess(task, userID, isAdmin); err != nil {
		if err == service.ErrUnauthorized {
			response.ForbiddenExport(c, "无权访问该导出任务")
			return
		}
		response.NotFoundExport(c, "导出任务不存在")
		return
	}

	response.Success(c, &export.DownloadResponse{URL: "/downloads/" + task.ID})
}

func (h *ExportHandler) GenerateDownloadURI(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}

	response.Success(c, "/downloads/"+id)
}

func (h *ExportHandler) Retry(c *gin.Context) {
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
		exportCenter.GET("/exportTasks", h.ExportTasks)
		exportCenter.GET("/download/:id", h.Download)
	}
}
