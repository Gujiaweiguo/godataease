package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// OuterParamsHandler handles HTTP requests for dashboard external parameters.
type OuterParamsHandler struct {
	service *service.OuterParamsService
}

func NewOuterParamsHandler(service *service.OuterParamsService) *OuterParamsHandler {
	return &OuterParamsHandler{service: service}
}

// QueryWithVisualizationId returns the full outer params config for a dashboard.
func (h *OuterParamsHandler) QueryWithVisualizationId(c *gin.Context) {
	defer recoverServicePanic(c)
	dvID := c.Param("dvId")
	if dvID == "" {
		response.Error(c, "500000", "Invalid dvId")
		return
	}
	result, err := h.service.QueryWithVisualizationId(dvID)
	if err != nil {
		response.Error(c, "500000", "Failed to query outer params: "+err.Error())
		return
	}
	response.Success(c, result)
}

// UpdateOuterParamsSet saves outer params configuration.
func (h *OuterParamsHandler) UpdateOuterParamsSet(c *gin.Context) {
	defer recoverServicePanic(c)
	var dto service.OuterParamsDTO
	if err := c.ShouldBindBodyWith(&dto, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.UpdateOuterParamsSet(&dto); err != nil {
		response.Error(c, "500000", "Failed to update outer params: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// GetOuterParamsInfo returns runtime outer params info for dashboard rendering.
func (h *OuterParamsHandler) GetOuterParamsInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	dvID := c.Param("dvId")
	if dvID == "" {
		response.Error(c, "500000", "Invalid dvId")
		return
	}
	result, err := h.service.GetOuterParamsInfo(dvID)
	if err != nil {
		response.Error(c, "500000", "Failed to get outer params info: "+err.Error())
		return
	}
	response.Success(c, result)
}

// QueryDsWithVisualizationId returns dataset groups with fields and chart views.
func (h *OuterParamsHandler) QueryDsWithVisualizationId(c *gin.Context) {
	defer recoverServicePanic(c)
	dvID := c.Param("dvId")
	if dvID == "" {
		response.Error(c, "500000", "Invalid dvId")
		return
	}
	result, err := h.service.QueryDsWithVisualizationId(dvID)
	if err != nil {
		response.Error(c, "500000", "Failed to query datasets: "+err.Error())
		return
	}
	response.Success(c, result)
}

// RegisterOuterParamsRoutes registers all outer params routes on the given router.
func RegisterOuterParamsRoutes(r gin.IRouter, h *OuterParamsHandler) {
	g := r.Group("/outerParams")
	{
		g.GET("/queryWithVisualizationId/:dvId", h.QueryWithVisualizationId)
		g.POST("/updateOuterParamsSet", h.UpdateOuterParamsSet)
		g.GET("/getOuterParamsInfo/:dvId", h.GetOuterParamsInfo)
		g.GET("/queryDsWithVisualizationId/:dvId", h.QueryDsWithVisualizationId)
	}
}
