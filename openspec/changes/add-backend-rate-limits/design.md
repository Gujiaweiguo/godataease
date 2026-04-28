## Context

The Go backend already protects sensitive request paths with JWT authentication, SSRF-safe outbound fetch rules, and local file path validation, but it does not currently constrain request frequency on abuse-prone endpoints. Three backend surfaces stand out:

- `POST /login/localLogin` and `POST /api/login/localLogin` are registered directly on the Gin engine in `apps/backend-go/internal/transport/http/handler/auth_handler.go` and have no request-throttling guard.
- `POST /api/ds/validate` and `GET /api/ds/validate/:id` are registered in `apps/backend-go/internal/transport/http/handler/datasource_handler.go` and already sit behind JWT auth via `router.go`, but every request can still trigger validation and TCP probing logic.
- `POST /api/audit/export` and `GET /api/audit/download` are registered in `apps/backend-go/internal/transport/http/handler/audit_handler.go` and already sit behind JWT auth, but they can still be abused to repeatedly generate export files and download them.

The change needs to fit the existing Gin middleware style, avoid changing normal success semantics, and remain narrow enough to land without widening the current security batch.

## Goals / Non-Goals

**Goals:**
- Add a reusable backend rate-limit middleware that can be attached to specific routes or route groups.
- Enforce rate limits on local login, datasource validate, and audit export/download with endpoint-appropriate keys.
- Return explicit throttling failures without changing happy-path request/response payloads for normal traffic.
- Add targeted backend verification for throttled and non-throttled behavior.

**Non-Goals:**
- Introduce a global rate limit for all backend routes.
- Change JWT authentication, RBAC, or existing SSRF/path validation behavior in the same change.
- Add distributed or Redis-backed rate limiting in this phase.
- Redesign audit export business rules or datasource validation semantics beyond throttling.

## Decisions

### 1. Use a dedicated Gin middleware in `internal/transport/http/middleware/ratelimit.go`

We will add a standalone middleware factory that returns `gin.HandlerFunc`, matching the lightweight pattern already used by `middleware.Auth(...)` and other transport-layer concerns.

**Why this approach:**
- Keeps rate limiting in the transport layer where request identity and path context already exist.
- Avoids pushing throttle state into handlers or services.
- Lets us attach limits narrowly to the target routes without touching unrelated domains.

**Alternatives considered:**
- **Handler-level checks inside each endpoint**: rejected because it duplicates logic and scatters throttle behavior across modules.
- **Global engine middleware**: rejected because the immediate goal is targeted hardening, not platform-wide throttling.

### 2. Start with in-process token buckets keyed by request identity

The middleware will use in-process token bucket state keyed by a stable identity string. For this phase, the implementation should prefer simplicity and narrow scope over cross-node coordination.

**Keying strategy:**
- **Login**: key by client IP, because no authenticated user identity exists yet.
- **Datasource validate**: key by authenticated user ID from Gin context, with IP fallback only if identity is unexpectedly absent.
- **Audit export/download**: key by authenticated user ID from Gin context, with IP fallback only if identity is unexpectedly absent.

**Why this approach:**
- Matches the immediate abuse models: brute-force from a source IP, repeated probing by a signed-in user, and repeated export/download by a signed-in user.
- Avoids bringing Redis or distributed coordination into this phase.
- Minimizes operational complexity while still materially reducing risk.

**Alternatives considered:**
- **Redis-backed distributed rate limiting**: rejected for this phase because it adds infrastructure coupling and rollout complexity.
- **Username-based login throttling**: rejected as the primary key because usernames are only available after request parsing and can be trivially varied by an attacker; IP is the safer first control.

### 3. Apply rate limiting at route-registration points, not in `router.go` global wiring

The current route setup already centralizes auth protection in `router.go`, then delegates concrete paths to `Register*Routes` helpers. We will preserve that split and attach throttling in the relevant registration helpers or per-route subgroups:

- `RegisterAuthRoutes(...)` for login
- `RegisterDatasourceRoutes(...)` for datasource validate routes
- `RegisterAuditRoutes(...)` for audit export/download routes

**Why this approach:**
- Keeps the throttling logic closest to the paths it governs.
- Avoids broad changes to the central router composition in `router.go`.
- Makes it easier to see which exact routes are intentionally limited.

**Alternatives considered:**
- **Wrapping entire `datasourceAPI` or `auditAPI` groups in `router.go`**: rejected because it would throttle unrelated routes and broaden behavior change.

### 4. Return explicit `429 Too Many Requests` with the existing response envelope shape

When a request is throttled, the middleware should return HTTP `429` and an explicit non-success envelope through the existing response helpers or a small dedicated helper.

**Why this approach:**
- Makes throttling observable to clients, tests, and logs.
- Preserves the backend’s established error-envelope conventions rather than introducing ad hoc plain-text failures.
- Avoids misclassifying throttled traffic as generic `500000` application errors.

**Alternatives considered:**
- **Silent delay or connection drop**: rejected because it is harder to diagnose and verify.
- **Reusing generic `500000` responses**: rejected because rate limiting is a distinct failure class.

### 5. Verify the middleware at handler/middleware level with targeted tests

The change will extend backend tests around the affected route surfaces instead of relying only on service tests:

- Login: add route/handler coverage for throttled local login attempts.
- Datasource validate: add route or handler coverage for repeated validate requests.
- Audit export/download: extend current audit handler coverage with throttled export/download behavior.

**Why this approach:**
- The change lives at the HTTP transport layer, so service-only tests are insufficient.
- Existing handler tests already cover path semantics for audit download and route auth boundaries, making them natural extension points.

## Risks / Trade-offs

- **[Risk] In-process limiter state is per-process only** → Mitigation: explicitly scope this phase to single-process protection semantics and record distributed throttling as future work if deployment topology requires it.
- **[Risk] Too-aggressive thresholds could block legitimate admin workflows** → Mitigation: keep the initial thresholds conservative, per-endpoint, and easy to tune in code/config during rollout.
- **[Risk] IP-based login throttling can affect multiple users behind the same NAT** → Mitigation: limit the first phase to modest burst control rather than heavy lockout behavior.
- **[Risk] New `429` responses may require client-side awareness later** → Mitigation: keep current scope server-side only and document explicit throttling semantics in specs so frontend handling can be improved separately if needed.
- **[Risk] Audit export/download throttling may be bypassed by distributing requests across identities** → Mitigation: accept this trade-off for phase 1 because the goal is targeted reduction of obvious abuse, not a full anti-abuse platform.

## Migration Plan

1. Add the reusable rate-limit middleware and attach it only to the selected routes.
2. Extend backend tests to cover throttled and non-throttled behavior for login, datasource validate, and audit export/download.
3. Roll out without changing route paths, auth boundaries, or success envelopes for normal traffic.
4. If thresholds prove too strict or behavior is incorrect, roll back by removing the new route-level middleware attachments while keeping unrelated endpoint behavior unchanged.

## Open Questions

- Should throttling thresholds be hard-coded for the first phase or sourced from backend config immediately?
- Should login throttling remain IP-only in phase 1, or should we also record username-aware counters once request parsing succeeds?
- Do audit export and audit download need distinct thresholds, or is a shared audit-export budget sufficient for the first rollout?
