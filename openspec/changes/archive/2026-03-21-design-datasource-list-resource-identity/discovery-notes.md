# Datasource List Resource Identity Discovery Notes

This note captures grounded findings for tasks `1.1`–`1.3` of `design-datasource-list-resource-identity`.

## Backend route and semantics inventory

### Canonical and compatibility aliases

The current datasource list runtime is exposed through three equivalent aliases:

- `POST /api/ds/list`
- `POST /api/datasource/list`
- `POST /de2api/datasource/list`

All three converge on the same runtime path:

- `apps/backend-go/internal/transport/http/handler/datasource_handler.go` → `DatasourceHandler.List()`
- `apps/backend-go/internal/service/datasource_service.go` → `DatasourceService.List()`
- `apps/backend-go/internal/repository/datasource_repo.go` → `DatasourceRepository.Query()`

### Current auth behavior

The list aliases are protected by JWT auth at the route level via `middleware.Auth(...)` in `apps/backend-go/internal/transport/http/router.go`.

What is currently proven:

- unauthenticated callers receive explicit auth failure
- no route-level permission middleware is attached to datasource list aliases

### Current permission model assumptions

Datasource detail/view permission checks already exist in `apps/backend-go/internal/transport/http/middleware/permission_middleware.go` through `CheckDatasourceView()`.

However, that middleware requires `extractResourceID()` to find a resource identifier from one of:

- path param `:id`
- query param `id` / `resourceId`
- request body field `id` / `resourceId`

Datasource list requests do not currently carry any of those fields in the runtime contract.

### Runtime request shape

`apps/backend-go/internal/domain/datasource/datasource.go` defines the current list request as:

- `keyword`
- `current`
- `size`

There is no stable resource identifier, datasource group identifier, workspace identifier, or other governing scope field in the current list contract.

### Current runtime semantics classification

Based on the live route shape and service/repository behavior, datasource list semantics are currently best described as:

## Current classification: auth-only list behavior

Why:

- route-level JWT auth exists
- route-level permission middleware does not exist
- service/repository list logic does not apply per-user filtering or permission-aware scoping
- the request contract does not expose a stable scope for explicit forbidden semantics

### Important distinction: middleware proof vs runtime behavior

Hardening work already added middleware-level forbidden proof for datasource list aliases by sending a synthetic request body with `{"id":123}`.

That proof is useful to show the permission boundary logic itself can produce `403`, but it does **not** mean the live runtime list endpoints currently support resource-bound forbidden semantics.

## Provisional conclusions for tasks 1.2 and 1.3

- **Task 1.2 (stable governing scope):** no existing stable governing scope has been found yet in the backend request shape.
- **Task 1.3 (current runtime semantics):** backend evidence currently supports `auth-only list` as the best-fit classification.

## Frontend caller inventory

### Primary wrapper functions

The current frontend does not call `POST /api/ds/list` directly. Production callers are centralized around datasource tree helpers that hit `/datasource/tree` with `busiFlag: 'datasource'`.

Key wrapper functions:

- `apps/frontend/src/api/datasource.ts`
  - `listDatasources()`
  - `getDsTree()`
- `apps/frontend/src/api/dataset.ts`
  - `getDatasourceList()`

All three wrappers inject `busiFlag: 'datasource'` and do not carry a stable datasource resource identity.

### Production caller sites

Current known production/runtime callers include:

- `apps/frontend/src/views/visualized/data/datasource/index.vue`
  - uses `interactiveStore.setInteractive({ busiFlag: 'datasource' })`
- `apps/frontend/src/views/visualized/data/datasource/form/CreatDsGroup.vue`
  - calls `listDatasources({ leaf: false, id, weight: 7 })`
- `apps/frontend/src/views/visualized/data/dataset/form/index.vue`
  - calls `getDatasourceList(weight)`
- `apps/frontend/src/views/visualized/data/dataset/form/AddSql.vue`
  - calls `getDatasourceList()`
- `apps/frontend/src/views/chart/components/editor/dataset-select/DatasetSelect.vue`
  - calls `getDatasourceList(null)` when `sourceType === 'datasource'`
- `apps/frontend/src/store/modules/interactive.ts`
  - maps datasource loading through `listDatasources()` via `apiMap[3]`

### Route-entry context

Datasource route entries currently converge on the same datasource page component:

- `/#/datasource-embedded`
- `/#/module-datasource`

This means a runtime semantics change must remain compatible with both entry points and with the shared interactive loading path.

### Test and support callers

Known test/runtime helper references include:

- `apps/frontend/tests/unit/datasource/api.test.ts`
- `apps/frontend/tests/unit/store/interactive.test.ts`
- `apps/frontend/e2e/datasource/datasource.spec.ts`
- `apps/frontend/e2e/recovery/core-reachability.spec.ts`
- `apps/frontend/e2e/interactive/interactive.spec.ts`
- `apps/frontend/e2e/embedding/embedding.spec.ts`

These tests rely on the same wrapper/route shape and therefore reinforce the current caller expectation that datasource list behaves like a broad collection load, not a resource-addressed detail request.

### Caller-side scope observations

The only caller-side fields that look remotely scope-like today are:

- `busiFlag: 'datasource'`
- optional `weight`
- occasional UI helper params such as `leaf` or local folder `id`

None of these behave like a stable datasource resource identity comparable to detail-path `id` or `resourceId` semantics.

## Combined discovery conclusions for tasks 1.1–1.3

- **Task 1.1 — caller inventory:** complete. Current datasource list callers are centralized in wrapper functions and interactive/page entry flows that all assume broad collection loading.
- **Task 1.2 — stable governing scope:** no existing stable governing scope has been found in either the backend list contract or current frontend wrapper usage.
- **Task 1.3 — current runtime semantics:** current datasource list behavior is best classified as **auth-only list behavior**, not scoped forbidden behavior and not explicit filtered-permission behavior.

## Selection handoff to tasks 2.x and 3.x

Based on the discovery above, the design now selects **Option C**:

- keep datasource list aliases as broad auth-only collection-load routes
- keep explicit forbidden runtime guarantees on detail/read paths that already carry stable resource identity
- treat list-alias forbidden `403` coverage as middleware-boundary proof only, not as a runtime route guarantee
