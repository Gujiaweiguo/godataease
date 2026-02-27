## Context

Current frontend initialization and admin workflows depend on legacy-compatible API routes under `/de2api/*`, which are bridged into Go runtime routes. Core menu bootstrap (`/roleRouter/query`, `/auth/menuResource`) is implemented, but permission-save and visualization-tree families are partially missing, producing `404` in role/menu management and dashboard resource-tree flows. The gap spans multiple modules (router compatibility, auth/permission handlers, role aliases, visualization handlers), so a cross-cutting design is required.

## Goals / Non-Goals

**Goals:**
- Restore frontend-critical endpoint availability for permission management and dashboard-tree operations.
- Keep compatibility behavior deterministic (`code/data/msg` contract) with explicit non-success semantics where business capability is intentionally unavailable.
- Centralize legacy-path aliasing strategy to reduce per-handler drift.
- Add verifiable endpoint matrix + regression checks for critical flows.

**Non-Goals:**
- Re-implement every historical Java endpoint in one release.
- Redesign frontend permission model or route generation mechanism.
- Introduce breaking API contract changes for existing working Go canonical routes.

## Decisions

### Decision 1: Introduce targeted compatibility endpoints first (P0 scope)
- **Choice**: Implement the minimum endpoint set that currently blocks role/menu management and dashboard resource tree.
- **Why**: This immediately removes high-frequency 404s without delaying on broad API reconstruction.
- **Alternatives considered**:
  - Implement full Java surface at once → rejected due to risk and long lead time.
  - Frontend fallback to ignore missing APIs → rejected because it hides auth defects and causes silent permission inconsistency.

### Decision 2: Preserve canonical handlers; add compatibility alias layer
- **Choice**: Add alias routes for `/system/role/*` and permission endpoints that map to canonical services where possible.
- **Why**: Reduces duplication and keeps business logic centralized in existing service layer.
- **Alternatives considered**:
  - Duplicate handlers with separate logic → rejected (maintenance and divergence risk).

### Decision 3: Visualization tree compatibility handled in visualization module boundary
- **Choice**: Add `/dataVisualization/tree` compatibility endpoint through visualization handler/service boundary.
- **Why**: Tree operations are first-class visualization behavior and should not live in generic router fallback.
- **Alternatives considered**:
  - Router-level synthetic responses → rejected due to placeholder-success anti-pattern.

### Decision 4: Governance via endpoint inventory and regression checks
- **Choice**: Maintain scoped endpoint inventory for this change and require curl/integration checks for non-404 + contract envelope.
- **Why**: Prevent recurrence of migration drift after this patch set.
- **Alternatives considered**:
  - One-off manual validation only → rejected as non-repeatable.

## Risks / Trade-offs

- [Risk] Compatibility alias behavior diverges from canonical semantics → **Mitigation**: route aliases reuse existing service methods and share response helper.
- [Risk] Partial implementation returns superficial success → **Mitigation**: enforce explicit non-success for unavailable logic, aligned with `api-compatibility-bridge` governance.
- [Risk] Endpoint additions increase surface for auth bypass bugs → **Mitigation**: keep middleware chain unchanged; add auth-required regression checks for protected routes.
- [Risk] Frontend expects additional fields not yet present → **Mitigation**: prioritize required fields used in current views; document gaps in endpoint matrix.

## Migration Plan

1. Add compatibility route registrations and aliases for P0 endpoint set.
2. Implement/bridge handlers to existing services; avoid introducing parallel business logic.
3. Validate endpoints through scripted checks (status, envelope, key data fields).
4. Run frontend regression on role/menu page and dashboard resource tree interactions.
5. If regressions occur, rollback by removing newly added alias registrations (canonical routes remain untouched).

## Open Questions

- Which non-P0 visualization endpoints should be promoted into next parity batch (`save`, `move`, `copy`, `nameCheck` priorities by traffic)?
- Do we need a dedicated compatibility error-code namespace for intentionally unavailable endpoints in this module set?
