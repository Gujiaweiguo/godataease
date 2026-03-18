package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/domain/visualization"
	pkgauth "dataease/backend/internal/pkg/auth"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

func TestResolveMenuTitle(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		locale string
		want   string
	}{
		{name: "exact route mapping zh", key: "menu", locale: localeZhCN, want: "菜单管理"},
		{name: "exact route mapping en", key: "menu", locale: localeEn, want: "Menu Management"},
		{name: "dotted key leaf mapping tw", key: "common.about", locale: localeTw, want: "關於"},
		{name: "dotted key namespace mapping en", key: "data_export.export_center", locale: localeEn, want: "Data Export Center"},
		{name: "unknown key falls back", key: "unknown.key", locale: localeZhCN, want: "unknown.key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveMenuTitle(tt.key, tt.locale); got != tt.want {
				t.Fatalf("resolveMenuTitle(%q, %q) = %q, want %q", tt.key, tt.locale, got, tt.want)
			}
		})
	}
}

func TestDisplayTitle(t *testing.T) {
	tests := []struct {
		name   string
		menu   *menu.MenuVO
		locale string
		want   string
	}{
		{
			name:   "prefers meta title when present",
			menu:   &menu.MenuVO{Meta: &menu.MenuMeta{Title: "commons.language"}, Name: "menu"},
			locale: localeEn,
			want:   "Language",
		},
		{
			name:   "falls back to name when meta missing",
			menu:   &menu.MenuVO{Name: "user.change_password"},
			locale: localeTw,
			want:   "修改密碼",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayTitle(tt.menu, tt.locale); got != tt.want {
				t.Fatalf("displayTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRequestLocale(t *testing.T) {
	tests := []struct {
		name         string
		header       string
		userLanguage string
		want         string
	}{
		{name: "defaults to zh cn", header: "", want: localeZhCN},
		{name: "maps english", header: "en-US", want: localeEn},
		{name: "maps traditional chinese", header: "zh-TW,zh;q=0.9", want: localeTw},
		{name: "maps simplified chinese", header: "zh-CN", want: localeZhCN},
		{name: "falls back to stored user language", header: "", userLanguage: "en-US", want: localeEn},
		{name: "unsupported header falls back to user language", header: "fr-FR", userLanguage: "zh-TW", want: localeTw},
		{name: "uses first supported locale from header list", header: "fr-FR,en-US;q=0.8", want: localeEn},
		{name: "uses first supported locale before user fallback", header: "de-DE,zh-TW;q=0.9", userLanguage: "en-US", want: localeTw},
		{name: "unsupported input falls back to default", header: "fr-FR", userLanguage: "de-DE", want: localeZhCN},
		{name: "normalizes stored tw alias", header: "", userLanguage: "tw", want: localeTw},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set("Accept-Language", tt.header)
			}
			c.Request = req
			if tt.userLanguage != "" {
				c.Set("user_id", uint64(7))
			}

			loader := userByIDLoader(nil)
			if tt.userLanguage != "" {
				userLanguage := tt.userLanguage
				loader = func(userID int64) (*user.SysUser, error) {
					return &user.SysUser{UserID: userID, Language: &userLanguage}, nil
				}
			}

			if got := requestLocale(c, loader); got != tt.want {
				t.Fatalf("requestLocale() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFrontendCompatHandler_LocalizedEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuTree := []*menu.MenuVO{
		{
			Path: "/mine",
			Meta: &menu.MenuMeta{Title: "commons.mine", Icon: "mine"},
			Children: []*menu.MenuVO{{
				Path: "mine/language",
				Meta: &menu.MenuMeta{Title: "commons.language", Icon: "lang"},
			}},
		},
	}

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) { return menuTree, nil },
		loadUserByID: func(userID int64) (*user.SysUser, error) {
			lang := "en-US"
			return &user.SysUser{UserID: userID, Language: &lang}, nil
		},
	}

	r := gin.New()
	r.GET("/api/roleRouter/query", func(c *gin.Context) {
		c.Set("user_id", uint64(7))
		h.GetRoleRouters(c)
	})
	r.GET("/api/auth/menuResource", func(c *gin.Context) {
		c.Set("user_id", uint64(7))
		h.GetMenuResource(c)
	})

	for _, path := range []string{"/api/roleRouter/query", "/api/auth/menuResource"} {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Accept-Language", "fr-FR")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("unexpected status for %s: %d", path, w.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response for %s: %v", path, err)
		}
		data := resp["data"].([]interface{})
		first := data[0].(map[string]interface{})
		meta := first["meta"].(map[string]interface{})
		if meta["title"] != "Mine" {
			t.Fatalf("expected localized title for %s, got %#v", path, meta["title"])
		}
		children := first["children"].([]interface{})
		childMeta := children[0].(map[string]interface{})["meta"].(map[string]interface{})
		if childMeta["title"] != "Language" {
			t.Fatalf("expected localized child title for %s, got %#v", path, childMeta["title"])
		}
		if childMeta["title"] == "commons.language" {
			t.Fatalf("expected translated child title for %s", path)
		}
	}
}

func TestFrontendCompatHandler_GetRoleRoutersUsesRoleFilteredMenus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{{Path: "/all", Meta: &menu.MenuMeta{Title: "all", Icon: "all"}}}, nil
		},
		queryMenuTreeByRoleIDs: func(roleIDs []int64) ([]*menu.MenuVO, error) {
			if len(roleIDs) != 1 || roleIDs[0] != 2 {
				t.Fatalf("unexpected role IDs: %#v", roleIDs)
			}
			return []*menu.MenuVO{{Path: "/authorized", Meta: &menu.MenuMeta{Title: "authorized", Icon: "authorized"}}}, nil
		},
		loadRoleIDsByUserID: func(userID int64) ([]int64, error) {
			if userID != 7 {
				t.Fatalf("unexpected user ID: %d", userID)
			}
			return []int64{2}, nil
		},
	}

	r := gin.New()
	r.GET("/api/roleRouter/query", func(c *gin.Context) {
		c.Set("user_id", uint64(7))
		h.GetRoleRouters(c)
	})

	req := httptest.NewRequest("GET", "/api/roleRouter/query", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	data := resp["data"].([]interface{})
	first := data[0].(map[string]interface{})
	if first["path"] != "/authorized" {
		t.Fatalf("expected role-filtered path, got %#v", first["path"])
	}
}

