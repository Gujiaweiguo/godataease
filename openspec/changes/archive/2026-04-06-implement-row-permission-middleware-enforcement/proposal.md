## Why

The current governance alignment work already made row and column permission behavior semantically correct in the service-layer runtime entry points, but `RowPermissionMiddleware()` is still a warning-only stub. That leaves a gap between the documented governed runtime model and the broader HTTP enforcement surface, so this change is needed to make row-permission enforcement real wherever middleware is supposed to guard governed access.

## What Changes

- Replace the current `RowPermissionMiddleware()` placeholder with real middleware behavior that enforces governed row-permission semantics instead of only logging that the framework exists.
- Define which runtime entry points must be protected by middleware and how middleware-enforced behavior composes with the existing dataset/chart service-layer permission logic.
- Align denial and failure semantics so middleware-protected flows do not silently bypass row-permission enforcement or return placeholder success behavior.
- Add focused backend verification for middleware-enforced paths, including success, denial, and error-propagation scenarios.
- **BREAKING**: routes or callers that currently rely on the middleware being non-operative may begin receiving real permission-denied or permission-error responses once enforcement is activated.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `permission-config`: strengthen the governed runtime enforcement requirement so row-permission behavior is enforced consistently at the middleware/runtime boundary instead of only inside a subset of service entry points.

## Impact

- **Affected backend modules**: `apps/backend-go/internal/transport/http/middleware/permission.go`, router wiring, and the governed dataset/chart runtime handlers and services that currently compose with row/column permission checks.
- **Affected runtime paths**: middleware-guarded dataset preview, chart query, and any other governed HTTP entry points that should not rely on a warning-only stub.
- **Affected tests**: backend middleware/handler/service tests and integration coverage for governed runtime enforcement behavior.
- **API impact**: some previously pass-through routes may now return explicit denial/error semantics when row-permission rules apply or permission lookup fails.
- **Dependencies**: no new infrastructure dependencies are expected; the change builds on the existing row-permission, column-permission, and permission middleware stack.
- **Rollback**: if the new middleware enforcement overreaches or breaks governed flows, revert to the last known good state where service-layer runtime checks remain active, then narrow the middleware scope before retrying.
