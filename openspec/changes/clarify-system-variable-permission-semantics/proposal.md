## Why

The current permission-center implementation has already converged T8 around governed menu, resource, and row/column permission workflows, but system-variable permission semantics are still ambiguous. Residual `sysParams`/system-variable traces in permission-facing DTOs, comments, and frontend filter components can make unsupported behavior look partially implemented, so this change is needed to turn that deferred state into an explicit and consistent contract.

## What Changes

- Clarify that the current unified permission center supports governed permission workflows for menu permissions, resource permissions, and row/column permissions for supported user/role targets only.
- Remove or explicitly mark misleading system-variable permission affordances in permission-facing contracts, validation, and user-facing copy where they currently imply support that does not exist.
- Align backend permission DTO semantics, validation responses, and frontend permission-facing selectors/messages so unsupported system-variable permission payloads fail explicitly instead of appearing accepted.
- Keep system variable definition/value management separate from permission assignment semantics until a future change intentionally expands permission-center scope.
- **BREAKING**: unsupported system-variable permission request shapes or UI paths that previously appeared available will become explicitly unsupported or hidden, so accidental callers can no longer treat them as partially supported behavior.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `permission-config`: clarify that the unified permission center does not currently provide completed system-variable permission assignment semantics, and that deferred targets must not be exposed as if they are governed and supported.
- `system-variable-management`: clarify that system variable definition/value management APIs remain separate from permission assignment workflows and do not by themselves imply permission-center support for system-variable authorization.

## Impact

- **Affected frontend modules**: permission-center and permission-adjacent auth target UI/contracts, especially views and shared filter components that still surface `sysParams`-style semantics.
- **Affected backend modules**: permission-facing domain/service/handler contracts and validation paths that still imply unsupported system-variable authorization targets.
- **Affected APIs**: permission save/query flows may change from ambiguous acceptance to explicit unsupported responses for system-variable permission semantics.
- **Affected specs/docs**: delta specs for `permission-config` and `system-variable-management`.
- **Dependencies**: no new infrastructure or external service dependencies.
- **Rollback**: if this clarification unexpectedly blocks an internal workflow that depends on the current ambiguous behavior, revert the UI/contract tightening while preserving the explicit deferred classification in follow-up planning so the unsupported path does not continue to look complete.
