package handler

import (
	"strconv"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

type WatermarkHandler struct {
	service *service.WatermarkService
}

func NewWatermarkHandler(service *service.WatermarkService) *WatermarkHandler {
	return &WatermarkHandler{service: service}
}

func (h *WatermarkHandler) Find(c *gin.Context) {
	if getAuthenticatedUserID(c) == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	result, err := h.service.Find()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *WatermarkHandler) Save(c *gin.Context) {
	userID := getAuthenticatedUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req visualization.WatermarkSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	updateBy := strconv.FormatUint(userID, 10)
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

func getAuthenticatedUserID(c *gin.Context) uint64 {
	if userID := middleware.GetUserID(c); userID > 0 {
		return userID
	}

	if userID, exists := c.Get("userId"); exists {
		switch v := userID.(type) {
		case uint64:
			return v
		case uint:
			return uint64(v)
		case int64:
			if v > 0 {
				return uint64(v)
			}
		case int:
			if v > 0 {
				return uint64(v)
			}
		case string:
			if parsed, err := strconv.ParseUint(v, 10, 64); err == nil {
				return parsed
			}
		}
	}

	return 0
}
