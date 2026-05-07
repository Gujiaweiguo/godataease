package handler

import (
	"encoding/json"
	"strconv"
	"strings"

	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type FrontendCompatHandler struct {
	menuService            *service.MenuService
	datasetService         *service.DatasetService
	datasourceService      *service.DatasourceService
	visualizationService   *service.VisualizationService
	linkageHandler         *LinkageHandler
	linkJumpHandler        *LinkJumpHandler
	loadUserByID           userByIDLoader
	loadRoleIDsByUserID    func(userID int64) ([]int64, error)
	queryMenuTree          func() ([]*menu.MenuVO, error)
	queryMenuTreeByRoleIDs func(roleIDs []int64) ([]*menu.MenuVO, error)
	loadDatasetTree        func(keyword *string) ([]dataset.TreeNode, error)
	loadDatasourceTree     func(keyword *string) ([]*datasource.CoreDatasource, error)
	loadVisualizationTree  func(busiFlag string) ([]*visualization.DataVisualizationInfo, error)
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

type interactiveRequest struct {
	BusiFlag string  `json:"busiFlag"`
	Leaf     *bool   `json:"leaf"`
	Keyword  *string `json:"keyword"`
}

const (
	interactiveBusiFlagDashboard  = "dashboard"
	interactiveBusiFlagDataV      = "dataV"
	interactiveBusiFlagDataset    = "dataset"
	interactiveBusiFlagDatasource = "datasource"
	interactiveMenuPathPanel      = "/panel"
	interactiveMenuPathScreen     = "/screen"
	interactiveMenuPathDataset    = "/data/dataset"
	interactiveMenuPathDatasource = "/data/datasource"
	interactivePanelAlias         = "panel"
	interactiveScreenAlias        = "screen"
)

func NewFrontendCompatHandler(
	menuService *service.MenuService,
	datasetService *service.DatasetService,
	datasourceService *service.DatasourceService,
	visualizationService *service.VisualizationService,
	userService *service.UserService,
	loadRoleIDsByUserID func(userID int64) ([]int64, error),
	linkageHandler *LinkageHandler,
	linkJumpHandler *LinkJumpHandler,
) *FrontendCompatHandler {
	h := &FrontendCompatHandler{
		menuService:          menuService,
		datasetService:       datasetService,
		datasourceService:    datasourceService,
		visualizationService: visualizationService,
		linkageHandler:       linkageHandler,
		linkJumpHandler:      linkJumpHandler,
		loadUserByID:         nil,
		loadRoleIDsByUserID:  loadRoleIDsByUserID,
	}
	if userService != nil {
		h.loadUserByID = userService.GetUserByID
	}
	if menuService != nil {
		h.queryMenuTree = menuService.Query
		h.queryMenuTreeByRoleIDs = menuService.QueryByRoleIDs
	}
	if visualizationService != nil {
		h.loadVisualizationTree = visualizationService.InteractiveTree
	}
	if datasetService != nil {
		h.loadDatasetTree = func(keyword *string) ([]dataset.TreeNode, error) {
			return datasetService.Tree(&dataset.TreeRequest{Keyword: keyword})
		}
	}
	if datasourceService != nil {
		h.loadDatasourceTree = func(keyword *string) ([]*datasource.CoreDatasource, error) {
			return datasourceService.Tree(&datasource.ListRequest{Keyword: keyword})
		}
	}
	return h
}

func (h *FrontendCompatHandler) GetRoleRouters(c *gin.Context) {
	defer recoverServicePanic(c)
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
	defer recoverServicePanic(c)
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
	if h.queryMenuTree == nil && h.queryMenuTreeByRoleIDs == nil {
		return []*menu.MenuVO{}, nil
	}

	userID := int64(middleware.GetUserID(c))
	if userID <= 0 || h.loadRoleIDsByUserID == nil || h.queryMenuTreeByRoleIDs == nil {
		if h.queryMenuTree == nil {
			return []*menu.MenuVO{}, nil
		}
		return h.queryMenuTree()
	}

	roleIDs, err := h.loadRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}

	return h.queryMenuTreeByRoleIDs(roleIDs)
}

func (h *FrontendCompatHandler) InteractiveTree(c *gin.Context) {
	defer recoverServicePanic(c)
	var requestMap map[string]interface{}
	if err := c.ShouldBindBodyWith(&requestMap, binding.JSON); err != nil {
		requestMap = make(map[string]interface{})
	}

	result := make(map[string]interface{})
	menus, err := h.loadRuntimeMenus(c)
	if err != nil {
		response.Error(c, "500000", "failed to load interactive tree")
		return
	}
	authorized := collectAuthorizedBusiFlags(menus)
	for busiFlag, rawReq := range requestMap {
		req := interactiveRequest{BusiFlag: busiFlag}
		if rawReq != nil {
			payload, _ := json.Marshal(rawReq)
			_ = json.Unmarshal(payload, &req)
			if strings.TrimSpace(req.BusiFlag) == "" {
				req.BusiFlag = busiFlag
			}
		}

		normalizedFlag := normalizeInteractiveBusiFlag(req.BusiFlag)
		if isVisualizationInteractiveBusiFlag(normalizedFlag) {
			result[busiFlag] = h.buildVisualizationInteractiveTree(normalizedFlag, req.Leaf, authorized[normalizedFlag])
			continue
		}
		if normalizedFlag == interactiveBusiFlagDataset {
			result[busiFlag] = h.buildDatasetInteractiveTree(req.Keyword, authorized[normalizedFlag])
			continue
		}
		if normalizedFlag == interactiveBusiFlagDatasource {
			result[busiFlag] = h.buildDatasourceInteractiveTree(req.Keyword, authorized[normalizedFlag])
			continue
		}

		result[busiFlag] = buildInteractiveTreeResponse(busiFlag, authorized[busiFlag])
	}
	response.Success(c, result)
}

func (h *FrontendCompatHandler) buildDatasetInteractiveTree(keyword *string, authorized bool) []interactiveTreeNode {
	if !authorized || h.loadDatasetTree == nil {
		return []interactiveTreeNode{}
	}
	items, err := h.loadDatasetTree(keyword)
	if err != nil {
		return []interactiveTreeNode{}
	}
	return convertDatasetTreeNodes(items)
}

func (h *FrontendCompatHandler) buildDatasourceInteractiveTree(keyword *string, authorized bool) []datasourceTreeNode {
	if !authorized || h.loadDatasourceTree == nil {
		return []datasourceTreeNode{}
	}
	items, err := h.loadDatasourceTree(keyword)
	if err != nil {
		return []datasourceTreeNode{}
	}
	return buildDatasourceTreeResponse(items)
}

func (h *FrontendCompatHandler) buildVisualizationInteractiveTree(
	busiFlag string,
	leaf *bool,
	authorized bool,
) []treeNode {
	if !authorized || h.loadVisualizationTree == nil {
		return []treeNode{}
	}
	items, err := h.loadVisualizationTree(busiFlag)
	if err != nil {
		return []treeNode{}
	}
	nodes, err := buildVisualizationTree(items, leaf)
	if err != nil {
		return []treeNode{}
	}
	return nodes
}

func (h *FrontendCompatHandler) FindTargetUrl(c *gin.Context) {
	defer recoverServicePanic(c)
	result := make(map[string]string)
	response.Success(c, result)
}

func (h *FrontendCompatHandler) QueryStore(c *gin.Context) {
	defer recoverServicePanic(c)
	response.Success(c, []map[string]any{})
}

func (h *FrontendCompatHandler) GetXpackContent(c *gin.Context) {
	defer recoverServicePanic(c)
	c.JSON(501, gin.H{
		"code": "501000",
		"msg":  "Not Implemented: xpackComponent requires enterprise license",
	})
}

func (h *FrontendCompatHandler) GetXpackPluginStaticInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	c.JSON(501, gin.H{
		"code": "501000",
		"msg":  "Not Implemented: xpackComponent requires enterprise license",
	})
}

