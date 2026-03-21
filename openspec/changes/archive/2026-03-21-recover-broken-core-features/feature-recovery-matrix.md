# Feature Recovery Matrix

This matrix freezes the current repository-observed state for the broken non-system-management flows governed by `recover-broken-core-features`.

The first execution wave is limited to the P0 BI recovery families:
- Datasource
- Dataset
- Dashboard
- Big-screen

Export-center and audit remain outside this first batch and must stay in P1 until the P0 BI flows are stabilized.

## Classification vocabulary

- **route/access regression**: feature code likely exists, but route registration, route matching, menu discovery, or access-path wiring blocks reachability
- **API mismatch**: frontend caller and backend route/response contract no longer line up
- **page-init failure**: page route is reachable, but bootstrap or initial query flow fails before the page reaches a usable state
- **state-sync failure**: frontend state, derived interactive state, or compatibility adapter no longer reflects the backend result correctly
- **real implementation gap**: the repository shows a concrete missing, unsupported, or intentionally unavailable sub-capability

## Priority vocabulary

- **P0**: critical user-path failure with bounded verification cost; required before widening scope
- **P1**: important but can wait until the P0 BI critical paths are stabilized
- **deferred**: not suitable for the current batch because of low user impact, unclear classification, or unbounded verification cost

## Failure-semantics vocabulary

- **unauthenticated**: request should fail as authentication-missing behavior, not degrade into forbidden or missing-route behavior
- **forbidden**: authenticated-but-unauthorized behavior must stay distinguishable from missing-route or missing-resource behavior
- **missing route/resource**: route or target resource is absent; must not be normalized into permission denial
- **explicit non-success**: dependency failure, unsupported behavior, real-gap behavior, or business failure must be surfaced explicitly instead of returning silent empty success

## P0 matrix

