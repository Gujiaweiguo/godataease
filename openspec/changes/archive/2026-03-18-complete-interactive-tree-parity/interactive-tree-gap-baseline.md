# Interactive Tree Gap Baseline

This document freezes the starting gap for `complete-interactive-tree-parity`.

## Current backend behavior

### Current route path

- `POST /api/dataVisualization/interactiveTree`
- `POST /dataVisualization/interactiveTree`
- `POST /de2api/dataVisualization/interactiveTree`

All three route entries are handled by `FrontendCompatHandler.InteractiveTree` in:

- `apps/backend-go/internal/transport/http/handler/frontend_compat_handler.go:103-120`

### Current implementation path

`InteractiveTree` currently does **not** query visualization resources.

Instead it:
1. loads authorized menus with `loadRuntimeMenus`
2. derives allowed BI flags using `collectAuthorizedBusiFlags` (`frontend_compat_handler.go:178-209`)
3. returns synthetic roots through `buildInteractiveTreeResponse` (`frontend_compat_handler.go:211-227`)

For authorized dashboard / dataV scopes, the returned node is a synthetic root:

- `id: "0"`
- `pid: "-1"`
- `name: busiFlag`
- `leaf: false`
- `weight: 9`
- `extraFlag: 1`
- `extraFlag1: 0`
- `children: []`

For unauthorized scopes, the route returns an empty list.

### Current test evidence of synthetic behavior

- `apps/backend-go/internal/transport/http/handler/frontend_compat_handler_test.go:270-328`
  - verifies `dashboard` returns a single synthetic node with `id="0"`, `pid="-1"`, and `name="dashboard"`
  - verifies `dataset` returns the same synthetic pattern
  - verifies unauthorized `datasource` returns an empty list

## Frontend contract requirements

The active frontend tree contract is defined by `BusiTreeNode` in:

- `apps/frontend/src/models/tree/TreeNode.ts:1-10`

Required fields consumed by current interactive consumers:

- `id`
- `pid`
- `name`
- `leaf`
- `weight`
- `extraFlag`
- `extraFlag1`
- `children`

Primary active consumers:

- `apps/frontend/src/store/modules/interactive.ts`
- `apps/frontend/src/views/common/DeResourceTree.vue`
- `apps/frontend/src/views/workbranch/index.vue`
- `apps/frontend/src/views/mobile/directory/index.vue`
- `apps/frontend/src/components/visualization/LinkJumpSet.vue`

Current store behavior relevant to parity:

- `interactiveStore.loadBusiInteractive()` batches dashboard / dataV / dataset / datasource calls through `queryBusiTreeApi`
- `interactiveStore.convertInteractive()` computes:
  - `rootManage`
  - `anyManage`
  - `leafNodeCount`
- `normalizeNodeIds()` converts `id` and `pid` to strings

This means parity work must keep the existing node contract intact even if backend starts returning real visualization resources.

## Existing real visualization tree logic

The closest reusable real tree implementation already exists in:

- `apps/backend-go/internal/transport/http/handler/visualization_handler.go:74-119`
- `apps/backend-go/internal/transport/http/handler/visualization_handler.go:139-199`

Key properties of the existing real tree path:

- `VisualizationHandler.Tree` resolves `dashboard` / `dataV` types with `resolveBusiTypes`
- it loads actual `DataVisualizationInfo` rows via `service.List`
- `buildVisualizationTree` converts persisted resources into hierarchical tree nodes
- returned nodes already match most frontend-required fields:
  - real `id`
  - real `pid`
  - real `name`
  - computed `leaf`
  - stable `children`

## Current repository/service gap

The shortest path to parity is blocked by one structural gap:

- `apps/backend-go/internal/repository/visualization_repo.go` has paginated `Query`, but no `ListAll` helper for tree assembly
- `apps/backend-go/internal/service/visualization_service.go` has CRUD/list/detail operations, but no tree-focused method reusable by `interactiveTree`

## Baseline difference between `tree` and `interactiveTree`

### `dataVisualization/tree`

- queries actual visualization resources
- builds real dashboard / dataV hierarchy
- returns a synthetic top-level `root` only as wrapper around real children

### `dataVisualization/interactiveTree`

- queries menu authorization only
- returns synthetic scope nodes instead of real resources
- does not expose actual dashboard or big-screen children
- remains marked `partial` in governed metadata because of this gap

## Follow-up implementation target

For this change to be complete:

1. `interactiveTree` must return real dashboard/dataV resources rather than only synthetic scope roots.
2. Authorization filtering must still apply.
3. The frontend tree node contract must remain stable.
4. Governance metadata can only move from `partial` to `full` after implementation and test evidence exist.
