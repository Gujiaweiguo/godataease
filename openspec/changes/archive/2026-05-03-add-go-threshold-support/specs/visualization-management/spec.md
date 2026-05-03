# visualization-management Specification (Delta)

> **Change**: `add-go-threshold-support`
> **Delta type**: additive — adds threshold-governed visualization behavior requirements alongside existing visualization spec.

## Purpose (unchanged)
Define data visualization lifecycle requirements for listing, saving, updating, and loading visualization canvases.

## Requirements

### Requirement: Visualization Core CRUD (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Visualization Listing (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Visualization Resource Tree Compatibility API (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Tree Response Contract Consistency (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Dashboard and Big-Screen Critical Flow Stability (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Resource Operation Precheck Stability (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Interactive Visualization Tree Parity (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Interactive Tree Authorization Filtering (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Visualization Entry-Chain Recovery (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Visualization Detail Hardening After Recovery (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Historical Visualization Resources Can Be Backfilled Into Governed Resource Identity (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

### Requirement: Backfilled Visualization Authorization Must Align With Permission-Center Results (unchanged)
> See main spec: `openspec/specs/visualization-management/spec.md`

---

### Requirement: Chart-Linked Threshold Data Is Governed Visualization Behavior (NEW)

Chart-linked threshold alert data is a governed part of visualization behavior. When a chart is deleted or a visualization resource is removed, associated threshold definitions SHALL be cleaned up via the supported threshold routes.

#### Scenario: Chart deletion triggers threshold cleanup via threshold routes
- **WHEN** a chart that has associated threshold definitions is deleted through visualization workflows
- **THEN** the threshold management routes (`/threshold/deleteWithChart/{chartId}/{resourceTable}`) SHALL be available to remove orphaned threshold records
- **AND** the cleanup SHALL NOT require the visualization module to directly access threshold persistence internals

#### Scenario: Threshold lookup does not break visualization rendering
- **WHEN** a visualization containing a chart with threshold definitions is rendered
- **THEN** the threshold lookup (`/threshold/anyThreshold/{chartId}/{resourceTable}`) SHALL be available for the frontend to display threshold indicators
- **AND** failure of threshold lookup SHALL NOT prevent chart rendering from completing
