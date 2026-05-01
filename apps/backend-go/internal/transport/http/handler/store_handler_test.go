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

func TestStoreHandler_ExecuteAndFavorited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterStoreRoutes(r, NewStoreHandler(nil), false)

	t.Run("execute_invalid_body", func(t *testing.T) {
		assertStoreHandlerErrorResponse(t, r, "POST", "/store/execute", "", 200, "10001", "Invalid request")
	})

	t.Run("execute_unauthorized", func(t *testing.T) {
		assertStoreHandlerErrorResponse(t, r, "POST", "/store/execute", `{"id":1,"type":"PANEL"}`, 401, "20001", "Unauthorized")
	})

	t.Run("favorited_invalid_id", func(t *testing.T) {
		assertStoreHandlerErrorResponse(t, r, "GET", "/store/favorited/abc", "", 200, "10001", "Invalid id")
	})

	t.Run("favorited_no_user_returns_false", func(t *testing.T) {
		assertStoreHandlerSuccessData(t, r, "GET", "/store/favorited/1", "", false)
	})
}

func TestStoreHandler_Query(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterStoreRoutes(r, NewStoreHandler(nil), false)

	t.Run("query_invalid_body", func(t *testing.T) {
		assertStoreHandlerErrorResponse(t, r, "POST", "/store/query", "", 200, "10001", "Invalid request")
	})

	t.Run("query_no_user_returns_empty_array", func(t *testing.T) {
		assertStoreHandlerSuccessData(t, r, "POST", "/store/query", `{"type":"PANEL"}`, []any{})
	})
}

func TestStoreHandler_QueryRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("skip_query_true_returns_404", func(t *testing.T) {
		r := gin.New()
		RegisterStoreRoutes(r, NewStoreHandler(nil), true)

		req := httptest.NewRequest("POST", "/store/query", strings.NewReader(`{"type":"PANEL"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("skip_query_false_route_registered", func(t *testing.T) {
		r := gin.New()
		RegisterStoreRoutes(r, NewStoreHandler(nil), false)

		req := httptest.NewRequest("POST", "/store/query", strings.NewReader(`{"type":"PANEL"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.NotEqual(t, 404, w.Code)
		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp["code"])
	})
}

func assertStoreHandlerErrorResponse(t *testing.T, r *gin.Engine, method, url, body string, expectedStatus int, expectedCode, expectedMessage string) {
	t.Helper()

	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, expectedStatus, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, expectedCode, resp["code"])
	assert.Contains(t, resp["msg"], expectedMessage)
}

func assertStoreHandlerSuccessData(t *testing.T, r *gin.Engine, method, url, body string, expectedData any) {
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
	assert.Equal(t, "000000", resp["code"])
	assert.Equal(t, expectedData, resp["data"])
}
