package handler

import (
	"strconv"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	service *service.RoleService
}

func NewRoleHandler(service *service.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

func (h *RoleHandler) Query(c *gin.Context) {
	var req role.RoleQueryRequest
	_ = c.ShouldBindJSON(&req)

	result, err := h.service.QueryRoles(&req)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, gin.H{"list": result})
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req role.RoleCreator
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	createBy := h.getCreateBy(c)
	id, err := h.service.CreateRole(&req, createBy)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, id)
}

func (h *RoleHandler) Edit(c *gin.Context) {
	var req role.RoleEditor
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	updateBy := h.getCreateBy(c)
	if err := h.service.EditRole(&req, updateBy); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid role ID")
		return
	}

	if err := h.service.DeleteRole(id); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *RoleHandler) Detail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid role ID")
		return
	}

	result, err := h.service.GetRoleByID(id)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *RoleHandler) getCreateBy(c *gin.Context) string {
	if userId, exists := c.Get("userId"); exists {
		switch v := userId.(type) {
		case string:
			return v
		case int64:
			return strconv.FormatInt(v, 10)
		case int:
			return strconv.Itoa(v)
		}
	}
	return "system"
}

// MountUser 绑定用户到角色
func (h *RoleHandler) MountUser(c *gin.Context) {
	var req role.MountUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.service.MountUsers(&req); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

// MountExternalUser 绑定组织外用户
func (h *RoleHandler) MountExternalUser(c *gin.Context) {
	var req role.MountExternalUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	// TODO: 从上下文获取组织ID，暂时使用1
	orgID := int64(1)
	if err := h.service.MountExternalUser(&req, orgID); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

// UnmountUser 解绑用户
func (h *RoleHandler) UnmountUser(c *gin.Context) {
	var req role.UnmountUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.service.UnmountUser(&req); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

// BeforeUnmountInfo 解绑前检查
func (h *RoleHandler) BeforeUnmountInfo(c *gin.Context) {
	var req role.UnmountUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	count, err := h.service.BeforeUnmountInfo(&req)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, count)
}

// SearchExternalUser 搜索组织外用户
func (h *RoleHandler) SearchExternalUser(c *gin.Context) {
	keyword := c.Param("keyword")
	// TODO: 从上下文获取排除的组织ID，暂时使用1
	excludeOrgID := int64(1)

	result, err := h.service.SearchExternalUser(keyword, excludeOrgID)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

// OptionForUser 用户可选角色
func (h *RoleHandler) OptionForUser(c *gin.Context) {
	var req role.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	// TODO: 从上下文获取组织ID，暂时使用1
	orgID := int64(1)
	result, err := h.service.OptionForUser(&req, orgID)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

// SelectedForUser 用户已选角色
func (h *RoleHandler) SelectedForUser(c *gin.Context) {
	var req role.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.SelectedForUser(&req)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

func RegisterRoleRoutes(r *gin.RouterGroup, h *RoleHandler) {
	roleGroup := r.Group("/role")
	{
		roleGroup.POST("/query", h.Query)
		roleGroup.POST("/byCurOrg", h.Query)
		roleGroup.POST("/create", h.Create)
		roleGroup.POST("/edit", h.Edit)
		roleGroup.POST("/delete/:id", h.Delete)
		roleGroup.POST("/mountUser", h.MountUser)
		roleGroup.POST("/mountExternalUser", h.MountExternalUser)
		roleGroup.POST("/unMountUser", h.UnmountUser)
		roleGroup.POST("/beforeUnmountInfo", h.BeforeUnmountInfo)
		roleGroup.GET("/searchExternalUser/:keyword", h.SearchExternalUser)
		roleGroup.POST("/user/option", h.OptionForUser)
		roleGroup.POST("/user/selected", h.SelectedForUser)
	}

	// System role routes - frontend compatibility aliases
	systemRoleGroup := r.Group("/system/role")
	{
		systemRoleGroup.POST("/create", h.Create)
		systemRoleGroup.POST("/update", h.Edit)  // Map to edit
		systemRoleGroup.POST("/delete/:id", h.Delete)
	}
}
