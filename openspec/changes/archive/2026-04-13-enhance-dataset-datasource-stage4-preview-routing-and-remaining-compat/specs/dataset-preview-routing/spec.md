## ADDED Requirements

### Requirement: SQL Preview Routing Must Be Explicit
The system SHALL classify SQL preview requests into an explicit execution path instead of silently ignoring datasource routing hints.

#### Scenario: Preview request uses local sync execution path
- **WHEN** a client calls `POST /datasetData/previewSql` for a preview that is defined to run against synchronized local dataset storage
- **THEN** the system MUST execute the preview against the local synchronized storage path
- **AND** the request MUST NOT be reinterpreted as an external datasource direct query

#### Scenario: Preview request uses external direct execution path
- **WHEN** a client calls `POST /datasetData/previewSql` with a datasource context that is explicitly supported for direct preview execution
- **THEN** the system MUST route the request through the external preview execution path for that datasource
- **AND** the result MUST remain distinguishable from local synchronized preview behavior

#### Scenario: Unsupported routing hint is rejected explicitly
- **WHEN** a client requests SQL preview with a datasource routing hint that is unsupported for direct execution
- **THEN** the system MUST return an explicit non-success result
- **AND** the system MUST NOT silently ignore the routing hint and continue with local preview execution

### Requirement: External Direct Preview Must Be Constrained
The system SHALL treat external direct SQL preview as a constrained capability governed by datasource type, authorization, and execution limits.

#### Scenario: External direct preview requires supported datasource type
- **WHEN** a client requests external direct SQL preview for a datasource type outside the supported preview matrix
- **THEN** the system MUST reject the request with explicit unsupported semantics
- **AND** the rejection MUST remain distinguishable from SQL validation failure and datasource-not-found failure

#### Scenario: External direct preview requires datasource access permission
- **WHEN** a client requests external direct SQL preview without sufficient permission to use the referenced datasource
- **THEN** the system MUST reject the request with explicit authorization-denied semantics
- **AND** the result MUST NOT degrade into a misleading missing-resource or empty-success response

#### Scenario: External direct preview enforces preview execution limits
- **WHEN** a client requests external direct SQL preview for a supported datasource
- **THEN** the system MUST apply bounded preview limits including timeout and result-size constraints
- **AND** the preview path MUST NOT behave as an unrestricted cross-datasource query interface

### Requirement: Preview Result Source And Failure Semantics Must Be Diagnosable
The system SHALL keep preview result origin and failure categories diagnosable for operators and callers.

#### Scenario: Caller can distinguish local and direct preview outcomes
- **WHEN** a preview request succeeds through either local synchronized execution or external direct execution
- **THEN** the system MUST preserve enough response or observable runtime semantics to distinguish which path was used
- **AND** troubleshooting MUST NOT rely on undocumented implementation knowledge

#### Scenario: Caller can distinguish unsupported direct preview from execution failure
- **WHEN** an external direct preview request fails because the path is unsupported, forbidden, or cannot establish a connection
- **THEN** the system MUST return a failure category that distinguishes routing/eligibility failure from SQL execution failure
- **AND** the failure MUST remain testable through deterministic compatibility semantics