func TestRegisterFrontendCompatRoutes_ProtectsMenuEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{}, nil
		},
	}

	jwtInstance := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "test-secret", Expire: 3600})
	token, err := jwtInstance.GenerateToken(7, "tester", "user")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	r := gin.New()
	protected := r.Group("")
	protected.Use(middleware.Auth(jwtInstance))
	RegisterFrontendCompatRoutes(r, protected, h)

	unauthorizedReq := httptest.NewRequest("GET", "/roleRouter/query", nil)
	unauthorizedResp := httptest.NewRecorder()
	r.ServeHTTP(unauthorizedResp, unauthorizedReq)
	if unauthorizedResp.Code != 401 {
		t.Fatalf("expected 401 without token, got %d", unauthorizedResp.Code)
	}

	authorizedReq := httptest.NewRequest("GET", "/roleRouter/query", nil)
	authorizedReq.Header.Set("Authorization", "Bearer "+token)
	authorizedResp := httptest.NewRecorder()
	r.ServeHTTP(authorizedResp, authorizedReq)
	if authorizedResp.Code != 200 {
		t.Fatalf("expected 200 with token, got %d", authorizedResp.Code)
	}

	interactiveReq := httptest.NewRequest("POST", "/dataVisualization/interactiveTree", bytes.NewBufferString(`{"dashboard":{"busiFlag":"dashboard"}}`))
	interactiveReq.Header.Set("Content-Type", "application/json")
	interactiveResp := httptest.NewRecorder()
	r.ServeHTTP(interactiveResp, interactiveReq)
	if interactiveResp.Code != 401 {
		t.Fatalf("expected 401 for interactiveTree without token, got %d", interactiveResp.Code)
	}
}

