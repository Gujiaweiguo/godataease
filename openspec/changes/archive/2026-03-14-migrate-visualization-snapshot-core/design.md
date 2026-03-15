## Context

The current Go visualization service stores metadata in `data_visualization_info` and mirrors it into `snapshot_data_visualization_info`. This matches only a small subset of legacy behavior. Repository evidence shows Go currently exposes `SaveSnapshot`, `SyncSnapshotFromMain`, `RestoreMainFromSnapshot`, and metadata-only snapshot updates in `apps/backend-go/internal/repository/visualization_repo.go`, while service methods such as `UpdatePublishStatus` and `RecoverToPublished` in `apps/backend-go/internal/service/visualization_service.go` only orchestrate metadata transitions.

Legacy Java uses a broader model. `DataVisualizationServer.saveCanvas` documents that all editing changes flow through snapshot tables, while publish copies snapshot data into core/main and recover rehydrates snapshot from published core. `CoreVisualizationManage.dvSnapshotRecover` and `dvRestore` copy metadata, views, jump data, linkage data, and outer-parameter data in a single transactional flow. Related services such as `ChartViewManege`, `VisualizationLinkJumpService`, and `VisualizationOuterParamsService` operate directly on snapshot-side child tables during editing.

Go already contains generated auto models for most visualization child-table pairs in `apps/backend-go/internal/domain/auto`, including linkage, jump, and outer-params tables and their snapshot variants. However, it does not currently contain a generated `snapshot_core_chart_view` model or repository/service orchestration for those child groups. This means Go has partial schema awareness but lacks the lifecycle coordinator that gives the legacy model its behavior.

## Goals / Non-Goals

**Goals:**
- Restore legacy-compatible lifecycle semantics where snapshot tables represent draft state and core/main tables represent published state.
- Introduce explicit multi-table orchestration for visualization metadata, chart views, linkage, jump, and outer-parameter table groups.
- Define stable transaction boundaries for save, publish, recover, delete, and copy operations.
- Stage the migration so the first phase corrects the most critical semantic gap before broader child-table coverage is added.

**Non-Goals:**
- Rebuild the entire visualization module around a new architectural abstraction unrelated to current Go layering.
- Migrate unrelated template, share, or frontend rendering logic as part of this change.
- Guarantee xpack threshold feature completeness in the first phase if the Go runtime integration remains incomplete; this change only reserves its orchestration boundary.

## Decisions

### 1. Keep the existing Go layering and add lifecycle coordinators inside the current repository/service structure

The repository already follows the project convention of `domain -> service -> repository -> transport`. This design will keep `VisualizationService` as the orchestration entry point and add table-group-specific repository methods rather than introducing a parallel subsystem. This minimizes migration risk and keeps the implementation aligned with current Go structure.

**Alternatives considered:**
- Introduce a brand new domain service tree for visualization aggregates. Rejected because it would mix architectural refactor with behavior migration.
- Keep child-table logic embedded in one repository file. Rejected because multi-table orchestration will become hard to verify and evolve.

### 2. Treat snapshot as draft state and core/main as published state across all visualization workflows

This follows legacy evidence directly: `saveCanvas` edits snapshot-side data, `updatePublishStatus` publishes by clearing core and restoring from snapshot, and `recoverToPublished` rebuilds snapshot from published core. Go must adopt the same direction before adding more tables; otherwise every later child-table migration would encode the wrong semantics.

**Alternatives considered:**
- Preserve the current Go interpretation and layer more tables onto it. Rejected because it diverges from legacy behavior and would make publish/recover semantics inconsistent across table groups.

### 3. Migrate by table groups, not by individual endpoints

The lifecycle behavior is organized around data groups, not isolated APIs. The implementation should therefore introduce repository methods for these groups:
- visualization metadata
- chart views (`core_chart_view` and `snapshot_core_chart_view`)
- linkage (`visualization_linkage`, `visualization_linkage_field`, and snapshot variants)
- jump (`visualization_link_jump`, `visualization_link_jump_info`, `visualization_link_jump_target_view_info`, and snapshot variants)
- outer parameters (`visualization_outer_params`, `visualization_outer_params_info`, `visualization_outer_params_target_view_info`, and snapshot variants)
- threshold integration hooks

**Alternatives considered:**
- Add one API at a time and patch the needed tables ad hoc. Rejected because publish/recover/delete/copy all span the same table groups and must stay transactionally consistent.

### 4. Split delivery into two implementation phases

Phase 1 will establish the minimum viable lifecycle correction: metadata plus chart-view snapshot/core orchestration. This phase requires adding the missing Go model/repository coverage for `snapshot_core_chart_view` and correcting save/publish/recover/delete semantics.

