package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type RoleMenuHandler struct {
	roleMenuService *service.RoleMenuService
}

func NewRoleMenuHandler(roleMenuService *service.RoleMenuService) *RoleMenuHandler {
	return &RoleMenuHandler{roleMenuService: roleMenuService}
}

func (h *RoleMenuHandler) GetRoleMenuAuth(c *gin.Context) {
	defer recoverServicePanic(c)
	roleID, ok := parseIDParamMsg(c, "roleId", "Invalid role ID")
	if !ok {
		return
	}
	var err error

	result, err := h.roleMenuService.GetRoleMenuAuth(roleID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *RoleMenuHandler) SaveRoleMenuAuth(c *gin.Context) {
	defer recoverServicePanic(c)
	var req service.SaveRoleMenuRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.roleMenuService.SaveRoleMenuAuth(&req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func RegisterRoleMenuRoutes(r *gin.RouterGroup, h *RoleMenuHandler) {
	roleMenuGroup := r.Group("/roleMenu")
	{
		roleMenuGroup.GET("/auth/:roleId", h.GetRoleMenuAuth)
		roleMenuGroup.POST("/auth", h.SaveRoleMenuAuth)
	}
}
