## 1. Dataset export execution lifecycle

- [x] 1.1 Audit the existing export-center query/download paths and define the minimal canonical contract for dataset-origin export tasks, including task identity, status lookup, and download handoff semantics.
- [x] 1.2 Implement backend dataset export lifecycle wiring so dataset-origin export tasks can be queried and consumed through the export-center flow after `POST /datasetTree/exportDataset`, with unit tests for task creation, status lookup, and failure diagnostics.
- [x] 1.3 Refine compatibility dataset identity handling for export flows so historical or compatibility dataset IDs resolve deterministically without polluting canonical dataset identity behavior, with regression coverage for true-not-found vs compat-resolved cases.

## 2. Dataset field dependency-aware delete blocking

- [x] 2.1 Audit dataset field downstream dependency sources (chart references, calculation fields, governed configuration consumers) and codify the blocking categories the backend must enforce.
- [x] 2.2 Implement repository/service dependency scanning for `POST /datasetField/delete/{id}` and `POST /datasetField/deleteByChartId/{id}`, returning explicit dependency-blocked semantics instead of generic failure when protected references exist.
- [x] 2.3 Add backend unit and MySQL integration tests covering dependency-blocked field deletion, dependency-free success, and distinguishable error semantics for missing field vs blocked field.

## 3. Datasource table status semantic upgrade

- [x] 3.1 Audit datasource runtime evidence sources and define the stage2 stable state set plus authoritative update-time precedence used by `POST /datasource/getTableStatus`.
- [x] 3.2 Implement backend datasource status mapping so table status returns the documented stable execution states and deterministic update timestamps without leaking raw task-system enums.
- [x] 3.3 Add backend unit and integration tests for each supported datasource table status branch, including incomplete-evidence fallback semantics.

## 4. Frontend contract alignment

- [x] 4.1 Update frontend dataset export flows to consume the stage2 export task lifecycle, including task lookup, failure handling, and any export-center integration points required by the new backend contract.
- [x] 4.2 Update frontend dataset field deletion flows to surface dependency-blocked outcomes with explicit user-facing feedback instead of generic failure handling.
- [x] 4.3 Update frontend datasource status rendering to consume the stage2 stable execution-state set and authoritative update-time semantics without depending on backend raw task enums.

## 5. Verification and rollout safety

- [x] 5.1 Run required backend validation (`make test` and `make test-integration` if persistence or runtime evidence handling changes) and fix any regressions introduced by stage2 execution semantics.
- [x] 5.2 Run required frontend validation (`npm run lint`, `npm run ts:check`, and affected datasource/dataset test suites) and fix any regressions introduced by stage2 UI contract updates.
- [x] 5.3 Perform manual regression verification for dataset export lifecycle consumption, dependency-blocked field deletion, and datasource table status display semantics; record any rollback levers required before merge.
