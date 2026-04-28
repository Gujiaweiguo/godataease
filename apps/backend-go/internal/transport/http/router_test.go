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
		name   string
		method string
		path   string
		body   string
	}{
		{name: "canonical datasource list", method: "POST", path: "/api/ds/list", body: "{"},
		{name: "canonical datasource tree", method: "POST", path: "/api/ds/tree", body: "{"},
		{name: "canonical datasource validate", method: "POST", path: "/api/ds/validate", body: "{"},
		{name: "canonical datasource hide password", method: "GET", path: "/api/ds/hidePw/not-a-number", body: ""},
		{name: "canonical datasource simple", method: "GET", path: "/api/ds/simple/not-a-number", body: ""},
		{name: "canonical datasource check repeat", method: "POST", path: "/api/ds/checkRepeat", body: "{"},
		{name: "canonical datasource check api datasource", method: "POST", path: "/api/ds/checkApiDatasource", body: "{"},
		{name: "canonical datasource save", method: "POST", path: "/api/ds/save", body: "{"},
		{name: "canonical datasource update", method: "POST", path: "/api/ds/update", body: "{"},
		{name: "canonical datasource delete", method: "POST", path: "/api/ds/delete/not-a-number", body: "{"},
		{name: "canonical datasource per delete", method: "POST", path: "/api/ds/perDelete/not-a-number", body: ""},
		{name: "canonical datasource move", method: "POST", path: "/api/ds/move", body: "{"},
		{name: "canonical datasource rename", method: "POST", path: "/api/ds/reName", body: "{"},
		{name: "canonical datasource create folder", method: "POST", path: "/api/ds/createFolder", body: "{"},
		{name: "canonical datasource tables", method: "POST", path: "/api/ds/tables", body: "{"},
		{name: "canonical datasource table status", method: "POST", path: "/api/ds/tableStatus", body: "{"},
		{name: "canonical datasource table field", method: "POST", path: "/api/ds/tableField", body: "{"},
		{name: "canonical datasource preview data", method: "POST", path: "/api/ds/previewData", body: "{"},
		{name: "canonical datasource sync api table", method: "POST", path: "/api/ds/syncApiTable", body: "{"},
		{name: "canonical datasource sync api ds", method: "POST", path: "/api/ds/syncApiDs", body: "{"},
		{name: "canonical datasource load remote file", method: "POST", path: "/api/ds/loadRemoteFile", body: "{"},
		{name: "canonical datasource upload file", method: "POST", path: "/api/ds/uploadFile", body: `{}`},
		{name: "api compatibility datasource list", method: "POST", path: "/api/datasource/list", body: "{"},
		{name: "api compatibility datasource validate", method: "POST", path: "/api/datasource/validate", body: "{"},
		{name: "api compatibility datasource check repeat", method: "POST", path: "/api/datasource/checkRepeat", body: "{"},
		{name: "api compatibility datasource check api datasource", method: "POST", path: "/api/datasource/checkApiDatasource", body: "{"},
		{name: "api compatibility datasource tables", method: "POST", path: "/api/datasource/getTables", body: "{"},
		{name: "api compatibility datasource table status", method: "POST", path: "/api/datasource/getTableStatus", body: "{"},
		{name: "api compatibility datasource table field", method: "POST", path: "/api/datasource/getTableField", body: "{"},
		{name: "api compatibility datasource preview data", method: "POST", path: "/api/datasource/previewData", body: "{"},
		{name: "api compatibility datasource sync api table", method: "POST", path: "/api/datasource/syncApiTable", body: "{"},
		{name: "api compatibility datasource sync api ds", method: "POST", path: "/api/datasource/syncApiDs", body: "{"},
		{name: "api compatibility datasource load remote file", method: "POST", path: "/api/datasource/loadRemoteFile", body: "{"},
		{name: "api compatibility datasource upload file", method: "POST", path: "/api/datasource/uploadFile", body: `{}`},
		{name: "de2api datasource list", method: "POST", path: "/de2api/datasource/list", body: "{"},
		{name: "de2api datasource validate", method: "POST", path: "/de2api/datasource/validate", body: "{"},
		{name: "de2api datasource check repeat", method: "POST", path: "/de2api/datasource/checkRepeat", body: "{"},
		{name: "de2api datasource check api datasource", method: "POST", path: "/de2api/datasource/checkApiDatasource", body: "{"},
		{name: "de2api datasource tables", method: "POST", path: "/de2api/datasource/getTables", body: "{"},
		{name: "de2api datasource table status", method: "POST", path: "/de2api/datasource/getTableStatus", body: "{"},
		{name: "de2api datasource table field", method: "POST", path: "/de2api/datasource/getTableField", body: "{"},
		{name: "de2api datasource preview data", method: "POST", path: "/de2api/datasource/previewData", body: "{"},
		{name: "de2api datasource sync api table", method: "POST", path: "/de2api/datasource/syncApiTable", body: "{"},
		{name: "de2api datasource sync api ds", method: "POST", path: "/de2api/datasource/syncApiDs", body: "{"},
		{name: "de2api datasource load remote file", method: "POST", path: "/de2api/datasource/loadRemoteFile", body: "{"},
		{name: "de2api datasource upload file", method: "POST", path: "/de2api/datasource/uploadFile", body: `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := tt.method
			if method == "" {
				method = "POST"
			}
			req := httptest.NewRequest(method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				if tt.path == "/api/ds/validate" || tt.path == "/api/datasource/validate" || tt.path == "/de2api/datasource/validate" {
					if w.Code != 403 {
						t.Fatalf("expected status 403 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
					}
					var resp struct {
						Code string `json:"code"`
						Msg  string `json:"msg"`
					}
					if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
						t.Fatalf("unmarshal response for %s failed: %v", tt.path, err)
					}
					if resp.Code != "70001" {
						t.Fatalf("expected code 70001 for %s, got %s", tt.path, resp.Code)
					}
					if !strings.Contains(resp.Msg, "No role assigned") {
						t.Fatalf("expected missing role message for %s, got %q", tt.path, resp.Msg)
					}
					return
				}
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
			if tt.path == "/api/ds/delete/not-a-number" || tt.path == "/api/ds/hidePw/not-a-number" || tt.path == "/api/ds/simple/not-a-number" || tt.path == "/api/ds/perDelete/not-a-number" {
				if !strings.Contains(resp.Msg, "Invalid datasource ID") {
					t.Fatalf("expected invalid datasource id message for %s, got %q", tt.path, resp.Msg)
				}
				return
			}
			if strings.Contains(tt.path, "/uploadFile") {
				if !strings.Contains(resp.Msg, "Failed to get uploaded file") {
					t.Fatalf("expected upload file binding error for %s, got %q", tt.path, resp.Msg)
				}
				return
			}
			if !strings.Contains(resp.Msg, "Invalid request") {
				t.Fatalf("expected invalid request message for %s, got %q", tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_DatasourceTableExplorationRoutesExistAcrossAliases(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	wantRoutes := map[string]bool{
		"POST /api/ds/tables":                        true,
		"POST /api/ds/tableStatus":                   true,
		"POST /api/ds/tableField":                    true,
		"POST /api/ds/schema":                        true,
		"POST /api/ds/previewData":                   true,
		"POST /api/ds/syncApiTable":                  true,
		"POST /api/ds/syncApiDs":                     true,
		"POST /api/ds/loadRemoteFile":                true,
		"POST /api/ds/uploadFile":                    true,
		"GET /api/ds/hidePw/:id":                     true,
		"GET /api/ds/simple/:id":                     true,
		"POST /api/ds/perDelete/:id":                 true,
		"GET /api/ds/validate/:id":                   true,
		"POST /api/ds/checkRepeat":                   true,
		"POST /api/ds/checkApiDatasource":            true,
		"POST /api/ds/move":                          true,
		"POST /api/ds/reName":                        true,
		"POST /api/ds/createFolder":                  true,
		"POST /api/datasource/getTables":             true,
		"GET /api/datasource/validate/:id":           true,
		"POST /api/datasource/checkRepeat":           true,
		"POST /api/datasource/checkApiDatasource":    true,
		"POST /api/datasource/getTableStatus":        true,
		"POST /api/datasource/getTableField":         true,
		"POST /api/datasource/getSchema":             true,
		"POST /api/datasource/previewData":           true,
		"POST /api/datasource/syncApiTable":          true,
		"POST /api/datasource/syncApiDs":             true,
		"POST /api/datasource/loadRemoteFile":        true,
		"POST /api/datasource/uploadFile":            true,
		"POST /de2api/datasource/getTables":          true,
		"POST /de2api/datasource/getTableStatus":     true,
		"POST /de2api/datasource/getTableField":      true,
		"POST /de2api/datasource/getSchema":          true,
		"POST /de2api/datasource/previewData":        true,
		"POST /de2api/datasource/syncApiTable":       true,
		"POST /de2api/datasource/syncApiDs":          true,
		"POST /de2api/datasource/loadRemoteFile":     true,
		"POST /de2api/datasource/uploadFile":         true,
		"GET /de2api/datasource/validate/:id":        true,
		"POST /de2api/datasource/checkRepeat":        true,
		"POST /de2api/datasource/checkApiDatasource": true,
	}

	for _, route := range router.Engine().Routes() {
		key := route.Method + " " + route.Path
		delete(wantRoutes, key)
	}

	if len(wantRoutes) > 0 {
		t.Fatalf("missing datasource table exploration routes: %+v", wantRoutes)
	}
}

func TestRegisterRoutes_DatasourceValidateByIDCanonicalRouteReturnsExplicitFailureForInvalidID(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	req := httptest.NewRequest("GET", "/api/ds/validate/not-a-number", nil)
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
		t.Fatalf("unmarshal response failed: %v; body=%s", err, w.Body.String())
	}

	if resp.Code != "20001" {
		t.Fatalf("expected code 20001, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "authentication required") {
		t.Fatalf("expected authentication required message, got %q", resp.Msg)
	}
}

func TestRegisterRoutes_DatasourceCheckRepeatSuccessEnvelopeAcrossAliases(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name string
		path string
	}{
		{name: "canonical datasource check repeat success envelope", path: "/api/ds/checkRepeat"},
		{name: "api compatibility datasource check repeat success envelope", path: "/api/datasource/checkRepeat"},
		{name: "de2api datasource check repeat success envelope", path: "/de2api/datasource/checkRepeat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(`{"type":"folder"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
				Data bool   `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}

			if resp.Code != "000000" {
				t.Fatalf("expected code 000000 for %s, got %s", tt.path, resp.Code)
			}
			if resp.Msg != "success" {
				t.Fatalf("expected success msg for %s, got %q", tt.path, resp.Msg)
			}
			if resp.Data {
				t.Fatalf("expected false repeat-check result for %s, got true", tt.path)
			}
		})
	}
}

func TestRegisterRoutes_DatasourceCheckRepeatCanonicalRouteReturnsExplicitFailureWhenStoreUnavailable(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	req := httptest.NewRequest("POST", "/api/ds/checkRepeat", strings.NewReader(`{"type":"mysql","configuration":"eyJob3N0IjoibG9jYWxob3N0IiwicG9ydCI6MzMwNn0="}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.Engine().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v; body=%s", err, w.Body.String())
	}

	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "repository is unavailable") {
		t.Fatalf("expected repository unavailable message, got %q", resp.Msg)
	}
}

