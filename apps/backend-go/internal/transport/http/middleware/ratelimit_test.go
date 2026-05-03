package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dataease/backend/internal/app"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit_RejectsRequestsOverBudget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/limited", RateLimit("test", 2, time.Hour, ClientIPKey), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	for range 2 {
		req := httptest.NewRequest(http.MethodGet, "/limited", nil)
		req.RemoteAddr = "192.0.2.10:1234"
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code)
	assert.Contains(t, resp.Body.String(), `"code":"429001"`)
}

func TestAuthenticatedUserKey_FallsBackToClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "198.51.100.20:4321"

	assert.Equal(t, "198.51.100.20", AuthenticatedUserKey(c))

	c.Set("user_id", uint64(42))
	assert.Equal(t, "42", AuthenticatedUserKey(c))
}

func TestInMemoryBackend_AllowReportsRemaining(t *testing.T) {
	backend := newInMemoryBackend()

	allowed, remaining, resetAt := backend.Allow("user-1", 2, time.Minute)
	assert.True(t, allowed)
	assert.Equal(t, 1, remaining)
	assert.WithinDuration(t, time.Now().Add(time.Minute), resetAt, 2*time.Second)

	allowed, remaining, _ = backend.Allow("user-1", 2, time.Minute)
	assert.True(t, allowed)
	assert.Equal(t, 0, remaining)

	allowed, remaining, _ = backend.Allow("user-1", 2, time.Minute)
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining)
}

func TestNewRateLimiterBackend_FallsBackToInMemory(t *testing.T) {
	backend := NewRateLimiterBackend(app.RateLimitConfig{UseRedis: true}, nil)
	_, ok := backend.(*InMemoryBackend)
	assert.True(t, ok)
}
