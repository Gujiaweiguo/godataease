package handler

import (
	"strings"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

type FrontendCompatHandler struct {
	menuService            *service.MenuService
	loadUserByID           userByIDLoader
	loadRoleIDsByUserID    func(userID int64) ([]int64, error)
	queryMenuTree          func() ([]*menu.MenuVO, error)
	queryMenuTreeByRoleIDs func(roleIDs []int64) ([]*menu.MenuVO, error)
}

type interactiveTreeNode struct {
	ID         string                `json:"id"`
	PID        string                `json:"pid"`
	Name       string                `json:"name"`
	Leaf       bool                  `json:"leaf"`
	Weight     int                   `json:"weight"`
	ExtraFlag  int                   `json:"extraFlag"`
	ExtraFlag1 int                   `json:"extraFlag1"`
	Children   []interactiveTreeNode `json:"children,omitempty"`
}

func NewFrontendCompatHandler(
	menuService *service.MenuService,
	userService *service.UserService,
	loadRoleIDsByUserID func(userID int64) ([]int64, error),
) *FrontendCompatHandler {
	h := &FrontendCompatHandler{
		menuService:         menuService,
		loadUserByID:        nil,
		loadRoleIDsByUserID: loadRoleIDsByUserID,
	}
	if userService != nil {
		h.loadUserByID = userService.GetUserByID
	}
	if menuService != nil {
		h.queryMenuTree = menuService.Query
		h.queryMenuTreeByRoleIDs = menuService.QueryByRoleIDs
	}
	return h
}

func (h *FrontendCompatHandler) GetRoleRouters(c *gin.Context) {
	locale := requestLocale(c, h.loadUserByID)
	menus, err := h.loadRuntimeMenus(c)
	if err != nil {
		response.Error(c, "500000", "failed to load role routers")
		return
	}

	routers := make([]map[string]interface{}, 0)
	for _, m := range menus {
		routers = append(routers, toRoleRouter(m, true, locale))
	}

	response.Success(c, routers)
}

func (h *FrontendCompatHandler) GetMenuResource(c *gin.Context) {
	locale := requestLocale(c, h.loadUserByID)
	menus, err := h.loadRuntimeMenus(c)
	if err != nil {
		response.Error(c, "500000", "failed to load menu resource")
		return
	}

	menuTree := make([]map[string]interface{}, 0)
	for _, m := range menus {
		menuTree = append(menuTree, toMenuResource(m, locale))
	}

	response.Success(c, menuTree)
}

func (h *FrontendCompatHandler) loadRuntimeMenus(c *gin.Context) ([]*menu.MenuVO, error) {
	if h.queryMenuTree == nil {
		return []*menu.MenuVO{}, nil
	}

	userID := int64(middleware.GetUserID(c))
	if userID <= 0 || h.loadRoleIDsByUserID == nil || h.queryMenuTreeByRoleIDs == nil {
		return h.queryMenuTree()
	}

	roleIDs, err := h.loadRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}

	return h.queryMenuTreeByRoleIDs(roleIDs)
}

