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

### Requirement: Datasource Table Status Must Reflect Real Synchronization State
The system SHALL return datasource table status using real synchronization evidence or explicit unknown-state semantics rather than a fixed success placeholder.

#### Scenario: Query table status with synchronization history
- **WHEN** a client calls `POST /datasource/getTableStatus` for a datasource table that has synchronization history or task records
- **THEN** the system MUST derive the returned status from the latest relevant synchronization evidence
- **AND** the response MUST include a stable status value and the latest available update time for frontend display

#### Scenario: Query table status without synchronization history
- **WHEN** a client calls `POST /datasource/getTableStatus` for a datasource table that has no usable synchronization evidence
- **THEN** the system MUST return an explicit unknown, uninitialized, or equivalent non-success-assuming state
- **AND** the endpoint MUST NOT report a synthetic success status solely because the route is reachable

### Requirement: Datasource Delete Compatibility And Canonical Write Paths Must Remain Consistent
The system SHALL provide a normative datasource delete write path while keeping historical compatibility delete routes behaviorally consistent during migration.

#### Scenario: Delete datasource through normative write route
- **WHEN** a client calls the normative datasource delete write route for a deletable datasource or datasource folder
- **THEN** the system MUST execute the same recursive delete business logic used by compatibility aliases
- **AND** the route MUST return deterministic success or failure semantics compatible with existing callers

#### Scenario: Compatibility delete route matches canonical delete semantics
- **WHEN** a client calls a historical datasource delete compatibility alias for the same target resource
- **THEN** the compatibility route MUST resolve through the same underlying delete logic as the normative write route
- **AND** both paths MUST preserve the same pre-delete validation, permission checks, and business failure semantics

#### Scenario: Datasource delete failure remains explicit
- **WHEN** datasource deletion is blocked by missing resource, authorization denial, dependent relations, or recursive delete failure
- **THEN** the system MUST return explicit non-success semantics
- **AND** the result MUST remain distinguishable from route absence and from a successful no-op response

### Requirement: Datasource Table Status Must Expose Stable Execution-State Semantics
The system SHALL expose datasource table status using a stable business-facing execution-state model rather than a minimally aggregated transport value.

#### Scenario: Datasource table status maps runtime evidence into a stable state set
- **WHEN** the system returns table status for a datasource table with available sync or execution evidence
- **THEN** it MUST map that evidence into a stable documented state set suitable for frontend rendering
- **AND** the frontend contract MUST NOT depend on raw task-system enum values

#### Scenario: Datasource table status includes authoritative update time
- **WHEN** table status is derived from multiple possible runtime timestamps such as create, start, or end time
- **THEN** the system MUST choose a deterministic authoritative update time according to documented precedence
- **AND** the chosen time MUST be suitable for stable frontend display and regression testing

#### Scenario: Datasource table status preserves unknown-state semantics
- **WHEN** runtime evidence is incomplete, missing, or insufficient to derive a meaningful execution state
- **THEN** the system MUST return an explicit unknown, warning, or equivalent non-success-assuming state
- **AND** it MUST NOT infer a completed state solely from route reachability or table existence

### Requirement: Datasource Detail Must Return Resolved Creator Name
The system SHALL return a resolved creator user name (not just the raw user ID) in datasource detail responses.

#### Scenario: Datasource detail includes creator display name
- **WHEN** a client calls `GET /datasource/get/{id}` or `GET /datasource/hidePw/{id}` for a valid datasource
- **THEN** the response MUST include a `creator` field containing the resolved user name corresponding to `create_by`
- **AND** if the user cannot be found, the system MUST fall back to the raw `create_by` value

#### Scenario: Datasource detail includes updater display name
- **WHEN** a datasource has been updated (non-empty `update_by` field)
- **THEN** the response MUST include an `updater` field containing the resolved user name corresponding to `update_by`
- **AND** if `update_by` is empty or the user cannot be found, the system MUST return an empty string

