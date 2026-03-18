# Core BI Endpoint Baseline

This document freezes the repository-observed baseline for tasks `1.1` through `1.5` of `stabilize-go-core-bi-compatibility`.

## Path entry model

- Frontend dev mode calls `VITE_API_BASEPATH=/api` (`apps/frontend/.env.dev`).
- Frontend base/distributed mode calls `VITE_API_BASEPATH="./de2api"` (`apps/frontend/.env.base`).
- Backend rewrites `/de2api/*` to `/api/*` in `rewriteCompatibilityPath` (`apps/backend-go/internal/transport/http/router.go`).
- As a result, BI traffic must be treated as a route family with two public entry paths:
  - canonical entry: `/api/...`
  - compatibility entry: `/de2api/...` → rewritten to `/api/...`

## Status vocabulary used in this baseline

- **full**: handler/service path exists and is already governed by whitelist metadata or strict compatibility checks.
- **partial**: handler exists, but current implementation is synthetic/limited or only partially governed.
- **stub**: route intentionally returns deterministic non-success or placeholder behavior instead of business behavior.
- **missing**: no supported route/handler path was found.

## Core BI inventory

| Domain | Frontend wrapper / caller evidence | Canonical Go route family | Compatibility / frontend route family | Backend handler + service | Observed status | Gate coverage status |
|---|---|---|---|---|---|---|
| Datasource list / validate / tree | `apps/frontend/src/api/datasource.ts`; callers in `views/visualized/data/datasource/index.vue`, `views/visualized/data/datasource/form/CreatDsGroup.vue`, `store/modules/interactive.ts` | `/api/ds/list`, `/api/ds/validate` via `DatasourceHandler` | `/api/datasource/list`, `/api/datasource/tree`, `/api/datasource/validate`, plus `/de2api/...` alias | `internal/transport/http/handler/datasource_handler.go` → `internal/service/datasource_service.go`; compatibility surface in `internal/transport/http/handler/compatibility_bridge_handler.go` | `full` for `/datasource/list`, `/datasource/tree`, `/datasource/validate`; canonical `/api/ds/*` implemented but not yet required-gate metadata | `critical-whitelist.yaml` already covers `/datasource/list`, `/datasource/tree`, `/datasource/validate`, `/datasource/getTables`, `/datasource/previewData`, `/datasource/save` |
| Dataset tree / fields / preview | `apps/frontend/src/api/dataset.ts`; callers in `views/dashboard/index.vue`, `views/data-visualization/index.vue`, `views/visualized/data/dataset/form/index.vue`, `store/modules/interactive.ts` | `/api/dataset/tree`, `/api/dataset/fields`, `/api/dataset/preview`, `/api/dataset/previewWithPerm` via `DatasetHandler` | `/api/datasetTree/tree`, `/api/datasetData/tableField`, `/api/datasetData/previewData`, plus `/de2api/...` alias | `internal/transport/http/handler/dataset_handler.go` → `internal/service/dataset_service.go`; compatibility surface in `internal/transport/http/handler/compatibility_bridge_handler.go` | `full` for tree / field / preview compatibility family; `previewWithPerm` exists only on canonical Go route family | `critical-whitelist.yaml` already covers `/datasetTree/tree`, `/datasetData/tableField`, `/datasetData/previewData`; governance test marks `/datasetData/previewSql` as `partial` |
| Dashboard + big-screen resource tree / detail | `apps/frontend/src/api/visualization/dataVisualization.ts`; callers in `views/common/DeResourceGroupOpt.vue`, `custom-component/de-screen/SelectScreenDialog.vue`, `components/visualization/LinkJumpSet.vue`, `utils/canvasUtils.ts` | `/api/dataVisualization/tree`, `/api/dataVisualization/findById`, `/api/dataVisualization/findDvType`, `/api/dataVisualization/move`, `/api/dataVisualization/copy`, `/api/dataVisualization/deleteLogic/:id/:busiFlag` via `VisualizationHandler` | `/de2api/dataVisualization/tree` aliases to `/api/dataVisualization/tree`; frontend also calls `/api/dataVisualization/interactiveTree` and `/de2api/dataVisualization/interactiveTree` | `internal/transport/http/handler/visualization_handler.go` → `internal/service/visualization_service.go`; compatibility-only interactive tree in `internal/transport/http/handler/frontend_compat_handler.go` | `full` for `/dataVisualization/tree` and `/dataVisualization/findById` implementation; `partial` for `/dataVisualization/interactiveTree` because current response is authorization-derived synthetic root nodes, not actual visualization resource data | strict compat script covers `/api/dataVisualization/tree` and `/de2api/dataVisualization/tree`; `critical-whitelist.yaml` does **not** currently cover visualization tree/detail/interactive-tree routes |
| BI navigation compatibility that gates datasource / dataset / dashboard / screen visibility | frontend permission/navigation consumers rely on runtime menu loading; evidence in `frontend_compat_handler.go` and menu-driven callers | n/a as standalone BI CRUD route family | `/api/roleRouter/query`, `/api/auth/menuResource`, `/api/dataVisualization/interactiveTree`, plus `/de2api/...` aliases | `internal/transport/http/handler/frontend_compat_handler.go` backed by `MenuService` and role lookup wiring in `router.go` | `partial` for BI resource discovery because menu visibility is dynamic, but interactive tree is synthetic and not tied to actual visualization records | strict compat script covers permission admin APIs and `/dataVisualization/tree`, but not `roleRouter/query`, `auth/menuResource`, or `interactiveTree` in required-gate metadata |

