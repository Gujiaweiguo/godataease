## Why

The Go backend already implements in-memory token-bucket rate limiting on 3 high-risk routes (login, datasource validation, audit export). However, the current implementation has gaps: (1) limits are per-process only and reset on restart — multi-instance deployments share no state; (2) no standard RateLimit headers (X-RateLimit-Limit/Remaining/Reset) are returned, making it harder for clients to adapt; (3) rate limit parameters are hardcoded in handler constructors rather than externally configurable; (4) there is no global default rate limit — only selectively applied per-route.

## What Changes

- Add Redis-backed rate limiter that uses the existing `go-redis/v9` dependency for distributed state sharing across multiple backend instances
- Add standard HTTP RateLimit response headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After`) to 429 responses
- Add `RateLimitConfig` to the application configuration (`config.yaml`) with per-route overrides and global defaults
- Implement automatic limiter backend selection: Redis when available, in-memory fallback when not
- Add global default rate limit middleware applied to all authenticated API routes (configurable, disabled by default)
- Migrate the 3 existing hardcoded rate limits to use the new configurable system

## Capabilities

### New Capabilities
_(none — all changes modify existing capabilities)_

### Modified Capabilities
- `api-rate-limiting`: Add distributed Redis-backed limiter, standard RateLimit headers, configurable rate limit parameters, and global default rate limiting
- `backend-go-architecture`: Add `RateLimitConfig` to the application configuration structure and environment binding

## Impact

- **Backend middleware**: `internal/transport/http/middleware/ratelimit.go` — add Redis limiter, headers, config integration
- **Backend config**: `internal/app/config.go` — add `RateLimitConfig` struct
- **Backend router**: `internal/transport/http/router.go` — wire global default limiter
- **Handler files**: `auth_handler.go`, `datasource_handler.go`, `audit_handler.go` — migrate hardcoded limits to config
- **Config files**: `configs/config.yaml`, `configs/config.example.yaml` — add rate limit section
- **Dependencies**: No new dependencies (already has `go-redis/v9`)
- **API contract**: No breaking changes — adds headers to existing 429 responses, no new endpoints
- **Frontend**: No changes required (429 handling is optional enhancement for future)
