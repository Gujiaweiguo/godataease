# Regression Evidence

This document records concrete verification evidence produced while implementing `recover-export-audit-operational-flows`.

## Audit evidence

### Handler boundary evidence

- 2026-03-21: `cd apps/backend-go && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -v -run 'TestAuditHandler_(CreateAuditLog|QueryAuditLogs|GetAuditLogByID)_Success' ./internal/transport/http/handler -count=1` → PASS.

Covered outcomes:
- `POST /audit/log` returns a valid success envelope for a well-formed audit log create request.
- `GET /audit/list` returns a valid success envelope with paginated audit log results.
- `GET /audit/:id` returns a valid success envelope for an existing audit log detail read.

Current proof limits:
- This first slice covers audit handler HTTP boundaries for create/list/detail happy paths only.
- Audit page entry and query initialization now have frontend unit coverage, but route-level smoke and broader explicit non-success UI paths are still open work.

### Frontend page-entry and query-init verification

- 2026-03-21: `cd apps/frontend && npm run test -- --run tests/unit/audit/index.test.ts` → PASS.

Covered outcomes:
- The audit page triggers `queryAuditLogsApi({})` on mount.
- Successful query initialization updates pagination metadata.
- Non-success query results surface explicit error messaging instead of silently rendering a misleading empty-success state.

### Route and authorization repair evidence

- 2026-03-21: `cd apps/backend-go && go test -v -run 'TestRegisterRoutes_AuditRoutesRequireAuthentication$' ./internal/transport/http` → PASS.
- 2026-03-21: after rebuilding the backend and restarting the dev container stack, live probes to `GET /api/audit/list` and `GET /api/audit/999999999` without authorization both returned `401 + 20001`.
- 2026-03-21: `cd apps/backend-go && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -v -run 'TestAuditHandler_(CreateAuditLog|QueryAuditLogs|GetAuditLogByID)_' ./internal/transport/http/handler -count=1` → PASS.

Covered outcomes:
- Governed audit list and detail-read routes no longer degrade unauthenticated access into business-shaped responses such as list success or `audit log not found`.
- Audit route semantics are now explicit at the HTTP boundary for unauthenticated callers.
- Existing audit detail reads now preserve a direct not-found envelope (`50001`, `Audit log not found`) instead of collapsing into a misleading success shape.

## Export-center evidence

### Frontend page-init/query verification

- 2026-03-21: `cd apps/frontend && npm run test -- --run tests/unit/dataset/exportExcel.test.ts` → PASS.
- 2026-03-21: `cd apps/frontend && npm run test -- --run tests/unit/dataset/api.test.ts` → PASS.

Covered outcomes:
- `ExportExcel.init({ activeName: 'FAILED' })` opens the drawer and issues the first export-center query with the requested tab state.
- `ExportExcel.init({ activeName: 'IN_PROGRESS' })` starts the polling loop and reissues `exportTasksRecords()` plus `exportTasks()` on the 5-second timer.
- Export-center API wrappers for `exportTasksRecords`, `exportTasks`, and `exportRetry` call the expected compatibility endpoints and preserve the response payloads consumed by the frontend.

### Route/contract repair evidence

- 2026-03-21: live probes showed bare `POST /exportCenter/exportTasks/records` and `POST /exportCenter/exportTasks/all/1/10` returning frontend HTML fallback, while `POST /api/exportCenter/exportTasks/records` and `POST /api/exportCenter/exportTasks/all/1/10` returned valid JSON envelopes.
- 2026-03-21: export-center frontend callers in `src/api/dataset.ts` were updated to use `/api/exportCenter/*` for `exportTasksRecords`, `exportTasks`, and `exportRetry`.
- 2026-03-21: `cd apps/frontend && npm run test -- --run tests/unit/dataset/api.test.ts tests/unit/dataset/exportExcel.test.ts` → PASS after the caller-path fix.
- 2026-03-21: `cd apps/backend-go && go test -v -run 'TestRegisterRoutes_ExportCenterRoutesRequireAuthentication$' ./internal/transport/http` → PASS.
- 2026-03-21: after rebuilding the backend and restarting the dev container stack, live probes to `POST /api/exportCenter/exportTasks/records` and `POST /api/exportCenter/exportTasks/all/1/10` without authorization both returned `401 + 20001`.

Current proof limits:
- This slice covers export-center frontend page-init/query behavior only.
- Retry wrapper routing and explicit unauthenticated query semantics are now covered, but download and broader route-level smoke are still open work.

## Cross-module operational verification

### Targeted recovered-scope verification

- 2026-03-21: `cd apps/frontend && npm run test -- --run tests/unit/dataset/exportExcel.test.ts tests/unit/dataset/api.test.ts tests/unit/audit/index.test.ts` → PASS.
- 2026-03-21: `cd apps/backend-go && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -v -run 'TestAuditHandler_(CreateAuditLog|QueryAuditLogs|GetAuditLogByID)_Success|TestAuditHandler_GetAuditLogByID_NotFound' ./internal/transport/http/handler -count=1 && go test -v -run 'TestRegisterRoutes_(AuditRoutesRequireAuthentication|ExportCenterRoutesRequireAuthentication)$' ./internal/transport/http -count=1` → PASS.

Covered outcomes:
- Export-center recovered query/retry caller paths remain stable under targeted frontend regression coverage and explicit auth-route checks.
- Audit recovered entry/query/detail paths remain stable under targeted frontend/backend regression coverage.

### Frontend closeout status

- 2026-03-21: `cd apps/frontend && npm run lint` → PASS.
- 2026-03-21: `cd apps/frontend && npm run ts:check` → PASS after clearing embedding, datasource test, app test, interactive test, permission test, and Date augmentation blockers.

Current proof limits:
- Frontend closeout is green for the current environment and recovered operational scope.

### Backend closeout status

- 2026-03-21: `cd apps/backend-go && make test` → PASS.
- 2026-03-21: `cd apps/backend-go && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test make test-integration` → PASS after serializing integration package execution in the Makefile with `-p 1` to avoid cross-package cleanup races against the shared MySQL test database.
- 2026-03-21: `cd apps/backend-go && make drift-check` → PASS.
- 2026-03-21: the export-center and audit recovered-scope backend tests remained green throughout this follow-up change.

Current proof limits:
- Release-readiness is no longer blocked in the current environment.
