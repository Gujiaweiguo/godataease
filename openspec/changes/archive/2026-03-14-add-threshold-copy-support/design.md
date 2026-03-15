## Design

### Context

The visualization copy flow currently handles:
- Views (core_chart_view, snapshot_core_chart_view)
- Linkage (visualization_linkage, visualization_linkage_field + snapshots)
- Jump (visualization_link_jump, visualization_link_jump_info, visualization_link_jump_target_view_info + snapshots)
- Outer params (visualization_outer_param, visualization_outer_params_info, visualization_outer_params_target_view_info + snapshots)
- Threshold (core_threshold, snapshot_core_threshold) - **NOT COPIED**

This leaves dashboards with threshold configurations lost when copying visualizations.

### Approach

1. Add `CopyCoreThresholdWithMap(sourceID, targetID, copyBatchID int64, viewMap map[int64]int64) error`
2. Add `CopySnapshotThresholdWithMap(sourceID, targetID, copyBatchID int64, viewMap map[int64]int64) error`
3. Call these methods in `CreateCopyWithChildren()` after copying other child tables
4. Use the same ID generation pattern: `oldID + 1000000` for deterministic IDs
5. Rewrite view-linked identifiers in threshold data using viewMap

### Threshold Table Schema

```sql
-- Core threshold
CREATE TABLE core_threshold (
    id BIGINT PRIMARY KEY,
    -- threshold-specific fields
);

-- Snapshot threshold  
CREATE TABLE snapshot_core_threshold (
    id BIGINT PRIMARY KEY,
    -- threshold-specific fields  
);
```
