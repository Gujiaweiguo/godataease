package handler

import (
	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type CustomGeoHandler struct {
	repo *repository.CustomGeoRepository
}

func NewCustomGeoHandler(repo *repository.CustomGeoRepository) *CustomGeoHandler {
	return &CustomGeoHandler{repo: repo}
}

func (h *CustomGeoHandler) ListGeoAreas(c *gin.Context) {
	defer recoverServicePanic(c)
	areas, err := h.repo.ListGeoAreas()
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	response.Success(c, areas)
}

func (h *CustomGeoHandler) GetGeoArea(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	subAreas, err := h.repo.GetGeoArea(id)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	response.Success(c, subAreas)
}

func (h *CustomGeoHandler) DeleteGeoArea(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	if err := h.repo.DeleteGeoArea(id); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *CustomGeoHandler) SaveGeoArea(c *gin.Context) {
	defer recoverServicePanic(c)
	var req struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	// Check for duplicate name
	exists, err := h.repo.CheckGeoAreaName(req.Name, req.ID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	if exists {
		response.Error(c, "500001", "区域名称已存在")
		return
	}

	area := &auto.CoreCustomGeoArea{
		ID:   req.ID,
		Name: req.Name,
	}
	if err := h.repo.SaveGeoArea(area); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *CustomGeoHandler) DeleteGeoSubArea(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "invalid id")
	if !ok {
		return
	}
	if err := h.repo.DeleteGeoSubArea(id); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *CustomGeoHandler) SaveGeoSubArea(c *gin.Context) {
	defer recoverServicePanic(c)
	var req struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Scope     string `json:"scope"`
		GeoAreaID string `json:"geoAreaId"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	// Check for duplicate name in same geo area
	exists, err := h.repo.CheckGeoSubAreaName(req.Name, req.GeoAreaID, req.ID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	if exists {
		response.Error(c, "500001", "子区域名称已存在")
		return
	}

	subArea := &auto.CoreCustomGeoSubArea{
		ID:        req.ID,
		Name:      req.Name,
		Scope:     req.Scope,
		GeoAreaID: req.GeoAreaID,
	}
	if err := h.repo.SaveGeoSubArea(subArea); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *CustomGeoHandler) ListSubAreaOptions(c *gin.Context) {
	defer recoverServicePanic(c)
	areas, err := h.repo.ListAreaOptions()
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	// Convert to AreaNode format
	nodes := make([]*areamap.AreaNode, 0, len(areas))
	for _, a := range areas {
		nodes = append(nodes, &areamap.AreaNode{
			ID:    a.ID,
			Level: a.Level,
			Name:  a.Name,
			Pid:   a.Pid,
		})
	}
	response.Success(c, nodes)
}

func RegisterCustomGeoRoutes(r *gin.RouterGroup, h *CustomGeoHandler) {
	customGeoGroup := r.Group("/customGeo")
	{
		customGeoGroup.GET("/geoArea/list", h.ListGeoAreas)
		customGeoGroup.GET("/geoArea/:id", h.GetGeoArea)
		customGeoGroup.DELETE("/geoArea/:id", h.DeleteGeoArea)
		customGeoGroup.POST("/geoArea/save", h.SaveGeoArea)
		customGeoGroup.DELETE("/geoSubArea/:id", h.DeleteGeoSubArea)
		customGeoGroup.POST("/geoSubArea/save", h.SaveGeoSubArea)
		customGeoGroup.GET("/geoSubArea/options", h.ListSubAreaOptions)
	}
}
