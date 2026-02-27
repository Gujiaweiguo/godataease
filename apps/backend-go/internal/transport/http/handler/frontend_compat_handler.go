package handler

import (
	"strings"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type FrontendCompatHandler struct {
	menuService *service.MenuService
}

func NewFrontendCompatHandler(menuService *service.MenuService) *FrontendCompatHandler {
	return &FrontendCompatHandler{menuService: menuService}
}

func (h *FrontendCompatHandler) GetRoleRouters(c *gin.Context) {
	routers := make([]map[string]interface{}, 0)
	if h.menuService != nil {
		menus, err := h.menuService.Query()
		if err != nil {
			response.Error(c, "500000", "failed to load role routers")
			return
		}
		for _, m := range menus {
			routers = append(routers, toRoleRouter(m, true))
		}
	}

	response.Success(c, routers)
}

func (h *FrontendCompatHandler) GetMenuResource(c *gin.Context) {
	menuTree := make([]map[string]interface{}, 0)
	if h.menuService != nil {
		menus, err := h.menuService.Query()
		if err != nil {
			response.Error(c, "500000", "failed to load menu resource")
			return
		}
		for _, m := range menus {
			menuTree = append(menuTree, toMenuResource(m))
		}
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
	c.JSON(200, gin.H{
		"websocket":     false,
		"origins":       []string{"*:*"},
		"cookie_needed": false,
		"entropy":       1,
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

func toRoleRouter(m *menu.MenuVO, isRoot bool) map[string]interface{} {
	path := normalizePath(m.Path, isRoot)
	result := map[string]interface{}{
		"path":     path,
		"name":     safeName(m.Name, path),
		"hidden":   m.Hidden,
		"inLayout": m.InLayout,
		"meta": map[string]interface{}{
			"title": displayTitle(m),
			"icon":  m.Meta.Icon,
		},
	}

	if m.Component != "" {
		result["component"] = m.Component
	}
	if m.Redirect != "" {
		result["redirect"] = m.Redirect
	}
	if m.IsPlugin {
		result["plugin"] = true
	}

	if len(m.Children) > 0 {
		children := make([]map[string]interface{}, 0, len(m.Children))
		for _, child := range m.Children {
			children = append(children, toRoleRouter(child, false))
		}
		result["children"] = children
	}

	return result
}

func toMenuResource(m *menu.MenuVO) map[string]interface{} {
	result := map[string]interface{}{
		"path": m.Path,
		"meta": map[string]interface{}{
			"title": displayTitle(m),
			"icon":  m.Meta.Icon,
		},
	}

	if len(m.Children) > 0 {
		children := make([]map[string]interface{}, 0, len(m.Children))
		for _, child := range m.Children {
			children = append(children, toMenuResource(child))
		}
		result["children"] = children
	}

	return result
}

func normalizePath(path string, isRoot bool) string {
	if isRoot {
		if strings.HasPrefix(path, "/") {
			return path
		}
		return "/" + path
	}
	return strings.TrimPrefix(path, "/")
}

func safeName(name string, path string) string {
	if name != "" {
		return name
	}
	return strings.ReplaceAll(strings.Trim(path, "/"), "/", "-")
}

func displayTitle(m *menu.MenuVO) string {
	if m.Meta != nil && m.Meta.Title != "" {
		if title, ok := menuTitleMap[m.Meta.Title]; ok {
			return title
		}
		return m.Meta.Title
	}
	if title, ok := menuTitleMap[m.Name]; ok {
		return title
	}
	return m.Name
}

var menuTitleMap = map[string]string{
	"workbranch":       "工作台",
	"panel":            "仪表板",
	"screen":           "数据大屏",
	"data":             "数据准备",
	"dataset":          "数据集",
	"datasource":       "数据源",
	"sys-setting":      "系统设置",
	"template-market":  "模板市场",
	"toolbox":          "工具箱",
	"template-setting": "模板管理",
	"msg":              "消息中心",
	"parameter":        "系统参数",
	"font":             "字体设置",
}
