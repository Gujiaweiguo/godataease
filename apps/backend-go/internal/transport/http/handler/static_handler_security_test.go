package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func staticTestJWT() *auth.JWT {
	return auth.NewJWT(&auth.JWTConfig{Secret: "test-secret-key-for-unittest", Expire: 3600})
}

func TestStaticHandler_FindResourceAsBase64_PathValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	r := gin.New()
	r.POST("/staticResource/findResourceAsBase64", h.FindResourceAsBase64)

	body := `{"resourcePathList":["../../etc/passwd","safe.png"]}`
	req := httptest.NewRequest(http.MethodPost, "/staticResource/findResourceAsBase64", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string            `json:"code"`
		Data map[string]string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, "", resp.Data["../../etc/passwd"])
}

func TestRegisterStaticAndFontRoutesRequireAuthOnProtectedAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtInstance := staticTestJWT()
	r := gin.New()
	protected := r.Group("/api")
	protected.Use(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"code": "20001", "msg": "missing authorization header"})
			c.Abort()
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")
		if _, err := jwtInstance.ParseToken(token); err != nil {
			c.JSON(401, gin.H{"code": "20001", "msg": "invalid token"})
			c.Abort()
			return
		}
		c.Next()
	})

	RegisterStaticRoutes(protected, NewStaticHandler(nil))
	RegisterFontRoutes(protected, NewFontHandler(nil))
	RegisterFontDownloadRoute(r.Group("/api"), NewFontHandler(nil))

	t.Run("static route rejects missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/staticResource/list", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("font route rejects missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/typeface/listFont", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("font download remains reachable without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/typeface/download/demo.ttf", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	})
}
