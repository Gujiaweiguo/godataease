package handler

import (
	"strings"

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

func (h *GeoHandler) Save(c *gin.Context) {
	code := c.PostForm("code")
	name := c.PostForm("name")
	pid := c.PostForm("pid")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, "400000", "geometry file is required")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".json") {
		response.Error(c, "400000", "only json format files are supported")
		return
	}

	fileContent := make([]byte, header.Size)
	if _, err := file.Read(fileContent); err != nil {
		response.Error(c, "500000", "failed to read file")
		return
	}

	if err := h.service.SaveMapGeo(code, name, pid, fileContent, header.Filename); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *GeoHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteGeo(id); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, nil)
}

func RegisterGeoRoutes(r *gin.RouterGroup, h *GeoHandler) {
	geoGroup := r.Group("/geometry")
	{
		geoGroup.GET("/areaList", h.ListAreas)
		geoGroup.GET("/area/:id", h.GetArea)
		geoGroup.POST("/save", h.Save)
		geoGroup.DELETE("/delete/:id", h.Delete)
	}
}
