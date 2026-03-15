package handler

import (
	"strconv"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type WatermarkHandler struct {
	service *service.WatermarkService
}

func NewWatermarkHandler(service *service.WatermarkService) *WatermarkHandler {
	return &WatermarkHandler{service: service}
}

func (h *WatermarkHandler) Find(c *gin.Context) {
	result, err := h.service.Find()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *WatermarkHandler) Save(c *gin.Context) {
	var req visualization.WatermarkSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	updateBy := getCreateByFromContext(c)
	result, err := h.service.Save(&req, updateBy)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterWatermarkRoutes(r gin.IRouter, h *WatermarkHandler) {
	group := r.Group("/watermark")
	{
		group.GET("/find", h.Find)
		group.POST("/save", h.Save)
	}
}

func getCreateByFromContext(c *gin.Context) string {
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
	if userID, exists := c.Get("user_id"); exists {
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
