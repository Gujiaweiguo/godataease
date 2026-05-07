package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditHandler_DownloadExportFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	r := gin.New()
	RegisterAuditRoutes(r.Group(""), h, nil)

	t.Run("rejects missing path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/audit/download", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"10001"`)
	})

	t.Run("rejects path outside temp dir", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/audit/download?path=/etc/passwd&format=csv", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"10001"`)
	})

	t.Run("rejects unexpected basename", func(t *testing.T) {
		path := filepath.Join(os.TempDir(), "not_audit.csv")
		req := httptest.NewRequest(http.MethodGet, "/audit/download?path="+path+"&format=csv", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"10001"`)
	})

	t.Run("rejects format mismatch", func(t *testing.T) {
		path := filepath.Join(os.TempDir(), "audit_logs_20260101010101.csv")
		req := httptest.NewRequest(http.MethodGet, "/audit/download?path="+path+"&format=json", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"10001"`)
	})

	t.Run("returns not found for missing file", func(t *testing.T) {
		path := filepath.Join(os.TempDir(), "audit_logs_20990101010101.csv")
		req := httptest.NewRequest(http.MethodGet, "/audit/download?path="+path+"&format=csv", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"50001"`)
	})

	t.Run("serves valid temp export file", func(t *testing.T) {
		file, err := os.CreateTemp(os.TempDir(), "audit_logs_*.csv")
		require.NoError(t, err)
		defer func() { _ = os.Remove(file.Name()) }()
		_, err = file.WriteString("id,action\n1,login\n")
		require.NoError(t, err)
		require.NoError(t, file.Close())

		req := httptest.NewRequest(http.MethodGet, "/audit/download?path="+file.Name()+"&format=csv", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Disposition"), "audit_logs.csv")
		assert.Contains(t, w.Body.String(), "id,action")
	})

	t.Run("rate limits export and download routes", func(t *testing.T) {
		for i := 0; i < auditExportRateLimitRequests; i++ {
			req := httptest.NewRequest(http.MethodPost, "/audit/export", strings.NewReader("{"))
			req.RemoteAddr = "203.0.113.40:1234"
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"code":"10001"`)
		}

		req := httptest.NewRequest(http.MethodGet, "/audit/download?path=/etc/passwd&format=csv", nil)
		req.RemoteAddr = "203.0.113.40:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"429001"`)
	})
}