| Flow family | Governed flow slice | Frontend evidence | Backend evidence | Current classification | Priority | Frontend caller / entry path | Backend owner | Expected usable state | Expected failure semantics | Current verification surface | Missing verification target | Current hypothesis |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Datasource | Datasource management entry, page initialization, and critical browse workflows | `apps/frontend/src/views/visualized/data/datasource/index.vue`, `apps/frontend/src/api/datasource.ts` | `apps/backend-go/internal/transport/http/handler/datasource_handler.go` | route/access regression / page-init failure | P0 | Entry paths are observable in `apps/frontend/src/router/index.ts`: `/#/module-datasource`, `/#/datasource-embedded`; the corresponding edit-path smoke target is defined in `apps/frontend/e2e/recovery/core-reachability.spec.ts` via `/#/data/datasource?id=<id>`. Current caller chain is `listDatasources()`, `listDatasourceTables()`, `previewData()`, and `validate()` in `apps/frontend/src/api/datasource.ts`. | `apps/backend-go/internal/transport/http/handler/datasource_handler.go` → `apps/backend-go/internal/service/datasource_service.go` → `apps/backend-go/internal/repository/datasource_repo.go` | Datasource page reaches usable initialized state for governed browse flows | unauthenticated / forbidden / missing route-resource / explicit non-success | Frontend smoke targets exist in `apps/frontend/e2e/datasource/datasource.spec.ts` and `apps/frontend/e2e/recovery/core-reachability.spec.ts`; backend coverage exists in `apps/backend-go/internal/service/datasource_service_test.go`, `apps/backend-go/internal/service/datasource_service_integration_test.go`, `apps/backend-go/internal/repository/datasource_repo_integration_test.go`, and `apps/backend-go/internal/domain/datasource/datasource_test.go`. | Add route/contract verification that explicitly covers the current frontend `/datasource/*` callers against the Go handler aliases and initialization failures, not just service-layer save/update behavior. | Current symptom is more likely entry/init instability than deleted backend capability |
| Dataset | Dataset management entry, initialization, browse, field, and preview workflows | `apps/frontend/src/views/visualized/data/dataset/index.vue`, `apps/frontend/src/api/dataset.ts` | `apps/backend-go/internal/transport/http/handler/dataset_handler.go` | route/access regression / page-init failure / API mismatch | P0 | Entry paths are observable in `apps/frontend/src/router/index.ts`: `/#/module-dataset`, `/#/dataset-embedded`, `/#/dataset-embedded-form`. Current caller chain is `getDatasetTree()`, `getTableField()`, `getPreviewData()`, `getDatasetDetails()`, and `getDsDetailsWithPerm()` in `apps/frontend/src/api/dataset.ts`. A frontend route-reachability smoke target exists in `apps/frontend/e2e/recovery/core-reachability.spec.ts`. | `apps/backend-go/internal/transport/http/handler/dataset_handler.go` → `apps/backend-go/internal/service/dataset_service.go` → `apps/backend-go/internal/repository/dataset_repo.go` | Dataset page reaches usable initialized state for browse, field, and preview flows | unauthenticated / forbidden / missing route-resource / explicit non-success | Frontend smoke currently appears in `apps/frontend/e2e/recovery/core-reachability.spec.ts`; backend coverage exists in `apps/backend-go/internal/service/dataset_service_test.go`, `apps/backend-go/internal/service/dataset_service_integration_test.go`, `apps/backend-go/internal/repository/dataset_repo_test.go`, `apps/backend-go/internal/repository/dataset_repo_integration_test.go`, and `apps/backend-go/internal/domain/dataset/dataset_test.go`. No dataset-specific frontend test file was identified during this pass. | Add explicit failing coverage for dataset entry init, field metadata, and preview semantics so the current frontend `/datasetTree/*`, `/datasetData/*`, and `/datasetField/*` callers are governed end-to-end rather than only through backend service/repository tests. | Current symptom is more likely route/contract/init drift, with deterministic failure semantics needing governance |
| Dashboard | Dashboard entry-chain plus governed list/tree/detail/discovery usability paths | `apps/frontend/src/views/dashboard/index.vue`, `apps/frontend/src/api/visualization/dataVisualization.ts` | `apps/backend-go/internal/transport/http/handler/visualization_handler.go` | route/access regression / page-init failure / API mismatch / state-sync failure | P0 | Entry paths are observable in `apps/frontend/src/router/index.ts`: `/#/dashboard`, `/#/dashboardPreview`; the corresponding edit-path smoke target is defined in `apps/frontend/e2e/recovery/core-reachability.spec.ts` via `/#/dashboard?resourceId=<id>`. Current caller chain is `findById()`, `queryTreeApi()`, and `queryBusiTreeApi()` in `apps/frontend/src/api/visualization/dataVisualization.ts`. | `apps/backend-go/internal/transport/http/handler/visualization_handler.go` → `apps/backend-go/internal/service/visualization_service.go` → `apps/backend-go/internal/repository/visualization_repo.go` | Dashboard flow reaches usable initialized state for governed entry and discovery paths | unauthenticated / forbidden / missing route-resource / explicit non-success | Frontend smoke currently appears in `apps/frontend/e2e/recovery/core-reachability.spec.ts`; backend coverage exists in `apps/backend-go/internal/transport/http/handler/visualization_handler_test.go`, `apps/backend-go/internal/service/visualization_service_integration_test.go`, `apps/backend-go/internal/repository/visualization_repo_integration_test.go`, and `apps/backend-go/internal/domain/visualization/visualization_test.go`. No dashboard-specific frontend unit test file was identified during this pass. | Current P0 coverage now ties dashboard entry-chain smoke to successful `findById()` consumption and discovery-layout initialization. Further tightening is optional hardening, not a current blocker. | Current symptom is more likely discovery/init/state-consumption instability than a true feature deletion |
| Big-screen | Big-screen entry-chain plus governed tree/detail/discovery usability paths | `apps/frontend/src/views/data-visualization/index.vue`, `apps/frontend/src/views/visualized/view/screen/index.vue`, `apps/frontend/src/api/visualization/dataVisualization.ts` | `apps/backend-go/internal/transport/http/handler/visualization_handler.go` | route/access regression / page-init failure / API mismatch / state-sync failure | P0 | Entry paths are observable in `apps/frontend/src/router/index.ts`: `/#/dvCanvas`, `/#/previewShow`, `/#/preview`, and `/#/de-link/:uuid`; `apps/frontend/src/views/visualized/view/screen/index.vue` shows an additional screen preview surface, but the main editor entry observable in router code is `/#/dvCanvas`. Current caller chain is `findById()`, `queryTreeApi()`, `queryBusiTreeApi()`, and related visualization save/update calls in `apps/frontend/src/api/visualization/dataVisualization.ts`. A frontend route-reachability smoke target exists in `apps/frontend/e2e/recovery/core-reachability.spec.ts` for `/#/dvCanvas`. | `apps/backend-go/internal/transport/http/handler/visualization_handler.go` → `apps/backend-go/internal/service/visualization_service.go` → `apps/backend-go/internal/repository/visualization_repo.go` | Big-screen flow reaches usable initialized state for governed entry and discovery paths | unauthenticated / forbidden / missing route-resource / explicit non-success | Frontend smoke currently appears in `apps/frontend/e2e/recovery/core-reachability.spec.ts`; backend coverage exists in `apps/backend-go/internal/transport/http/handler/visualization_handler_test.go`, `apps/backend-go/internal/service/visualization_service_integration_test.go`, `apps/backend-go/internal/repository/visualization_repo_integration_test.go`, and `apps/backend-go/internal/domain/visualization/visualization_test.go`. No big-screen-specific frontend unit test file was identified during this pass. | Current P0 coverage now distinguishes big-screen preview/discovery entry from route-only reachability and proves payload-consumable discovery layout. Deeper editor/detail hardening is optional, not a current blocker. | Current symptom is more likely route/discovery compatibility drift than missing product behavior |

## P1 matrix stub

| Flow family | Current priority | Why outside P0 |
|---|---|---|
| Export-center | P1 | Operational flow, but not part of the first BI critical-path stabilization batch |
| Audit | P1 | Operational flow, but should remain out of the first batch until datasource, dataset, and visualization are stabilized |