### Requirement: Datasource Preview Eligibility Must Be Explicit
The system SHALL explicitly define whether a datasource can participate in direct SQL preview execution.

#### Scenario: Supported datasource is eligible for direct preview
- **WHEN** a datasource type and connection shape are included in the direct preview support matrix
- **THEN** the system MUST treat that datasource as eligible for direct SQL preview routing
- **AND** the eligibility decision MUST be stable enough for regression verification

#### Scenario: Unsupported datasource is rejected explicitly for direct preview
- **WHEN** a client requests direct SQL preview for a datasource outside the supported preview matrix
- **THEN** the system MUST return explicit unsupported semantics
- **AND** the rejection MUST remain distinguishable from datasource-not-found and permission-denied outcomes

### Requirement: Datasource Direct Preview Must Respect Runtime Authorization Boundaries
The system SHALL apply datasource authorization semantics consistently to direct SQL preview requests.

#### Scenario: Unauthorized datasource cannot be used for direct preview
- **WHEN** a client requests direct SQL preview for a datasource without sufficient permission
- **THEN** the system MUST reject the request with authorization-denied semantics
- **AND** the outcome MUST remain consistent with datasource runtime permission behavior in other governed flows

#### Scenario: Authorized datasource preview failure remains diagnosable
- **WHEN** a client requests direct SQL preview for an authorized datasource but connection establishment or execution fails
- **THEN** the system MUST return explicit non-success semantics for connection or execution failure
- **AND** operators MUST be able to distinguish that failure from unsupported-datasource and unauthorized-datasource cases

### Requirement: Datasource Table Exploration Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource table exploration capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource table exploration uses canonical routes consistently
- **WHEN** a client performs datasource table exploration for table listing, schema lookup, status lookup, or field inspection
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with the corresponding compatibility datasource routes

### Requirement: Datasource Exploration Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource exploration response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource table exploration remains contract-safe
- **WHEN** the frontend datasource API layer switches table exploration callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource exploration pages MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** invalid datasource, invalid table, or unavailable backend outcomes MUST remain explicit and testable rather than silently degrading into empty success responses

### Requirement: Datasource Preview And Sync Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource preview and synchronization capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource preview and sync use canonical routes consistently
- **WHEN** a client performs datasource preview data retrieval or synchronization actions
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with corresponding compatibility datasource routes

### Requirement: Datasource Preview And Sync Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource preview and sync response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource preview and sync remain contract-safe
- **WHEN** the frontend datasource API layer switches preview and sync callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource preview/sync flows MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** invalid datasource, invalid sync target, or unavailable backend outcomes MUST remain explicit and testable rather than silently degrading into empty success responses

### Requirement: Datasource File Ingest Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource file upload and remote file loading capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource file ingest uses canonical routes consistently
- **WHEN** a client performs datasource file upload or remote file loading
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with corresponding compatibility datasource routes

### Requirement: Datasource File Ingest Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource file ingest response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource file ingest remains contract-safe
- **WHEN** the frontend datasource API layer switches file ingest callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource ingest flows MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** invalid upload input, invalid remote source, or unavailable backend outcomes MUST remain explicit and testable rather than silently degrading into empty success responses

### Requirement: Datasource Validation And Checking Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource validate-by-ID, name/type repeat checking, and API datasource validity checking capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource validation and checking uses canonical routes consistently
- **WHEN** a client performs datasource validation by ID, name/type repeat checking, or API datasource validity checking
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with corresponding compatibility datasource routes

### Requirement: Datasource Validation And Checking Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource validation and checking response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource validation and checking remains contract-safe
- **WHEN** the frontend datasource API layer switches validation/checking callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource validation/checking flows MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** invalid datasource ID, duplicate name/type, invalid API datasource, or unavailable backend outcomes MUST remain explicit and testable rather than silently degrading into empty success responses
