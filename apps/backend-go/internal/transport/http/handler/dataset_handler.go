package handler

import (
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

type DatasetHandler struct {
	service *service.DatasetService
}

func NewDatasetHandler(service *service.DatasetService) *DatasetHandler {
	return &DatasetHandler{service: service}
}

func (h *DatasetHandler) Tree(c *gin.Context) {
	var req dataset.TreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Tree(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) Fields(c *gin.Context) {
	var req dataset.FieldsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Fields(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) Preview(c *gin.Context) {
	var req dataset.PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Preview(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) PreviewWithPermission(c *gin.Context) {
	var req dataset.PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	userID := int64(middleware.GetUserID(c))
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	result, err := h.service.PreviewWithPermission(&req, userID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func RegisterDatasetRoutes(r *gin.RouterGroup, h *DatasetHandler) {
	datasetGroup := r.Group("/dataset")
	{
		datasetGroup.POST("/tree", h.Tree)
		datasetGroup.POST("/fields", h.Fields)
		datasetGroup.POST("/preview", h.Preview)
		datasetGroup.POST("/previewWithPerm", h.PreviewWithPermission)
	}
}
