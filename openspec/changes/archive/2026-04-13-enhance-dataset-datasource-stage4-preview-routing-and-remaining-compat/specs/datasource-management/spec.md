## ADDED Requirements

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
