package handler

import (
	"strconv"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PermissionCompatHandler struct {
	menuService         *service.MenuService
	permService         *service.PermService
	roleMenuService     *service.RoleMenuService
	resourcePermService *service.ResourcePermissionService
}

func NewPermissionCompatHandler(
	menuService *service.MenuService,
	permService *service.PermService,
	roleMenuService *service.RoleMenuService,
	resourcePermService *service.ResourcePermissionService,
) *PermissionCompatHandler {
	return &PermissionCompatHandler{
		menuService:         menuService,
		permService:         permService,
		roleMenuService:     roleMenuService,
		resourcePermService: resourcePermService,
	}
}

type permissionSaveRequest struct {
	RoleID  int64   `json:"roleId" binding:"required"`
	PermIDs []int64 `json:"permIds"`
}

type menuSaveRequest struct {
	RoleID  int64   `json:"roleId" binding:"required"`
	MenuIDs []int64 `json:"menuIds"`
}

type targetPermissionRequest struct {
	RoleID int64 `json:"roleId" binding:"required"`
}

type targetPermissionSaveRequest struct {
	RoleID      int64                    `json:"roleId" binding:"required"`
	TargetPerms []map[string]interface{} `json:"targetPerms"`
}

type roleIDQuery struct {
	RoleID int64 `json:"roleId"`
}

func parseRoleIDQuery(c *gin.Context) int64 {
	var req roleIDQuery
	_ = c.ShouldBindJSON(&req)
	if req.RoleID > 0 {
		return req.RoleID
	}
	if roleIDStr := c.Query("roleId"); roleIDStr != "" {
		if roleID, err := strconv.ParseInt(roleIDStr, 10, 64); err == nil {
			return roleID
		}
	}
	return 0
}

func (h *PermissionCompatHandler) MenuPermission(c *gin.Context) {
	menus, err := h.menuService.Query()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	roleID := parseRoleIDQuery(c)

	menuIDs := make([]int64, 0)
	if roleID > 0 {
		auth, authErr := h.roleMenuService.GetRoleMenuAuth(roleID)
		if authErr == nil && auth != nil {
			menuIDs = auth.MenuIDs
		}
	}

	response.Success(c, gin.H{
		"menuTree": menus,
		"menuIds":  menuIDs,
	})
}

func (h *PermissionCompatHandler) SaveMenuPer(c *gin.Context) {
	var req menuSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.roleMenuService.SaveRoleMenuAuth(&service.SaveRoleMenuRequest{
		RoleID:  req.RoleID,
		MenuIDs: req.MenuIDs,
	}); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *PermissionCompatHandler) BusiPermission(c *gin.Context) {
	roleID := parseRoleIDQuery(c)

	list, err := h.permService.ListPerms(&permission.PermQueryRequest{Current: 1, Size: 1000})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	permIDs := make([]int64, 0)
	if roleID > 0 {
		permIDs, err = h.resourcePermService.GetRolePermissionIDs(roleID)
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
	}

	response.Success(c, gin.H{
		"permissions": list.List,
		"permIds":     permIDs,
	})
}

func (h *PermissionCompatHandler) SaveBusiPer(c *gin.Context) {
	var req permissionSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.saveRolePerms(req.RoleID, req.PermIDs); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *PermissionCompatHandler) SaveRolePermission(c *gin.Context) {
	h.SaveBusiPer(c)
}

func (h *PermissionCompatHandler) BusiResource(c *gin.Context) {
	list, err := h.permService.ListPerms(&permission.PermQueryRequest{Current: 1, Size: 1000})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, list.List)
}

func (h *PermissionCompatHandler) MenuTargetPermission(c *gin.Context) {
	var req targetPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	response.Error(c, "501000", "Not Implemented: menu target permission compatibility is unavailable")
}

func (h *PermissionCompatHandler) BusiTargetPermission(c *gin.Context) {
	var req targetPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	response.Error(c, "501000", "Not Implemented: business target permission compatibility is unavailable")
}

func (h *PermissionCompatHandler) SaveMenuTargetPer(c *gin.Context) {
	var req targetPermissionSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	response.Error(c, "501000", "Not Implemented: save menu target permission compatibility is unavailable")
}

func (h *PermissionCompatHandler) SaveBusiTargetPer(c *gin.Context) {
	var req targetPermissionSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	response.Error(c, "501000", "Not Implemented: save business target permission compatibility is unavailable")
}

func (h *PermissionCompatHandler) saveRolePerms(roleID int64, targetPermIDs []int64) error {
	target := make(map[int64]struct{}, len(targetPermIDs))
	for _, id := range targetPermIDs {
		if id > 0 {
			target[id] = struct{}{}
		}
	}

	currentPermIDs, err := h.resourcePermService.GetRolePermissionIDs(roleID)
	if err != nil {
		return err
	}

	current := make(map[int64]struct{}, len(currentPermIDs))
	for _, id := range currentPermIDs {
		current[id] = struct{}{}
	}

	for id := range target {
		if _, exists := current[id]; !exists {
			if grantErr := h.resourcePermService.GrantPermissionToRole(roleID, id); grantErr != nil {
				return grantErr
			}
		}
	}

	for id := range current {
		if _, keep := target[id]; !keep {
			if revokeErr := h.resourcePermService.RevokePermissionFromRole(roleID, id); revokeErr != nil {
				return revokeErr
			}
		}
	}

	return nil
}

func RegisterPermissionCompatRoutes(r *gin.RouterGroup, h *PermissionCompatHandler) {
	if h == nil {
		return
	}

	authGroup := r.Group("/auth")
	{
		authGroup.GET("/menuPermission", h.MenuPermission)
		authGroup.POST("/menuPermission", h.MenuPermission)
		authGroup.GET("/busiPermission", h.BusiPermission)
		authGroup.POST("/busiPermission", h.BusiPermission)
		authGroup.GET("/busiResource/:flag", h.BusiResource)
		authGroup.POST("/menuTargetPermission", h.MenuTargetPermission)
		authGroup.POST("/busiTargetPermission", h.BusiTargetPermission)
		authGroup.POST("/saveMenuPer", h.SaveMenuPer)
		authGroup.POST("/saveBusiPer", h.SaveBusiPer)
		authGroup.POST("/saveMenuTargetPer", h.SaveMenuTargetPer)
		authGroup.POST("/saveBusiTargetPer", h.SaveBusiTargetPer)
	}

	roleGroup := r.Group("/role")
	{
		roleGroup.POST("/permission/save", h.SaveRolePermission)
	}

	systemRoleGroup := r.Group("/system/role")
	{
		systemRoleGroup.POST("/permission/save", h.SaveRolePermission)
	}
}
