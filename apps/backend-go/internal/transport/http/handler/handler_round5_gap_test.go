package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/governance"
	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func respCode(body []byte) string {
	var r struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(body, &r)
	return r.Code
}

func TestRound5_FrontendCompat_NewFrontendCompatHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("nil services produce safe defaults", func(t *testing.T) {
		h := NewFrontendCompatHandler(nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NotNil(t, h)
		assert.Nil(t, h.queryMenuTree)
		assert.Nil(t, h.loadUserByID)
	})
}

func TestRound5_FrontendCompat_GetRoleRouters_LoadRuntimeMenusFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{{Path: "/fallback", Meta: &menu.MenuMeta{Title: "fallback", Icon: "fb"}}}, nil
		},
	}

	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		h.GetRoleRouters(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestRound5_FrontendCompat_GetRoleRouters_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) {
			return nil, fmt.Errorf("menu error")
		},
	}

	r := gin.New()
	r.GET("/test", h.GetRoleRouters)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_FrontendCompat_GetMenuResource_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{{Path: "/sys", Meta: &menu.MenuMeta{Title: "system", Icon: "sys"}}}, nil
		},
	}

	r := gin.New()
	r.GET("/test", h.GetMenuResource)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestRound5_FrontendCompat_InteractiveTree_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) { return nil, nil },
	}

	r := gin.New()
	r.POST("/test", h.InteractiveTree)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_FrontendCompat_InteractiveTree_LoadRuntimeError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTreeByRoleIDs: func(roleIDs []int64) ([]*menu.MenuVO, error) {
			return nil, fmt.Errorf("role menu error")
		},
		loadRoleIDsByUserID: func(userID int64) ([]int64, error) {
			return []int64{1}, nil
		},
	}

	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		h.InteractiveTree(c)
	})

	body := bytes.NewBufferString(`{"dashboard":{"busiFlag":"dashboard"}}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_FrontendCompat_FindTargetUrl(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{}
	r := gin.New()
	r.GET("/test", h.FindTargetUrl)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.Empty(t, data)
}

func TestRound5_FrontendCompat_QueryStore(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{}
	r := gin.New()
	r.POST("/test", h.QueryStore)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_FrontendCompat_GetXpackContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{}
	r := gin.New()
	r.GET("/test/:id", h.GetXpackContent)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test/xpack-1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 501, w.Code)
}

func TestRound5_FrontendCompat_GetXpackPluginStaticInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{}
	r := gin.New()
	r.GET("/test/:id", h.GetXpackPluginStaticInfo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test/plugin-1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 501, w.Code)
}

func TestRound5_FrontendCompat_GetWebSocketInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{}
	r := gin.New()
	r.GET("/test", h.GetWebSocketInfo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["websocket"])
}

func TestRound5_FrontendCompat_StubEmptyData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{}
	r := gin.New()
	r.GET("/test", h.StubEmptyData)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_FrontendCompat_RegisterFrontendCompatRoutes_De2apiAliases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) { return []*menu.MenuVO{}, nil },
	}

	r := gin.New()
	protected := r.Group("")
	RegisterFrontendCompatRoutes(r, protected, h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/de2api/roleRouter/query", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestRound5_FrontendCompat_RegisterFrontendCompatRoutes_PublicEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{}
	r := gin.New()
	protected := r.Group("")
	RegisterFrontendCompatRoutes(r, protected, h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/websocket/info", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/api/aiBase/findTargetUrl", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
}

func TestRound5_FrontendCompat_normalizeInteractiveBusiFlag(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"panel", "dashboard"},
		{"screen", "dataV"},
		{"dashboard", "dashboard"},
		{"dataV", "dataV"},
		{"dataset", "dataset"},
		{"", ""},
		{"  panel  ", "dashboard"},
	}
	for _, tt := range tests {
		got := normalizeInteractiveBusiFlag(tt.input)
		assert.Equal(t, tt.want, got, "normalizeInteractiveBusiFlag(%q)", tt.input)
	}
}

func TestRound5_FrontendCompat_isVisualizationInteractiveBusiFlag(t *testing.T) {
	assert.True(t, isVisualizationInteractiveBusiFlag("dashboard"))
	assert.True(t, isVisualizationInteractiveBusiFlag("dataV"))
	assert.False(t, isVisualizationInteractiveBusiFlag("dataset"))
	assert.False(t, isVisualizationInteractiveBusiFlag("datasource"))
	assert.False(t, isVisualizationInteractiveBusiFlag(""))
}

func TestRound5_FrontendCompat_convertDatasetTreeNodes(t *testing.T) {
	items := []dataset.TreeNode{
		{
			ID: 1, Name: "Root Folder", NodeType: "folder",
			Children: []dataset.TreeNode{
				{ID: 2, Name: "Leaf DS", NodeType: "dataset"},
			},
		},
	}
	nodes := convertDatasetTreeNodes(items)
	require.Len(t, nodes, 1)

	root := nodes[0]
	assert.Equal(t, "1", root.ID)
	assert.Equal(t, "0", root.PID)
	assert.Equal(t, "Root Folder", root.Name)
	assert.False(t, root.Leaf)
	assert.Equal(t, 9, root.Weight)
	require.Len(t, root.Children, 1)

	child := root.Children[0]
	assert.Equal(t, "2", child.ID)
	assert.Equal(t, "1", child.PID)
	assert.True(t, child.Leaf)
}

func TestRound5_FrontendCompat_convertDatasetTreeNodes_Empty(t *testing.T) {
	nodes := convertDatasetTreeNodes(nil)
	assert.NotNil(t, nodes)
	assert.Empty(t, nodes)
}

func TestRound5_FrontendCompat_buildInteractiveTreeResponse(t *testing.T) {
	t.Run("authorized returns placeholder node", func(t *testing.T) {
		nodes := buildInteractiveTreeResponse("custom-flag", true)
		require.Len(t, nodes, 1)
		assert.Equal(t, "0", nodes[0].ID)
		assert.Equal(t, "-1", nodes[0].PID)
		assert.Equal(t, "custom-flag", nodes[0].Name)
		assert.False(t, nodes[0].Leaf)
		assert.Equal(t, 1, nodes[0].ExtraFlag)
	})

	t.Run("unauthorized returns empty", func(t *testing.T) {
		nodes := buildInteractiveTreeResponse("custom-flag", false)
		assert.Empty(t, nodes)
	})
}

func TestRound5_FrontendCompat_InteractiveTree_NonStandardBusiFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{}, nil
		},
	}

	r := gin.New()
	r.POST("/test", h.InteractiveTree)

	body := bytes.NewBufferString(`{"customFlag":{"busiFlag":"customFlag"}}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
	// Non-standard busi flag → unauthorized → empty
	data := resp["data"].(map[string]interface{})
	customNodes := data["customFlag"].([]interface{})
	assert.Empty(t, customNodes)
}

