package cache

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"dataease/backend/internal/app"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestCache creates an in-memory miniredis instance and sets the package-level
// client variable. It restores the original client on cleanup.
func setupTestCache(t *testing.T) (*miniredis.Miniredis, context.Context) {
	t.Helper()
	mr := miniredis.RunT(t)

	origClient := client
	t.Cleanup(func() { client = origClient })

	client = redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return mr, context.Background()
}

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------

func TestInit_Success(t *testing.T) {
	mr := miniredis.RunT(t)

	origClient := client
	t.Cleanup(func() { client = origClient })

	host, portStr, _ := net.SplitHostPort(mr.Addr())
	port, _ := strconv.Atoi(portStr)
	cfg := &app.RedisConfig{
		Host: host,
		Port: port,
		DB:   0,
	}

	gotClient, err := Init(cfg)
	require.NoError(t, err)
	require.NotNil(t, gotClient)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	assert.NoError(t, gotClient.Ping(ctx).Err())
}

func TestInit_Failure(t *testing.T) {
	origClient := client
	t.Cleanup(func() { client = origClient })

	cfg := &app.RedisConfig{
		Host: "127.0.0.1",
		Port: 1, // nobody listening here
		DB:   0,
	}

	gotClient, err := Init(cfg)
	assert.Nil(t, gotClient)
	assert.ErrorContains(t, err, "failed to connect redis")
}

// ---------------------------------------------------------------------------
// GetClient
// ---------------------------------------------------------------------------

func TestGetClient_AfterInit(t *testing.T) {
	_, _ = setupTestCache(t)

	got := GetClient()
	assert.NotNil(t, got)
}

func TestGetClient_Nil(t *testing.T) {
	origClient := client
	t.Cleanup(func() { client = origClient })
	client = nil

	assert.Nil(t, GetClient())
}

// ---------------------------------------------------------------------------
// Close
// ---------------------------------------------------------------------------

func TestClose_WithClient(t *testing.T) {
	_, _ = setupTestCache(t)

	assert.NoError(t, Close())
	// After close, client reference is still set but connection is closed.
	// Reset so subsequent tests aren't affected — t.Cleanup already restores origClient.
}

func TestClose_NilClient(t *testing.T) {
	origClient := client
	t.Cleanup(func() { client = origClient })
	client = nil

	assert.NoError(t, Close())
}

// ---------------------------------------------------------------------------
// Set / Get
// ---------------------------------------------------------------------------

func TestSet_and_Get(t *testing.T) {
	_, ctx := setupTestCache(t)

	err := Set(ctx, "foo", "bar", 0)
	require.NoError(t, err)

	val, err := Get(ctx, "foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", val)
}

func TestGet_MissingKey(t *testing.T) {
	_, ctx := setupTestCache(t)

	val, err := Get(ctx, "nonexistent")
	assert.ErrorIs(t, err, redis.Nil)
	assert.Empty(t, val)
}

func TestSet_WithExpiration(t *testing.T) {
	mr, ctx := setupTestCache(t)

	err := Set(ctx, "ttlkey", "val", 5*time.Second)
	require.NoError(t, err)

	val, err := Get(ctx, "ttlkey")
	assert.NoError(t, err)
	assert.Equal(t, "val", val)

	// Verify TTL is set via miniredis
	ttl := mr.TTL("ttlkey")
	assert.True(t, ttl > 0, "expected positive TTL, got %v", ttl)
}

// ---------------------------------------------------------------------------
// Del
// ---------------------------------------------------------------------------

