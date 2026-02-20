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

	response.Success(c, result)
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

func RegisterRoleRoutes(r *gin.RouterGroup, h *RoleHandler) {
	roleGroup := r.Group("/role")
	{
		roleGroup.POST("/query", h.Query)
		roleGroup.POST("/create", h.Create)
		roleGroup.POST("/edit", h.Edit)
		roleGroup.POST("/delete/:id", h.Delete)
		roleGroup.GET("/detail/:id", h.Detail)
	}
}
