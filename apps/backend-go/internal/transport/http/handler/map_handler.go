package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type MapHandler struct {
	service *service.MapService
}

func NewMapHandler(service *service.MapService) *MapHandler {
	return &MapHandler{service: service}
}

func (h *MapHandler) GetWorldTree(c *gin.Context) {
	result, err := h.service.GetWorldTree()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterMapRoutes(r *gin.RouterGroup, h *MapHandler) {
	mapGroup := r.Group("/map")
	{
		mapGroup.GET("/worldTree", h.GetWorldTree)
	}
}
