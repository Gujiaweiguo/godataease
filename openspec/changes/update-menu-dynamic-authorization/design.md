## Context
Current Go backend migration includes:
- `core_menu`-based menu query (`/api/menu/query`), but no CRUD/authorization lifecycle.
- Compatibility endpoint `/api/auth/menuResource` with hardcoded menu items.
- Compatibility endpoint `/api/roleRouter/query` now partially dynamic but still includes fallback hardcoding.
- Permission spec expects role-based menu authorization, but runtime lacks explicit role-menu relation and management API.

This creates an architecture split where frontend-visible menu behavior is not fully governed by persisted authorization data.

## Goals / Non-Goals
- Goals:
  - Single source of truth for menus and menu authorization from persistent storage.
  - Role-menu authorization model that drives both compatibility and canonical APIs.
  - Admin-operable menu management and role-menu assignment APIs.
  - Backward-compatible migration for existing frontend routes.
- Non-Goals:
  - Full redesign of resource-level (`sys_perm`) authorization semantics.
  - Enterprise/xpack feature implementation.
  - Rebuilding all frontend admin UX in one iteration beyond required menu/role-menu management.

## Decisions
- Decision: Introduce explicit role-menu relation (`sys_role_menu`) and make menu visibility role-derived.
  - Why: Existing schema has `core_menu` but no persisted role-menu relation; hardcoded endpoints cannot scale.
  - Alternative considered: infer permissions via generic `sys_perm` only.
  - Rejected because current frontend and compatibility contracts require direct menu tree visibility decisions.

- Decision: Unify `/api/roleRouter/query` and `/api/auth/menuResource` on the same menu assembly service.
  - Why: Prevent drift and contradictory menu trees.
  - Alternative considered: keep one endpoint hardcoded as fallback.
  - Rejected due recurring inconsistency and debug complexity.

- Decision: Keep safe bootstrap fallback behind explicit runtime switch, default off after migration verification.
  - Why: Avoid admin lockout during first deployment while still removing permanent hardcoding.
  - Alternative considered: immediate hard cut without fallback.
  - Rejected due operational risk.

## Risks / Trade-offs
- Risk: Incorrect role-menu seed data can hide required menus for administrators.
  - Mitigation: migration seed grants full menu set to bootstrap admin role and provides rollback SQL.
- Risk: Compatibility clients rely on exact field shape.
  - Mitigation: keep response contract parity tests between canonical and compatibility routes.
- Risk: Route authorization changes may produce more 403 responses where users previously saw menus.
  - Mitigation: explicit spec scenarios for unauthorized direct URL and frontend guard behavior.

## Migration Plan
1. Add relation schema and seed admin role-menu mappings.
2. Implement unified menu assembly service and route handlers.
3. Switch compatibility endpoints to service output.
4. Validate parity and role-based visibility in integration tests.
5. Disable fallback hardcoding by default after verification.

## Open Questions
- Role identity for bootstrap mapping in environments with custom seed data (need deterministic lookup rule).
- Whether menu sorting supports drag-drop incremental updates or full-tree replacement in first iteration.
