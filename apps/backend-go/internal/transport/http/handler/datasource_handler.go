package handler

import (
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type DatasourceHandler struct {
	service *service.DatasourceService
}

func NewDatasourceHandler(service *service.DatasourceService) *DatasourceHandler {
	return &DatasourceHandler{service: service}
}

func (h *DatasourceHandler) List(c *gin.Context) {
	var req datasource.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.List(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasourceHandler) Validate(c *gin.Context) {
	var req datasource.ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Validate(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func RegisterDatasourceRoutes(r *gin.RouterGroup, h *DatasourceHandler) {
	dsGroup := r.Group("/ds")
	{
		dsGroup.POST("/list", h.List)
		dsGroup.POST("/validate", h.Validate)
	}
}
