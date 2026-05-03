package cache

import (
	"context"
	"time"
)

type RedisCacheBackend struct{}

func NewRedisCacheBackend() *RedisCacheBackend {
	return &RedisCacheBackend{}
}

func (b *RedisCacheBackend) Get(ctx context.Context, key string) (string, error) {
	return Get(ctx, key)
}

func (b *RedisCacheBackend) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return Set(ctx, key, value, expiration)
}

func (b *RedisCacheBackend) Del(ctx context.Context, keys ...string) error {
	return Del(ctx, keys...)
}

func (b *RedisCacheBackend) DelByPattern(ctx context.Context, pattern string) error {
	return DelByPattern(ctx, pattern)
}

func (b *RedisCacheBackend) Exists(ctx context.Context, keys ...string) (int64, error) {
	return Exists(ctx, keys...)
}
