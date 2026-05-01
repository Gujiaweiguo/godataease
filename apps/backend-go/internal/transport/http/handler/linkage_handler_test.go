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

func assertHandlerErrorResponse(t *testing.T, r *gin.Engine, method, url, body string) {
	t.Helper()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func assertHandlerRouteReachableWithRecoveredPanic(t *testing.T, r *gin.Engine, method, url, body string) {
	t.Helper()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
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

func TestLinkageHandler_InvalidInputsAndRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkageHandler(nil)
	r := gin.New()
	RegisterLinkageRoutes(r.Group("/api"), h)

	t.Run("invalid request body", func(t *testing.T) {
		for _, route := range []string{
			"/api/linkage/getViewLinkageGather",
			"/api/linkage/getViewLinkageGatherArray",
			"/api/linkage/saveLinkage",
			"/api/linkage/updateLinkageActive",
			"/api/linkage/removeLinkage",
		} {
			assertHandlerErrorResponse(t, r, "POST", route, "{")
		}
	})

	t.Run("invalid dvId", func(t *testing.T) {
		assertHandlerErrorResponse(t, r, "GET", "/api/linkage/getVisualizationAllLinkageInfo/not-a-number/resource_table", "")
	})

	t.Run("registered routes reachable", func(t *testing.T) {
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/linkage/getViewLinkageGather", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/linkage/getViewLinkageGatherArray", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/linkage/saveLinkage", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/linkage/getVisualizationAllLinkageInfo/1/resource_table", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/linkage/updateLinkageActive", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/linkage/removeLinkage", "{}")
	})
}
