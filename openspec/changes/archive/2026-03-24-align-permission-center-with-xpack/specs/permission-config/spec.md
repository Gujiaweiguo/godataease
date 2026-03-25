## ADDED Requirements

### Requirement: Unified Permission Center Consumes a Single Effective Authorization State
The system SHALL present menu, resource, and row/column permission workflows from one governed permission center backed by a single effective authorization state.

#### Scenario: Administrator switches between permission tabs
- **WHEN** an administrator moves between menu permission, resource permission, and row/column permission tabs
- **THEN** the system MUST keep the user inside one governed permission-management entry path
- **AND** each tab MUST reflect authorization data derived from the same effective authorization state

#### Scenario: Administrator compares by-user and by-resource views
- **WHEN** an administrator inspects authorization from the by-user and by-resource perspectives
- **THEN** both perspectives MUST resolve to the same underlying grants for the same scope
- **AND** the system MUST NOT allow the two views to drift semantically

### Requirement: Resource Permission Inheritance Must Remain Governed In the Unified Center
The system SHALL manage resource-group inheritance behavior from the unified permission center so new governed resources receive consistent inherited authorization state.

#### Scenario: New governed resource is created inside an authorized group
- **WHEN** a new datasource, dataset, dashboard, or big-screen resource is created inside a governed resource group
- **THEN** the resource MUST inherit the effective authorization state of that group by default
- **AND** the inherited result MUST be visible from the same unified permission workflow
