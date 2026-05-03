package middleware

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"dataease/backend/internal/app"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RateLimitKeyFunc func(*gin.Context) string

type RateLimiterBackend interface {
	Allow(key string, limit int, window time.Duration) (allowed bool, remaining int, resetAt time.Time)
}

type tokenBucketLimiter struct {
	mu            sync.Mutex
	buckets       map[string]*tokenBucket
	capacity      float64
	refillPerSec  float64
	now           func() time.Time
	rateLimitName string
}

type tokenBucket struct {
	tokens float64
	last   time.Time
}

type InMemoryBackend struct {
	mu       sync.Mutex
	limiters map[string]*tokenBucketLimiter
	now      func() time.Time
}

type RedisBackend struct {
	client *redis.Client
	now    func() time.Time
	ctx    context.Context
}

func newTokenBucketLimiter(name string, maxRequests int, window time.Duration, now func() time.Time) *tokenBucketLimiter {
	if maxRequests <= 0 {
		maxRequests = 1
	}
	if window <= 0 {
		window = time.Second
	}
	if now == nil {
		now = time.Now
	}
	return &tokenBucketLimiter{
		buckets:       make(map[string]*tokenBucket),
		capacity:      float64(maxRequests),
		refillPerSec:  float64(maxRequests) / window.Seconds(),
		now:           now,
		rateLimitName: name,
	}
}

func (l *tokenBucketLimiter) allow(key string) bool {
	allowed, _, _ := l.allowWithDetails(key, time.Second)
	return allowed
}

func (l *tokenBucketLimiter) allowWithDetails(key string, window time.Duration) (bool, int, time.Time) {
	current := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	resetAt := current.Add(window)

	bucket, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &tokenBucket{tokens: l.capacity - 1, last: current}
		return true, int(math.Max(0, l.capacity-1)), resetAt
	}

	elapsed := current.Sub(bucket.last).Seconds()
	if elapsed > 0 {
		bucket.tokens = math.Min(l.capacity, bucket.tokens+elapsed*l.refillPerSec)
		bucket.last = current
	}

	if bucket.tokens < 1 {
		return false, 0, resetAt
	}

	bucket.tokens--
	return true, int(math.Max(0, math.Floor(bucket.tokens))), resetAt
}

func newInMemoryBackend() *InMemoryBackend {
	return &InMemoryBackend{
		limiters: make(map[string]*tokenBucketLimiter),
		now:      time.Now,
	}
}

func (b *InMemoryBackend) Allow(key string, limit int, window time.Duration) (bool, int, time.Time) {
	if b == nil {
		return true, 0, time.Now().Add(window)
	}

	limiterKey := fmt.Sprintf("%d:%d", limit, window)

	b.mu.Lock()
	limiter, ok := b.limiters[limiterKey]
	if !ok {
		limiter = newTokenBucketLimiter(limiterKey, limit, window, b.now)
		b.limiters[limiterKey] = limiter
	}
	b.mu.Unlock()

	return limiter.allowWithDetails(key, window)
}

func (b *RedisBackend) WithContext(ctx context.Context) *RedisBackend {
	if b == nil {
		return nil
	}
	clone := *b
	clone.ctx = ctx
	return &clone
}

func (b *RedisBackend) context() context.Context {
	if b != nil && b.ctx != nil {
		return b.ctx
	}
	return context.TODO()
}

func (b *RedisBackend) Allow(key string, limit int, window time.Duration) (bool, int, time.Time) {
	current := time.Now()
	if b != nil && b.now != nil {
		current = b.now()
	}
	if b == nil || b.client == nil {
		return true, limit, current.Add(window)
	}

	ctx := b.context()
	redisKey := fmt.Sprintf("ratelimit:%s", key)
	count, err := b.client.Incr(ctx, redisKey).Result()
	if err != nil {
		logger.Warn("redis rate limit incr failed",
			zap.String("key", redisKey),
			zap.Error(err),
		)
		return true, limit, current.Add(window)
	}

	if count == 1 {
		if err := b.client.Expire(ctx, redisKey, window).Err(); err != nil {
			logger.Warn("redis rate limit expire failed",
				zap.String("key", redisKey),
				zap.Error(err),
			)
			return true, maxInt(limit-1, 0), current.Add(window)
		}
	}

	resetAt := current.Add(window)
	ttl, err := b.client.TTL(ctx, redisKey).Result()
	if err != nil {
		logger.Warn("redis rate limit ttl failed",
			zap.String("key", redisKey),
			zap.Error(err),
		)
	} else if ttl > 0 {
		resetAt = current.Add(ttl)
	}

	if count > int64(limit) {
		return false, 0, resetAt
	}

	return true, maxInt(limit-int(count), 0), resetAt
}

func NewRateLimiterBackend(cfg app.RateLimitConfig, redisClient *redis.Client) RateLimiterBackend {
	if cfg.UseRedis && redisClient != nil {
		return &RedisBackend{client: redisClient, now: time.Now}
	}
	return newInMemoryBackend()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func RateLimit(name string, maxRequests int, window time.Duration, keyFunc RateLimitKeyFunc) gin.HandlerFunc {
	limiter := newTokenBucketLimiter(name, maxRequests, window, time.Now)
	return func(c *gin.Context) {
		key := strings.TrimSpace(keyFunc(c))
		if key == "" {
			key = strings.TrimSpace(c.ClientIP())
		}
		if key == "" {
			key = "anonymous"
		}

		if !limiter.allow(key) {
			response.TooManyRequests(c, "rate limit exceeded")
			return
		}

		c.Next()
	}
}

func ClientIPKey(c *gin.Context) string {
	return strings.TrimSpace(c.ClientIP())
}

func AuthenticatedUserKey(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		switch value := userID.(type) {
		case uint64:
			if value > 0 {
				return strconv.FormatUint(value, 10)
			}
		case int64:
			if value > 0 {
				return strconv.FormatInt(value, 10)
			}
		case int:
			if value > 0 {
				return strconv.Itoa(value)
			}
		}
	}

	return ClientIPKey(c)
}
