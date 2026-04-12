## ADDED Requirements

### Requirement: Dataset Preview Compatibility Must Preserve Explicit Routing Semantics
The system SHALL preserve Java-compatible preview entry points while making datasource-aware routing behavior explicit.

#### Scenario: Preview request with datasource context does not silently fall back
- **WHEN** a client calls `POST /datasetData/previewSql` with datasource context that implies external preview routing behavior
- **THEN** the system MUST either execute the supported datasource-aware preview path or return an explicit non-success result
- **AND** the route MUST NOT silently ignore datasource context while reporting a normal preview success

#### Scenario: Preview request without datasource direct routing remains stable
- **WHEN** a client calls `POST /datasetData/previewSql` under the existing synchronized-local preview model
- **THEN** the system MUST preserve deterministic preview semantics for rows, fields, SQL validation, and timeout handling
- **AND** stage4 routing work MUST NOT regress the established local preview baseline

### Requirement: Permission-Aware Field Enum Compatibility Must Be Executable
The system SHALL provide executable compatibility semantics for permission-aware multi-field value enumeration.

#### Scenario: Query field values under permission constraints
- **WHEN** a client calls `POST /datasetField/multFieldValuesForPermissions` for one or more dataset fields within a permission-aware context
- **THEN** the system MUST return only field values permitted by the governing field and permission filters
- **AND** the response MUST remain compatible with the existing field enumeration contract expected by migrated callers

#### Scenario: Permission-aware field enumeration distinguishes empty result from failure
- **WHEN** no field values are visible after applying permission-aware filtering
- **THEN** the system MUST return an empty successful enumeration result
- **AND** the endpoint MUST NOT misclassify the outcome as a route absence or generic execution failure

### Requirement: Copilot Field Compatibility Must Use Governed Dataset Field Metadata
The system SHALL provide `copilotFields` compatibility behavior using governed dataset field metadata rather than an isolated side-channel field model.

#### Scenario: Copilot field query returns governed dataset fields
- **WHEN** a client calls `POST /datasetField/copilotFields` for a supported dataset context
- **THEN** the system MUST return fields derived from the governed dataset field metadata model
- **AND** the response MUST remain consistent with field visibility and compatibility naming semantics used elsewhere in dataset management

#### Scenario: Copilot field query preserves explicit failure semantics
- **WHEN** a copilot field query targets a missing dataset, unauthorized context, or unsupported request shape
- **THEN** the system MUST return explicit non-success semantics
- **AND** the failure MUST remain distinguishable from a successful empty field set
