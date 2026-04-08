## Why

The current permission-center alignment already enforces resource-level authorization for selected visualization routes such as detail lookup and canvas updates, but several visualization write paths still rely on Auth-only protection. That leaves save, copy, move, publish-status, and related resource-management flows with weaker enforcement than the governed dashboard and big-screen model now expects.

## What Changes

- Align governed visualization write routes with real resource-level permission middleware instead of leaving them behind Auth-only protection.
- Define which visualization write operations must require governed edit-level authorization before the handler continues.
- Extend visualization route coverage in small slices, starting from the highest-risk write paths and preserving explicit scope boundaries for non-write discovery/list routes.
- Add focused backend verification for allowed, denied, and missing-context cases on visualization write operations.
- **BREAKING**: visualization write requests that previously passed through with Auth-only protection may begin returning explicit permission-denied or permission-error responses when governed authorization is absent or insufficient.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `permission-config`: strengthen governed visualization write-route authorization requirements so dashboard and big-screen mutation flows enforce resource-level permission checks consistently.
- `visualization-management`: require governed authorization semantics for in-scope visualization write operations rather than treating successful authenticated access as sufficient.

## Impact

- **Affected backend modules**: `apps/backend-go/internal/transport/http/router.go`, visualization-related handlers, and any middleware or request-resolution helpers needed to identify governed visualization resources for write flows.
- **Affected runtime paths**: in-scope visualization write routes such as save, copy, move, metadata updates, publish-state changes, and recovery flows on canonical or legacy-compatible paths selected by the change tasks.
- **Affected tests**: backend router, handler, and visualization authorization regression coverage for happy-path and denial-path write operations.
- **API impact**: some visualization write requests that previously reached mutation handlers may now fail closed with explicit authorization errors.
- **Rollback**: if the first enforcement slice proves too broad or breaks legitimate governed workflows, revert the new route wiring, preserve the authorization regression tests, and narrow the write-route scope before retrying.
