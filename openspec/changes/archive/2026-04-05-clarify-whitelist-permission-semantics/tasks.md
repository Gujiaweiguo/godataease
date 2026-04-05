## 1. Backend whitelist contract clarification

- [x] 1.1 Audit the row-permission read/write contract surfaces that still expose whitelist semantics (`RowPermissionForm`, `RowPermissionDTO`, `DataPermRow`, and row-permission page shaping) and confirm which fields stay compatibility-safe versus which messages must change.
- [x] 1.2 Replace the current `whiteList is not supported in T8` validation with a stable deferred-semantics error message, keeping supported user/role row-permission behavior unchanged.
- [x] 1.3 Add or update backend tests covering the stable deferred whitelist rejection contract and any agreed read-path shaping expectations for whitelist-related fields.

## 2. Unified permission-center boundary clarification

- [x] 2.1 Audit the unified permission-center row/column UI and confirm that governed whitelist editing remains unavailable while adjacent legacy flows are treated as explicit deferred boundaries.
- [x] 2.2 Update frontend permission-center wording or regression coverage as needed so the governed row-permission flow does not imply whitelist editing support.
- [x] 2.3 Add or update affected frontend tests/specs to prove the unified permission center still does not offer whitelist editing and remains consistent with the backend deferred contract.

## 3. Verification and scope control

- [x] 3.1 Run backend verification for the touched whitelist-contract files (`make test`, and `TEST_DB_HOST=127.0.0.1 make test-integration` only if persistence-facing behavior changes beyond validation/response semantics).
- [x] 3.2 Run frontend verification for the touched permission-center files (`npm run lint`, `npm run ts:check`, and affected tests/specs such as the data-permission smoke path).
- [x] 3.3 Perform a final scope check confirming this change only clarifies deferred whitelist semantics and does not implement whitelist persistence, runtime whitelist enforcement, middleware replacement, or broader P2 cleanup.
