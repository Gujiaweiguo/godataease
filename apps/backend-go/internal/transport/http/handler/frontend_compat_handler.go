package handler

import (
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type FrontendCompatHandler struct{}

func NewFrontendCompatHandler() *FrontendCompatHandler {
	return &FrontendCompatHandler{}
}

func (h *FrontendCompatHandler) GetRoleRouters(c *gin.Context) {
	routers := []map[string]interface{}{
		{
			"path":      "/system",
			"name":      "system",
			"component": "Layout",
			"redirect":  "/system/user",
			"top":       true,
			"inLayout":  true,
			"meta": map[string]interface{}{
				"title": "系统管理",
			},
			"children": []map[string]interface{}{
				{
					"path":      "user",
					"name":      "system-user",
					"component": "system/user",
					"meta": map[string]interface{}{
						"title": "用户管理",
						"icon":  "peoples",
					},
				},
				{
					"path":      "role",
					"name":      "system-role",
					"component": "system/role",
					"meta": map[string]interface{}{
						"title": "角色管理",
						"icon":  "auth",
					},
				},
				{
					"path":      "org",
					"name":      "system-org",
					"component": "system/org",
					"meta": map[string]interface{}{
						"title": "组织管理",
						"icon":  "org",
					},
				},
				{
					"path":      "permission",
					"name":      "system-permission",
					"component": "system/permission",
					"meta": map[string]interface{}{
						"title": "权限管理",
						"icon":  "icon_security",
					},
				},
			},
		},
	}

	response.Success(c, routers)
}

func (h *FrontendCompatHandler) GetMenuResource(c *gin.Context) {
	menuTree := []map[string]interface{}{
		{
			"path": "user",
			"meta": map[string]interface{}{
				"title": "用户管理",
				"icon":  "peoples",
			},
		},
		{
			"path": "role",
			"meta": map[string]interface{}{
				"title": "角色管理",
				"icon":  "auth",
			},
		},
		{
			"path": "org",
			"meta": map[string]interface{}{
				"title": "组织管理",
				"icon":  "org",
			},
		},
		{
			"path": "permission",
			"meta": map[string]interface{}{
				"title": "权限管理",
				"icon":  "icon_security",
			},
		},
	}

	response.Success(c, menuTree)
}

func (h *FrontendCompatHandler) InteractiveTree(c *gin.Context) {
	var requestMap map[string]interface{}
	if err := c.ShouldBindJSON(&requestMap); err != nil {
		requestMap = make(map[string]interface{})
	}

	result := make(map[string]interface{})
	response.Success(c, result)
}

func (h *FrontendCompatHandler) FindTargetUrl(c *gin.Context) {
	result := make(map[string]string)
	response.Success(c, result)
}

func (h *FrontendCompatHandler) GetXpackContent(c *gin.Context) {
	c.JSON(501, gin.H{
		"code": "501000",
		"msg":  "Not Implemented: xpackComponent requires enterprise license",
	})
}

func (h *FrontendCompatHandler) GetXpackPluginStaticInfo(c *gin.Context) {
	c.JSON(501, gin.H{
		"code": "501000",
		"msg":  "Not Implemented: xpackComponent requires enterprise license",
	})
}

func (h *FrontendCompatHandler) GetWebSocketInfo(c *gin.Context) {
	c.JSON(501, gin.H{
		"code": "501000",
		"msg":  "Not Implemented: WebSocket endpoint",
	})
}

func RegisterFrontendCompatRoutes(engine *gin.Engine, h *FrontendCompatHandler) {
	engine.GET("/roleRouter/query", h.GetRoleRouters)
	engine.GET("/auth/menuResource", h.GetMenuResource)
	engine.POST("/dataVisualization/interactiveTree", h.InteractiveTree)
	engine.GET("/aiBase/findTargetUrl", h.FindTargetUrl)
	engine.GET("/xpackComponent/content/:id", h.GetXpackContent)
	engine.GET("/xpackComponent/pluginStaticInfo/:id", h.GetXpackPluginStaticInfo)
	engine.GET("/websocket/info", h.GetWebSocketInfo)

	engine.GET("/api/roleRouter/query", h.GetRoleRouters)
	engine.GET("/api/auth/menuResource", h.GetMenuResource)
	engine.POST("/api/dataVisualization/interactiveTree", h.InteractiveTree)
	engine.GET("/api/aiBase/findTargetUrl", h.FindTargetUrl)
	engine.GET("/api/xpackComponent/content/:id", h.GetXpackContent)
	engine.GET("/api/xpackComponent/pluginStaticInfo/:id", h.GetXpackPluginStaticInfo)
	engine.GET("/api/websocket/info", h.GetWebSocketInfo)
}
