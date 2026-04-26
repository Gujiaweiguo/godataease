package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFontHandler_Download_PathValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	r := gin.New()
	r.GET("/typeface/download/:file", h.Download)

	t.Run("rejects path traversal", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/typeface/download/..", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"500000"`)
	})

	t.Run("serves file inside font dir", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("FONT_DIR", dir)
		fontPath := filepath.Join(dir, "demo.ttf")
		require.NoError(t, os.WriteFile(fontPath, []byte("font-data"), 0644))

		req := httptest.NewRequest(http.MethodGet, "/typeface/download/demo.ttf", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Disposition"), "demo.ttf")
		assert.Equal(t, "font-data", w.Body.String())
	})
}
