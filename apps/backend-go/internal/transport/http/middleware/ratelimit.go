package middleware

import (
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type RateLimitKeyFunc func(*gin.Context) string

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
	current := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &tokenBucket{tokens: l.capacity - 1, last: current}
		return true
	}

	elapsed := current.Sub(bucket.last).Seconds()
	if elapsed > 0 {
		bucket.tokens = math.Min(l.capacity, bucket.tokens+elapsed*l.refillPerSec)
		bucket.last = current
	}

	if bucket.tokens < 1 {
		return false
	}

	bucket.tokens--
	return true
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