## Envelope and error semantics baseline

Repository-wide response semantics are defined in `apps/backend-go/internal/pkg/response/response.go`:

- success: HTTP `200`, `code="000000"`, `msg="success"`
- generic validation / business error: HTTP `200`, non-success code such as `500000`
- unauthenticated: HTTP `401`, `code="20001"`
- forbidden: HTTP `403`, `code="70001"`
- not-found semantic helper exists as `code="50001"`, but core BI routes should avoid misclassifying permission denial as generic missing-resource behavior

This change will treat the following as the frozen baseline for tasks `1.3` and `1.4`:

1. Canonical Go BI routes must preserve `code/data/msg` envelopes even when entered through `/de2api` rewrite.
2. Unauthorized BI access must remain distinguishable from not-found routes.
3. Visualization interactive-tree compatibility is currently weaker than visualization tree/detail and should be treated as **partial**, not silently assumed equivalent.

## Current gate metadata findings

### Already governed

- `apps/backend-go/testdata/contract-diff/critical-whitelist.yaml`
  - includes datasource compatibility routes such as `/datasource/list`, `/datasource/tree`, `/datasource/validate`
  - includes dataset compatibility routes such as `/datasetTree/tree`, `/datasetData/tableField`, `/datasetData/previewData`
- `apps/backend-go/scripts/compat-checks/run_auth_visualization_compat.sh`
  - executes `/api/...` and `/de2api/...` checks for permission APIs and `/dataVisualization/tree`
- `apps/backend-go/internal/transport/http/handler/compatibility_governance_test.go`
  - asserts placeholder-success governance and documents partial endpoints like `/datasetData/previewSql`

### Not yet governed as required-gate scope

- `/api/ds/list`
- `/api/ds/validate`
- `/api/dataset/tree`
- `/api/dataset/fields`
- `/api/dataset/preview`
- `/api/dataVisualization/tree`
- `/api/dataVisualization/findById`
- `/api/dataVisualization/interactiveTree`
- `/api/roleRouter/query`
- `/api/auth/menuResource`

## Required-gate candidates for this change

These routes should become explicit required-gate candidates before release hardening is considered complete:

### Tier A — core BI route families

1. `POST /datasource/tree`
2. `POST /datasource/validate`
3. `POST /datasetTree/tree`
4. `POST /datasetData/tableField`
5. `POST /datasetData/previewData`
6. `POST /dataVisualization/tree`
7. `POST /dataVisualization/findById`

### Tier B — canonical Go parity coverage

1. `POST /api/ds/list`
2. `POST /api/ds/validate`
3. `POST /api/dataset/tree`
4. `POST /api/dataset/fields`
5. `POST /api/dataset/preview`
6. `POST /api/dataVisualization/tree`
7. `POST /api/dataVisualization/findById`

### Tier C — compatibility discovery and permission surfaces

1. `POST /api/dataVisualization/interactiveTree`
2. `GET /api/roleRouter/query`
3. `GET /api/auth/menuResource`

Tier C should not be treated as optional, because the frontend uses these routes to discover whether dashboard, big-screen, dataset, and datasource resources are visible at all.

## Immediate implementation implications

- Tasks `1.1` and `1.2` are satisfied by the inventory above.
- Task `1.3` is satisfied by freezing repository-observed statuses (`full` / `partial`) for the in-scope BI families.
- Task `1.4` is satisfied by freezing `code/data/msg`, `401`, and `403` semantics from `response.go` as the contract baseline.
- Task `1.5` is satisfied by the required-gate candidate lists above; actual whitelist and gate updates are deferred to task group `7.x`.
