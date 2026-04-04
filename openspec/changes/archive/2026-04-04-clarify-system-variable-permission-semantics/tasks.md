## 1. Backend permission contract clarification

- [x] 1.1 Audit permission-facing backend DTOs, comments, and validation paths for residual `sysParams` / system-variable authorization-target semantics, and list the exact files that must be clarified without changing supported `/sysVariable/*` CRUD behavior.
- [x] 1.2 Update backend permission-facing contracts and validation so unsupported system-variable permission targets are explicitly rejected or documented as unsupported, while keeping supported `user` / `role` behavior unchanged.
- [x] 1.3 Add or update backend tests covering explicit rejection/unsupported handling for system-variable permission-assignment semantics and confirming existing system-variable management CRUD contracts still pass.

## 2. Frontend permission-affordance clarification

- [x] 2.1 Audit permission-center and permission-adjacent frontend components that still expose `sysParams` semantics, and classify each path as governed user flow vs internal/deferred-only usage.
- [x] 2.2 Remove or replace misleading governed-flow `sysParams` permission affordances with explicit unsupported messaging or hidden states, without regressing supported permission-center tabs.
- [x] 2.3 Add or update frontend tests for the clarified UI/contract behavior, and rerun the affected frontend quality gates for the touched modules.

## 3. Verification and rollout

- [x] 3.1 Run backend verification for this change (`make test`, and `TEST_DB_HOST=127.0.0.1 make test-integration` if persistence-facing permission contracts are touched) and record any pre-existing failures separately from change-induced failures.
- [x] 3.2 Run frontend verification for this change (`npm run lint`, `npm run ts:check`, and affected tests/specs for clarified permission-center flows) and confirm no supported menu/resource/row-column workflows regress.
- [x] 3.3 Perform a final scope check confirming this change only clarifies deferred system-variable permission semantics and does not implement middleware replacement, whitelist persistence, or broader P2 work.
