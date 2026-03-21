package handler

import (
	"strconv"

	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type DataPermissionHandler struct {
	service *service.DataPermissionAdminService
}

func NewDataPermissionHandler(service *service.DataPermissionAdminService) *DataPermissionHandler {
	return &DataPermissionHandler{service: service}
}

func (h *DataPermissionHandler) RowPermissionPager(c *gin.Context) {
	datasetID, page, size, ok := parseDatasetPagerParams(c)
	if !ok {
		return
	}

	result, err := h.service.RowPermissionPage(datasetID, page, size)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataPermissionHandler) SaveRowPermission(c *gin.Context) {
	var req service.RowPermissionForm
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.service.SaveRowPermission(&req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataPermissionHandler) DeleteRowPermission(c *gin.Context) {
	var req service.DeletePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.service.DeleteRowPermission(req.ID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataPermissionHandler) ColumnPermissionPager(c *gin.Context) {
	datasetID, page, size, ok := parseDatasetPagerParams(c)
	if !ok {
		return
	}

	result, err := h.service.ColumnPermissionPage(datasetID, page, size)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataPermissionHandler) SaveColumnPermission(c *gin.Context) {
	var req service.ColumnPermissionForm
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.service.SaveColumnPermission(&req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataPermissionHandler) DeleteColumnPermission(c *gin.Context) {
	var req service.DeletePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.service.DeleteColumnPermission(req.ID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func RegisterDataPermissionRoutes(r *gin.RouterGroup, h *DataPermissionHandler) {
	if h == nil {
		return
	}

	datasetGroup := r.Group("/dataset")
	{
		datasetGroup.GET("/rowPermissions/pager/:datasetId/:page/:limit", h.RowPermissionPager)
		datasetGroup.POST("/rowPermissions/save", h.SaveRowPermission)
		datasetGroup.POST("/rowPermissions/delete", h.DeleteRowPermission)
		datasetGroup.GET("/columnPermissions/pager/:datasetId/:page/:limit", h.ColumnPermissionPager)
		datasetGroup.POST("/columnPermissions/save", h.SaveColumnPermission)
		datasetGroup.POST("/columnPermissions/delete", h.DeleteColumnPermission)
	}
}

func parseDatasetPagerParams(c *gin.Context) (int64, int, int, bool) {
	datasetID, err := strconv.ParseInt(c.Param("datasetId"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return 0, 0, 0, false
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		response.Error(c, "500000", "Invalid page")
		return 0, 0, 0, false
	}

	size, err := strconv.Atoi(c.Param("limit"))
	if err != nil {
		response.Error(c, "500000", "Invalid limit")
		return 0, 0, 0, false
	}

	return datasetID, page, size, true
}
