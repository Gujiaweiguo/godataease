# Frontend Parity Assessment

This note records the frontend-side assessment for `complete-interactive-tree-parity`.

## Result

No production frontend code change was required to consume real dashboard/dataV resource nodes from `queryBusiTreeApi`.

## Why no runtime code change was required

- `apps/frontend/src/store/modules/interactive.ts` already normalizes incoming tree nodes generically via `normalizeNodeIds()`.
- `convertInteractive()` computes derived state (`rootManage`, `anyManage`, `leafNodeCount`) from `leaf`, `weight`, and `children`, not from synthetic-root-only semantics.
- Active consumers rely on stable node fields rather than on a hard requirement that the top-level node be a synthetic authorization placeholder.

## Verified consumer expectations

Key frontend consumers continue to work with the existing `BusiTreeNode` contract:

- `apps/frontend/src/store/modules/interactive.ts`
- `apps/frontend/src/views/common/DeResourceTree.vue`
- `apps/frontend/src/views/workbranch/index.vue`
- `apps/frontend/src/views/mobile/directory/index.vue`

These consumers require the following fields to remain stable:

- `id`
- `pid`
- `name`
- `leaf`
- `weight`
- `extraFlag`
- `extraFlag1`
- `children`

## Regression evidence

- `apps/frontend/tests/unit/store/interactive.test.ts`
  - verifies real dashboard/dataV nodes from `queryBusiTreeApi` are preserved
  - verifies derived store state still computes correctly from real resource trees

## Conclusion

Frontend callers do not depend on synthetic-root-only semantics.
The effective parity requirement is therefore: preserve `BusiTreeNode` contract shape while upgrading backend `interactiveTree` behavior to return real visualization resource nodes.
