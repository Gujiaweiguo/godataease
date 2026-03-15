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

func TestSyncHandler_InvalidInputSmoke(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSyncRoutes(r.Group("/api"), NewSyncHandler(nil))

	t.Run("task_get_invalid_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sync/task/get/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})

	t.Run("task_log_invalid_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sync/task/log/detail/abc/0", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})

	t.Run("task_pager_invalid_page", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/sync/task/pager/0/10", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})

	t.Run("datasource_batch_del_invalid_id", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/sync/datasource/batchDel", strings.NewReader("[\"bad\"]"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})
}