func TestFrontendCompatHandler_InteractiveTreeUsesAuthorizedMenus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dashboardType := interactiveBusiFlagDashboard
	folderType := visualizationNodeTypeFolder
	panelType := visualizationNodeTypePanel
	rootID := int64(10)

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{}, nil
		},
		queryMenuTreeByRoleIDs: func(roleIDs []int64) ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{
				{Path: interactiveMenuPathPanel + "/index"},
				{Path: interactiveMenuPathDataset},
			}, nil
		},
		loadRoleIDsByUserID: func(userID int64) ([]int64, error) {
			return []int64{2}, nil
		},
		loadDatasetTree: func(keyword *string) ([]dataset.TreeNode, error) {
			return []dataset.TreeNode{{ID: 12, Name: "Dataset Folder", NodeType: visualizationNodeTypeFolder}}, nil
		},
		loadVisualizationTree: func(busiFlag string) ([]*visualization.DataVisualizationInfo, error) {
			if busiFlag != interactiveBusiFlagDashboard {
				return []*visualization.DataVisualizationInfo{}, nil
			}
			return []*visualization.DataVisualizationInfo{
				{ID: rootID, Name: "Dashboard Folder", NodeType: &folderType, Type: &dashboardType},
				{ID: 11, PID: &rootID, Name: "Revenue Dashboard", NodeType: &panelType, Type: &dashboardType},
			}, nil
		},
	}

	r := gin.New()
	r.POST("/dataVisualization/interactiveTree", func(c *gin.Context) {
		c.Set("user_id", uint64(7))
		h.InteractiveTree(c)
	})

	req := httptest.NewRequest("POST", "/dataVisualization/interactiveTree", bytes.NewBufferString(`{"dashboard":{"busiFlag":"dashboard"},"dataset":{"busiFlag":"dataset"},"datasource":{"busiFlag":"datasource"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal interactive tree response: %v", err)
	}
	data := resp["data"].(map[string]interface{})
	assertDashboardInteractiveTree(t, data)
	assertDatasetInteractiveTree(t, data)
	assertEmptyInteractiveTreeScope(t, data, interactiveBusiFlagDatasource)
}

func TestFrontendCompatHandler_InteractiveTreeReturnsRealDataVNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataVType := interactiveBusiFlagDataV
	panelType := visualizationNodeTypePanel

	h := &FrontendCompatHandler{
		queryMenuTreeByRoleIDs: func(roleIDs []int64) ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{{Path: interactiveMenuPathScreen + "/index"}}, nil
		},
		loadRoleIDsByUserID: func(userID int64) ([]int64, error) {
			return []int64{2}, nil
		},
		loadVisualizationTree: func(busiFlag string) ([]*visualization.DataVisualizationInfo, error) {
			if busiFlag != interactiveBusiFlagDataV {
				return []*visualization.DataVisualizationInfo{}, nil
			}
			return []*visualization.DataVisualizationInfo{{ID: 21, Name: "Executive Screen", NodeType: &panelType, Type: &dataVType}}, nil
		},
	}

	r := gin.New()
	r.POST("/dataVisualization/interactiveTree", func(c *gin.Context) {
		c.Set("user_id", uint64(7))
		h.InteractiveTree(c)
	})

	req := httptest.NewRequest("POST", "/dataVisualization/interactiveTree", bytes.NewBufferString(`{"dataV":{"busiFlag":"dataV"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal interactive tree response: %v", err)
	}
	nodes := resp["data"].(map[string]interface{})["dataV"].([]interface{})
	if len(nodes) != 1 {
		t.Fatalf("expected one real dataV node, got %#v", nodes)
	}
	node := nodes[0].(map[string]interface{})
	if node["id"] != "21" || node["name"] != "Executive Screen" || node["leaf"] != true {
		t.Fatalf("unexpected real dataV node contract: %#v", node)
	}
}

func TestFrontendCompatHandler_InteractiveTreeFiltersUnauthorizedVisualizationScopes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dashboardType := interactiveBusiFlagDashboard
	panelType := visualizationNodeTypePanel

	h := &FrontendCompatHandler{
		queryMenuTreeByRoleIDs: func(roleIDs []int64) ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{{Path: interactiveMenuPathPanel + "/index"}}, nil
		},
		loadRoleIDsByUserID: func(userID int64) ([]int64, error) {
			return []int64{2}, nil
		},
		loadVisualizationTree: func(busiFlag string) ([]*visualization.DataVisualizationInfo, error) {
			return []*visualization.DataVisualizationInfo{{ID: 31, Name: busiFlag + "-node", NodeType: &panelType, Type: &dashboardType}}, nil
		},
	}

	r := gin.New()
	r.POST("/dataVisualization/interactiveTree", func(c *gin.Context) {
		c.Set("user_id", uint64(7))
		h.InteractiveTree(c)
	})

	req := httptest.NewRequest("POST", "/dataVisualization/interactiveTree", bytes.NewBufferString(`{"dashboard":{"busiFlag":"dashboard"},"dataV":{"busiFlag":"dataV"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal interactive tree response: %v", err)
	}
	data := resp["data"].(map[string]interface{})
	if len(data["dashboard"].([]interface{})) != 1 {
		t.Fatalf("expected authorized dashboard nodes, got %#v", data["dashboard"])
	}
	if len(data["dataV"].([]interface{})) != 0 {
		t.Fatalf("expected unauthorized dataV scope to be empty, got %#v", data["dataV"])
	}
}

func TestFrontendCompatHandler_InteractiveTreeReturnsDatasetAndDatasourceNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTreeByRoleIDs: func(roleIDs []int64) ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{{Path: interactiveMenuPathDataset}, {Path: interactiveMenuPathDatasource}}, nil
		},
		loadRoleIDsByUserID: func(userID int64) ([]int64, error) {
			return []int64{2}, nil
		},
		loadDatasetTree: func(keyword *string) ([]dataset.TreeNode, error) {
			return []dataset.TreeNode{{
				ID:       101,
				Name:     "Dataset Folder",
				NodeType: visualizationNodeTypeFolder,
				Children: []dataset.TreeNode{{ID: 102, Name: "Orders Dataset", NodeType: interactiveBusiFlagDataset}},
			}}, nil
		},
		loadDatasourceTree: func(keyword *string) ([]*datasource.CoreDatasource, error) {
			rootID := int64(201)
			folderPID := int64(201)
			return []*datasource.CoreDatasource{
				{ID: rootID, Name: "Datasource Folder", Type: datasource.TypeFolder},
				{ID: 202, PID: &folderPID, Name: "MySQL DS", Type: "mysql"},
			}, nil
		},
	}

	r := gin.New()
	r.POST("/dataVisualization/interactiveTree", func(c *gin.Context) {
		c.Set("user_id", uint64(7))
		h.InteractiveTree(c)
	})

	req := httptest.NewRequest("POST", "/dataVisualization/interactiveTree", bytes.NewBufferString(`{"dataset":{"busiFlag":"dataset"},"datasource":{"busiFlag":"datasource"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal interactive tree response: %v", err)
	}
	data := resp["data"].(map[string]interface{})

	datasetNodes := data["dataset"].([]interface{})
	if len(datasetNodes) != 1 {
		t.Fatalf("expected one dataset root node, got %#v", datasetNodes)
	}
	datasetNode := datasetNodes[0].(map[string]interface{})
	if datasetNode["id"] != "101" || datasetNode["name"] != "Dataset Folder" || datasetNode["leaf"] != false {
		t.Fatalf("unexpected dataset node contract: %#v", datasetNode)
	}
	datasetChildren := datasetNode["children"].([]interface{})
	if len(datasetChildren) != 1 || datasetChildren[0].(map[string]interface{})["id"] != "102" {
		t.Fatalf("unexpected dataset child nodes: %#v", datasetChildren)
	}

	datasourceNodes := data["datasource"].([]interface{})
	if len(datasourceNodes) != 1 {
		t.Fatalf("expected datasource root wrapper, got %#v", datasourceNodes)
	}
	datasourceRoot := datasourceNodes[0].(map[string]interface{})
	if datasourceRoot["id"] != "0" || datasourceRoot["name"] != "root" {
		t.Fatalf("unexpected datasource root wrapper: %#v", datasourceRoot)
	}
	rootChildren := datasourceRoot["children"].([]interface{})
	if len(rootChildren) != 1 {
		t.Fatalf("expected one datasource folder child, got %#v", rootChildren)
	}
}

func TestFrontendCompatHandler_InteractiveTreeHandlesDatasetAndDatasourceLoaderErrorsDeterministically(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTreeByRoleIDs: func(roleIDs []int64) ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{{Path: interactiveMenuPathDataset}, {Path: interactiveMenuPathDatasource}}, nil
		},
		loadRoleIDsByUserID: func(userID int64) ([]int64, error) {
			return []int64{2}, nil
		},
		loadDatasetTree: func(keyword *string) ([]dataset.TreeNode, error) {
			return nil, errors.New("dataset unavailable")
		},
		loadDatasourceTree: func(keyword *string) ([]*datasource.CoreDatasource, error) {
			return nil, errors.New("datasource unavailable")
		},
	}

	r := gin.New()
	r.POST("/dataVisualization/interactiveTree", func(c *gin.Context) {
		c.Set("user_id", uint64(7))
		h.InteractiveTree(c)
	})

	req := httptest.NewRequest("POST", "/dataVisualization/interactiveTree", bytes.NewBufferString(`{"dataset":{"busiFlag":"dataset"},"datasource":{"busiFlag":"datasource"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal interactive tree response: %v", err)
	}
	data := resp["data"].(map[string]interface{})
	if len(data["dataset"].([]interface{})) != 0 {
		t.Fatalf("expected dataset loader failure to yield empty list, got %#v", data["dataset"])
	}
	if len(data["datasource"].([]interface{})) != 0 {
		t.Fatalf("expected datasource loader failure to yield empty list, got %#v", data["datasource"])
	}
}

func assertDashboardInteractiveTree(t *testing.T, data map[string]interface{}) {
	t.Helper()
	if len(data[interactiveBusiFlagDashboard].([]interface{})) != 1 {
		t.Fatalf("expected authorized dashboard tree")
	}
	dashboardNode := data[interactiveBusiFlagDashboard].([]interface{})[0].(map[string]interface{})
	if dashboardNode["id"] != "10" || dashboardNode["pid"] != "0" || dashboardNode["name"] != "Dashboard Folder" {
		t.Fatalf("unexpected dashboard node contract: %#v", dashboardNode)
	}
	if dashboardNode["leaf"] != false {
		t.Fatalf("expected dashboard node to remain non-leaf: %#v", dashboardNode)
	}
	if dashboardNode["weight"] != float64(9) || dashboardNode["extraFlag"] != float64(0) || dashboardNode["extraFlag1"] != float64(0) {
		t.Fatalf("unexpected dashboard node extras: %#v", dashboardNode)
	}
	children := dashboardNode["children"].([]interface{})
	if len(children) != 1 {
		t.Fatalf("expected one real dashboard child node, got %#v", children)
	}
	child := children[0].(map[string]interface{})
	if child["id"] != "11" || child["pid"] != "10" || child["name"] != "Revenue Dashboard" || child["leaf"] != true {
		t.Fatalf("unexpected real dashboard child node contract: %#v", child)
	}
}

func assertDatasetInteractiveTree(t *testing.T, data map[string]interface{}) {
	t.Helper()
	if len(data[interactiveBusiFlagDataset].([]interface{})) != 1 {
		t.Fatalf("expected authorized dataset tree")
	}
	datasetNode := data[interactiveBusiFlagDataset].([]interface{})[0].(map[string]interface{})
	if datasetNode["id"] != "12" || datasetNode["pid"] != "0" || datasetNode["name"] != "Dataset Folder" {
		t.Fatalf("unexpected dataset node contract: %#v", datasetNode)
	}
	if datasetNode["leaf"] != false {
		t.Fatalf("expected dataset folder node to be non-leaf: %#v", datasetNode)
	}
}

func assertEmptyInteractiveTreeScope(t *testing.T, data map[string]interface{}, busiFlag string) {
	t.Helper()
	if len(data[busiFlag].([]interface{})) != 0 {
		t.Fatalf("expected %s scope to be empty", busiFlag)
	}
}

func TestFrontendCompatHandler_MenuQueryError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) { return nil, errors.New("boom") },
	}

	r := gin.New()
	r.GET("/api/auth/menuResource", h.GetMenuResource)

	req := httptest.NewRequest("GET", "/api/auth/menuResource", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if resp["code"] != "500000" {
		t.Fatalf("expected error code, got %#v", resp["code"])
	}
}
