package handler

import (
	"strconv"
	"strings"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type PermissionCompatHandler struct {
	menuService         *service.MenuService
	permService         *service.PermService
	roleMenuService     *service.RoleMenuService
	resourcePermService *service.ResourcePermissionService
	roleService         *service.RoleService
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

func (h *PermissionCompatHandler) SetRoleService(roleService *service.RoleService) {
	h.roleService = roleService
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

type businessTargetPermissionRequest struct {
	ID     int64  `json:"id"`
	Type   int    `json:"type"`
	Flag   string `json:"flag"`
	RoleID int64  `json:"roleId"`
}

type userPerspectiveCompatRequest struct {
	UserID       int64  `json:"userId" binding:"required"`
	ResourceID   int64  `json:"resourceId"`
	ResourceType string `json:"resourceType"`
}

type targetPermissionSaveRequest struct {
	ID          int64                    `json:"id"`
	Type        int                      `json:"type"`
	Flag        string                   `json:"flag"`
	RoleID      int64                    `json:"roleId"`
	TargetPerms []targetPermissionTarget `json:"targetPerms"`
}

type roleIDQuery struct {
	RoleID int64 `json:"roleId"`
}

func parseRoleIDQuery(c *gin.Context) (int64, error) {
	var req roleIDQuery
	if err := shouldBindOptionalJSON(c, &req); err != nil {
		return 0, err
	}
	if req.RoleID > 0 {
		return req.RoleID, nil
	}
	if roleIDStr := c.Query(middleware.ContextRoleID); roleIDStr != "" {
		if roleID, err := strconv.ParseInt(roleIDStr, 10, 64); err == nil {
			return roleID, nil
		}
	}
	return 0, nil
}

func (h *PermissionCompatHandler) MenuPermission(c *gin.Context) {
	defer recoverServicePanic(c)
	scope, err := buildPermissionScope(c)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	menus, err := h.menuService.Query()
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	roleID, err := parseRoleIDQuery(c)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}

	menuIDs := make([]int64, 0)
	if roleID > 0 {
		auth, authErr := h.roleMenuService.GetRoleMenuAuth(roleID, scope)
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
	defer recoverServicePanic(c)
	var req menuSaveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}

	if err := h.roleMenuService.SaveRoleMenuAuth(&service.SaveRoleMenuRequest{
		RoleID:  req.RoleID,
		MenuIDs: req.MenuIDs,
	}); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *PermissionCompatHandler) BusiPermission(c *gin.Context) {
	defer recoverServicePanic(c)
	scope, err := buildPermissionScope(c)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	roleID, err := parseRoleIDQuery(c)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}

	list, err := h.permService.ListPerms(&permission.PermQueryRequest{Current: 1, Size: 1000}, scope)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	permIDs := make([]int64, 0)
	if roleID > 0 {
		permIDs, err = h.resourcePermService.GetRolePermissionIDs(roleID, scope)
		if err != nil {
			response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
			return
		}
	}

	response.Success(c, gin.H{
		"permissions": list.List,
		"permIds":     permIDs,
	})
}

func (h *PermissionCompatHandler) SaveBusiPer(c *gin.Context) {
	defer recoverServicePanic(c)
	scope, err := buildPermissionScope(c)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	var req permissionSaveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	if h.roleService != nil {
		if err := h.roleService.ValidatePermissionInheritance(req.RoleID, req.PermIDs); err != nil {
			response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
			return
		}
	}

	if err := h.saveRolePerms(req.RoleID, req.PermIDs, scope); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *PermissionCompatHandler) SaveRolePermission(c *gin.Context) {
	defer recoverServicePanic(c)
	h.SaveBusiPer(c)
}

func (h *PermissionCompatHandler) BusiResource(c *gin.Context) {
	defer recoverServicePanic(c)
	scope, err := buildPermissionScope(c)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	flag := strings.TrimSpace(c.Param("flag"))
	list, err := h.permService.ListPerms(&permission.PermQueryRequest{Current: 1, Size: 1000}, scope)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	if flag != "" && flag != "1" && !strings.EqualFold(flag, "all") {
		perms, ok := list.List.([]*permission.SysPerm)
		if !ok {
			response.Error(c, response.CodeInternalError, "Failed: unexpected permission payload")
			return
		}

		filtered := make([]*permission.SysPerm, 0, len(perms))
		prefix := flag + ":"
		for _, perm := range perms {
			if perm == nil {
				continue
			}
			if strings.HasPrefix(perm.PermKey, prefix) {
				filtered = append(filtered, perm)
			}
		}

		response.Success(c, filtered)
		return
	}

	response.Success(c, list.List)
}

func (h *PermissionCompatHandler) MenuTargetPermission(c *gin.Context) {
	defer recoverServicePanic(c)
	scope, err := buildPermissionScope(c)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	var req targetPermissionRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	if req.RoleID <= 0 {
		response.Error(c, response.CodeInternalError, "Invalid request: roleId is required")
		return
	}
	if h.menuService == nil || h.roleMenuService == nil {
		response.Error(c, response.CodeInternalError, "menu target permission compatibility is not configured")
		return
	}

	menus, err := h.menuService.Query()
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	auth, err := h.roleMenuService.GetRoleMenuAuth(req.RoleID, scope)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"menuTree": menus,
		"menuIds":  auth.MenuIDs,
	})
}

