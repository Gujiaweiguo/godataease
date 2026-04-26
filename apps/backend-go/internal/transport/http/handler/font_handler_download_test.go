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

	t.Run("rejects disallowed extension", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("FONT_DIR", dir)
		filePath := filepath.Join(dir, "demo.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("not-font"), 0644))

		req := httptest.NewRequest(http.MethodGet, "/typeface/download/demo.txt", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"500000"`)
	})

	t.Run("rejects missing extension", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("FONT_DIR", dir)
		filePath := filepath.Join(dir, "demo")
		require.NoError(t, os.WriteFile(filePath, []byte("font-data"), 0644))

		req := httptest.NewRequest(http.MethodGet, "/typeface/download/demo", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"500000"`)
	})

	t.Run("rejects disguised backup suffix", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("FONT_DIR", dir)
		filePath := filepath.Join(dir, "demo.woff2.bak")
		require.NoError(t, os.WriteFile(filePath, []byte("font-data"), 0644))

		req := httptest.NewRequest(http.MethodGet, "/typeface/download/demo.woff2.bak", nil)
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

	t.Run("serves allowed lowercase font extensions", func(t *testing.T) {
		for _, name := range []string{"demo.ttf", "demo.otf", "demo.woff", "demo.woff2"} {
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				t.Setenv("FONT_DIR", dir)
				fontPath := filepath.Join(dir, name)
				require.NoError(t, os.WriteFile(fontPath, []byte("font-data"), 0644))

				req := httptest.NewRequest(http.MethodGet, "/typeface/download/"+name, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Contains(t, w.Header().Get("Content-Disposition"), name)
				assert.Equal(t, "font-data", w.Body.String())
			})
		}
	})

	t.Run("serves allowed uppercase extension", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("FONT_DIR", dir)
		fontPath := filepath.Join(dir, "demo.WOFF2")
		require.NoError(t, os.WriteFile(fontPath, []byte("font-data"), 0644))

		req := httptest.NewRequest(http.MethodGet, "/typeface/download/demo.WOFF2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Disposition"), "demo.WOFF2")
		assert.Equal(t, "font-data", w.Body.String())
	})
}
