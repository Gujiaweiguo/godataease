package http

import (
	"dataease/backend/internal/app"
	pkgauth "dataease/backend/internal/pkg/auth"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func newRouterWithJWTConfig() *Router {
	return NewRouter(&app.Application{
		Config: &app.Config{
			JWT: app.JWTConfig{Secret: "test-secret", Expire: 3600},
		},
	}, nil)
}

func TestRegisterRoutes_RegistersVisualizationCompatibilityRoutes(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	wantRoutes := map[string]bool{
		"GET /dataVisualization/findDvType/:id":             true,
		"POST /dataVisualization/tree":                      true,
		"POST /dataVisualization/nameCheck":                 true,
		"POST /dataVisualization/checkCanvasChange":         true,
		"POST /dataVisualization/findById":                  true,
		"POST /dataVisualization/list":                      true,
		"POST /dataVisualization/updateBase":                true,
		"POST /dataVisualization/move":                      true,
		"POST /dataVisualization/updatePublishStatus":       true,
		"POST /dataVisualization/recoverToPublished":        true,
		"POST /dataVisualization/saveCanvas":                true,
		"POST /dataVisualization/updateCanvas":              true,
		"POST /dataVisualization/deleteLogic/:id":           true,
		"POST /dataVisualization/deleteLogic/:id/:busiFlag": true,
		"GET /de2api/auth/menuPermission":                   true,
		"POST /de2api/auth/menuPermission":                  true,
		"GET /de2api/auth/busiPermission":                   true,
		"POST /de2api/auth/busiPermission":                  true,
		"POST /de2api/auth/saveMenuPer":                     true,
		"POST /de2api/auth/saveBusiPer":                     true,
		"POST /de2api/system/role/permission/save":          true,
		"POST /de2api/system/role/create":                   true,
		"POST /de2api/system/role/update":                   true,
		"POST /de2api/system/role/delete/:id":               true,
	}

	for _, route := range router.Engine().Routes() {
		key := route.Method + " " + route.Path
		delete(wantRoutes, key)
	}

	if len(wantRoutes) > 0 {
		t.Fatalf("missing visualization compatibility routes: %+v", wantRoutes)
	}
}

func TestRegisterRoutes_DatasourceCanonicalAndCompatibilityContracts(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name string
		path string
	}{
		{name: "canonical datasource list", path: "/api/ds/list"},
		{name: "canonical datasource validate", path: "/api/ds/validate"},
		{name: "api compatibility datasource list", path: "/api/datasource/list"},
		{name: "api compatibility datasource validate", path: "/api/datasource/validate"},
		{name: "de2api datasource list", path: "/de2api/datasource/list"},
		{name: "de2api datasource validate", path: "/de2api/datasource/validate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader("{"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v", tt.path, err)
			}

			if resp.Code != "500000" {
				t.Fatalf("expected code 500000 for %s, got %s", tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, "Invalid request") {
				t.Fatalf("expected invalid request message for %s, got %q", tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_DatasourceValidateSuccessEnvelopeAcrossAliases(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name string
		path string
	}{
		{name: "canonical datasource validate success envelope", path: "/api/ds/validate"},
		{name: "api compatibility datasource validate success envelope", path: "/api/datasource/validate"},
		{name: "de2api datasource validate success envelope", path: "/de2api/datasource/validate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(`{"type":"folder","configuration":"e30="}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string                 `json:"code"`
				Msg  string                 `json:"msg"`
				Data map[string]interface{} `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v", tt.path, err)
			}

			if resp.Code != "000000" {
				t.Fatalf("expected code 000000 for %s, got %s", tt.path, resp.Code)
			}
			if resp.Msg != "success" {
				t.Fatalf("expected msg success for %s, got %q", tt.path, resp.Msg)
			}
			if resp.Data["status"] != "Success" {
				t.Fatalf("expected datasource validation status Success for %s, got %#v", tt.path, resp.Data["status"])
			}
		})
	}
}

func TestRegisterRoutes_DatasourceListReturnsExplicitErrorEnvelopeAcrossAliasesWhenStoreUnavailable(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name string
		path string
	}{
		{name: "canonical datasource list explicit error envelope", path: "/api/ds/list"},
		{name: "api compatibility datasource list explicit error envelope", path: "/api/datasource/list"},
		{name: "de2api datasource list explicit error envelope", path: "/de2api/datasource/list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}

			if resp.Code != "500000" {
				t.Fatalf("expected code 500000 for %s, got %s", tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, "repository is unavailable") {
				t.Fatalf("expected unavailable-store message for %s, got %q", tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_DatasourceListAliasesRequireAuthentication(t *testing.T) {
	router := newRouterWithJWTConfig()
	router.RegisterRoutes()

	tests := []struct {
		name string
		path string
	}{
		{name: "canonical datasource list requires auth", path: "/api/ds/list"},
		{name: "api compatibility datasource list requires auth", path: "/api/datasource/list"},
		{name: "de2api datasource list requires auth", path: "/de2api/datasource/list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 401 {
				t.Fatalf("expected status 401 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}

			if resp.Code != "20001" {
				t.Fatalf("expected code 20001 for %s, got %s", tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, "authorization") {
				t.Fatalf("expected authorization message for %s, got %q", tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_DatasourceListAliasesReturnExplicitErrorAfterAuthenticationWhenStoreUnavailable(t *testing.T) {
	router := newRouterWithJWTConfig()
	router.RegisterRoutes()

	jwtInstance := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "test-secret", Expire: 3600})
	token, err := jwtInstance.GenerateToken(1, "admin", "admin")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{name: "canonical datasource list explicit error envelope after auth", path: "/api/ds/list"},
		{name: "api compatibility datasource list explicit error envelope after auth", path: "/api/datasource/list"},
		{name: "de2api datasource list explicit error envelope after auth", path: "/de2api/datasource/list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}

			if resp.Code != "500000" {
				t.Fatalf("expected code 500000 for %s, got %s", tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, "repository is unavailable") {
				t.Fatalf("expected unavailable-store message for %s, got %q", tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_LoginRefreshCompatibilityRouteReturnsRefreshedToken(t *testing.T) {
	router := newRouterWithJWTConfig()
	router.RegisterRoutes()

	jwtInstance := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "test-secret", Expire: 3600})
	token, err := jwtInstance.GenerateTokenWithOrgID(7, "refresh-user", "", 3)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	tests := []string{"/login/refresh", "/api/login/refresh"}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path+"?time=1", nil)
			req.Header.Set("X-DE-TOKEN", token)
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200 for %s, got %d with body %s", path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Data struct {
					Token string `json:"token"`
					Exp   int64  `json:"exp"`
				} `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v", path, err)
			}
			if resp.Code != "000000" {
				t.Fatalf("expected code 000000 for %s, got %s", path, resp.Code)
			}
			if resp.Data.Token == "" || resp.Data.Exp == 0 {
				t.Fatalf("expected refreshed token payload for %s, got %#v", path, resp.Data)
			}

			claims, err := jwtInstance.ParseToken(resp.Data.Token)
			if err != nil {
				t.Fatalf("parse refreshed token for %s failed: %v", path, err)
			}
			if claims.UserID != 7 || claims.OrgID != 3 {
				t.Fatalf("expected refreshed token to preserve user/org for %s, got %#v", path, claims)
			}
		})
	}
}

func TestRegisterRoutes_DatasetCanonicalAndCompatibilityContracts(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name string
		path string
	}{
		{name: "canonical dataset tree", path: "/api/dataset/tree"},
		{name: "canonical dataset fields", path: "/api/dataset/fields"},
		{name: "canonical dataset preview", path: "/api/dataset/preview"},
		{name: "api compatibility dataset tree", path: "/api/datasetTree/tree"},
		{name: "api compatibility dataset fields", path: "/api/datasetData/tableField"},
		{name: "api compatibility dataset preview", path: "/api/datasetData/previewData"},
		{name: "de2api dataset tree", path: "/de2api/datasetTree/tree"},
		{name: "de2api dataset fields", path: "/de2api/datasetData/tableField"},
		{name: "de2api dataset preview", path: "/de2api/datasetData/previewData"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader("{"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v", tt.path, err)
			}

			if resp.Code != "500000" {
				t.Fatalf("expected code 500000 for %s, got %s", tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, "Invalid request") {
				t.Fatalf("expected invalid request message for %s, got %q", tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_DatasetAliasesRequireAuthentication(t *testing.T) {
	router := newRouterWithJWTConfig()
	router.RegisterRoutes()

	tests := []struct {
		name string
		path string
	}{
		{name: "canonical dataset tree requires auth", path: "/api/dataset/tree"},
		{name: "canonical dataset fields requires auth", path: "/api/dataset/fields"},
		{name: "canonical dataset preview requires auth", path: "/api/dataset/preview"},
		{name: "api compatibility dataset tree requires auth", path: "/api/datasetTree/tree"},
		{name: "api compatibility dataset fields requires auth", path: "/api/datasetData/tableField"},
		{name: "api compatibility dataset preview requires auth", path: "/api/datasetData/previewData"},
		{name: "de2api dataset tree requires auth", path: "/de2api/datasetTree/tree"},
		{name: "de2api dataset fields requires auth", path: "/de2api/datasetData/tableField"},
		{name: "de2api dataset preview requires auth", path: "/de2api/datasetData/previewData"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 401 {
				t.Fatalf("expected status 401 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}

			if resp.Code != "20001" {
				t.Fatalf("expected code 20001 for %s, got %s", tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, "authorization") {
				t.Fatalf("expected authorization message for %s, got %q", tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_DatasetAliasesReturnExplicitErrorAfterAuthenticationWhenStoreUnavailable(t *testing.T) {
	router := newRouterWithJWTConfig()
	router.RegisterRoutes()

	jwtInstance := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "test-secret", Expire: 3600})
	token, err := jwtInstance.GenerateToken(1, "admin", "admin")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	tests := []struct {
		name string
		path string
		msg  string
	}{
		{name: "canonical dataset tree explicit error after auth", path: "/api/dataset/tree", msg: "dataset repository is unavailable"},
		{name: "canonical dataset fields explicit error after auth", path: "/api/dataset/fields", msg: "Invalid request"},
		{name: "canonical dataset preview explicit error after auth", path: "/api/dataset/preview", msg: "Invalid request"},
		{name: "api compatibility dataset tree explicit error after auth", path: "/api/datasetTree/tree", msg: "dataset repository is unavailable"},
		{name: "api compatibility dataset fields explicit error after auth", path: "/api/datasetData/tableField", msg: "Invalid request"},
		{name: "api compatibility dataset preview explicit error after auth", path: "/api/datasetData/previewData", msg: "Invalid request"},
		{name: "de2api dataset tree explicit error after auth", path: "/de2api/datasetTree/tree", msg: "dataset repository is unavailable"},
		{name: "de2api dataset fields explicit error after auth", path: "/de2api/datasetData/tableField", msg: "Invalid request"},
		{name: "de2api dataset preview explicit error after auth", path: "/de2api/datasetData/previewData", msg: "Invalid request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}

			if resp.Code != "500000" {
				t.Fatalf("expected code 500000 for %s, got %s", tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, tt.msg) {
				t.Fatalf("expected message containing %q for %s, got %q", tt.msg, tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_VisualizationCanonicalAndCompatibilityContracts(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantCode   string
		msgPart    string
	}{
		{name: "api visualization tree", path: "/api/dataVisualization/tree", wantStatus: 200, wantCode: "500000", msgPart: "Invalid request"},
		{name: "api visualization detail", path: "/api/dataVisualization/findById", wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "de2api visualization tree", path: "/de2api/dataVisualization/tree", wantStatus: 200, wantCode: "500000", msgPart: "Invalid request"},
		{name: "de2api visualization detail", path: "/de2api/dataVisualization/findById", wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader("{"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d for %s, got %d with body %s", tt.wantStatus, tt.path, w.Code, w.Body.String())
			}

			body := w.Body.String()
			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil {
				if resp.Code != tt.wantCode {
					t.Fatalf("expected code %s for %s, got %s", tt.wantCode, tt.path, resp.Code)
				}
				if !strings.Contains(resp.Msg, tt.msgPart) {
					t.Fatalf("expected message part %q for %s, got %q", tt.msgPart, tt.path, resp.Msg)
				}
				return
			}

			if !strings.Contains(body, tt.wantCode) {
				t.Fatalf("expected body for %s to contain code %s, got %s", tt.path, tt.wantCode, body)
			}
			if !strings.Contains(body, tt.msgPart) {
				t.Fatalf("expected body for %s to contain message part %q, got %s", tt.path, tt.msgPart, body)
			}
		})
	}
}

func TestRegisterRoutes_VisualizationDe2apiDetailReturnsSingleUnauthorizedEnvelope(t *testing.T) {
	router := newRouterWithJWTConfig()
	router.RegisterRoutes()

	req := httptest.NewRequest("POST", "/de2api/dataVisualization/findById", strings.NewReader(`{"id":123,"busiFlag":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.Engine().ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d with body %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected a single valid JSON envelope, got error %v with body %s", err, w.Body.String())
	}
	if resp.Code != "20001" {
		t.Fatalf("expected code 20001, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "authorization") {
		t.Fatalf("expected authorization-related message, got %q", resp.Msg)
	}
}

func TestRegisterRoutes_AuditRoutesRequireAuthentication(t *testing.T) {
	router := newRouterWithJWTConfig()
	router.RegisterRoutes()

	tests := []struct {
		name string
		path string
	}{
		{name: "audit list requires auth", path: "/api/audit/list"},
		{name: "audit detail requires auth", path: "/api/audit/123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 401 {
				t.Fatalf("expected status 401 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("expected a valid JSON envelope for %s: %v, body=%s", tt.path, err, w.Body.String())
			}
			if resp.Code != "20001" {
				t.Fatalf("expected code 20001 for %s, got %s", tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, "authorization") {
				t.Fatalf("expected authorization-related message for %s, got %q", tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_ExportCenterRoutesRequireAuthentication(t *testing.T) {
	router := newRouterWithJWTConfig()
	router.RegisterRoutes()

	tests := []struct {
		name string
		path string
	}{
		{name: "export-center records requires auth", path: "/api/exportCenter/exportTasks/records"},
		{name: "export-center pager requires auth", path: "/api/exportCenter/exportTasks/all/1/10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 401 {
				t.Fatalf("expected status 401 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("expected a valid JSON envelope for %s: %v, body=%s", tt.path, err, w.Body.String())
			}
			if resp.Code != "20001" {
				t.Fatalf("expected code 20001 for %s, got %s", tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, "authorization") {
				t.Fatalf("expected authorization-related message for %s, got %q", tt.path, resp.Msg)
			}
		})
	}
}
