## ADDED Requirements

### Requirement: Datasource Critical Flow Release Readiness
The system SHALL treat datasource list and validation APIs as release-blocking surfaces for the migrated core BI flow.

#### Scenario: Datasource list remains release-ready
- **WHEN** release verification is executed for datasource management
- **THEN** `POST /api/ds/list` and any governed compatibility alias MUST both be routable
- **AND** both routes MUST return Java-compatible `code/data/msg` envelopes for supported datasource queries

#### Scenario: Datasource validation returns deterministic failure semantics
- **WHEN** datasource validation fails because of timeout, invalid parameters, or backend connectivity problems
- **THEN** `POST /api/ds/validate` MUST return explicit non-success semantics with actionable error detail
- **AND** the endpoint MUST NOT return placeholder success or generic static-route fallback behavior

### Requirement: Datasource Permission-Aware Stability
Datasource APIs used by downstream dataset flows SHALL keep permission semantics distinguishable from resource absence during migration.

#### Scenario: Unauthorized datasource access is not misclassified
- **WHEN** an authenticated user requests a datasource operation without sufficient permission
- **THEN** the system MUST return authorization-denied semantics
- **AND** MUST NOT degrade the result into a misleading `404` or an empty success payload
