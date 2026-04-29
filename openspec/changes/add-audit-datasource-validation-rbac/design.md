## Context

Audit export/download and datasource validation are currently protected by JWT authentication, and both recently gained route-scoped rate limits. The remaining gap is authorization depth:

- `RegisterAuditRoutes(...)` attaches rate limiting to `POST /api/audit/export` and `GET /api/audit/download`, but the enclosing `/api/audit/*` group is still only gated by `middleware.Auth(...)` in `apps/backend-go/internal/transport/http/router.go`.
- `RegisterDatasourceRoutes(...)` attaches rate limiting to `POST /api/ds/validate` and `GET /api/ds/validate/:id`, but the datasource API group is likewise only gated by JWT auth in `router.go`.
- Compatibility datasource validation aliases under `/api/datasource/*` and `/de2api/datasource/*` still route into the same handler methods, so they need the same authorization semantics to avoid preserving a weaker side door.

The codebase already has two useful authorization patterns:

- `MenuAuthMiddleware.RequireMenuAuth(menuPath)` for menu-governed system surfaces.
- `PermissionMiddleware.CheckDatasourceView()` / `CheckDatasourceEdit()` for datasource resource-level permission checks.

This change should reuse those existing patterns instead of introducing a third RBAC mechanism.

## Goals / Non-Goals

**Goals:**
- Restrict audit export/download to users who hold the intended audit menu authorization instead of allowing any authenticated user.
- Restrict datasource validation to users with datasource-specific management authority, with explicit handling for both validation-by-ID and payload-based validation.
- Keep canonical and compatibility datasource validation routes behaviorally aligned.
- Add targeted backend tests covering both allowed and denied authorization outcomes.

**Non-Goals:**
- Redesign the overall role/permission data model.
- Change rate-limit thresholds or limiter identity strategies introduced in the previous phase.
- Expand datasource CRUD RBAC beyond the validation surface.
- Rework frontend permission UX or menu assignment flows in this change.

## Decisions

### 1. Audit export/download will use existing menu-path authorization

The audit export and download routes are system-level operations rather than resource-instance operations. We will reuse `MenuAuthMiddleware.RequireMenuAuth(...)` for those routes instead of adding a new audit-specific permission layer.

**Why this approach:**
- The codebase already uses menu-path authorization for protected system-management routes.
- Audit export/download maps naturally to the governed audit feature surface rather than an individual row/resource instance.
- This keeps admin bypass and role-menu authorization semantics consistent with existing backend behavior.

**Alternatives considered:**
- **JWT-only protection**: rejected because it preserves the current overly broad access model.
- **Datasource-style resource permission checks**: rejected because audit export/download does not target a stable resource instance in the same way.
- **Admin-only hard gate**: rejected because it is coarser than the existing role-menu model and would unnecessarily limit delegated audit operators.

### 2. Datasource validation will use datasource resource permissions when a datasource identity is present

When validation targets an existing datasource (for example, `GET /validate/:id` or payload-based validation that includes datasource identity), the backend should enforce datasource resource permissions through the existing `PermissionMiddleware` datasource checks.

**Why this approach:**
- Datasource resource permission infrastructure already exists and is the backend’s established fine-grained pattern for datasource access control.
- Existing-datasource validation is a stronger action than plain visibility because it can trigger outbound connection checks; it should not remain available to broad authenticated traffic.
- Reusing datasource permission middleware avoids inventing new tables or permission keys.

**Alternatives considered:**
- **Menu-only protection for all datasource validation**: rejected because it is too coarse when a concrete datasource resource ID is already available.
- **View-level access for all validation**: rejected as the default because validation can act as an active probe, which is closer to management/edit semantics than passive read semantics.

### 3. Payload-based datasource validation without a stable datasource ID will fall back to datasource-management menu authorization

Some payload-based validation requests may not carry a stable datasource resource ID because they validate a draft before persistence. In those cases, the backend should require datasource menu authorization equivalent to datasource management access rather than silently allowing JWT-only traffic.

**Why this approach:**
- It closes the probe surface for unaffiliated authenticated users when resource-level permission checks cannot resolve a datasource ID.
- It reuses existing menu-governed authorization infrastructure instead of introducing ad hoc body parsing rules for every draft shape.
- It preserves a path for legitimate datasource creators/editors to validate draft configurations before save.

**Alternatives considered:**
- **Require resource ID for every POST validate request**: rejected because it breaks legitimate pre-save validation flows.
- **Allow JWT-only fallback**: rejected because it keeps the current authorization gap intact.

### 4. Compatibility datasource validation aliases must mirror canonical authorization semantics

The `/api/datasource/*` and `/de2api/datasource/*` validation aliases will adopt the same authorization checks as `/api/ds/*` so there is no weaker compatibility path.

**Why this approach:**
- Compatibility routes remain actively used and must not bypass new protections.
- It preserves route parity and rollback safety across canonical and compatibility callers.

**Alternatives considered:**
- **Harden only canonical routes**: rejected because compatibility aliases would remain an authorization bypass.

## Risks / Trade-offs

- **[Risk] Audit menu path resolution may differ from assumed route naming** → Mitigation: verify the concrete audit menu path during implementation and keep the change scoped to whichever existing menu path already governs the audit page.
- **[Risk] Datasource draft validation may not always expose a stable datasource ID** → Mitigation: explicitly document and test the menu-auth fallback path for draft validation.
- **[Risk] Using edit/manage-style datasource authorization may block users who previously only had read visibility** → Mitigation: treat the route as an active management surface and cover denied behavior in tests so the stricter policy is deliberate and diagnosable.
- **[Risk] Compatibility routes may be missed when wiring middleware** → Mitigation: require route-level verification for canonical and alias endpoints in the task list.

## Migration Plan

1. Wire audit export/download through menu-based authorization in addition to existing JWT and rate-limit protection.
2. Wire datasource validation routes through datasource resource permission checks when a datasource identity is available, with explicit datasource-menu fallback for draft validation.
3. Apply the same datasource validation checks to compatibility aliases.
4. Extend route-level backend tests for both authorized and forbidden cases.
5. Roll back, if needed, by removing the new route-level authorization middleware while preserving existing JWT and rate-limit behavior.

## Open Questions

- What exact audit menu path should be treated as the governing authorization path in production data: `/audit`, `/system/audit-log`, or another existing seeded path?
- Should existing-datasource validation require datasource `edit` permission, or is a dedicated `manage`/custom permission needed later if policy becomes stricter?
- For payload validation that includes a datasource ID field, should the implementation resolve that ID directly in middleware, or route such requests through a dedicated handler-side authorization helper?
