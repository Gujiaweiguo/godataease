package handler

import (
	"dataease/backend/internal/domain/engine"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EngineHandler struct {
	service *service.EngineService
}

func NewEngineHandler(service *service.EngineService) *EngineHandler {
	return &EngineHandler{service: service}
}

func (h *EngineHandler) GetEngine(c *gin.Context) {
	result, err := h.service.GetEngine()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *EngineHandler) Validate(c *gin.Context) {
	var req engine.ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.Validate(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *EngineHandler) ValidateByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid id")
		return
	}
	result, err := h.service.ValidateByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *EngineHandler) SupportSetKey(c *gin.Context) {
	result, err := h.service.SupportSetKey()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterEngineRoutes(r *gin.RouterGroup, h *EngineHandler) {
	eg := r.Group("/engine")
	{
		eg.GET("/getEngine", h.GetEngine)
		eg.POST("/validate", h.Validate)
		eg.POST("/validate/:id", h.ValidateByID)
		eg.GET("/supportSetKey", h.SupportSetKey)
	}
}
