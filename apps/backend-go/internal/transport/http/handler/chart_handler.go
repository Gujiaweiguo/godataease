package handler

import (
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

type ChartHandler struct {
	service       *service.ChartService
	exportService *service.ChartExportService
}

func NewChartHandler(svc *service.ChartService) *ChartHandler {
	return &ChartHandler{service: svc, exportService: service.NewChartExportService(svc)}
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

	userID := int64(middleware.GetUserID(c))
	var result *chart.ChartDataResponse
	var err error
	if userID > 0 {
		result, err = h.service.QueryDataWithPermission(&req, userID)
	} else {
		result, err = h.service.QueryData(&req)
	}
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
