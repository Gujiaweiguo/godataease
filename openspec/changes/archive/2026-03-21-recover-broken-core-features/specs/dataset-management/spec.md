## ADDED Requirements

### Requirement: Dataset Entry and Initialization Recovery
The system SHALL keep dataset entry paths, initialization flows, and governed preview/browse operations recoverable as a broken-feature surface.

#### Scenario: Dataset page initializes from governed entry paths
- **WHEN** a user enters dataset management through an in-scope menu, route, or governed compatibility path
- **THEN** the page MUST complete its required initialization sequence for browsing workflows
- **AND** initialization failure MUST NOT be disguised as a successful but empty business state

#### Scenario: Dataset recovery preserves deterministic failure semantics
- **WHEN** a dataset browse, field, or preview operation fails during stabilization
- **THEN** the system MUST preserve distinguishable semantics for authorization failure, dependency failure, missing route, and business execution failure
- **AND** targeted regression evidence MUST exist for the recovered path