func (h *FrontendCompatHandler) GetWebSocketInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	c.JSON(200, gin.H{
		"websocket":     false,
		"origins":       []string{"*:*"},
		"cookie_needed": false,
		"entropy":       1,
	})
}

// StubEmptyData returns an empty success response for stubbed frontend compat routes.
func (h *FrontendCompatHandler) StubEmptyData(c *gin.Context) {
	defer recoverServicePanic(c)
	response.Success(c, map[string]interface{}{})
}

// RegisterFrontendCompatRoutes registers frontend-facing compatibility routes.
//
// P4 Legacy Compat Contract classification (C1 policy lock):
//
//	PERMANENT SHIM (保留 shim):
//	  - /de2api/* routes: external system/plugin dependency, cannot migrate frontend
//	  - /api/xpackComponent/*: enterprise plugin dependency, returns 501
//	  - /api/aiBase/findTargetUrl: external AI integration dependency
//	  - /api/websocket/info: protocol negotiation endpoint
//
//	DUAL-SUPPORT TRANSITION (C1 keep, C3 migrate):
//	  - /roleRouter/query, /api/roleRouter/query: frontend router resolution
//	  - /auth/menuResource, /api/auth/menuResource: menu resource loading
//	  - /dataVisualization/interactiveTree, /api/dataVisualization/interactiveTree: interactive tree
//	  - /store/query, /api/store/query: store compatibility
//
//	NOTE: Non-/api/ prefixed versions (e.g. /roleRouter/query) are aliases for /api/* routes.
//	Both exist for backward compatibility with legacy frontend builds.
//
//nolint:dupl // route registration pattern is intentionally similar
func RegisterFrontendCompatRoutes(engine *gin.Engine, protected gin.IRoutes, h *FrontendCompatHandler) {
	// Dual-support transition: /de2api/* prefix (aliased for /api/*)
	protected.GET("/de2api/roleRouter/query", h.GetRoleRouters)
	protected.GET("/de2api/auth/menuResource", h.GetMenuResource)
	protected.POST("/de2api/dataVisualization/interactiveTree", h.InteractiveTree)
	protected.POST("/de2api/store/query", h.QueryStore)

	// Permanent shim: xpack/plugin/AI integration
	engine.GET("/aiBase/findTargetUrl", h.FindTargetUrl)
	engine.GET("/xpackComponent/content/:id", h.GetXpackContent)
	engine.GET("/xpackComponent/contentPlugin/:id", h.GetXpackContent)
	engine.GET("/xpackComponent/pluginStaticInfo/:id", h.GetXpackPluginStaticInfo)
	engine.GET("/websocket/info", h.GetWebSocketInfo)

	// Dual-support transition: /api/* prefix (primary compat form)
	protected.GET("/api/roleRouter/query", h.GetRoleRouters)
	protected.GET("/api/auth/menuResource", h.GetMenuResource)
	protected.POST("/api/dataVisualization/interactiveTree", h.InteractiveTree)
	protected.POST("/api/store/query", h.QueryStore)
	engine.GET("/api/aiBase/findTargetUrl", h.FindTargetUrl)
	engine.GET("/api/xpackComponent/content/:id", h.GetXpackContent)
	engine.GET("/api/xpackComponent/contentPlugin/:id", h.GetXpackContent)
	engine.GET("/api/xpackComponent/pluginStaticInfo/:id", h.GetXpackPluginStaticInfo)
	engine.GET("/api/websocket/info", h.GetWebSocketInfo)

	// Legacy prefix aliases (non-/api/) — kept for backward compat, will be removed after C3
	protected.GET("/roleRouter/query", h.GetRoleRouters)
	protected.GET("/auth/menuResource", h.GetMenuResource)
	protected.POST("/dataVisualization/interactiveTree", h.InteractiveTree)

	// Stub routes for dashboard canvas linkage/jump info (returns empty data)
	// Note: Root-level /linkage/... and /linkJump/... are already registered by
	// RegisterLinkageRoutes and RegisterLinkJumpRoutes in registerRootRoutes().
	// Only /api/ prefixed aliases are needed here for frontend compat.
	protected.GET("/api/linkage/getVisualizationAllLinkageInfo/:dvId/:resourceTable", h.linkageHandler.GetVisualizationAllLinkageInfo)
	protected.GET("/api/linkJump/queryVisualizationJumpInfo/:dvId/:resourceTable", h.linkJumpHandler.QueryVisualizationJumpInfo)
}

