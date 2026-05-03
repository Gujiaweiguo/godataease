## Context

The Go backend already enforces route-scoped rate limits with an in-memory token bucket on login, datasource validation, and audit export/download routes. That implementation is simple and well-tested, but it is process-local, hardcoded in handler registration, and does not expose standard client-facing rate limit metadata. The change needs to preserve the existing middleware pattern, avoid new dependencies, and fit the current Gin + Viper + Redis architecture already used elsewhere in the backend.

This enhancement adds a distributed rate-limit option for multi-instance deployments while keeping the current in-memory behavior as a safe fallback. It also moves rate-limit policy into application configuration so operators can enable a global authenticated-route default, keep existing route-specific protections, and override limits without changing Go source.

## Goals / Non-Goals

**Goals:**
- Introduce a dual-backend rate limiter with Redis as the primary shared-state backend and the existing in-memory limiter as the fallback path.
- Preserve the current middleware factory style by routing all enforcement through a single `RateLimit()` middleware entry point.
- Add standard `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, and `Retry-After` headers without changing existing response bodies or endpoints.
- Add `RateLimitConfig` and route override support to the backend configuration model and environment binding flow.
- Support a configurable global default limiter for authenticated API routes while preserving route-specific overrides for the currently protected endpoints.

**Non-Goals:**
- Changing API response payload schemas or introducing new throttling endpoints.
- Replacing Gin middleware composition or the existing auth middleware contract.
- Introducing a new external dependency beyond the existing `github.com/redis/go-redis/v9` client.
- Designing tenant-specific, role-specific, or dynamic admin-managed rate-limit policies in this change.

## Decisions

### 1. Use a backend interface to decouple policy enforcement from storage strategy

Define a `RateLimiterBackend` interface in `internal/transport/http/middleware/ratelimit.go`:

```go
type RateLimiterBackend interface {
	Allow(key string, limit int, window time.Duration) (allowed bool, remaining int, resetAt time.Time)
}
```

This keeps the middleware focused on key derivation, config lookup, and header emission while letting storage-specific logic live behind a single contract.

**Why this approach:**
- Matches the repository's factory-driven middleware pattern.
- Makes Redis and in-memory implementations testable in isolation.
- Keeps future backend additions possible without reshaping handlers or router wiring.

**Alternatives considered:**
- Extending the current `tokenBucketLimiter` directly with Redis branches inside the same type: rejected because it would couple process-local state and distributed state into one implementation and make tests harder to reason about.
- Keeping separate middleware functions for Redis and memory: rejected because configuration-driven backend selection would then leak into route registration.

### 2. Preserve the current token bucket as `InMemoryBackend`

The existing `tokenBucketLimiter` remains the fallback implementation, wrapped or renamed conceptually as `InMemoryBackend`. It will be enhanced to return `remaining` and `resetAt` so headers can be generated for both accepted and rejected requests.

**Why this approach:**
- The current implementation is already covered by tests and fits the current route-scoped usage.
- It provides a low-risk fallback when Redis is disabled or temporarily unavailable.
- It avoids behavioral drift for single-instance deployments.

**Alternatives considered:**
- Replacing the token bucket with a fixed-window counter in memory: rejected because it changes local throttling semantics unnecessarily.

### 3. Implement `RedisBackend` with `INCR` + `EXPIRE` window tracking

The distributed backend will use Redis `INCR` with an expiry window per rate-limit key. The middleware will compose a stable Redis key from the route name and identity key. On the first increment, the backend sets the TTL; on subsequent increments it reads or preserves the TTL and derives `resetAt` from the remaining expiry. This provides the shared per-window enforcement model required for multi-instance deployments while keeping the implementation compact and operationally simple.

**Why this approach:**
- Uses existing Redis primitives already available through `go-redis/v9`.
- Produces deterministic `remaining` and `resetAt` values needed for standard headers.
- Minimizes implementation complexity and operational risk compared with Lua-based or sorted-set sliding windows.

**Alternatives considered:**
- Redis sorted sets for true sliding windows: rejected for initial rollout due to higher implementation and test complexity.
- Lua scripts for atomic multi-step logic: deferred because the initial use case can be covered with simpler Redis commands and graceful fallback.

### 4. Select backend through a dedicated factory

Add `NewRateLimiterBackend(cfg RateLimitConfig, redisClient *redis.Client) RateLimiterBackend`.

Factory rules:
- Return an in-memory backend when rate limiting is enabled but Redis use is disabled.
- Return an in-memory backend when Redis use is requested but no Redis client is available.
- Return a Redis backend when `UseRedis` is true and a client is configured.

**Why this approach:**
- Centralizes environment-sensitive backend selection.
- Prevents route handlers and router wiring from containing backend-availability branches.
- Makes fallback behavior explicit and testable.

### 5. Introduce configuration-first policy objects

Add the following structures to `internal/app/config.go`:

```go
type RateLimitConfig struct {
	Enabled              bool                        `mapstructure:"enabled"`
	DefaultMaxRequests   int                         `mapstructure:"default_max_requests"`
	DefaultWindowSeconds int                         `mapstructure:"default_window_seconds"`
	UseRedis             bool                        `mapstructure:"use_redis"`
	RouteOverrides       map[string]RouteLimitConfig `mapstructure:"route_overrides"`
}

