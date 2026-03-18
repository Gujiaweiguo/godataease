package http

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

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