func setupRound5RoleRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &org.SysOrg{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}))

	repo := repository.NewRoleRepository(db)
	orgRepo := repository.NewOrgRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	require.NoError(t, orgRepo.Create(&org.SysOrg{OrgID: 1, OrgName: "Default Org", ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}))
	systemType := "system"
	now := time.Unix(300, 0)
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "Admin", RoleCode: "admin", RoleType: &systemType, Status: role.StatusEnabled, CreateTime: &now}))
	require.NoError(t, userRepo.Create(&user.SysUser{UserID: 50, Username: "testuser", NickName: "Test User", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}))
	require.NoError(t, userRepo.Create(&user.SysUser{UserID: 51, Username: "extuser", NickName: "Ext User", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}))
	require.NoError(t, db.Create(&user.SysUserRole{UserID: 1, RoleID: 1, OrgID: 1}).Error)

	policySvc := service.NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), nil)
	svc := service.NewRoleService(repo, userRepo, userRoleRepo, orgRepo, policySvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", uint64(1))
		c.Set("user_id", uint64(1))
		c.Next()
	})
	h := NewRoleHandler(svc)
	h.SetGovernancePolicyService(policySvc)
	h.SetAdminChecker(middleware.NewDefaultAdminChecker([]int64{1}))
	RegisterRoleRoutes(r.Group("/api"), h)
	return r
}

func TestRound5_Role_NewRoleHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

func TestRound5_Role_SetGovernancePolicyService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	assert.Nil(t, h.governancePolicySvc)
	h.SetGovernancePolicyService(nil)
	assert.Nil(t, h.governancePolicySvc)
}

func TestRound5_Role_SetAdminChecker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	assert.Nil(t, h.adminChecker)
	h.SetAdminChecker(nil)
	assert.Nil(t, h.adminChecker)
}

