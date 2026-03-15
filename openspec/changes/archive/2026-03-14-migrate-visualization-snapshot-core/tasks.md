## 1. Spec and schema alignment

- [x] 1.1 Confirm whether `snapshot_core_chart_view` already exists in the Go runtime database schema and add or generate the corresponding Go model.
- [x] 1.2 Add repository coverage for visualization metadata and chart-view snapshot/core table groups needed by save, publish, recover, delete, and copy flows.

## 2. Phase 1 lifecycle correction

- [x] 2.1 Refactor visualization save flow so draft editing writes metadata to both metadata tables and writes editable chart-view data to snapshot-side tables.
- [x] 2.2 Refactor visualization publish flow so it copies snapshot draft data into core/main tables and preserves snapshot as the post-publish draft baseline.
- [x] 2.3 Refactor visualization recover flow so it rebuilds snapshot draft data from core/main published data.
- [x] 2.4 Extend delete handling so metadata and chart-view snapshot/core tables are cleaned together.
- [x] 2.5 Add integration tests that prove snapshot=draft and core=published for save, publish, recover, and delete workflows.

## 3. Phase 2 child-table orchestration

- [x] 3.1 Add repository methods for linkage and linkage-field snapshot/core synchronization.
- [x] 3.2 Add repository methods for jump, jump-info, and jump-target snapshot/core synchronization.
- [x] 3.3 Add repository methods for outer-params, outer-params-info, and outer-params-target snapshot/core synchronization.
- [x] 3.4 Attach threshold publish/recover/delete hook points to the visualization lifecycle coordinator.
- [x] 3.5 Add copy-flow support that duplicates child-table groups and rewrites copied view-linked identifiers.

## 4. Verification and rollout

- [x] 4.1 Add integration coverage for copy and delete cleanup across metadata, chart views, linkage, jump, and outer-params groups.
- [x] 4.2 Verify compatibility handlers continue to expose stable API paths while using the new lifecycle coordinator.
- [x] 4.3 Document any deferred threshold behavior or schema prerequisites in the change notes before implementation begins.
