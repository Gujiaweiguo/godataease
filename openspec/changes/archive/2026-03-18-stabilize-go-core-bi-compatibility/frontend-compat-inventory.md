# Frontend Compatibility Inventory

This inventory captures the frontend compatibility surface verified during tasks `6.1` through `6.5`.

## Active BI compatibility callers

### Visualization tree and detail

- `apps/frontend/src/api/visualization/dataVisualization.ts`
  - `queryTreeApi(data)` → `/dataVisualization/tree`
  - `queryBusiTreeApi(data)` → `/dataVisualization/interactiveTree`
  - `findById(dvId, busiFlag, attachInfo)` → `/dataVisualization/findById`
  - `findDvType(dvId)` → `/dataVisualization/findDvType/:id`

### Frontend consumers of compatibility-shaped BI payloads

- `apps/frontend/src/store/modules/interactive.ts`
  - consumes `queryTreeApi`, `queryBusiTreeApi`, `getDatasetTree`, `listDatasources`
  - assumes tree payloads can be normalized into `{ id, pid, leaf, weight, children }`
- `apps/frontend/src/views/common/DeResourceGroupOpt.vue`
  - consumes `queryTreeApi`
- `apps/frontend/src/custom-component/de-screen/SelectScreenDialog.vue`
  - consumes `queryTreeApi`
- `apps/frontend/src/components/visualization/LinkJumpSet.vue`
  - consumes `queryTreeApi`, `findDvType`
- `apps/frontend/src/utils/canvasUtils.ts`
  - consumes `findById`

## Permission / admin compatibility callers

- `apps/frontend/src/api/auth.ts`
  - actively used wrappers:
    - `menuTreeApi()` → `/auth/menuResource`
    - `resourceTreeApi(flag)` → `/auth/busiResource/:flag`
    - `resourcePerSaveApi(data)` → `/system/role/permission/save`
  - compatibility wrappers present but not observed in active call sites during this pass:
    - `/auth/busiPermission`
    - `/auth/menuPermission`
    - `/auth/busiTargetPermission`
    - `/auth/menuTargetPermission`
    - `/auth/saveBusiTargetPer`
    - `/auth/saveMenuTargetPer`

## Compatibility assumptions verified

1. `interactiveStore` previously assumed BI tree API calls would always resolve successfully; this session hardened it to fall back to empty state on reject.
2. `interactiveStore.loadBusiInteractive()` previously only populated returned keys; this session changed it to backfill missing BI families with empty trees.
3. Frontend axios handling already treats non-`000000` responses as failures, so backend `501000` permission-target compatibility responses surface as explicit non-success without extra axios changes.
4. Target permission compatibility endpoints are currently not used by active frontend pages, so the main runtime compatibility surface remains the BI tree/detail APIs and role/menu resource discovery endpoints.

## Test evidence added

- `apps/frontend/tests/unit/store/interactive.test.ts`
  - API reject fallback coverage
  - missing compatibility tree key backfill coverage
- `apps/frontend/e2e/interactive/interactive.spec.ts`
  - tagged unauthenticated BI tree smoke coverage for dashboard / screen / dataset / datasource under `@system-smoke`
