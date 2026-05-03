## MODIFIED Requirements

### Requirement: Chart-Linked Threshold Data Is Governed Visualization Behavior

Chart-linked threshold alert data is a governed part of visualization behavior. When a chart is deleted or a visualization resource is removed, associated threshold definitions SHALL be cleaned up automatically through the threshold service dependency.

#### Scenario: Chart deletion triggers threshold cleanup via threshold routes
- **WHEN** a chart that has associated threshold definitions is deleted through visualization workflows
- **THEN** the threshold management routes (`/threshold/deleteWithChart/{chartId}/{resourceTable}`) SHALL be available to remove orphaned threshold records
- **AND** the cleanup SHALL NOT require the visualization module to directly access threshold persistence internals

#### Scenario: Threshold lookup does not break visualization rendering
- **WHEN** a visualization containing a chart with threshold definitions is rendered
- **THEN** the threshold lookup (`/threshold/anyThreshold/{chartId}/{resourceTable}`) SHALL be available for the frontend to display threshold indicators
- **AND** failure of threshold lookup SHALL NOT prevent chart rendering from completing

#### Scenario: Visualization deletion cleans up thresholds for all contained charts
- **WHEN** a visualization is soft-deleted through `VisualizationService.DeleteLogic`
- **THEN** the system parses the visualization's ComponentData to enumerate all chart IDs contained within
- **AND** calls `ThresholdService.DeleteWithChart` for each chart ID to remove associated threshold records
- **AND** the deletion proceeds even if threshold cleanup encounters an error for any individual chart
- **AND** the cleanup uses the "core" resource table value for each chart

#### Scenario: Visualization deletion handles missing threshold service gracefully
- **WHEN** a visualization is deleted and the ThresholdService dependency is not wired (nil)
- **THEN** the visualization deletion completes successfully without threshold cleanup
- **AND** the system logs a warning indicating that threshold cleanup was skipped

#### Scenario: Visualization deletion handles unparsable ComponentData
- **WHEN** a visualization is deleted and its ComponentData cannot be parsed to extract chart IDs
- **THEN** the visualization deletion completes successfully
- **AND** threshold cleanup is skipped for that visualization without failing the deletion