## Known issue-cluster triage

This section classifies the currently documented issue clusters and proof gaps into execution batches. It is intentionally limited to issues already surfaced by the active feature matrix, semantic matrix, evidence baseline, and cut-line documents.

| Issue cluster | Source of record | Current family | Execution batch | Why it belongs there |
|---|---|---|---|---|
| Datasource frontend caller coverage does not yet bind `/datasource/*` callers to Go route aliases and page-init failure behavior | `Missing verification target` in datasource row; datasource sections in `regression-evidence.md` | Datasource | P0 | It directly affects governed datasource entry and browse usability and has bounded verification scope in the first batch |
| Datasource unauthenticated / forbidden semantics are strongest at middleware view paths, not yet at every frontend-facing datasource API alias | datasource row in `permission-semantic-regression-matrix.md`; datasource proof limits in `regression-evidence.md` | Datasource | P0 | It is a governed failure-semantic gap on a P0 family and should be frozen before datasource recovery is considered complete |
| Dataset route reachability is observable, but caller-bound frontend proof for field metadata and preview behavior is still weak | `Missing verification target` in dataset row; dataset frontend proof limits in `regression-evidence.md` | Dataset | P0 | Dataset field and preview behavior are explicitly included in the P0 cut line and remain bounded enough to verify in this batch |
| Dataset 401/403 semantics are strongest on `previewWithPerm`, while tree / fields / preview route families still lack caller-bound semantic freezing | dataset row in `permission-semantic-regression-matrix.md`; dataset semantic proof limits in `regression-evidence.md` | Dataset | P0 | This is a governed semantic gap on critical dataset flows and should not slip to P1 while dataset remains a P0 family |
| Dashboard entry smoke exists, but handler-level proof for missing-resource vs permission-denial behavior on detail paths remains incomplete | dashboard row in `permission-semantic-regression-matrix.md`; visualization route-contract limits in `regression-evidence.md` | Dashboard | P0 | Dashboard detail and discovery semantics are part of the first BI recovery batch and the gap is still bounded around existing handler and contract surfaces |
| Dashboard caller-facing payload-consumption and bootstrap verification is still weaker than route and service proof | `Missing verification target` in dashboard row; visualization frontend proof limits in `regression-evidence.md` | Dashboard | P0 | Payload consumability for governed dashboard entry-chain flows is explicitly in scope for P0 |
| Big-screen route reachability is observable, but direct edit/detail and missing-resource semantics remain weaker than dashboard coverage | big-screen row in `permission-semantic-regression-matrix.md`; visualization semantic proof limits in `regression-evidence.md` | Big-screen | P0 | Big-screen is a first-batch BI family, and these semantics directly affect governed entry-chain recovery rather than optional enhancements |
| Big-screen payload-consumption and deeper editor bootstrap verification remain weaker than shared route smoke | `Missing verification target` in big-screen row; visualization frontend proof limits in `regression-evidence.md` | Big-screen | P0 | The gap is still on a governed P0 critical path and should be addressed before widening scope |
| Export-center query / retry / download recovery | `P1 matrix stub`; `p0-cut-line.md` In P1 section | Export-center | P1 | Explicitly held outside the first BI batch to preserve scope control |
| Audit page reachability / filter-query / detail-read recovery | `P1 matrix stub`; `p0-cut-line.md` In P1 section | Audit | P1 | Explicitly held outside the first BI batch to preserve scope control |
| Standalone real implementation gaps that do not block governed BI usability | `p0-cut-line.md` Deferred section | Cross-family | deferred | They are not suitable for the current batch unless they directly block a governed P0 path and remain bounded |
| Low-value UI glitches, broad compatibility rewrites, and scope-reopening product work | `p0-cut-line.md` Deferred section | Cross-family | deferred | They fail the current P0 bounded-cost rule and should not be silently pulled into recovery |

## Freeze checklist per P0 row

Each P0 row is not considered frozen until all of the following are filled:

- governed user-facing entry path
- governed frontend caller file(s)
- governed backend owner file(s)
- current symptom reproduction note
- current classification
- current verification surface
- missing verification target
- expected usable state
- expected failure semantics
- P0 confirmation

Standalone real implementation gaps should remain P1 or deferred unless they directly block a governed BI critical path and have bounded verification cost inside the P0 batch.

## Batch-entry rule

No P0 repair should begin for a governed flow until:
1. its matrix row is frozen
2. its failure classification is agreed
3. its missing or failing verification target is named
4. its expected usable state and failure semantics are explicit

## Immediate interpretation

The current active change should proceed from the narrower hypothesis that the P0 BI families are primarily suffering from recoverable access-path, contract, initialization, or state-consumption regressions.

This matrix is the execution baseline for:
1. freezing the P0 cut line
2. defining failing verification targets
3. sequencing datasource, dataset, dashboard, and big-screen recovery lanes
4. separating true implementation gaps from governed recoverable regressions
