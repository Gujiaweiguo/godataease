package middleware

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MenuAuthMiddleware struct {
	roleMenuService *service.RoleMenuService
	menuService     *service.MenuService
}

func NewMenuAuthMiddleware(roleMenuService *service.RoleMenuService, menuService *service.MenuService) *MenuAuthMiddleware {
	return &MenuAuthMiddleware{
		roleMenuService: roleMenuService,
		menuService:     menuService,
	}
}

// CheckMenuAccess 检查用户是否有访问指定菜单的权限
// 用法: router.GET("/some-menu-path", menuAuthMiddleware.CheckMenuAccess(), handler)
func (m *MenuAuthMiddleware) CheckMenuAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色 IDs
		roleIDs, exists := c.Get("role_ids")
		if !exists {
			response.Forbidden(c, "No role assigned")
			c.Abort()
			return
		}

		roleIDSlice, ok := roleIDs.([]int64)
		if !ok || len(roleIDSlice) == 0 {
			response.Forbidden(c, "No valid role")
			c.Abort()
			return
		}

		// 检查是否是管理员 (role_id = 1)
		for _, id := range roleIDSlice {
			if id == 1 {
				c.Next()
				return
			}
		}

		// 从请求路径或参数中获取菜单 ID
		menuIDStr := c.Param("menuId")
		if menuIDStr == "" {
			menuIDStr = c.Query("menuId")
		}

		if menuIDStr == "" {
			// 没有指定菜单 ID，允许通过（由具体 handler 处理）
			c.Next()
			return
		}

		menuID, err := strconv.ParseInt(menuIDStr, 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid menu ID")
			c.Abort()
			return
		}

		// 检查角色是否有该菜单的授权
		authorized, err := m.roleMenuService.IsMenuAuthorized(roleIDSlice, menuID)
		if err != nil {
			response.Error(c, "500000", "Failed to check menu authorization")
			c.Abort()
			return
		}

		if !authorized {
			response.Forbidden(c, "Menu access denied")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireMenuAuth 返回 403 如果用户没有菜单授权
func (m *MenuAuthMiddleware) RequireMenuAuth(menuPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIDs, exists := c.Get("role_ids")
		if !exists {
			response.Forbidden(c, "No role assigned")
			c.Abort()
			return
		}

		roleIDSlice, ok := roleIDs.([]int64)
		if !ok || len(roleIDSlice) == 0 {
			response.Forbidden(c, "No valid role")
			c.Abort()
			return
		}

		// 管理员绕过
		for _, id := range roleIDSlice {
			if id == 1 {
				c.Next()
				return
			}
		}

		// 通过路径查找菜单并检查授权
		// 这里需要根据实际菜单路径检查逻辑实现
		// 简化版本：直接返回 403
		response.Forbidden(c, "Menu access denied: "+menuPath)
		c.Abort()
	}
}