func TestDel_SingleKey(t *testing.T) {
	_, ctx := setupTestCache(t)

	require.NoError(t, Set(ctx, "k1", "v1", 0))
	require.NoError(t, Del(ctx, "k1"))

	_, err := Get(ctx, "k1")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestDel_MultipleKeys(t *testing.T) {
	_, ctx := setupTestCache(t)

	require.NoError(t, Set(ctx, "k1", "v1", 0))
	require.NoError(t, Set(ctx, "k2", "v2", 0))
	require.NoError(t, Set(ctx, "k3", "v3", 0))

	require.NoError(t, Del(ctx, "k1", "k2"))

	_, err1 := Get(ctx, "k1")
	assert.ErrorIs(t, err1, redis.Nil)

	_, err2 := Get(ctx, "k2")
	assert.ErrorIs(t, err2, redis.Nil)

	// k3 should still exist
	val, err3 := Get(ctx, "k3")
	assert.NoError(t, err3)
	assert.Equal(t, "v3", val)
}

func TestDel_NonExistentKey(t *testing.T) {
	_, ctx := setupTestCache(t)

	assert.NoError(t, Del(ctx, "does_not_exist"))
}

// ---------------------------------------------------------------------------
// DelByPattern
// ---------------------------------------------------------------------------

func TestDelByPattern_Basic(t *testing.T) {
	_, ctx := setupTestCache(t)

	require.NoError(t, Set(ctx, "prefix:a", "1", 0))
	require.NoError(t, Set(ctx, "prefix:b", "2", 0))
	require.NoError(t, Set(ctx, "other:c", "3", 0))

	err := DelByPattern(ctx, "prefix:*")
	require.NoError(t, err)

	_, err1 := Get(ctx, "prefix:a")
	assert.ErrorIs(t, err1, redis.Nil)

	_, err2 := Get(ctx, "prefix:b")
	assert.ErrorIs(t, err2, redis.Nil)

	// "other:c" must survive
	val, err3 := Get(ctx, "other:c")
	assert.NoError(t, err3)
	assert.Equal(t, "3", val)
}

func TestDelByPattern_NoMatch(t *testing.T) {
	_, ctx := setupTestCache(t)

	err := DelByPattern(ctx, "nomatch:*")
	assert.NoError(t, err)
}

func TestDelByPattern_MultiplePages(t *testing.T) {
	_, ctx := setupTestCache(t)

	// Create 150 keys with a common prefix — SCAN uses count=100 so this
	// forces at least two cursor iterations.
	const n = 150
	for i := 0; i < n; i++ {
		require.NoError(t, Set(ctx, fmt.Sprintf("mp:key:%d", i), "v", 0))
	}

	err := DelByPattern(ctx, "mp:key:*")
	require.NoError(t, err)

	// All 150 keys must be gone.
	count, err := Exists(ctx, make([]string, 0)...)
	// No specific assertion on count when no keys given (returns 0).
	_ = count
	_ = err

	// Spot-check a few keys.
	for _, k := range []string{"mp:key:0", "mp:key:75", "mp:key:149"} {
		_, getErr := Get(ctx, k)
		assert.ErrorIs(t, getErr, redis.Nil, "expected key %s to be deleted", k)
	}
}

// ---------------------------------------------------------------------------
// Exists
// ---------------------------------------------------------------------------

func TestExists_Found(t *testing.T) {
	_, ctx := setupTestCache(t)

	require.NoError(t, Set(ctx, "ex1", "v", 0))

	n, err := Exists(ctx, "ex1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestExists_NotFound(t *testing.T) {
	_, ctx := setupTestCache(t)

	n, err := Exists(ctx, "nothere")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestExists_Multiple(t *testing.T) {
	_, ctx := setupTestCache(t)

	require.NoError(t, Set(ctx, "e1", "v", 0))
	require.NoError(t, Set(ctx, "e2", "v", 0))

	n, err := Exists(ctx, "e1", "e2", "e3")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), n)
}

// ---------------------------------------------------------------------------
// Expire
// ---------------------------------------------------------------------------

func TestExpire_KeyExpires(t *testing.T) {
	mr, ctx := setupTestCache(t)

	require.NoError(t, Set(ctx, "expkey", "val", 0))
	require.NoError(t, Expire(ctx, "expkey", 2*time.Second))

	// Key should still exist
	val, err := Get(ctx, "expkey")
	assert.NoError(t, err)
	assert.Equal(t, "val", val)

	// Fast-forward miniredis clock past the TTL
	mr.FastForward(3 * time.Second)

	_, err = Get(ctx, "expkey")
	assert.ErrorIs(t, err, redis.Nil)
}

// ---------------------------------------------------------------------------
// SetNX
// ---------------------------------------------------------------------------

func TestSetNX_New(t *testing.T) {
	_, ctx := setupTestCache(t)

	ok, err := SetNX(ctx, "nxkey", "val", 0)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestSetNX_Existing(t *testing.T) {
	_, ctx := setupTestCache(t)

	require.NoError(t, Set(ctx, "nxkey", "first", 0))

	ok, err := SetNX(ctx, "nxkey", "second", 0)
	assert.NoError(t, err)
	assert.False(t, ok)

	// Original value unchanged
	val, err := Get(ctx, "nxkey")
	assert.NoError(t, err)
	assert.Equal(t, "first", val)
}

// ---------------------------------------------------------------------------
// RedisCacheBackend
// ---------------------------------------------------------------------------

func TestNewRedisCacheBackend(t *testing.T) {
	b := NewRedisCacheBackend()
	assert.NotNil(t, b)
}

func TestRedisCacheBackend_Get_Set(t *testing.T) {
	_, _ = setupTestCache(t)
	ctx := context.Background()
	b := NewRedisCacheBackend()

	err := b.Set(ctx, "bkey", "bval", 0)
	require.NoError(t, err)

	val, err := b.Get(ctx, "bkey")
	assert.NoError(t, err)
	assert.Equal(t, "bval", val)
}

func TestRedisCacheBackend_Del(t *testing.T) {
	_, _ = setupTestCache(t)
	ctx := context.Background()
	b := NewRedisCacheBackend()

	require.NoError(t, b.Set(ctx, "bdel", "v", 0))
	require.NoError(t, b.Del(ctx, "bdel"))

	_, err := b.Get(ctx, "bdel")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestRedisCacheBackend_DelByPattern(t *testing.T) {
	_, _ = setupTestCache(t)
	ctx := context.Background()
	b := NewRedisCacheBackend()

	require.NoError(t, b.Set(ctx, "bp:a", "1", 0))
	require.NoError(t, b.Set(ctx, "bp:b", "2", 0))

	require.NoError(t, b.DelByPattern(ctx, "bp:*"))

	_, err := b.Get(ctx, "bp:a")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestRedisCacheBackend_Exists(t *testing.T) {
	_, _ = setupTestCache(t)
	ctx := context.Background()
	b := NewRedisCacheBackend()

	require.NoError(t, b.Set(ctx, "bex", "v", 0))

	n, err := b.Exists(ctx, "bex")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)
}
