## Proposal: Add Threshold Copy Support

### Summary

Add copy support for visualization threshold tables, When a visualization with thresholds is copied, the threshold configurations are currently lost. This change implements the `CopyCoreThresholdWithMap()` and `CopySnapshotThresholdWithMap()` methods to ensure thresholds are preserved during copy operations.

### Motivation

During the `migrate-visualization-snapshot-core` implementation, Oracle review identified that threshold tables are not copied during visualization copy operations. While the design explicitly allows threshold to remain a no-op adapter in the phase, the users copying dashboards with threshold configurations would lose those settings.

### Goals

1. Implement `CopyCoreThresholdWithMap()` method to copy core threshold tables with view ID rewriting
2. Implement `CopySnapshotThresholdWithMap()` method to copy snapshot threshold tables with view ID rewriting  
3. Integrate threshold copy into `CreateCopyWithChildren()` orchestration
4. Add integration test to verify threshold copy behavior

### Non-Goals

- Implement full threshold business logic (only copy support)
- Modify threshold sync/delete behavior (already implemented)
- Add new threshold-related API endpoints

### Impact

- **Files to modify**:
  - `internal/repository/visualization_threshold_repo.go` - Add copy methods
  - `internal/repository/visualization_repo.go` - Call threshold copy in `CreateCopyWithChildren()`
  - `internal/repository/visualization_repo_integration_test.go` - Add test

### Risks

- Threshold table schema may differ from expected
- ID collision with deterministic ID generation
