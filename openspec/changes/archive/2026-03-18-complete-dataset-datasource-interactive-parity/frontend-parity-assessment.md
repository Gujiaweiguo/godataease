# Frontend Parity Assessment

This note records the frontend-side assessment for `complete-dataset-datasource-interactive-parity`.

## Result

The interactive aggregate loader no longer needs to special-case dataset and datasource discovery during bootstrap. `initInteractive()` now relies on the batched interactive loading path so all four BI domains initialize through the same aggregate discovery model.

## What changed

- `apps/frontend/src/store/modules/interactive.ts`
  - `initInteractive()` now delegates bootstrap loading to `loadBusiInteractive()` when any interactive domain is missing
  - dataset and datasource no longer require direct-tree bootstrap calls during aggregate initialization

## What did not change

- Existing direct API wrappers remain available:
  - `getDatasetTree`
  - `listDatasources`
- Production callers that explicitly use dataset/datasource direct APIs are not rewritten by this change.
- The interactive store still consumes the same normalized `BusiTreeNode` contract.

## Regression evidence

- `apps/frontend/tests/unit/store/interactive.test.ts`
  - verifies batched interactive loading preserves dataset and datasource nodes
  - verifies `initInteractive()` no longer boots through direct dataset/datasource tree calls

## Conclusion

Dataset and datasource aggregate discovery is now aligned with dashboard/dataV aggregate discovery for frontend bootstrap behavior, while preserving direct APIs for existing page-level callers.
