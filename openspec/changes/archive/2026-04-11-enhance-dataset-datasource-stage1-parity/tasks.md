## 1. Dataset export and field-delete backend implementation

- [x] 1.1 Audit the existing Go export-center and dataset service wiring, then define the minimal dataset export request/response path for `POST /datasetTree/exportDataset` with unit-testable success and failure cases.
- [x] 1.2 Implement dataset export service and compatibility-bridge handler wiring so `POST /datasetTree/exportDataset` starts a real export workflow instead of returning `not supported`, and add backend tests for task creation / delegation semantics.
- [x] 1.3 Add repository support for deleting dataset fields by field ID and by chart ID, including scoped lookup helpers needed to avoid deleting unrelated fields, with repository tests.
- [x] 1.4 Implement dataset service delete flows for `POST /datasetField/delete/{id}` and `POST /datasetField/deleteByChartId/{id}`, including missing-resource, dependency-blocked, and success paths, with service unit/integration tests.
- [x] 1.5 Replace compatibility stubs in the dataset compatibility bridge with the new service-backed behavior and add handler/router regression coverage for canonical and compatibility aliases.

## 2. Datasource table-status and delete-path implementation

- [x] 2.1 Audit available datasource sync evidence sources (task logs, sync records, table metadata) and define the first-stage status mapping set (`success`, `running`, `failed`, `unknown` or equivalent) with backend tests.
- [x] 2.2 Implement `POST /datasource/getTableStatus` using real synchronization evidence plus explicit unknown-state fallback, and add service/repository tests for each status branch.
- [x] 2.3 Add a normative datasource delete write route in router/handler wiring while preserving the existing compatibility delete alias, with both paths delegated to the same datasource service delete logic.
- [x] 2.4 Add regression tests proving the normative datasource delete route and historical compatibility delete route share the same pre-delete validation, permission semantics, and failure behavior.

## 3. Frontend API and flow alignment

- [x] 3.1 Update `apps/frontend/src/api/dataset.ts` and affected dataset views so dataset export and field deletion call the real backend paths and handle deterministic non-success responses, with affected frontend tests where available.
- [x] 3.2 Update `apps/frontend/src/api/datasource.ts` and affected datasource views so table-status rendering consumes the new real status values and datasource deletion prefers the normative write route while keeping compatibility-safe fallback behavior if needed.

## 4. End-to-end verification and rollout safety

- [x] 4.1 Add or update backend unit tests and MySQL integration tests covering dataset export delegation, dataset field deletion, datasource table status aggregation, and dual delete-route consistency.
- [x] 4.2 Run required backend validation (`make test`, plus `make test-integration` if persistence changes require it) and fix any regressions introduced by this change.
- [x] 4.3 Run required frontend validation (`npm run lint`, `npm run ts:check`, and `npm run test:affected:datasource` or the closest affected suite) and fix any regressions introduced by this change.
- [x] 4.4 Perform manual regression verification for the primary UI flows: dataset export, dataset field delete, datasource table status display, and datasource delete from the management page; record any fallback or rollback actions needed before merge.