type RouteLimitConfig struct {
	Enabled       *bool `mapstructure:"enabled"`
	MaxRequests   int   `mapstructure:"max_requests"`
	WindowSeconds int   `mapstructure:"window_seconds"`
}
```

Defaults:
- `Enabled=false` for backward compatibility.
- `DefaultMaxRequests=100`.
- `DefaultWindowSeconds=60`.
- `UseRedis=true`.

The `RouteOverrides` map is keyed by stable route-limit names so the three existing hardcoded limits can migrate without changing endpoint semantics.

**Why this approach:**
- Matches the current nested config + `mapstructure` + `bindEnvKeys` pattern.
- Supports global defaults and route-specific exceptions in a single config surface.
- Keeps enable/disable decisions declarative for deployment and rollback.

**Alternatives considered:**
- Separate top-level config sections per handler: rejected because it spreads one concern across unrelated config branches.

### 6. Apply standard rate-limit headers on both success and rejection paths

The middleware will set:
- `X-RateLimit-Limit`: configured maximum requests for the evaluated policy.
- `X-RateLimit-Remaining`: requests left in the active window after the current evaluation.
- `X-RateLimit-Reset`: Unix timestamp for when the current window resets.
- `Retry-After`: seconds until reset, only on `429 Too Many Requests`.

Headers are emitted before calling `c.Next()` for allowed requests and before `response.TooManyRequests(...)` for rejected requests.

**Why this approach:**
- Gives clients enough information to back off without changing JSON response contracts.
- Keeps behavior consistent across in-memory and Redis implementations.

### 7. Add a global authenticated-route default after auth middleware

Global default rate limiting will be wired in `router.go` for authenticated API route groups after auth middleware has populated `user_id`. The middleware identity order is:
1. `AuthenticatedUserKey`
2. `ClientIPKey` fallback
3. `anonymous` final fallback if neither resolves

This preserves the existing per-route identity model while allowing a broader authenticated-route safety net when enabled.

**Why this approach:**
- Authenticated user identity is the most stable budget key for protected routes.
- Positioning after auth keeps the middleware contract consistent with existing handler-level rate limits.
- Default disablement avoids surprising behavioral change for existing deployments.

### 8. Migrate the current hardcoded limits into route overrides

The existing login, datasource validation, and audit export/download rate limits remain conceptually unchanged, but their numeric values move from handler constructor arguments into `RateLimitConfig.RouteOverrides` entries. The middleware name passed today becomes the route override key.

**Why this approach:**
- Preserves current effective limits during migration.
- Allows operators to tune limits without code edits.
- Reduces duplication across handler registration code.

## Risks / Trade-offs

- **[Redis unavailable or misconfigured]** → Fallback to `InMemoryBackend` when no Redis client is present or Redis usage is disabled; document that cross-instance consistency is lost in fallback mode.
- **[Redis window semantics differ from local token bucket smoothing]** → Keep the fallback path unchanged and document the implementation semantics in code comments and tests so operators understand the difference.
- **[Global default limiter may throttle routes unintentionally]** → Ship with `rate_limit.enabled=false` by default and require explicit route overrides for sensitive endpoints during rollout.
- **[Header values become inconsistent across implementations]** → Require backend-specific tests to assert `remaining`, `resetAt`, and `Retry-After` calculations.
- **[Configuration sprawl or invalid overrides]** → Validate numeric defaults and override values during config loading, rejecting zero or negative windows when a limit is enabled.

## Migration Plan

1. Add `RateLimitConfig` and env bindings to the backend config model with safe defaults.
2. Introduce the backend interface and factory while preserving current in-memory behavior behind the new abstraction.
3. Implement Redis-backed enforcement and header generation tests.
4. Move the three existing hardcoded rate-limit policies into config route overrides with identical values.
5. Wire the global authenticated-route default in `router.go`, gated by config.
6. Roll out with `Enabled=false` by default; operators can enable globally or per override in config.

Rollback strategy:
- Disable rate limiting globally via config if rollout issues occur.
- Set `UseRedis=false` to force in-memory enforcement if Redis-specific issues appear.
- Because endpoint contracts remain unchanged, rollback is configuration-first and does not require schema migration.

## Open Questions

- Should Redis backend failures during request handling fall back silently to in-memory for that request, or fail closed/open after startup selected Redis mode? The implementation should prefer predictable behavior and explicit logging.
- Should route override keys map to Gin route names, middleware names, or a repository-defined constants set? A stable naming convention is needed to avoid config drift.
- Should the global authenticated-route default exclude specific high-volume endpoints by default, or rely entirely on explicit per-route disable overrides?
