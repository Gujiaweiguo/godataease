package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFontHandlerDeleteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.CoreFont{}))
	return db
}

func TestFontHandler_Delete_PathValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("removes font file inside font dir", func(t *testing.T) {
		db := setupFontHandlerDeleteTestDB(t)
		repo := repository.NewTypefaceRepository(db)
		h := NewFontHandler(repo)
		r := gin.New()
		r.POST("/typeface/delete/:id", h.Delete)

		fontDir := filepath.Join(t.TempDir(), "fonts")
		require.NoError(t, os.MkdirAll(fontDir, 0o755))
		t.Setenv("FONT_DIR", fontDir)

		fontPath := filepath.Join(fontDir, "demo.ttf")
		require.NoError(t, os.WriteFile(fontPath, []byte("font-data"), 0o644))
		require.NoError(t, repo.CreateFont(&auto.CoreFont{ID: 1, Name: "demo", FileTransName: "demo.ttf", UpdateTime: 1}))

		req := httptest.NewRequest(http.MethodPost, "/typeface/delete/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"000000"`)
		_, err := os.Stat(fontPath)
		assert.ErrorIs(t, err, os.ErrNotExist)

		var count int64
		require.NoError(t, db.Model(&auto.CoreFont{}).Where("id = ?", 1).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("skips deleting path traversal target while removing db record", func(t *testing.T) {
		db := setupFontHandlerDeleteTestDB(t)
		repo := repository.NewTypefaceRepository(db)
		h := NewFontHandler(repo)
		r := gin.New()
		r.POST("/typeface/delete/:id", h.Delete)

		baseDir := t.TempDir()
		fontDir := filepath.Join(baseDir, "fonts")
		require.NoError(t, os.MkdirAll(fontDir, 0o755))
		t.Setenv("FONT_DIR", fontDir)

		outsidePath := filepath.Join(baseDir, "escape.ttf")
		require.NoError(t, os.WriteFile(outsidePath, []byte("outside-data"), 0o644))
		require.NoError(t, repo.CreateFont(&auto.CoreFont{ID: 2, Name: "escape", FileTransName: "../escape.ttf", UpdateTime: 1}))

		req := httptest.NewRequest(http.MethodPost, "/typeface/delete/2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"000000"`)
		_, err := os.Stat(outsidePath)
		require.NoError(t, err)

		var count int64
		require.NoError(t, db.Model(&auto.CoreFont{}).Where("id = ?", 2).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("skips deleting absolute path target while removing db record", func(t *testing.T) {
		db := setupFontHandlerDeleteTestDB(t)
		repo := repository.NewTypefaceRepository(db)
		h := NewFontHandler(repo)
		r := gin.New()
		r.POST("/typeface/delete/:id", h.Delete)

		fontDir := filepath.Join(t.TempDir(), "fonts")
		require.NoError(t, os.MkdirAll(fontDir, 0o755))
		t.Setenv("FONT_DIR", fontDir)

		absolutePath := filepath.Join(t.TempDir(), "absolute.woff2")
		require.NoError(t, os.WriteFile(absolutePath, []byte("outside-data"), 0o644))
		require.NoError(t, repo.CreateFont(&auto.CoreFont{ID: 3, Name: "absolute", FileTransName: absolutePath, UpdateTime: 1}))

		req := httptest.NewRequest(http.MethodPost, "/typeface/delete/3", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"000000"`)
		_, err := os.Stat(absolutePath)
		require.NoError(t, err)

		var count int64
		require.NoError(t, db.Model(&auto.CoreFont{}).Where("id = ?", 3).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("skips deleting disallowed extension while removing db record", func(t *testing.T) {
		db := setupFontHandlerDeleteTestDB(t)
		repo := repository.NewTypefaceRepository(db)
		h := NewFontHandler(repo)
		r := gin.New()
		r.POST("/typeface/delete/:id", h.Delete)

		fontDir := filepath.Join(t.TempDir(), "fonts")
		require.NoError(t, os.MkdirAll(fontDir, 0o755))
		t.Setenv("FONT_DIR", fontDir)

		filePath := filepath.Join(fontDir, "demo.exe")
		require.NoError(t, os.WriteFile(filePath, []byte("not-font"), 0o644))
		require.NoError(t, repo.CreateFont(&auto.CoreFont{ID: 4, Name: "bad-ext", FileTransName: "demo.exe", UpdateTime: 1}))

		req := httptest.NewRequest(http.MethodPost, "/typeface/delete/4", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"000000"`)
		_, err := os.Stat(filePath)
		require.NoError(t, err)

		var count int64
		require.NoError(t, db.Model(&auto.CoreFont{}).Where("id = ?", 4).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})
}
