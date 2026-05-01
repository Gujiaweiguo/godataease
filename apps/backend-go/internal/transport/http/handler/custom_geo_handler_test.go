package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomGeoHandler_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCustomGeoRoutes(r.Group("/api"), NewCustomGeoHandler(nil))

	tests := []struct {
		name         string
		method       string
		url          string
		body         string
		expectedCode string
	}{
		{name: "save_geo_area_empty_body", method: "POST", url: "/api/customGeo/geoArea/save", body: "", expectedCode: "400000"},
		{name: "save_geo_sub_area_empty_body", method: "POST", url: "/api/customGeo/geoSubArea/save", body: "", expectedCode: "400000"},
		{name: "delete_geo_sub_area_invalid_id", method: "DELETE", url: "/api/customGeo/geoSubArea/abc", expectedCode: "400000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCustomGeoHandlerErrorResponse(t, r, tt.method, tt.url, tt.body, tt.expectedCode)
		})
	}
}

func TestCustomGeoHandler_NilRepoRoutesReachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCustomGeoRoutes(r.Group("/api"), NewCustomGeoHandler(nil))

	tests := []struct {
		name   string
		method string
		url    string
	}{
		{name: "list_geo_areas_registered", method: "GET", url: "/api/customGeo/geoArea/list"},
		{name: "get_geo_area_registered", method: "GET", url: "/api/customGeo/geoArea/test-id"},
		{name: "delete_geo_area_registered", method: "DELETE", url: "/api/customGeo/geoArea/test-id"},
		{name: "list_sub_area_options_registered", method: "GET", url: "/api/customGeo/geoSubArea/options"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCustomGeoHandlerRouteReachableWithRecoveredPanic(t, r, tt.method, tt.url, "")
		})
	}
}

func assertCustomGeoHandlerErrorResponse(t *testing.T, r *gin.Engine, method, url, body, expectedCode string) {
	t.Helper()

	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, expectedCode, resp["code"])
}

func assertCustomGeoHandlerRouteReachableWithRecoveredPanic(t *testing.T, r *gin.Engine, method, url, body string) {
	t.Helper()

	req := httptest.NewRequest(method, url, strings.NewReader(body))
	w := httptest.NewRecorder()

	didPanic := false
	func() {
		defer func() {
			if recover() != nil {
				didPanic = true
			}
		}()
		r.ServeHTTP(w, req)
	}()

	assert.NotEqual(t, 404, w.Code)
	assert.True(t, didPanic || w.Code == 200)
}
