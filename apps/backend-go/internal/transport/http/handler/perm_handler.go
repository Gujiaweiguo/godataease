package handler

import (
	"strconv"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PermHandler struct {
	permService *service.PermService
}

func NewPermHandler(permService *service.PermService) *PermHandler {
	return &PermHandler{
		permService: permService,
	}
}

func (h *PermHandler) ListPerms(c *gin.Context) {
	var req permission.PermQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 10
	}

	result, err := h.permService.ListPerms(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *PermHandler) CreatePerm(c *gin.Context) {
	var req permission.PermCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	permID, err := h.permService.CreatePerm(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, permID)
}

func (h *PermHandler) UpdatePerm(c *gin.Context) {
	var req permission.PermUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	err := h.permService.UpdatePerm(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *PermHandler) DeletePerm(c *gin.Context) {
	permIDStr := c.Param("permId")
	permID, err := strconv.ParseInt(permIDStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid permission ID")
		return
	}

	err = h.permService.DeletePerm(permID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func RegisterPermRoutes(r *gin.RouterGroup, h *PermHandler) {
	permGroup := r.Group("/system/permission")
	{
		permGroup.POST("/list", h.ListPerms)
		permGroup.POST("/create", h.CreatePerm)
		permGroup.POST("/update", h.UpdatePerm)
		permGroup.POST("/delete/:permId", h.DeletePerm)
	}
}