func (h *PermissionCompatHandler) BusiTargetPermission(c *gin.Context) {
	defer recoverServicePanic(c)
	scope, err := buildPermissionScope(c)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	var req businessTargetPermissionRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}

	resourceID := req.ID
	if resourceID == 0 {
		resourceID = req.RoleID
	}
	if resourceID <= 0 || req.Flag == "" {
		response.Error(c, response.CodeInternalError, "Invalid request: id and flag are required")
		return
	}

	targetPermissions, err := h.resourcePermService.GetResourcePerspective(resourceID, req.Flag, scope)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, targetPermissions)
}

func (h *PermissionCompatHandler) UserPerspective(c *gin.Context) {
	defer recoverServicePanic(c)
	scope, scopeErr := buildPermissionScope(c)
	if scopeErr != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+scopeErr.Error())
		return
	}
	var req userPerspectiveCompatRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}

	var (
		userPermissions []*permission.UserResourcePermVO
		err             error
	)
	if req.ResourceID > 0 {
		resourcePermissions, resourceErr := h.resourcePermService.GetResourcePerspective(req.ResourceID, req.ResourceType, scope)
		if resourceErr != nil {
			response.Error(c, response.CodeInternalError, "Failed: "+resourceErr.Error())
			return
		}
		userPermissions = make([]*permission.UserResourcePermVO, 0)
		for _, item := range resourcePermissions {
			if item == nil || item.UserID != req.UserID {
				continue
			}
			userPermissions = append(userPermissions, &permission.UserResourcePermVO{
				ResourceID:   req.ResourceID,
				ResourceType: req.ResourceType,
				PermKey:      item.PermKey,
				PermName:     item.PermName,
				SourceType:   item.SourceType,
				SourceID:     item.SourceID,
				SourceName:   item.SourceName,
			})
		}
	} else {
		userPermissions, err = h.resourcePermService.GetUserPerspective(req.UserID, req.ResourceType, scope)
	}
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, userPermissions)
}

func (h *PermissionCompatHandler) SaveMenuTargetPer(c *gin.Context) {
	defer recoverServicePanic(c)
	var req targetPermissionSaveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	if h.roleMenuService == nil {
		response.Error(c, response.CodeInternalError, "save menu target permission compatibility is not configured")
		return
	}

	roleID := req.RoleID
	menuIDs := make([]int64, 0)
	for _, target := range req.TargetPerms {
		targetType := normalizeTargetType(target.TargetType, target.SourceType)
		if targetType != permission.AuthTargetTypeRole {
			response.Error(c, response.CodeInternalError, "Invalid request: only role targets are supported")
			return
		}
		if roleID == 0 {
			roleID = normalizeTargetID(target.TargetID, target.SourceID)
		}
		menuIDs = append(menuIDs, extractTargetPermIDs(target)...)
	}
	if roleID <= 0 {
		response.Error(c, response.CodeInternalError, "Invalid request: roleId is required")
		return
	}

	if err := h.roleMenuService.SaveRoleMenuAuth(&service.SaveRoleMenuRequest{RoleID: roleID, MenuIDs: uniqueInt64(menuIDs)}); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *PermissionCompatHandler) SaveBusiTargetPer(c *gin.Context) {
	defer recoverServicePanic(c)
	scope, err := buildPermissionScope(c)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	var req targetPermissionSaveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}

	resourceID := req.ID
	if resourceID == 0 {
		resourceID = req.RoleID
	}
	if resourceID <= 0 || req.Flag == "" {
		response.Error(c, response.CodeInternalError, "Invalid request: id and flag are required")
		return
	}

	resourcePermIDs := make([]int64, 0)
	preservedDirectPermIDs, err := h.collectDirectResourcePermIDs(resourceID, req.Flag, scope)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	resourcePermIDs = append(resourcePermIDs, preservedDirectPermIDs...)
	for _, target := range req.TargetPerms {
		matchedPermIDs, err := h.collectMatchedTargetPermIDs(target, req.Flag)
		if err != nil {
			response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
			return
		}
		resourcePermIDs = append(resourcePermIDs, matchedPermIDs...)

		if err := h.saveRolePermsForResourceType(normalizeTargetID(target.TargetID, target.SourceID), extractTargetPermIDs(target), req.Flag, scope); err != nil {
			response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
			return
		}
	}

	if err := h.resourcePermService.ReplaceResourcePermissions(resourceID, req.Flag, uniqueInt64(resourcePermIDs)); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// RegisterPermissionCompatRoutes registers permission and menu compatibility routes.