func (h *FrontendCompatHandler) InteractiveTree(c *gin.Context) {
	var requestMap map[string]interface{}
	if err := c.ShouldBindJSON(&requestMap); err != nil {
		requestMap = make(map[string]interface{})
	}

	result := make(map[string]interface{})
	menus, err := h.loadRuntimeMenus(c)
	if err != nil {
		response.Error(c, "500000", "failed to load interactive tree")
		return
	}
	authorized := collectAuthorizedBusiFlags(menus)
	for busiFlag := range requestMap {
		result[busiFlag] = buildInteractiveTreeResponse(busiFlag, authorized[busiFlag])
	}
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

func RegisterFrontendCompatRoutes(engine *gin.Engine, protected gin.IRoutes, h *FrontendCompatHandler) {
	protected.GET("/roleRouter/query", h.GetRoleRouters)
	protected.GET("/auth/menuResource", h.GetMenuResource)
	protected.POST("/dataVisualization/interactiveTree", h.InteractiveTree)
	engine.GET("/aiBase/findTargetUrl", h.FindTargetUrl)
	engine.GET("/xpackComponent/content/:id", h.GetXpackContent)
	engine.GET("/xpackComponent/pluginStaticInfo/:id", h.GetXpackPluginStaticInfo)
	engine.GET("/websocket/info", h.GetWebSocketInfo)

	protected.GET("/api/roleRouter/query", h.GetRoleRouters)
	protected.GET("/api/auth/menuResource", h.GetMenuResource)
	protected.POST("/api/dataVisualization/interactiveTree", h.InteractiveTree)
	engine.GET("/api/aiBase/findTargetUrl", h.FindTargetUrl)
	engine.GET("/api/xpackComponent/content/:id", h.GetXpackContent)
	engine.GET("/api/xpackComponent/pluginStaticInfo/:id", h.GetXpackPluginStaticInfo)
	engine.GET("/api/websocket/info", h.GetWebSocketInfo)
}

func collectAuthorizedBusiFlags(menus []*menu.MenuVO) map[string]bool {
	authorized := map[string]bool{
		"dashboard":  false,
		"dataV":      false,
		"dataset":    false,
		"datasource": false,
	}
	var walk func(nodes []*menu.MenuVO)
	walk = func(nodes []*menu.MenuVO) {
		for _, node := range nodes {
			if node == nil {
				continue
			}
			path := strings.TrimSpace(node.Path)
			switch {
			case strings.HasPrefix(path, "/panel"):
				authorized["dashboard"] = true
			case strings.HasPrefix(path, "/screen"):
				authorized["dataV"] = true
			case strings.HasPrefix(path, "/data/dataset"):
				authorized["dataset"] = true
			case strings.HasPrefix(path, "/data/datasource"):
				authorized["datasource"] = true
			}
			if len(node.Children) > 0 {
				walk(node.Children)
			}
		}
	}
	walk(menus)
	return authorized
}

func buildInteractiveTreeResponse(busiFlag string, authorized bool) []interactiveTreeNode {
	if !authorized {
		return []interactiveTreeNode{}
	}
	return []interactiveTreeNode{
		{
			ID:         "0",
			PID:        "-1",
			Name:       busiFlag,
			Leaf:       false,
			Weight:     9,
			ExtraFlag:  1,
			ExtraFlag1: 0,
			Children:   []interactiveTreeNode{},
		},
	}
}

func toRoleRouter(m *menu.MenuVO, isRoot bool, locale string) map[string]interface{} {
	path := normalizePath(m.Path, isRoot)
	result := map[string]interface{}{
		"path":     path,
		"name":     safeName(m.Name, path),
		"hidden":   m.Hidden,
		"inLayout": m.InLayout,
		"meta": map[string]interface{}{
			"title": displayTitle(m, locale),
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
			children = append(children, toRoleRouter(child, false, locale))
		}
		result["children"] = children
	}

	return result
}

func toMenuResource(m *menu.MenuVO, locale string) map[string]interface{} {
	result := map[string]interface{}{
		"path": m.Path,
		"meta": map[string]interface{}{
			"title": displayTitle(m, locale),
			"icon":  m.Meta.Icon,
		},
	}

	if len(m.Children) > 0 {
		children := make([]map[string]interface{}, 0, len(m.Children))
		for _, child := range m.Children {
			children = append(children, toMenuResource(child, locale))
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

func displayTitle(m *menu.MenuVO, locale string) string {
	if m.Meta != nil && m.Meta.Title != "" {
		return resolveMenuTitle(m.Meta.Title, locale)
	}
	return resolveMenuTitle(m.Name, locale)
}

func resolveMenuTitle(key string, locale string) string {
	titleMap := menuTitleMapForLocale(locale)
	leafMap := menuTitleLeafMapForLocale(locale)

	if title, ok := titleMap[key]; ok {
		return title
	}

	lastDot := strings.LastIndex(key, ".")
	if lastDot >= 0 && lastDot < len(key)-1 {
		leaf := key[lastDot+1:]
		if title, ok := leafMap[leaf]; ok {
			return title
		}
	}

	return key
}

func menuTitleMapForLocale(locale string) map[string]string {
	switch locale {
	case localeEn:
		return menuTitleMapEn
	case localeTw:
		return menuTitleMapTw
	default:
		return menuTitleMapZhCN
	}
}

func menuTitleLeafMapForLocale(locale string) map[string]string {
	switch locale {
	case localeEn:
		return menuTitleLeafMapEn
	case localeTw:
		return menuTitleLeafMapTw
	default:
		return menuTitleLeafMapZhCN
	}
}

var menuTitleMapZhCN = map[string]string{
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
	"system":           "系统管理",
	"menu":             "菜单管理",
	"user":             "用户管理",
	"role":             "角色管理",
	"org":              "组织管理",
	"permission":       "权限管理",
	"audit":            "审计日志",
	"audit-dashboard":  "审计仪表板",
	"audit-settings":   "审计设置",
	"datasource-form":  "数据源表单",
	"dataset-form":     "数据集表单",
}

var menuTitleMapEn = map[string]string{
	"workbranch":       "Workbench",
	"panel":            "Dashboard",
	"screen":           "Data Screen",
	"data":             "Data Preparation",
	"dataset":          "Dataset",
	"datasource":       "Data Source",
	"sys-setting":      "System Settings",
	"template-market":  "Template Market",
	"toolbox":          "Toolbox",
	"template-setting": "Template Management",
	"msg":              "Message Center",
	"parameter":        "System Parameters",
	"font":             "Font Settings",
	"system":           "System Management",
	"menu":             "Menu Management",
	"user":             "User Management",
	"role":             "Role Management",
	"org":              "Organization Management",
	"permission":       "Permission Management",
	"audit":            "Audit Log",
	"audit-dashboard":  "Audit Dashboard",
	"audit-settings":   "Audit Settings",
	"datasource-form":  "Data Source Form",
	"dataset-form":     "Dataset Form",
}

var menuTitleMapTw = map[string]string{
	"workbranch":       "工作台",
	"panel":            "儀表板",
	"screen":           "數據大屏",
	"data":             "數據準備",
	"dataset":          "數據集",
	"datasource":       "數據源",
	"sys-setting":      "系統設置",
	"template-market":  "模板市場",
	"toolbox":          "工具箱",
	"template-setting": "模板管理",
	"msg":              "消息中心",
	"parameter":        "系統參數",
	"font":             "字體設置",
	"system":           "系統管理",
	"menu":             "菜單管理",
	"user":             "用戶管理",
	"role":             "角色管理",
	"org":              "組織管理",
	"permission":       "權限管理",
	"audit":            "審計日誌",
	"audit-dashboard":  "審計儀表板",
	"audit-settings":   "審計設置",
	"datasource-form":  "數據源表單",
	"dataset-form":     "數據集表單",
}

var menuTitleLeafMapZhCN = map[string]string{
	"about":                    "关于",
	"change_password":          "修改密码",
	"enterprise_edition_trial": "企业版试用",
	"exit_system":              "退出系统",
	"export_center":            "数据导出中心",
	"help_documentation":       "帮助文档",
	"language":                 "语言",
	"mine":                     "我的",
	"product_forum":            "产品论坛",
	"system_setting":           "系统设置",
	"technical_blog":           "技术博客",
}

var menuTitleLeafMapEn = map[string]string{
	"about":                    "About",
	"change_password":          "Change Password",
	"enterprise_edition_trial": "Enterprise Trial",
	"exit_system":              "Log Out",
	"export_center":            "Data Export Center",
	"help_documentation":       "Help Documentation",
	"language":                 "Language",
	"mine":                     "Mine",
	"product_forum":            "Product Forum",
	"system_setting":           "System Settings",
	"technical_blog":           "Technical Blog",
}

var menuTitleLeafMapTw = map[string]string{
	"about":                    "關於",
	"change_password":          "修改密碼",
	"enterprise_edition_trial": "企業版試用",
	"exit_system":              "退出系統",
	"export_center":            "資料匯出中心",
	"help_documentation":       "幫助文檔",
	"language":                 "語言",
	"mine":                     "我的",
	"product_forum":            "產品論壇",
	"system_setting":           "系統設定",
	"technical_blog":           "技術部落格",
}
