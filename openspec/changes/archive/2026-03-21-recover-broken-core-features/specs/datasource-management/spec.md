## ADDED Requirements

### Requirement: Datasource Entry and Initialization Recovery
The system SHALL keep datasource entry paths and initialization flows recoverable as a governed broken-feature surface.

#### Scenario: Datasource page remains reachable from governed entry paths
- **WHEN** a user enters datasource management through an in-scope menu, route, or governed compatibility path
- **THEN** the datasource page MUST initialize without route-loss or bootstrap-loss behavior
- **AND** any failure MUST be reported as an explicit non-success outcome instead of a silent empty state

#### Scenario: Datasource recovery work distinguishes real gaps from access-path regressions
- **WHEN** datasource functionality appears broken during stabilization
- **THEN** the recovery record MUST classify the symptom as route/access, API contract, page-init, or real implementation gap
- **AND** the classification MUST be testable or otherwise verifiable before the issue is considered closed
