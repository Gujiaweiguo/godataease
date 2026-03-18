## Verification Report: stabilize-go-core-bi-compatibility

### Summary

| Dimension | Status |
|---|---|
| Completeness | 44/44 tasks complete; 5/5 delta spec files present |
| Correctness | 5/5 capability deltas have implementation evidence and regression evidence |
| Coherence | Design decisions followed; 1 documented follow-up suggestion |

### Completeness

- **Tasks:** all checkboxes in `openspec/changes/stabilize-go-core-bi-compatibility/tasks.md` are complete.
- **Artifacts:** `proposal.md`, `design.md`, `specs/**/*.md`, and `tasks.md` are all `done` per `openspec status --change "stabilize-go-core-bi-compatibility" --json`.
- **Evidence:** implementation and verification records exist in:
  - `openspec/changes/stabilize-go-core-bi-compatibility/regression-evidence.md`
  - `openspec/changes/stabilize-go-core-bi-compatibility/implementation-status.md`
  - `openspec/changes/stabilize-go-core-bi-compatibility/frontend-compat-inventory.md`
  - `openspec/changes/stabilize-go-core-bi-compatibility/permission-semantic-regression-matrix.md`

### Correctness

- **Datasource stability requirements** are backed by route, service, and permission tests in:
  - `apps/backend-go/internal/transport/http/router_test.go`
  - `apps/backend-go/internal/service/datasource_service_test.go`
  - `apps/backend-go/internal/transport/http/middleware/permission_integration_test.go`
- **Dataset stability requirements** are backed by deterministic preview/field/tree and datasource-dependency checks in:
  - `apps/backend-go/internal/service/dataset_service.go`
  - `apps/backend-go/internal/service/dataset_service_integration_test.go`
  - `apps/backend-go/internal/transport/http/handler/dataset_handler.go`
- **Visualization stability requirements** are backed by route/tree/detail tests in:
  - `apps/backend-go/internal/transport/http/router_test.go`
  - `apps/backend-go/internal/transport/http/handler/visualization_handler_test.go`
  - `apps/backend-go/internal/service/visualization_service_integration_test.go`
- **Compatibility bridge requirements** are backed by whitelist/governance/evidence updates in:
  - `apps/backend-go/testdata/contract-diff/critical-whitelist.yaml`
  - `apps/backend-go/internal/transport/http/handler/compatibility_governance_test.go`
  - `apps/backend-go/scripts/check-status-drift.sh`
- **Permission stability requirements** are backed by:
  - `apps/backend-go/internal/transport/http/middleware/permission_integration_test.go`
  - `apps/backend-go/internal/transport/http/handler/permission_compat_handler.go`
  - `apps/backend-go/internal/transport/http/handler/permission_compat_handler_test.go`
- **Frontend compatibility convergence** is backed by:
  - `apps/frontend/src/store/modules/interactive.ts`
  - `apps/frontend/tests/unit/store/interactive.test.ts`
  - `apps/frontend/e2e/interactive/interactive.spec.ts`

### Coherence

- **Design decision followed: contract-first hardening**
  - Evidence: route-contract regression coverage and evidence-first docs were added before broader smoke claims.
- **Design decision followed: compatibility aliases and canonical routes jointly governed**
  - Evidence: datasource, dataset, and visualization route tests cover `/api/...`, compatibility aliases, and `/de2api/...` entry paths.
- **Design decision followed: permission semantics are first-class**
  - Evidence: explicit `401`/`403` regression matrix and handler/service behavior for datasource, dataset, dashboard, and screen flows.

### Issues by Priority

#### CRITICAL

- None.

#### WARNING

- None.

#### SUGGESTION

- `dataVisualization/interactiveTree` remains intentionally marked `partial` in governed metadata (`apps/backend-go/testdata/contract-diff/critical-whitelist.yaml`).
  - Recommendation: treat full parity for this endpoint as a follow-up change if the frontend needs more than the current authorized synthetic-root behavior.

### Final Assessment

No critical issues found. No warnings found. The change is ready for archive once you decide whether to sync the delta specs into the main specs first.
