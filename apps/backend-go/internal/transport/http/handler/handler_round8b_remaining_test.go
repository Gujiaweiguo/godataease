package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---------- 1. QueryAuditLogs ----------

func TestRound8B_AuditHandler_QueryAuditLogs_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.QueryAuditLogs(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8B_AuditHandler_QueryAuditLogs_WithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/?userId=1&username=admin&actionType=LOGIN&resourceType=USER&organizationId=2&status=SUCCESS&startTime=2024-01-01T00:00:00Z&endTime=2024-12-31T23:59:59Z&page=2&pageSize=50", "")
	h.QueryAuditLogs(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8B_AuditHandler_QueryAuditLogs_DefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.QueryAuditLogs(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8B_AuditHandler_QueryAuditLogs_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/?userId=bad", "")
	h.QueryAuditLogs(c)
	resp := parseRound8Resp(t, w)
	// Invalid userID is silently ignored, still calls service which is nil → panic
	assert.Equal(t, "500000", resp["code"])
}

// ---------- 2. RegisterGeoRoutes ----------

func TestRound8B_RegisterGeoRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewGeoHandler(nil)
	RegisterGeoRoutes(apiGroup, h)

	routes := r.Routes()
	routeMap := make(map[string]string)
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = route.Path
	}

	assert.Contains(t, routeMap, "GET:/api/geometry/areaList")
	assert.Contains(t, routeMap, "GET:/api/geometry/area/:id")
	assert.Contains(t, routeMap, "POST:/api/geometry/save")
	assert.Contains(t, routeMap, "DELETE:/api/geometry/delete/:id")
}

// ---------- 3. RegisterMapRoutes ----------

func TestRound8B_RegisterMapRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewMapHandler(nil)
	RegisterMapRoutes(apiGroup, h)

	routes := r.Routes()
	routeMap := make(map[string]string)
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = route.Path
	}

	assert.Contains(t, routeMap, "GET:/api/map/worldTree")
}

// ---------- 4. RegisterPdfTemplateRoutes ----------

func TestRound8B_RegisterPdfTemplateRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewPdfTemplateHandler()
	RegisterPdfTemplateRoutes(apiGroup, h)

	routes := r.Routes()
	routeMap := make(map[string]string)
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = route.Path
	}

	assert.Contains(t, routeMap, "GET:/api/pdf-template/queryAll")
}

// ---------- 5. SetRoleService ----------

func TestRound8B_PermissionCompat_SetRoleService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	// SetRoleService is a simple setter — verify it doesn't panic
	assert.NotPanics(t, func() {
		h.SetRoleService(nil)
	})
	assert.Nil(t, h.roleService)
}

func TestRound8B_PermissionCompat_SetRoleService_SetsField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	// Verify field starts nil
	assert.Nil(t, h.roleService)
	// We can't create a real RoleService without a DB, but we can verify the setter doesn't panic
	h.SetRoleService(nil)
	assert.Nil(t, h.roleService)
}

// ---------- 6. BusiPermission ----------

func TestRound8B_PermissionCompat_BusiPermission_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	// Need to set up context for buildPermissionScope
	c.Set("role", "admin")
	c.Set("user_id", uint64(1))
	h.BusiPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8B_PermissionCompat_BusiPermission_MissingScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	// No role/user_id set — buildPermissionScope returns error
	h.BusiPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

// ---------- 7. SaveRolePermission ----------

func TestRound8B_PermissionCompat_SaveRolePermission_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", `{"roleId":1,"permIds":[1,2,3]}`)
	c.Set("role", "admin")
	c.Set("user_id", uint64(1))
	h.SaveRolePermission(c)
	resp := parseRound8Resp(t, w)
	// SaveRolePermission delegates to SaveBusiPer which needs roleMenuService — nil → panic
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8B_PermissionCompat_SaveRolePermission_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	c.Set("role", "admin")
	c.Set("user_id", uint64(1))
	h.SaveRolePermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

// ---------- 8. RegisterPermissionRoutes ----------

func TestRound8B_RegisterPermissionRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	RegisterPermissionRoutes(apiGroup, h)

	routes := r.Routes()
	routeMap := make(map[string]string)
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = route.Path
	}

	expectedRoutes := []string{
		"GET:/api/system/permission/menuPermission",
		"POST:/api/system/permission/menuPermission",
		"GET:/api/system/permission/busiPermission",
		"POST:/api/system/permission/busiPermission",
		"GET:/api/system/permission/busiResource/:flag",
		"POST:/api/system/permission/userPerspective",
		"POST:/api/system/permission/menuTargetPermission",
		"POST:/api/system/permission/busiTargetPermission",
		"POST:/api/system/permission/saveMenuPer",
		"POST:/api/system/permission/saveBusiPer",
		"POST:/api/system/permission/saveMenuTargetPer",
		"POST:/api/system/permission/saveBusiTargetPer",
	}
	for _, expected := range expectedRoutes {
		assert.Contains(t, routeMap, expected, "missing route: %s", expected)
	}
}