Phase 2 will add the child-table groups that depend on stable view identifiers: linkage, jump, and outer parameters. Threshold orchestration hooks are attached in this phase as well, even if the underlying behavior initially lands behind a minimal adapter.

**Alternatives considered:**
- Ship all table groups in one phase. Rejected because the missing snapshot view layer is the primary blocker and should be corrected first.

### 5. Define transaction boundaries around user-visible lifecycle actions

Each of the following operations will execute as a single coordinated transaction in service orchestration:
- **Save / saveCanvas**: persist metadata to both meta tables and write draft child data to snapshot-side tables.
- **Publish**: update status/meta, prune invalid snapshot views if needed, clear core child tables, restore snapshot into core, then run threshold publish hooks.
- **Recover to published**: clear snapshot child tables, copy core/published data back into snapshot, update status to published.
- **Delete**: delete or logically delete both metadata tables and clear child-table groups for all affected resource IDs.
- **Copy**: create a new visualization ID, duplicate core and snapshot child tables, rewrite view-linked identifiers, then persist new metadata.

This matches the legacy behavior more closely than isolated repository updates.

## Risks / Trade-offs

- **[Risk] Missing `snapshot_core_chart_view` model in Go blocks faithful lifecycle migration** → Mitigation: make this the first deliverable in Phase 1 and do not extend publish/recover further before it exists.
- **[Risk] Existing Go publish/recover semantics are reversed relative to legacy** → Mitigation: correct semantics before wiring additional child tables, and add integration tests that assert snapshot/core direction explicitly.
- **[Risk] Child tables reference view IDs, so partial migration can create dangling references** → Mitigation: complete metadata + view orchestration before migrating linkage/jump/outer-params groups.
- **[Risk] Threshold behavior may be partially unavailable in Go runtime** → Mitigation: define threshold hook interfaces and invocation points now, but allow the first implementation to be a no-op adapter until full threshold support is ready.
- **[Risk] Delete and copy paths can leave orphaned child rows if only meta tables are updated** → Mitigation: require grouped repository methods for cleanup and duplication, and verify them with integration tests.

## Migration Plan

1. Add proposal-backed delta specs defining lifecycle semantics for visualization draft/publish/recover/delete/copy behavior.
2. Implement Phase 1: add `snapshot_core_chart_view` support, repository methods for metadata + views, and correct publish/recover/save/delete transaction flow.
3. Add integration tests that prove snapshot is draft and core is published for save/publish/recover transitions.
4. Implement Phase 2: migrate linkage, jump, and outer-parameter table groups into the same lifecycle coordinator; add threshold hooks.
5. Extend copy/delete tests to assert child-table duplication and cleanup.
6. Roll out behind existing compatibility endpoints so external API paths remain stable.

Rollback strategy: keep endpoint paths unchanged and isolate behavior changes to service/repository orchestration. If a phase introduces regressions, revert the new table-group coordinator while keeping already-stable metadata compatibility fixes intact.

## Open Questions

- Does the Go database schema already contain `snapshot_core_chart_view`, requiring only a generated model, or does it also need migration work outside this change?
- Which threshold behaviors are required immediately in Go versus deferred behind an adapter?
- Should copy support for visualization child tables land in the same implementation phase as delete cleanup, or can copy safely follow once publish/recover semantics are stable?

## Change Notes

### Threshold Behavior (Deferred)

Per design decision #5 (Risk mitigation), threshold publish/recover/delete hooks are attached to the visualization lifecycle coordinator but implemented as **no-op adapters** in this phase. The hook invocation points exist in:

- `visualization_threshold_repo.go`: `SyncCoreThresholdFromSnapshot()`, `SyncSnapshotThresholdFromCore()`
- These methods are called during publish/recover flows but perform no database operations

**Rationale**: Threshold behavior requires additional schema validation and business logic that is deferred until full threshold support is ready. The no-op adapters ensure:
1. Lifecycle orchestration boundaries are correctly positioned
2. Future threshold implementation can be dropped in without refactoring coordinator logic
3. Current visualization workflows remain stable

**Schema Prerequisites for Full Threshold Support**:
- `core_threshold` and `snapshot_core_threshold` tables must exist
- Threshold entity models must be generated in `internal/domain/auto/`
- Threshold copy methods (`CopyCoreThresholdWithMap`, `CopySnapshotThresholdWithMap`) are not yet implemented

### Open Questions Resolution

1. **snapshot_core_chart_view**: Confirmed - model exists in Go runtime, no external migration needed.
2. **Threshold behaviors**: Sync hooks implemented as no-op; copy not yet implemented.
3. **Copy support timing**: Implemented in Phase 2 alongside delete cleanup, as both require consistent child-table orchestration.
