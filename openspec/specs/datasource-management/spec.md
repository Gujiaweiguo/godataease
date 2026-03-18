# datasource-management Specification

## Purpose
Define datasource management requirements for connection lifecycle, validation, and sync-related operations.
## Requirements
### Requirement: Datasource List Query
The system SHALL provide datasource list query capability in Go backend.

#### Scenario: Query datasource list
- **WHEN** client calls `POST /api/ds/list` with filter conditions
- **THEN** the system returns datasource records with pagination metadata
- **AND** response format uses `code/data/msg` compatible with Java backend

### Requirement: Datasource Connectivity Validation
The system SHALL validate datasource connection parameters before dataset usage.

#### Scenario: Validate datasource connection
- **WHEN** client calls `POST /api/ds/validate` with connection config
- **THEN** the system tests connectivity with timeout control
- **AND** returns success or failure with clear error message

### Requirement: Datasource Migration Baseline
The system SHALL keep datasource behavior parity baseline between Java and Go for first-wave migration.

#### Scenario: Parity verification for first-wave datasource APIs
- **WHEN** migration verification is executed for first-wave datasource APIs
- **THEN** request/response contracts remain compatible with Java implementation
- **AND** unsupported datasource types are explicitly documented in this change scope

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

### Requirement: Datasource Interactive Aggregate Parity
The system SHALL provide datasource tree data for the frontend interactive aggregate view with the same governed contract stability expected by the other BI discovery domains.

#### Scenario: Interactive aggregate can load datasource tree consistently
- **WHEN** the frontend interactive aggregate requests datasource discovery data
- **THEN** the system MUST return datasource tree nodes using the established `BusiTreeNode` contract
- **AND** the interactive loader MUST NOT require a datasource-specific fallback behavior that breaks aggregate consistency

#### Scenario: Datasource interactive nodes remain structurally valid
- **WHEN** datasource tree data is consumed through the interactive aggregate flow
- **THEN** returned nodes MUST preserve valid identifiers, parent relationships, `leaf` semantics, and required children structure