//
// P4 Legacy Compat Contract classification (C1 policy lock):
//
//	DUAL-SUPPORT TRANSITION (C1 keep, C3 migrate):
//	  All routes in this handler are dual-support transition routes.
//	  They bridge the legacy permission/menu API shape to the new service layer.
//	  C3 will migrate the frontend to use canonical permission endpoints directly.
//
//	Route families:
//	  /auth/menuPermission     — menu permission tree + role menu IDs
//	  /auth/busiPermission     — business permission list + role permission IDs
//	  /auth/busiResource/:flag — business resource permission list
//	  /auth/userPerspective    — user perspective permission query
//	  /auth/menuTargetPermission  — menu target permission (role-scoped)
//	  /auth/busiTargetPermission  — business target permission (role-scoped)
//	  /auth/saveMenuPer        — save role menu permission
//	  /auth/saveBusiPer        — save role business permission
//	  /auth/saveMenuTargetPer  — save menu target permission
//	  /auth/saveBusiTargetPer  — save business target permission
//	  /role/permission/save    — role permission save (alias for saveBusiPer)
//	  /system/role/permission/save — system role permission save (alias for saveBusiPer)
func RegisterPermissionCompatRoutes(r *gin.RouterGroup, h *PermissionCompatHandler) {
	if h == nil {
		return
	}

	// Dual-support transition: auth permission/menu routes
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/menuPermission", h.MenuPermission)
		authGroup.POST("/menuPermission", h.MenuPermission)
		authGroup.GET("/busiPermission", h.BusiPermission)
		authGroup.POST("/busiPermission", h.BusiPermission)
		authGroup.GET("/busiResource/:flag", h.BusiResource)
		authGroup.POST("/userPerspective", h.UserPerspective)
		authGroup.POST("/menuTargetPermission", h.MenuTargetPermission)
		authGroup.POST("/busiTargetPermission", h.BusiTargetPermission)
		authGroup.POST("/saveMenuPer", h.SaveMenuPer)
		authGroup.POST("/saveBusiPer", h.SaveBusiPer)
		authGroup.POST("/saveMenuTargetPer", h.SaveMenuTargetPer)
		authGroup.POST("/saveBusiTargetPer", h.SaveBusiTargetPer)
	}

	// Dual-support transition: role permission save aliases
	roleGroup := r.Group("/role")
	{
		roleGroup.POST("/permission/save", h.SaveRolePermission)
	}

	systemRoleGroup := r.Group("/system/role")
	{
		systemRoleGroup.POST("/permission/save", h.SaveRolePermission)
	}
}

// RegisterPermissionRoutes registers canonical permission routes under /system/permission.
func RegisterPermissionRoutes(r *gin.RouterGroup, h *PermissionCompatHandler) {
	if h == nil {
		return
	}

	permGroup := r.Group("/system/permission")
	{
		permGroup.GET("/menuPermission", h.MenuPermission)
		permGroup.POST("/menuPermission", h.MenuPermission)
		permGroup.GET("/busiPermission", h.BusiPermission)
		permGroup.POST("/busiPermission", h.BusiPermission)
		permGroup.GET("/busiResource/:flag", h.BusiResource)
		permGroup.POST("/userPerspective", h.UserPerspective)
		permGroup.POST("/menuTargetPermission", h.MenuTargetPermission)
		permGroup.POST("/busiTargetPermission", h.BusiTargetPermission)
		permGroup.POST("/saveMenuPer", h.SaveMenuPer)
		permGroup.POST("/saveBusiPer", h.SaveBusiPer)
		permGroup.POST("/saveMenuTargetPer", h.SaveMenuTargetPer)
		permGroup.POST("/saveBusiTargetPer", h.SaveBusiTargetPer)
	}
}
