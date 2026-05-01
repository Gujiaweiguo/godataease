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

func TestShareHandler_EditRoutesRejectInvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterShareRoutes(r.Group("/api"), NewShareHandler(nil))

	t.Run("edit_uuid_invalid_payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/share/editUuid", strings.NewReader(`{"resourceId":"bad"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "10001", resp["code"])
	})

	t.Run("edit_exp_invalid_payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/share/editExp", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "10001", resp["code"])
	})

	t.Run("edit_pwd_invalid_payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/share/editPwd", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "10001", resp["code"])
	})
}