func TestRound8B_RegisterPermissionRoutes_NilHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	RegisterPermissionRoutes(apiGroup, nil)
	// Should register no routes when handler is nil
	routes := r.Routes()
	assert.Empty(t, routes)
}

// ---------- 9. RegisterRelationRoutes ----------

func TestRound8B_RegisterRelationRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewRelationHandler(nil)
	RegisterRelationRoutes(apiGroup, h)

	routes := r.Routes()
	routeMap := make(map[string]string)
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = route.Path
	}

	assert.Contains(t, routeMap, "POST:/api/relation/datasource/:id")
	assert.Contains(t, routeMap, "POST:/api/relation/dataset/:id")
	assert.Contains(t, routeMap, "POST:/api/relation/dv/:id")
	assert.Contains(t, routeMap, "POST:/api/resource/checkPermission/:id")
}

// ---------- 10. Detail (RoleHandler) ----------

func TestRound8B_RoleHandler_Detail_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Detail(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8B_RoleHandler_Detail_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	h.Detail(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8B_RoleHandler_Detail_NegativeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "-1"}}
	h.Detail(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

// ---------- 11. getThresholdUserInfo (unexported helper) ----------

func TestRound8B_GetThresholdUserInfo_NoContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	userID, userName, oid := getThresholdUserInfo(c)
	assert.Equal(t, int64(0), userID)
	assert.Equal(t, "", userName)
	assert.Equal(t, int64(0), oid)
}

func TestRound8B_GetThresholdUserInfo_WithValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("user_id", uint64(42))
	c.Set("username", "testuser")
	c.Set("org_id", uint64(10))
	userID, userName, oid := getThresholdUserInfo(c)
	assert.Equal(t, int64(42), userID)
	assert.Equal(t, "testuser", userName)
	assert.Equal(t, int64(10), oid)
}

func TestRound8B_GetThresholdUserInfo_Int64OrgID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("user_id", uint64(7))
	c.Set("username", "another")
	c.Set("org_id", int64(99))
	userID, userName, oid := getThresholdUserInfo(c)
	assert.Equal(t, int64(7), userID)
	assert.Equal(t, "another", userName)
	assert.Equal(t, int64(99), oid)
}

// ---------- 12. RegisterVisualizationBackgroundRoutes ----------

func TestRound8B_RegisterVisualizationBackgroundRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewVisualizationBackgroundHandler(nil)
	RegisterVisualizationBackgroundRoutes(apiGroup, h)

	routes := r.Routes()
	routeMap := make(map[string]string)
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = route.Path
	}

	assert.Contains(t, routeMap, "GET:/api/visualizationBackground/findAll")
}

// ---------- Bonus: RegisterPermissionCompatRoutes (nil handler guard) ----------

func TestRound8B_RegisterPermissionCompatRoutes_NilHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	RegisterPermissionCompatRoutes(apiGroup, nil)
	routes := r.Routes()
	assert.Empty(t, routes)
}

func TestRound8B_RegisterPermissionCompatRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	RegisterPermissionCompatRoutes(apiGroup, h)

	routes := r.Routes()
	routeMap := make(map[string]string)
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = route.Path
	}

	expectedRoutes := []string{
		"GET:/api/auth/menuPermission",
		"POST:/api/auth/menuPermission",
		"GET:/api/auth/busiPermission",
		"POST:/api/auth/busiPermission",
		"GET:/api/auth/busiResource/:flag",
		"POST:/api/auth/userPerspective",
		"POST:/api/auth/menuTargetPermission",
		"POST:/api/auth/busiTargetPermission",
		"POST:/api/auth/saveMenuPer",
		"POST:/api/auth/saveBusiPer",
		"POST:/api/auth/saveMenuTargetPer",
		"POST:/api/auth/saveBusiTargetPer",
		"POST:/api/role/permission/save",
		"POST:/api/system/role/permission/save",
	}
	for _, expected := range expectedRoutes {
		assert.Contains(t, routeMap, expected, "missing route: %s", expected)
	}
}

// ---------- Route registration integration: verify handlers are wired ----------

func TestRound8B_RegisterGeoRoutes_RoutesReachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewGeoHandler(nil)
	RegisterGeoRoutes(apiGroup, h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/geometry/areaList", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound8B_RegisterMapRoutes_RoutesReachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewMapHandler(nil)
	RegisterMapRoutes(apiGroup, h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/map/worldTree", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound8B_RegisterPdfTemplateRoutes_RoutesReachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewPdfTemplateHandler()
	RegisterPdfTemplateRoutes(apiGroup, h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/pdf-template/queryAll", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
}

func TestRound8B_RegisterVisualizationBackgroundRoutes_RoutesReachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewVisualizationBackgroundHandler(nil)
	RegisterVisualizationBackgroundRoutes(apiGroup, h)

	routes := r.Routes()
	assert.Len(t, routes, 1)
	assert.Equal(t, "/api/visualizationBackground/findAll", routes[0].Path)
}

func TestRound8B_RegisterRelationRoutes_RouteCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	h := NewRelationHandler(nil)
	RegisterRelationRoutes(apiGroup, h)

	routes := r.Routes()
	assert.Len(t, routes, 4)
}