// busiFlagPathRule maps a normalized menu path segment to the busiFlag it authorizes.
type busiFlagPathRule struct {
	segment  string
	busiFlag string
}

var busiFlagPathRules = []busiFlagPathRule{
	{segment: "panel", busiFlag: interactiveBusiFlagDashboard},
	{segment: "screen", busiFlag: interactiveBusiFlagDataV},
	{segment: "dataset", busiFlag: interactiveBusiFlagDataset},
	{segment: "datasource", busiFlag: interactiveBusiFlagDatasource},
}

func collectAuthorizedBusiFlags(menus []*menu.MenuVO) map[string]bool {
	authorized := map[string]bool{
		interactiveBusiFlagDashboard:  false,
		interactiveBusiFlagDataV:      false,
		interactiveBusiFlagDataset:    false,
		interactiveBusiFlagDatasource: false,
	}
	var walk func(nodes []*menu.MenuVO)
	walk = func(nodes []*menu.MenuVO) {
		for _, node := range nodes {
			if node == nil {
				continue
			}
			p := strings.TrimPrefix(strings.TrimSpace(node.Path), "/")
			for _, rule := range busiFlagPathRules {
				if p == rule.segment || strings.HasPrefix(p, rule.segment+"/") || strings.HasSuffix(p, "/"+rule.segment) {
					authorized[rule.busiFlag] = true
				}
			}
			if len(node.Children) > 0 {
				walk(node.Children)
			}
		}
	}
	walk(menus)
	return authorized
}

func normalizeInteractiveBusiFlag(busiFlag string) string {
	flag := strings.TrimSpace(busiFlag)
	switch flag {
	case interactivePanelAlias:
		return interactiveBusiFlagDashboard
	case interactiveScreenAlias:
		return interactiveBusiFlagDataV
	default:
		return flag
	}
}

func isVisualizationInteractiveBusiFlag(busiFlag string) bool {
	return busiFlag == interactiveBusiFlagDashboard || busiFlag == interactiveBusiFlagDataV
}

func convertDatasetTreeNodes(items []dataset.TreeNode) []interactiveTreeNode {
	result := make([]interactiveTreeNode, 0, len(items))
	for _, item := range items {
		leaf := !strings.EqualFold(strings.TrimSpace(item.NodeType), "folder")
		node := interactiveTreeNode{
			ID:         strconv.FormatInt(item.ID, 10),
			PID:        "0",
			Name:       item.Name,
			Leaf:       leaf,
			Weight:     9,
			ExtraFlag:  0,
			ExtraFlag1: 0,
			Children:   convertDatasetTreeNodes(item.Children),
		}
		for idx := range node.Children {
			node.Children[idx].PID = node.ID
		}
		result = append(result, node)
	}
	return result
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
