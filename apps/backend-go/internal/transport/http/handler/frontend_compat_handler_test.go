package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/user"
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

	h := &FrontendCompatHandler{
		queryMenuTree: func() ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{}, nil
		},
		queryMenuTreeByRoleIDs: func(roleIDs []int64) ([]*menu.MenuVO, error) {
			return []*menu.MenuVO{
				{Path: "/panel/index"},
				{Path: "/data/dataset"},
			}, nil
		},
		loadRoleIDsByUserID: func(userID int64) ([]int64, error) {
			return []int64{2}, nil
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
	if len(data["dashboard"].([]interface{})) != 1 {
		t.Fatalf("expected authorized dashboard tree")
	}
	if len(data["dataset"].([]interface{})) != 1 {
		t.Fatalf("expected authorized dataset tree")
	}
	if len(data["datasource"].([]interface{})) != 0 {
		t.Fatalf("expected unauthorized datasource tree to be empty")
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
