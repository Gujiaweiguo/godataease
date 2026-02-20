package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type GeoHandler struct {
	service *service.GeoService
}

func NewGeoHandler(service *service.GeoService) *GeoHandler {
	return &GeoHandler{service: service}
}

func (h *GeoHandler) ListAreas(c *gin.Context) {
	result, err := h.service.ListAreas()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *GeoHandler) GetArea(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetArea(id)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterGeoRoutes(r *gin.RouterGroup, h *GeoHandler) {
	geoGroup := r.Group("/geometry")
	{
		geoGroup.GET("/areaList", h.ListAreas)
		geoGroup.GET("/area/:id", h.GetArea)
	}
}
