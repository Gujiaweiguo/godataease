package handler

import (
	"fmt"
	"strconv"
	"strings"

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

type targetPermissionTarget struct {
	TargetType  string                  `json:"targetType"`
	SourceType  string                  `json:"sourceType"`
	TargetID    int64                   `json:"targetId"`
	SourceID    int64                   `json:"sourceId"`
	PermIDs     []int64                 `json:"permIds"`
	Permissions []targetPermissionEntry `json:"permissions"`
}

type targetPermissionEntry struct {
	ID int64 `json:"id"`
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
	if roleIDStr := c.Query("roleId"); roleIDStr != "" {
		if roleID, err := strconv.ParseInt(roleIDStr, 10, 64); err == nil {
			return roleID, nil
		}
	}
	return 0, nil
}

func (h *PermissionCompatHandler) MenuPermission(c *gin.Context) {
	menus, err := h.menuService.Query()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	roleID, err := parseRoleIDQuery(c)
	if err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

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
	roleID, err := parseRoleIDQuery(c)
	if err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

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
	if h.roleService != nil {
		if err := h.roleService.ValidatePermissionInheritance(req.RoleID, req.PermIDs); err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
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
	flag := strings.TrimSpace(c.Param("flag"))
	list, err := h.permService.ListPerms(&permission.PermQueryRequest{Current: 1, Size: 1000})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	if flag != "" && flag != "1" && !strings.EqualFold(flag, "all") {
		perms, ok := list.List.([]*permission.SysPerm)
		if !ok {
			response.Error(c, "500000", "Failed: unexpected permission payload")
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
	var req targetPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if req.RoleID <= 0 {
		response.Error(c, "500000", "Invalid request: roleId is required")
		return
	}
	if h.menuService == nil || h.roleMenuService == nil {
		response.Error(c, "500000", "menu target permission compatibility is not configured")
		return
	}

	menus, err := h.menuService.Query()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	auth, err := h.roleMenuService.GetRoleMenuAuth(req.RoleID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"menuTree": menus,
		"menuIds":  auth.MenuIDs,
	})
}

func (h *PermissionCompatHandler) BusiTargetPermission(c *gin.Context) {
	var req businessTargetPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	resourceID := req.ID
	if resourceID == 0 {
		resourceID = req.RoleID
	}
	if resourceID <= 0 || req.Flag == "" {
		response.Error(c, "500000", "Invalid request: id and flag are required")
		return
	}

	targetPermissions, err := h.resourcePermService.GetResourcePerspective(resourceID, req.Flag)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, targetPermissions)
}

func (h *PermissionCompatHandler) UserPerspective(c *gin.Context) {
	var req userPerspectiveCompatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	var (
		userPermissions []*permission.UserResourcePermVO
		err             error
	)
	if req.ResourceID > 0 {
		resourcePermissions, resourceErr := h.resourcePermService.GetResourcePerspective(req.ResourceID, req.ResourceType)
		if resourceErr != nil {
			response.Error(c, "500000", "Failed: "+resourceErr.Error())
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
		userPermissions, err = h.resourcePermService.GetUserPerspective(req.UserID, req.ResourceType)
	}
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, userPermissions)
}

func (h *PermissionCompatHandler) SaveMenuTargetPer(c *gin.Context) {
	var req targetPermissionSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if h.roleMenuService == nil {
		response.Error(c, "500000", "save menu target permission compatibility is not configured")
		return
	}

	roleID := req.RoleID
	menuIDs := make([]int64, 0)
	for _, target := range req.TargetPerms {
		targetType := normalizeTargetType(target.TargetType, target.SourceType)
		if targetType != permission.AuthTargetTypeRole {
			response.Error(c, "500000", "Invalid request: only role targets are supported")
			return
		}
		if roleID == 0 {
			roleID = normalizeTargetID(target.TargetID, target.SourceID)
		}
		menuIDs = append(menuIDs, extractTargetPermIDs(target)...)
	}
	if roleID <= 0 {
		response.Error(c, "500000", "Invalid request: roleId is required")
		return
	}

	if err := h.roleMenuService.SaveRoleMenuAuth(&service.SaveRoleMenuRequest{RoleID: roleID, MenuIDs: uniqueInt64(menuIDs)}); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *PermissionCompatHandler) SaveBusiTargetPer(c *gin.Context) {
	var req targetPermissionSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	resourceID := req.ID
	if resourceID == 0 {
		resourceID = req.RoleID
	}
	if resourceID <= 0 || req.Flag == "" {
		response.Error(c, "500000", "Invalid request: id and flag are required")
		return
	}

	resourcePermIDs := make([]int64, 0)
	preservedDirectPermIDs, err := h.collectDirectResourcePermIDs(resourceID, req.Flag)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	resourcePermIDs = append(resourcePermIDs, preservedDirectPermIDs...)
	for _, target := range req.TargetPerms {
		matchedPermIDs, err := h.collectMatchedTargetPermIDs(target, req.Flag)
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		resourcePermIDs = append(resourcePermIDs, matchedPermIDs...)

		if err := h.saveRolePermsForResourceType(normalizeTargetID(target.TargetID, target.SourceID), extractTargetPermIDs(target), req.Flag); err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
	}

	if err := h.resourcePermService.ReplaceResourcePermissions(resourceID, req.Flag, uniqueInt64(resourcePermIDs)); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *PermissionCompatHandler) collectMatchedTargetPermIDs(target targetPermissionTarget, resourceType string) ([]int64, error) {
	targetType := normalizeTargetType(target.TargetType, target.SourceType)
	targetID := normalizeTargetID(target.TargetID, target.SourceID)
	if targetType != permission.AuthTargetTypeRole || targetID <= 0 {
		return nil, fmt.Errorf("only role targets are supported in the current resource-perspective save slice")
	}

	permIDs := extractTargetPermIDs(target)
	matchedPermIDs := make([]int64, 0, len(permIDs))
	for _, permID := range permIDs {
		matches, err := h.permissionMatchesResourceType(permID, resourceType)
		if err != nil {
			return nil, err
		}
		if matches {
			matchedPermIDs = append(matchedPermIDs, permID)
		}
	}
	return matchedPermIDs, nil
}

func extractTargetPermIDs(target targetPermissionTarget) []int64 {
	if len(target.PermIDs) > 0 || len(target.Permissions) == 0 {
		return target.PermIDs
	}

	permIDs := make([]int64, 0, len(target.Permissions))
	for _, item := range target.Permissions {
		if item.ID > 0 {
			permIDs = append(permIDs, item.ID)
		}
	}
	return permIDs
}

func (h *PermissionCompatHandler) collectDirectResourcePermIDs(resourceID int64, resourceType string) ([]int64, error) {
	items, err := h.resourcePermService.GetResourcePerspective(resourceID, resourceType)
	if err != nil {
		return nil, err
	}
	permIDs := make([]int64, 0)
	seen := make(map[int64]struct{})
	for _, item := range items {
		if item == nil || item.SourceType != "direct" || strings.TrimSpace(item.PermKey) == "" {
			continue
		}
		perm, resolveErr := h.resourcePermService.ResolvePermission(resourceType, item.PermKey)
		if resolveErr != nil || perm == nil || perm.PermID <= 0 {
			continue
		}
		if _, ok := seen[perm.PermID]; ok {
			continue
		}
		seen[perm.PermID] = struct{}{}
		permIDs = append(permIDs, perm.PermID)
	}
	return permIDs, nil
}

func uniqueInt64(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
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

func (h *PermissionCompatHandler) saveRolePermsForResourceType(roleID int64, targetPermIDs []int64, resourceType string) error {
	target := make(map[int64]struct{}, len(targetPermIDs))
	for _, id := range targetPermIDs {
		if id <= 0 {
			continue
		}
		matches, err := h.permissionMatchesResourceType(id, resourceType)
		if err != nil {
			return err
		}
		if !matches {
			return fmt.Errorf("permission %d does not belong to resource type %s", id, resourceType)
		}
		target[id] = struct{}{}
	}

	currentPermIDs, err := h.resourcePermService.GetRolePermissionIDs(roleID)
	if err != nil {
		return err
	}

	current := make(map[int64]struct{}, len(currentPermIDs))
	for _, id := range currentPermIDs {
		matches, matchErr := h.permissionMatchesResourceType(id, resourceType)
		if matchErr != nil {
			return matchErr
		}
		if matches {
			current[id] = struct{}{}
		}
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

func (h *PermissionCompatHandler) permissionMatchesResourceType(permID int64, resourceType string) (bool, error) {
	if h.resourcePermService == nil {
		return false, fmt.Errorf("resource permission service is unavailable")
	}

	perm, err := h.resourcePermService.GetPermissionByID(permID)
	if err != nil {
		return false, err
	}
	if perm == nil {
		return false, fmt.Errorf("permission %d not found", permID)
	}

	return strings.HasPrefix(perm.PermKey, resourceType+":"), nil
}

func normalizeTargetType(targetType, sourceType string) string {
	if targetType != "" {
		return targetType
	}
	return sourceType
}

func normalizeTargetID(targetID, sourceID int64) int64 {
	if targetID > 0 {
		return targetID
	}
	return sourceID
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
