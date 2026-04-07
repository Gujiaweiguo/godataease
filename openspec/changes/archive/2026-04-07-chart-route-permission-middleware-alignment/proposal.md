## Why

The current permission-center alignment work already moved part of chart runtime authorization into governed service-layer paths, but the HTTP chart entry points still are not consistently guarded by resource-level permission middleware. That leaves canonical chart routes and compatibility bridge routes with uneven enforcement boundaries, including fail-open compatibility behavior that can still fall back to an admin identity when request context is missing.

## What Changes

- Align governed chart runtime routes with real resource-level permission middleware instead of relying on Auth-only route protection.
- Define the chart canonical and compatibility bridge entry points that must enforce governed permission checks before chart data execution proceeds.
- Tighten compatibility bridge permission boundaries so missing user context does not silently fall back to default admin behavior for chart-related permission flows.
- Add focused backend verification for governed chart middleware enforcement, including allowed, denied, and context-missing/error scenarios.
- **BREAKING**: chart routes that previously passed through with Auth-only protection or compatibility fallbacks may begin returning explicit permission-denied or permission-error responses when governed authorization is missing or invalid.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `permission-config`: strengthen governed chart-route authorization requirements so chart runtime entry points enforce resource-level permission middleware consistently across canonical and compatibility paths.

## Impact

- **Affected backend modules**: `apps/backend-go/internal/transport/http/router.go`, `apps/backend-go/internal/transport/http/handler/chart_handler.go`, `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`, and the chart-related permission middleware stack.
- **Affected runtime paths**: canonical `/api/chart/*` routes, compatibility `/chartData/*` and `/chart/*` chart runtime routes, and any shared permission-resolution helpers they depend on.
- **Affected tests**: backend middleware, handler, and chart authorization coverage, plus compatibility-bridge regression tests for fail-open behavior.
- **API impact**: some chart requests that previously reached execution may now fail closed with explicit authorization errors when governed permission context is absent or insufficient.
- **Rollback**: if enforcement scope proves too broad or breaks legitimate governed flows, revert to the last known good state, preserve the new tests that expose the gap, and narrow route coverage before retrying.
