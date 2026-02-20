package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	service *service.MenuService
}

func NewMenuHandler(service *service.MenuService) *MenuHandler {
	return &MenuHandler{service: service}
}

func (h *MenuHandler) Query(c *gin.Context) {
	result, err := h.service.Query()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterMenuRoutes(r *gin.RouterGroup, h *MenuHandler) {
	menuGroup := r.Group("/menu")
	{
		menuGroup.GET("/query", h.Query)
	}
}
