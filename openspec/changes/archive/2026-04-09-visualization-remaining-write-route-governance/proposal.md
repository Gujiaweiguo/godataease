## Why

The previous visualization write-route alignment closed the highest-risk governed gaps, but four remaining root visualization mutation routes still rely on Auth-only protection. That leaves `updateBase`, `move`, `updatePublishStatus`, and `recoverToPublished` weaker than the governed dashboard and big-screen write model already enforced on adjacent visualization routes.

## What Changes

- Extend visualization-aware governed edit authorization to the remaining root visualization write routes that mutate an existing visualization resource by `id`.
- Reuse the existing dashboard-vs-screen resource resolution path instead of introducing a new permission type or a route-specific authorization model.
- Add focused backend route and middleware regression coverage for allowed, denied, missing-auth, and unresolved-target cases on the newly governed routes.
- Preserve explicit scope boundaries by leaving visualization list, tree, helper, and create/copy flows unchanged in this change.
- **BREAKING**: the affected root visualization mutation routes may begin returning explicit permission-denied or permission-error responses where Auth-only access previously allowed the request to proceed.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `permission-config`: extend governed visualization write-route authorization requirements to the remaining root legacy routes that mutate an existing visualization resource.
- `visualization-management`: require governed authorization semantics for the remaining in-scope root visualization mutation flows before state changes proceed.

## Impact

- **Affected backend modules**: `apps/backend-go/internal/transport/http/handler/visualization_handler.go`, visualization permission middleware, and route-level regression coverage for visualization write paths.
- **Affected runtime paths**: root legacy visualization routes `updateBase`, `move`, `updatePublishStatus`, and `recoverToPublished`.
- **Affected tests**: backend middleware and router/handler regression tests covering visualization write authorization behavior.
- **API impact**: root visualization mutation requests that previously relied on Auth-only access may now fail closed with explicit authorization errors.
- **Rollback**: revert the newly added route middleware wiring and keep the regression coverage so the remaining governance gap stays explicit until a narrower retry is ready.
