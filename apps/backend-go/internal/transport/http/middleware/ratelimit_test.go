package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"dataease/backend/internal/app"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRedisBackend_SharesBudgetAcrossCalls(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	backend := &RedisBackend{client: client, now: time.Now}

	allowed, remaining, _ := backend.Allow("shared-budget", 2, time.Minute)
	assert.True(t, allowed)
	assert.Equal(t, 1, remaining)

	allowed, remaining, _ = backend.Allow("shared-budget", 2, time.Minute)
	assert.True(t, allowed)
	assert.Equal(t, 0, remaining)

	allowed, remaining, _ = backend.Allow("shared-budget", 2, time.Minute)
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining)
}

func TestRedisBackend_ResetsAfterExpiry(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	backend := &RedisBackend{client: client, now: time.Now}

	allowed, _, _ := backend.Allow("expiring-budget", 1, time.Minute)
	require.True(t, allowed)

	allowed, _, _ = backend.Allow("expiring-budget", 1, time.Minute)
	assert.False(t, allowed)

	mr.FastForward(time.Minute)

	allowed, remaining, _ := backend.Allow("expiring-budget", 1, time.Minute)
	assert.True(t, allowed)
	assert.Equal(t, 0, remaining)
}

func TestRedisBackend_FailsOpenOnRedisError(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	backend := &RedisBackend{client: client, now: time.Now}
	mr.Close()
	t.Cleanup(func() { _ = client.Close() })

	allowed, remaining, resetAt := backend.Allow("redis-error", 3, time.Minute)
	assert.True(t, allowed)
	assert.Equal(t, 3, remaining)
	assert.WithinDuration(t, time.Now().Add(time.Minute), resetAt, 2*time.Second)
}

func TestConfigurableRateLimit_SetsHeadersForAllowedAndRejectedRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	backend := newInMemoryBackend()
	r.GET("/limited", ConfigurableRateLimit("headers", 2, time.Minute, backend, ClientIPKey), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	request := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/limited", nil)
		req.RemoteAddr = "192.0.2.88:1234"
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		return resp
	}

	first := request()
	assert.Equal(t, http.StatusOK, first.Code)
	assert.Equal(t, "2", first.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "1", first.Header().Get("X-RateLimit-Remaining"))
	_, err := strconv.ParseInt(first.Header().Get("X-RateLimit-Reset"), 10, 64)
	require.NoError(t, err)
	assert.Empty(t, first.Header().Get("Retry-After"))

	second := request()
	assert.Equal(t, http.StatusOK, second.Code)
	assert.Equal(t, "0", second.Header().Get("X-RateLimit-Remaining"))

	third := request()
	assert.Equal(t, http.StatusTooManyRequests, third.Code)
	assert.Equal(t, "2", third.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", third.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, third.Header().Get("Retry-After"))
	_, err = strconv.Atoi(third.Header().Get("Retry-After"))
	require.NoError(t, err)
}

func TestResolveRouteLimit(t *testing.T) {
	disabled := false
	tests := []struct {
		name        string
		cfg         app.RateLimitConfig
		routeName   string
		defMax      int
		defWindow   time.Duration
		wantEnabled bool
		wantMax     int
		wantWindow  time.Duration
	}{
		{
			name: "no overrides uses defaults", cfg: app.RateLimitConfig{}, routeName: "login",
			defMax: 100, defWindow: time.Minute, wantEnabled: true, wantMax: 100, wantWindow: time.Minute,
		},
		{
			name: "override max and window", cfg: app.RateLimitConfig{
				RouteOverrides: map[string]app.RouteLimitConfig{
					"login": {MaxRequests: 10, WindowSeconds: 30},
				},
			}, routeName: "login",
			defMax: 100, defWindow: time.Minute, wantEnabled: true, wantMax: 10, wantWindow: 30 * time.Second,
		},
		{
			name: "override disabled", cfg: app.RateLimitConfig{
				RouteOverrides: map[string]app.RouteLimitConfig{
					"login": {Enabled: &disabled},
				},
			}, routeName: "login",
			defMax: 100, defWindow: time.Minute, wantEnabled: false, wantMax: 100, wantWindow: time.Minute,
		},
		{
			name: "missing override uses defaults", cfg: app.RateLimitConfig{
				RouteOverrides: map[string]app.RouteLimitConfig{
					"other": {MaxRequests: 5},
				},
			}, routeName: "unknown",
			defMax: 50, defWindow: 30 * time.Second, wantEnabled: true, wantMax: 50, wantWindow: 30 * time.Second,
		},
		{
			name: "zero defaults get floored", cfg: app.RateLimitConfig{}, routeName: "x",
			defMax: 0, defWindow: 0, wantEnabled: true, wantMax: 1, wantWindow: time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enabled, maxReq, window := ResolveRouteLimit(tt.cfg, tt.routeName, tt.defMax, tt.defWindow)
			assert.Equal(t, tt.wantEnabled, enabled)
			assert.Equal(t, tt.wantMax, maxReq)
			assert.Equal(t, tt.wantWindow, window)
		})
	}
}

func TestRedisBackend_WithContext_NilReceiver(t *testing.T) {
	var b *RedisBackend
	assert.Nil(t, b.WithContext(context.Background()))
}

func TestRedisBackend_WithContext_SetsContext(t *testing.T) {
	ctx := context.Background()
	b := &RedisBackend{client: nil, now: time.Now}
	b2 := b.WithContext(ctx)
	require.NotNil(t, b2)
	assert.Equal(t, ctx, b2.ctx)
	assert.Nil(t, b.ctx, "original should be unchanged")
}

func TestRedisBackend_WithContext_PreservesOtherFields(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	b := &RedisBackend{client: client, now: time.Now}
	b2 := b.WithContext(ctx)
	require.NotNil(t, b2)
	assert.Equal(t, client, b2.client)
}
