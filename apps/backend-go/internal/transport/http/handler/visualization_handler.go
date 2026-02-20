package handler

import (
	"strconv"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type VisualizationHandler struct {
	service *service.VisualizationService
}

func NewVisualizationHandler(service *service.VisualizationService) *VisualizationHandler {
	return &VisualizationHandler{service: service}
}

func (h *VisualizationHandler) FindByID(c *gin.Context) {
	var req visualization.DetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Detail(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) List(c *gin.Context) {
	var req visualization.ListRequest
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

func (h *VisualizationHandler) SaveCanvas(c *gin.Context) {
	var req visualization.SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	updateBy := h.getUpdateBy(c)
	id, err := h.service.Save(&req, updateBy)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, strconv.FormatInt(id, 10))
}

func (h *VisualizationHandler) UpdateCanvas(c *gin.Context) {
	var req visualization.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	updateBy := h.getUpdateBy(c)
	if err := h.service.Update(&req, updateBy); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *VisualizationHandler) DeleteLogic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid ID")
		return
	}

	updateBy := h.getUpdateBy(c)
	if err = h.service.DeleteLogic(id, updateBy); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *VisualizationHandler) getUpdateBy(c *gin.Context) string {
	if userID, exists := c.Get("userId"); exists {
		switch v := userID.(type) {
		case int64:
			return strconv.FormatInt(v, 10)
		case int:
			return strconv.Itoa(v)
		case string:
			return v
		}
	}
	return "system"
}

func RegisterVisualizationRoutes(r *gin.RouterGroup, h *VisualizationHandler) {
	vg := r.Group("/dataVisualization")
	{
		vg.POST("/findById", h.FindByID)
		vg.POST("/list", h.List)
		vg.POST("/saveCanvas", h.SaveCanvas)
		vg.POST("/updateCanvas", h.UpdateCanvas)
		vg.POST("/deleteLogic/:id", h.DeleteLogic)
	}
}
