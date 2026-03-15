package handler

import (
	"dataease/backend/internal/service"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemVariableHandler_InvalidRequestSmoke(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSystemVariableRoutes(r.Group("/api"), NewSystemVariableHandler(service.NewSystemVariableService(nil)))

	t.Run("create_invalid", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/sysVariable/create", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})

	t.Run("detail_invalid_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sysVariable/detail/bad", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})

	t.Run("value_page_invalid_page", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/sysVariable/value/selected/0/10", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})
}
