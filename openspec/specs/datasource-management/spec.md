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

### Requirement: Datasource List Permission Model Must Be Explicitly Defined
The system SHALL explicitly define the runtime permission semantics for datasource list endpoints and their compatibility aliases.

#### Scenario: Datasource list runtime semantics are chosen deliberately
- **WHEN** datasource list endpoints are hardened after recovery
- **THEN** the system MUST choose and document whether list behavior is filtered, scope-bound with explicit forbidden outcomes, or intentionally auth-only
- **AND** the chosen behavior MUST be justified against existing callers and compatibility paths

#### Scenario: Datasource list permission behavior remains consistent with caller expectations
- **WHEN** a client calls a datasource list route through a governed canonical or compatibility alias
- **THEN** the runtime permission behavior MUST remain consistent across those aliases
- **AND** regression coverage MUST exist for the selected permission model

### Requirement: Historical Datasources Can Be Backfilled Into Governed Resource Identity
The system SHALL support backfilling historical datasource records into the governed resource identity model used by unified permission management.

#### Scenario: Backfill a historical datasource with valid governing scope
- **WHEN** a historical datasource has a valid organization boundary and parent governing scope
- **THEN** the system MUST register a stable governed resource identity for that datasource
- **AND** repeated backfill execution MUST NOT create duplicate governed resource records

#### Scenario: Historical datasource keeps stable identity across repeated backfill runs
- **WHEN** the backfill process is re-executed for a datasource that has already been governed
- **THEN** the system MUST reuse the existing governed resource identity
- **AND** the process MUST NOT create duplicate inheritance state or duplicate authorization mappings

#### Scenario: Historical datasource lacks safe governing scope
- **WHEN** a historical datasource cannot be mapped safely to a governed parent or organization scope
- **THEN** the system MUST skip automatic registration
- **AND** the result MUST be recorded for remediation
- **AND** the datasource MUST NOT be represented as fully governed in unified permission workflows

### Requirement: Backfilled Datasource Authorization Must Match Runtime Semantics
The system SHALL ensure that a backfilled historical datasource uses the same effective authorization semantics as newly governed datasource resources.

#### Scenario: Backfilled datasource appears consistently in permission and runtime flows
- **WHEN** a historical datasource has been backfilled successfully
- **THEN** unified permission queries, user/resource perspectives, and datasource runtime authorization checks MUST resolve against the same governed authorization state

#### Scenario: Backfilled datasource inherits governed parent permission
- **WHEN** a historical datasource is backfilled beneath a parent resource group that already has effective grants
- **THEN** the datasource MUST inherit those grants through the governed authorization model
- **AND** permission queries MUST return the inherited result without requiring manual re-authorization

#### Scenario: Backfilled datasource denial remains distinguishable
- **WHEN** a user accesses a backfilled datasource without sufficient permission
- **THEN** the system MUST return authorization-denied semantics
- **AND** the result MUST remain distinguishable from missing datasource or missing route behavior