func TestRound5_Role_Query_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	body := bytes.NewBufferString(`{"keyword":"Admin"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/query", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_Role_Query_InvalidJSON(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/query", bytes.NewBuffer([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_QueryByCurrentOrg_NoOrgContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &org.SysOrg{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}))

	repo := repository.NewRoleRepository(db)
	orgRepo := repository.NewOrgRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	require.NoError(t, orgRepo.Create(&org.SysOrg{OrgID: 1, OrgName: "O1", ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}))
	policySvc := service.NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), nil)
	svc := service.NewRoleService(repo, userRepo, userRoleRepo, orgRepo, policySvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", uint64(0))
		c.Set("user_id", uint64(1))
		c.Next()
	})
	h := NewRoleHandler(svc)
	RegisterRoleRoutes(r.Group("/api"), h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/byCurOrg", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
	assert.Equal(t, "Invalid org context", resp["msg"])
}

func TestRound5_Role_Create_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	body := bytes.NewBufferString(`{"roleName":"NewRole"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/create", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_Role_Create_InvalidJSON(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/create", bytes.NewBuffer([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_Edit_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	createBody := bytes.NewBufferString(`{"roleName":"EditableRole"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/create", createBody)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var createResp struct {
		Data float64 `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	require.Equal(t, "000000", respCode(w.Body.Bytes()))
	roleID := int64(createResp.Data)

	editBody := bytes.NewBufferString(fmt.Sprintf(`{"roleId":%d,"roleName":"EditedName"}`, roleID))
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/role/edit", editBody)
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, "000000", respCode(w2.Body.Bytes()))
}

func TestRound5_Role_Edit_InvalidJSON(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/edit", bytes.NewBuffer([]byte("xxx")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_Delete_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	createBody := bytes.NewBufferString(`{"roleName":"ToBeDeleted"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/create", createBody)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var createResp struct {
		Data float64 `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	roleID := int64(createResp.Data)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", fmt.Sprintf("/api/role/delete/%d", roleID), nil)
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_Role_Delete_InvalidID(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/delete/bad", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_Detail_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/delete/999999", nil)
	_ = req
	req2, _ := http.NewRequest("GET", "/api/role/queryWithOid/1", nil)
	r.ServeHTTP(w, req2)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_Role_Detail_InvalidID(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/role/queryWithOid/bad", nil)
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_MountUser_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	body := bytes.NewBufferString(`{"rid":1,"uids":[50],"orgId":1}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/mountUser", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_Role_MountUser_InvalidJSON(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/mountUser", bytes.NewBuffer([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_MountExternalUser_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	body := bytes.NewBufferString(`{"rid":1,"uid":51}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/mountExternalUser", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_Role_MountExternalUser_InvalidJSON(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/mountExternalUser", bytes.NewBuffer([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_MountExternalUser_NoOrgContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &org.SysOrg{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}))

	repo := repository.NewRoleRepository(db)
	orgRepo := repository.NewOrgRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	require.NoError(t, orgRepo.Create(&org.SysOrg{OrgID: 1, OrgName: "O1", ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}))
	policySvc := service.NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), nil)
	svc := service.NewRoleService(repo, userRepo, userRoleRepo, orgRepo, policySvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", uint64(0))
		c.Set("user_id", uint64(1))
		c.Next()
	})
	h := NewRoleHandler(svc)
	RegisterRoleRoutes(r.Group("/api"), h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/mountExternalUser", bytes.NewBufferString(`{"rid":1,"uid":51}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_UnmountUser_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	mountBody := bytes.NewBufferString(`{"rid":1,"uids":[50],"orgId":1}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/mountUser", mountBody)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)

	unmountBody := bytes.NewBufferString(`{"rid":1,"uid":50,"orgId":1}`)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/role/unMountUser", unmountBody)
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_Role_UnmountUser_InvalidJSON(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/unMountUser", bytes.NewBuffer([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_BeforeUnmountInfo_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	mountBody := bytes.NewBufferString(`{"rid":1,"uids":[50],"orgId":1}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/mountUser", mountBody)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)

	infoBody := bytes.NewBufferString(`{"rid":1,"uid":50,"orgId":1}`)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/role/beforeUnmountInfo", infoBody)
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_Role_BeforeUnmountInfo_InvalidJSON(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/beforeUnmountInfo", bytes.NewBuffer([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_SearchExternalUser_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/role/searchExternalUser/testuser", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_Role_SearchExternalUser_NoOrgContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &org.SysOrg{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}))

	repo := repository.NewRoleRepository(db)
	orgRepo := repository.NewOrgRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	require.NoError(t, orgRepo.Create(&org.SysOrg{OrgID: 1, OrgName: "O1", ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}))
	policySvc := service.NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), nil)
	svc := service.NewRoleService(repo, userRepo, userRoleRepo, orgRepo, policySvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", uint64(0))
		c.Set("user_id", uint64(1))
		c.Next()
	})
	h := NewRoleHandler(svc)
	RegisterRoleRoutes(r.Group("/api"), h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/role/searchExternalUser/test", nil)
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_OptionForUser_Success(t *testing.T) {
	r := setupRound5RoleRouter(t)

	body := bytes.NewBufferString(`{}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/user/option", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRound5_Role_OptionForUser_InvalidJSON(t *testing.T) {
	r := setupRound5RoleRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/user/option", bytes.NewBuffer([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRound5_Role_OptionForUser_NoOrgContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &org.SysOrg{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}))

	repo := repository.NewRoleRepository(db)
	orgRepo := repository.NewOrgRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	require.NoError(t, orgRepo.Create(&org.SysOrg{OrgID: 1, OrgName: "O1", ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}))
	policySvc := service.NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), nil)
	svc := service.NewRoleService(repo, userRepo, userRoleRepo, orgRepo, policySvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", uint64(0))
		c.Set("user_id", uint64(1))
		c.Next()
	})
	h := NewRoleHandler(svc)
	RegisterRoleRoutes(r.Group("/api"), h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/user/option", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}
