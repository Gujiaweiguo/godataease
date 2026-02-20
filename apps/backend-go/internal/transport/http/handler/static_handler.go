package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type StaticHandler struct {
	service *service.StaticService
}

func NewStaticHandler(service *service.StaticService) *StaticHandler {
	return &StaticHandler{service: service}
}

func (h *StaticHandler) ListResources(c *gin.Context) {
	result, err := h.service.ListResources()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *StaticHandler) GetResource(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetResource(id)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *StaticHandler) ListStores(c *gin.Context) {
	result, err := h.service.ListStores()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *StaticHandler) ListTypefaces(c *gin.Context) {
	result, err := h.service.ListTypefaces()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

// ListFont returns font list for frontend compatibility
func (h *StaticHandler) ListFont(c *gin.Context) {
	result, err := h.service.ListTypefaces()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

// DefaultFont returns default font for frontend compatibility
func (h *StaticHandler) DefaultFont(c *gin.Context) {
	result, err := h.service.ListTypefaces()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	// Return first font as default, or empty object if none
	if len(result) > 0 {
		response.Success(c, result[0])
		return
	}
	response.Success(c, map[string]interface{}{})
}

// XpackModel returns xpack model status
func (h *StaticHandler) XpackModel(c *gin.Context) {
	response.Success(c, false)
}

func RegisterStaticRoutes(r *gin.RouterGroup, h *StaticHandler) {
	staticGroup := r.Group("/staticResource")
	{
		staticGroup.GET("/list", h.ListResources)
		staticGroup.GET("/:id", h.GetResource)
	}

	storeGroup := r.Group("/store")
	{
		storeGroup.GET("/list", h.ListStores)
	}

	typefaceGroup := r.Group("/typeface")
	{
		typefaceGroup.GET("/list", h.ListTypefaces)
		typefaceGroup.GET("/listFont", h.ListFont)
		typefaceGroup.GET("/defaultFont", h.DefaultFont)
	}

	// Xpack model endpoint
	r.GET("/xpackModel", h.XpackModel)
}
