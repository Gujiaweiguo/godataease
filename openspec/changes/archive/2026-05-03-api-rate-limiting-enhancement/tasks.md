## 1. Configuration

- [ ] 1.1 Add `RateLimitConfig` and `RouteLimitConfig` to `internal/app/config.go` with validation and env binding entries.
- [ ] 1.2 Add rate-limit defaults and route override examples to `apps/backend-go/configs/config.yaml`, `config.example.yaml`, and `config.local.yaml`.

## 2. Backend Rate Limiter Abstractions

- [ ] 2.1 Define the `RateLimiterBackend` interface and route policy resolution helpers in `internal/transport/http/middleware/ratelimit.go`.
- [ ] 2.2 Refactor the existing `tokenBucketLimiter` into the in-memory backend path so it returns `allowed`, `remaining`, and `resetAt` values.
- [ ] 2.3 Implement `RedisBackend` with Redis `INCR` + `EXPIRE` window tracking using the existing `go-redis/v9` client.
- [ ] 2.4 Add `NewRateLimiterBackend(cfg, redisClient)` to choose Redis or in-memory enforcement based on configuration and availability.

## 3. Middleware and Routing Integration

- [ ] 3.1 Update `RateLimit()` middleware to emit `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, and `Retry-After` headers.
- [ ] 3.2 Replace the three hardcoded handler rate-limit values with `RateLimitConfig.RouteOverrides` entries while preserving their current budgets.
- [ ] 3.3 Wire a configurable global default rate-limit middleware into authenticated API routes in `router.go` after auth middleware.

## 4. Verification

- [ ] 4.1 Add unit tests covering Redis backend budget sharing, expiry handling, and fallback behavior.
- [ ] 4.2 Add middleware tests covering header generation for both allowed and rejected requests.
- [ ] 4.3 Add an integration-style backend test that verifies config-driven global defaults and route overrides are loaded and applied correctly.