func TestRegisterRoutes_DatasourceCheckAPIDatasourceRoutesReturnExplicitEnvelopesAcrossAliases(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name         string
		path         string
		body         string
		wantCode     string
		wantDataType string
		wantMessage  string
	}{
		{
			name:         "canonical datasource check api datasource success envelope",
			path:         "/api/ds/checkApiDatasource",
			body:         `{"data":"eyJ1cmwiOiJodHRwOi8vZXhhbXBsZS5jb20ifQ==","type":"apiStructure"}`,
			wantCode:     "000000",
			wantDataType: "table",
		},
		{
			name:         "api compatibility datasource check api datasource success envelope",
			path:         "/api/datasource/checkApiDatasource",
			body:         `{"data":"eyJ1cmwiOiJodHRwOi8vZXhhbXBsZS5jb20ifQ==","type":"apiStructure"}`,
			wantCode:     "000000",
			wantDataType: "table",
		},
		{
			name:         "de2api datasource check api datasource success envelope",
			path:         "/de2api/datasource/checkApiDatasource",
			body:         `{"data":"eyJ1cmwiOiJodHRwOi8vZXhhbXBsZS5jb20ifQ==","type":"apiStructure"}`,
			wantCode:     "000000",
			wantDataType: "table",
		},
		{
			name:        "canonical datasource check api datasource explicit failure",
			path:        "/api/ds/checkApiDatasource",
			body:        `{}`,
			wantCode:    "500000",
			wantMessage: "request is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			if tt.wantCode == "000000" {
				var resp struct {
					Code string         `json:"code"`
					Msg  string         `json:"msg"`
					Data map[string]any `json:"data"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
				}
				if resp.Code != tt.wantCode {
					t.Fatalf("expected code %s for %s, got %s", tt.wantCode, tt.path, resp.Code)
				}
				if resp.Msg != "success" {
					t.Fatalf("expected success msg for %s, got %q", tt.path, resp.Msg)
				}
				if resp.Data["type"] != tt.wantDataType {
					t.Fatalf("expected type %q for %s, got %#v", tt.wantDataType, tt.path, resp.Data["type"])
				}
				if resp.Data["showApiStructure"] != true {
					t.Fatalf("expected showApiStructure true for %s, got %#v", tt.path, resp.Data["showApiStructure"])
				}
				return
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}
			if resp.Code != tt.wantCode {
				t.Fatalf("expected code %s for %s, got %s", tt.wantCode, tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, tt.wantMessage) {
				t.Fatalf("expected message containing %q for %s, got %q", tt.wantMessage, tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_DatasourceValidateAliasesRequireAuthorizationWithoutRoleContext(t *testing.T) {
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

			if w.Code != 403 {
				t.Fatalf("expected status 403 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string         `json:"code"`
				Msg  string         `json:"msg"`
				Data map[string]any `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v", tt.path, err)
			}

			if resp.Code != "70001" {
				t.Fatalf("expected code 70001 for %s, got %s", tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, "No role assigned") {
				t.Fatalf("expected missing role message for %s, got %q", tt.path, resp.Msg)
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
		{name: "canonical datasource tree requires auth", path: "/api/ds/tree"},
		{name: "canonical datasource save requires auth", path: "/api/ds/save"},
		{name: "canonical datasource update requires auth", path: "/api/ds/update"},
		{name: "canonical datasource delete requires auth", path: "/api/ds/delete/1"},
		{name: "canonical datasource tables requires auth", path: "/api/ds/tables"},
		{name: "canonical datasource table status requires auth", path: "/api/ds/tableStatus"},
		{name: "canonical datasource table field requires auth", path: "/api/ds/tableField"},
		{name: "canonical datasource schema requires auth", path: "/api/ds/schema"},
		{name: "canonical datasource preview data requires auth", path: "/api/ds/previewData"},
		{name: "canonical datasource sync api table requires auth", path: "/api/ds/syncApiTable"},
		{name: "canonical datasource sync api ds requires auth", path: "/api/ds/syncApiDs"},
		{name: "canonical datasource load remote file requires auth", path: "/api/ds/loadRemoteFile"},
		{name: "canonical datasource upload file requires auth", path: "/api/ds/uploadFile"},
		{name: "api compatibility datasource list requires auth", path: "/api/datasource/list"},
		{name: "api compatibility datasource tables requires auth", path: "/api/datasource/getTables"},
		{name: "api compatibility datasource table status requires auth", path: "/api/datasource/getTableStatus"},
		{name: "api compatibility datasource table field requires auth", path: "/api/datasource/getTableField"},
		{name: "api compatibility datasource schema requires auth", path: "/api/datasource/getSchema"},
		{name: "api compatibility datasource preview data requires auth", path: "/api/datasource/previewData"},
		{name: "api compatibility datasource sync api table requires auth", path: "/api/datasource/syncApiTable"},
		{name: "api compatibility datasource sync api ds requires auth", path: "/api/datasource/syncApiDs"},
		{name: "api compatibility datasource load remote file requires auth", path: "/api/datasource/loadRemoteFile"},
		{name: "api compatibility datasource upload file requires auth", path: "/api/datasource/uploadFile"},
		{name: "de2api datasource list requires auth", path: "/de2api/datasource/list"},
		{name: "de2api datasource tables requires auth", path: "/de2api/datasource/getTables"},
		{name: "de2api datasource table status requires auth", path: "/de2api/datasource/getTableStatus"},
		{name: "de2api datasource table field requires auth", path: "/de2api/datasource/getTableField"},
		{name: "de2api datasource schema requires auth", path: "/de2api/datasource/getSchema"},
		{name: "de2api datasource preview data requires auth", path: "/de2api/datasource/previewData"},
		{name: "de2api datasource sync api table requires auth", path: "/de2api/datasource/syncApiTable"},
		{name: "de2api datasource sync api ds requires auth", path: "/de2api/datasource/syncApiDs"},
		{name: "de2api datasource load remote file requires auth", path: "/de2api/datasource/loadRemoteFile"},
		{name: "de2api datasource upload file requires auth", path: "/de2api/datasource/uploadFile"},
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
		body string
	}{
		{name: "canonical datasource list explicit error envelope after auth", path: "/api/ds/list", body: `{}`},
		{name: "canonical datasource tree explicit error envelope after auth", path: "/api/ds/tree", body: `{}`},
		{name: "canonical datasource save explicit error envelope after auth", path: "/api/ds/save", body: `{"name":"demo","type":"folder"}`},
		{name: "canonical datasource update explicit error envelope after auth", path: "/api/ds/update", body: `{"id":1,"name":"demo"}`},
		{name: "canonical datasource delete explicit error envelope after auth", path: "/api/ds/delete/1", body: `{}`},
		{name: "api compatibility datasource list explicit error envelope after auth", path: "/api/datasource/list", body: `{}`},
		{name: "de2api datasource list explicit error envelope after auth", path: "/de2api/datasource/list", body: `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
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

func TestRegisterRoutes_DatasourceTableExplorationAliasesReturnSuccessEnvelopeAfterAuthentication(t *testing.T) {
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
		{name: "canonical datasource tables success envelope after auth", path: "/api/ds/tables"},
		{name: "canonical datasource table status success envelope after auth", path: "/api/ds/tableStatus"},
		{name: "canonical datasource table field success envelope after auth", path: "/api/ds/tableField"},
		{name: "canonical datasource preview data success envelope after auth", path: "/api/ds/previewData"},
		{name: "api compatibility datasource tables success envelope after auth", path: "/api/datasource/getTables"},
		{name: "api compatibility datasource table status success envelope after auth", path: "/api/datasource/getTableStatus"},
		{name: "api compatibility datasource table field success envelope after auth", path: "/api/datasource/getTableField"},
		{name: "api compatibility datasource preview data success envelope after auth", path: "/api/datasource/previewData"},
		{name: "de2api datasource tables success envelope after auth", path: "/de2api/datasource/getTables"},
		{name: "de2api datasource table status success envelope after auth", path: "/de2api/datasource/getTableStatus"},
		{name: "de2api datasource table field success envelope after auth", path: "/de2api/datasource/getTableField"},
		{name: "de2api datasource preview data success envelope after auth", path: "/de2api/datasource/previewData"},
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
				Code string          `json:"code"`
				Msg  string          `json:"msg"`
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}

			if resp.Code != "000000" {
				t.Fatalf("expected code 000000 for %s, got %s", tt.path, resp.Code)
			}
			if resp.Msg != "success" {
				t.Fatalf("expected success msg for %s, got %q", tt.path, resp.Msg)
			}
			if len(resp.Data) == 0 {
				t.Fatalf("expected non-empty data payload for %s, got empty", tt.path)
			}
		})
	}
}

func TestRegisterRoutes_DatasourcePreviewAndSyncAliasesReturnExplicitErrorAfterAuthenticationWhenStoreUnavailable(t *testing.T) {
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
		body string
		msg  string
	}{
		{name: "canonical datasource sync api table explicit error envelope after auth", path: "/api/ds/syncApiTable", body: `{}`, msg: "Failed to sync api table: request is required"},
		{name: "canonical datasource sync api ds explicit error envelope after auth", path: "/api/ds/syncApiDs", body: `{}`, msg: "Failed to sync api datasource: request is required"},
		{name: "api compatibility datasource sync api table explicit error envelope after auth", path: "/api/datasource/syncApiTable", body: `{}`, msg: "Failed to sync api table: request is required"},
		{name: "api compatibility datasource sync api ds explicit error envelope after auth", path: "/api/datasource/syncApiDs", body: `{}`, msg: "Failed to sync api datasource: request is required"},
		{name: "de2api datasource sync api table explicit error envelope after auth", path: "/de2api/datasource/syncApiTable", body: `{}`, msg: "Failed to sync api table: request is required"},
		{name: "de2api datasource sync api ds explicit error envelope after auth", path: "/de2api/datasource/syncApiDs", body: `{}`, msg: "Failed to sync api datasource: request is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
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
			if resp.Msg != tt.msg {
				t.Fatalf("expected message %q for %s, got %q", tt.msg, tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_DatasourceFileIngestAliasesReturnExplicitErrorsAfterAuthentication(t *testing.T) {
	router := newRouterWithJWTConfig()
	router.RegisterRoutes()

	jwtInstance := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "test-secret", Expire: 3600})
	token, err := jwtInstance.GenerateToken(1, "admin", "admin")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	tests := []struct {
		name          string
		path          string
		contentType   string
		body          string
		expectContain string
	}{
		{name: "canonical datasource load remote file explicit error after auth", path: "/api/ds/loadRemoteFile", contentType: "application/json", body: `{"url":"://bad"}`, expectContain: "Failed to load remote file:"},
		{name: "api compatibility datasource load remote file explicit error after auth", path: "/api/datasource/loadRemoteFile", contentType: "application/json", body: `{"url":"://bad"}`, expectContain: "Failed to load remote file:"},
		{name: "de2api datasource load remote file explicit error after auth", path: "/de2api/datasource/loadRemoteFile", contentType: "application/json", body: `{"url":"://bad"}`, expectContain: "Failed to load remote file:"},
		{name: "canonical datasource upload file explicit error after auth", path: "/api/ds/uploadFile", contentType: "application/json", body: `{}`, expectContain: "Failed to get uploaded file:"},
		{name: "api compatibility datasource upload file explicit error after auth", path: "/api/datasource/uploadFile", contentType: "application/json", body: `{}`, expectContain: "Failed to get uploaded file:"},
		{name: "de2api datasource upload file explicit error after auth", path: "/de2api/datasource/uploadFile", contentType: "application/json", body: `{}`, expectContain: "Failed to get uploaded file:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
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
			if !strings.Contains(resp.Msg, tt.expectContain) {
				t.Fatalf("expected message containing %q for %s, got %q", tt.expectContain, tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_DatasourceCanonicalGetRouteContracts(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	t.Run("canonical datasource get invalid id returns explicit error envelope", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/ds/not-a-number", nil)
		w := httptest.NewRecorder()

		router.Engine().ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}

		var resp struct {
			Code string `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
		if resp.Code != "500000" {
			t.Fatalf("expected code 500000, got %s", resp.Code)
		}
		if !strings.Contains(resp.Msg, "Invalid datasource ID") {
			t.Fatalf("expected invalid datasource id message, got %q", resp.Msg)
		}
	})
}

func TestRegisterRoutes_DatasourceCanonicalGetRequiresAuthentication(t *testing.T) {
	router := newRouterWithJWTConfig()
	router.RegisterRoutes()

	req := httptest.NewRequest("GET", "/api/ds/1", nil)
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
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "20001" {
		t.Fatalf("expected code 20001, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "authorization") {
		t.Fatalf("expected authorization message, got %q", resp.Msg)
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

func TestRegisterRoutes_ChartCanonicalContracts(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		wantCode   string
		msgPart    string
	}{
		{name: "chart query remains handler-owned", path: "/api/chart/query", body: "{", wantStatus: 200, wantCode: "500000", msgPart: "Invalid request"},
		{name: "chart data now requires governed auth", path: "/api/chart/data", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d for %s, got %d with body %s", tt.wantStatus, tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}
			if resp.Code != tt.wantCode {
				t.Fatalf("expected code %s for %s, got %s", tt.wantCode, tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, tt.msgPart) {
				t.Fatalf("expected message containing %q for %s, got %q", tt.msgPart, tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_ChartCompatibilityContracts(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		wantCode   string
		msgPart    string
	}{
		{name: "compat chartData getData now requires governed auth", path: "/api/chartData/getData", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "compat chart getData now requires governed auth", path: "/api/chart/getData", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "compat chart listByDQ now requires governed auth", path: "/api/chart/listByDQ/11/9", body: `{"type":"bar"}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d for %s, got %d with body %s", tt.wantStatus, tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}
			if resp.Code != tt.wantCode {
				t.Fatalf("expected code %s for %s, got %s", tt.wantCode, tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, tt.msgPart) {
				t.Fatalf("expected message containing %q for %s, got %q", tt.msgPart, tt.path, resp.Msg)
			}
		})
	}
}

func TestRegisterRoutes_RootChartCompatibilityContracts(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		wantCode   string
		msgPart    string
	}{
		{name: "root chartData getData now requires governed auth", path: "/chartData/getData", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root chart getData now requires governed auth", path: "/chart/getData", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root chart listByDQ now requires governed auth", path: "/chart/listByDQ/11/9", body: `{"type":"bar"}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d for %s, got %d with body %s", tt.wantStatus, tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}
			if resp.Code != tt.wantCode {
				t.Fatalf("expected code %s for %s, got %s", tt.wantCode, tt.path, resp.Code)
			}
			if !strings.Contains(resp.Msg, tt.msgPart) {
				t.Fatalf("expected message containing %q for %s, got %q", tt.msgPart, tt.path, resp.Msg)
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
		body       string
		wantStatus int
		wantCode   string
		msgPart    string
	}{
		{name: "api visualization tree", path: "/api/dataVisualization/tree", body: "{", wantStatus: 200, wantCode: "500000", msgPart: "Invalid request"},
		{name: "api visualization detail", path: "/api/dataVisualization/findById", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "api visualization saveCanvas now governed by parent scope", path: "/api/dataVisualization/saveCanvas", body: `{"pid":123,"type":"dashboard"}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "api visualization updateCanvas", path: "/api/dataVisualization/updateCanvas", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "api visualization deleteLogic", path: "/api/dataVisualization/deleteLogic/123", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "de2api visualization tree", path: "/de2api/dataVisualization/tree", body: "{", wantStatus: 200, wantCode: "500000", msgPart: "Invalid request"},
		{name: "de2api visualization detail", path: "/de2api/dataVisualization/findById", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization tree remains ungated", path: "/dataVisualization/tree", body: "{", wantStatus: 200, wantCode: "500000", msgPart: "Invalid request"},
		{name: "root visualization detail now governed", path: "/dataVisualization/findById", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization save alias now governed by parent scope", path: "/dataVisualization/save", body: `{"pid":123,"type":"dashboard","name":"demo"}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization copy now governed by source and destination scopes", path: "/dataVisualization/copy", body: `{"id":123,"pid":456,"name":"copy","type":"dashboard"}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization saveCanvas now governed by parent scope", path: "/dataVisualization/saveCanvas", body: `{"pid":123,"type":"dashboard","name":"demo"}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization updateBase now governed", path: "/dataVisualization/updateBase", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization move now governed", path: "/dataVisualization/move", body: `{"id":123,"pid":456}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization updatePublishStatus now governed", path: "/dataVisualization/updatePublishStatus", body: `{"id":123,"status":1}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization recoverToPublished now governed", path: "/dataVisualization/recoverToPublished", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization updateCanvas now governed", path: "/dataVisualization/updateCanvas", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization deleteLogic now governed", path: "/dataVisualization/deleteLogic/123", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
		{name: "root visualization deleteLogic with busiFlag now governed", path: "/dataVisualization/deleteLogic/123/dashboard", body: `{"id":123}`, wantStatus: 401, wantCode: "20001", msgPart: "authentication required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
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

func TestRegisterRoutes_DatasourceUIMetadataCanonicalRoutes(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	t.Run("GET /api/ds/types returns hardcoded type list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/ds/types", nil)
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
		var resp struct {
			Code string              `json:"code"`
			Data []map[string]string `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
		if resp.Code != "000000" {
			t.Fatalf("expected code 000000, got %s", resp.Code)
		}
		if len(resp.Data) != 5 {
			t.Fatalf("expected 5 types, got %d", len(resp.Data))
		}
	})

	t.Run("GET /api/ds/showFinishPage returns success envelope", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/ds/showFinishPage", nil)
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
	})

	t.Run("POST /api/ds/showFinishPage returns success envelope", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/ds/showFinishPage", nil)
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
	})

	t.Run("POST /api/ds/latestUse returns success envelope", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/ds/latestUse", nil)
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
	})

	t.Run("POST /api/ds/syncRecord/1/1/10 returns success envelope", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/ds/syncRecord/1/1/10", nil)
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
	})
}

func TestRegisterRoutes_DatasetTreeCRUDCanonicalRoutes(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	t.Run("POST /api/dataset/save returns success envelope", func(t *testing.T) {
		body := strings.NewReader(`{"name":"test","nodeType":"dataset"}`)
		req := httptest.NewRequest("POST", "/api/dataset/save", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
	})

	t.Run("POST /api/dataset/create returns success envelope", func(t *testing.T) {
		body := strings.NewReader(`{"name":"test","nodeType":"dataset"}`)
		req := httptest.NewRequest("POST", "/api/dataset/create", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
	})

	t.Run("POST /api/dataset/rename returns error for missing id", func(t *testing.T) {
		body := strings.NewReader(`{"name":"test"}`)
		req := httptest.NewRequest("POST", "/api/dataset/rename", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if resp["code"] != "500000" {
			t.Fatalf("expected error code 500000, got %v", resp["code"])
		}
	})

	t.Run("POST /api/dataset/move returns error for missing id", func(t *testing.T) {
		body := strings.NewReader(`{}`)
		req := httptest.NewRequest("POST", "/api/dataset/move", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if resp["code"] != "500000" {
			t.Fatalf("expected error code 500000, got %v", resp["code"])
		}
	})

	t.Run("POST /api/dataset/delete/1 returns success envelope", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/dataset/delete/1", nil)
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
	})

	t.Run("POST /api/dataset/perDelete/1 returns success envelope", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/dataset/perDelete/1", nil)
		w := httptest.NewRecorder()
		router.Engine().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d with body %s", w.Code, w.Body.String())
		}
	})
}

func TestRegisterRoutes_DatasetQueryCanonicalRoutes(t *testing.T) {
	router := NewRouter(nil, nil)
	router.RegisterRoutes()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "canonical dataset get detail", method: "POST", path: "/api/dataset/get/1", body: ""},
		{name: "canonical dataset get detail invalid id", method: "POST", path: "/api/dataset/get/not-a-number", body: ""},
		{name: "canonical dataset details", method: "POST", path: "/api/dataset/details/1", body: ""},
		{name: "canonical dataset dsDetails", method: "POST", path: "/api/dataset/dsDetails", body: `{}`},
		{name: "canonical dataset getSqlParams", method: "POST", path: "/api/dataset/getSqlParams", body: `{}`},
		{name: "canonical dataset barInfo", method: "GET", path: "/api/dataset/barInfo/1", body: ""},
		{name: "canonical dataset barInfo invalid id", method: "GET", path: "/api/dataset/barInfo/not-a-number", body: ""},
		{name: "canonical dataset getDatasetTotal", method: "POST", path: "/api/dataset/getDatasetTotal", body: `{"id":1}`},
		{name: "canonical dataset previewSql", method: "POST", path: "/api/dataset/previewSql", body: `{}`},
		{name: "canonical dataset enumValueObj", method: "POST", path: "/api/dataset/enumValueObj", body: `{}`},
		{name: "canonical dataset enumValueDs", method: "POST", path: "/api/dataset/enumValueDs", body: `{}`},
		{name: "canonical dataset enumValue", method: "POST", path: "/api/dataset/enumValue", body: `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected status 200 for %s, got %d with body %s", tt.path, w.Code, w.Body.String())
			}

			var resp struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response for %s failed: %v; body=%s", tt.path, err, w.Body.String())
			}
		})
	}
}
