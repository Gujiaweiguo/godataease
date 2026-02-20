package handler

import (
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ChartHandler struct {
	service *service.ChartService
}

func NewChartHandler(service *service.ChartService) *ChartHandler {
	return &ChartHandler{service: service}
}

func (h *ChartHandler) Query(c *gin.Context) {
	var req chart.ChartQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Query(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *ChartHandler) Data(c *gin.Context) {
	var req chart.ChartDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.QueryData(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterChartRoutes(r *gin.RouterGroup, h *ChartHandler) {
	chartGroup := r.Group("/chart")
	{
		chartGroup.POST("/query", h.Query)
		chartGroup.POST("/data", h.Data)
	}
}
