## MODIFIED Requirements

### Requirement: Visualization Core CRUD
The system SHALL provide visualization core CRUD capability in Go backend.

#### Scenario: Create or update visualization
- **WHEN** client submits visualization definition payload
- **THEN** the system persists visualization metadata and content
- **AND** returns success with Java-compatible response envelope

#### Scenario: Query visualization detail
- **WHEN** client requests visualization detail by identifier
- **THEN** the system returns complete visualization definition for rendering

### Requirement: Visualization draft and publish lifecycle
The system SHALL treat `snapshot_*` visualization data as draft/edit state and core/main visualization data as published state.

#### Scenario: Save visualization draft
- **WHEN** client saves a non-folder visualization through the compatibility save flow
- **THEN** the system MUST persist visualization metadata in both metadata tables
- **AND** MUST persist chart-view and dependent editable child data to snapshot-side tables used for draft editing
- **AND** MUST leave core/main child tables as the published representation until a publish action occurs

#### Scenario: Publish visualization draft
- **WHEN** client publishes a visualization
- **THEN** the system MUST update visualization publish status
- **AND** MUST copy the current snapshot-side visualization child data into the core/main table groups in one coordinated operation
- **AND** MUST preserve snapshot-side data as the editable draft baseline after publish

#### Scenario: Recover draft from published visualization
- **WHEN** client requests recovery to the published version
- **THEN** the system MUST rebuild snapshot-side visualization data from the current core/main published data
- **AND** MUST mark the visualization as published after recovery completes

### Requirement: Visualization multi-table consistency
Visualization lifecycle operations SHALL keep metadata, chart views, and view-dependent child tables consistent across snapshot/core table groups.

#### Scenario: Delete visualization with child data
- **WHEN** client deletes a visualization resource
- **THEN** the system MUST clear or logically delete both metadata records and all associated child-table groups for the affected visualization scope
- **AND** MUST NOT leave orphaned snapshot/core rows for chart views, linkage, jump, or outer-parameter data

#### Scenario: Copy visualization with dependent tables
- **WHEN** client copies a visualization resource
- **THEN** the system MUST duplicate both snapshot/core visualization data sets required by the copied resource
- **AND** MUST rewrite copied view-linked identifiers so dependent linkage, jump, and outer-parameter records reference the copied visualization data instead of the source resource

#### Scenario: Publish and recover run atomically per lifecycle action
- **WHEN** the system executes publish or recover for a visualization
- **THEN** metadata status changes and associated child-table synchronization MUST complete as one coordinated lifecycle action
- **AND** partial completion that mixes old and new snapshot/core table groups MUST be treated as failure
