# dataset-management Specification

## Purpose
Define dataset tree, field, and preview operation requirements including permission-aware access paths.
## Requirements
### Requirement: Dataset Tree Query
The system SHALL provide dataset tree query capability in Go backend.

#### Scenario: Query dataset tree
- **WHEN** client calls `POST /api/dataset/tree`
- **THEN** the system returns hierarchical dataset nodes
- **AND** response format uses `code/data/msg` compatible with Java backend

### Requirement: Dataset Field Metadata Query
The system SHALL provide dataset field metadata query capability.

#### Scenario: Query dataset fields
- **WHEN** client calls `POST /api/dataset/fields` with dataset identifier
- **THEN** the system returns field list including name, type, and aggregation metadata
- **AND** field type mapping follows defined Java-Go compatibility mapping

### Requirement: Dataset Preview Query
The system SHALL provide dataset preview query capability for development and verification.

#### Scenario: Preview dataset data
- **WHEN** client calls `POST /api/dataset/preview` with preview parameters
- **THEN** the system returns sampled rows under configurable row limit
- **AND** query timeout and error handling are consistent with migration baseline

### Requirement: Dataset Critical Flow Release Readiness
The system SHALL treat dataset tree, field metadata, and preview APIs as release-blocking surfaces for the migrated core BI flow.

#### Scenario: Dataset tree contract remains stable
- **WHEN** release verification is executed for dataset browsing workflows
- **THEN** `POST /api/dataset/tree` and any governed compatibility alias MUST return structurally valid hierarchical nodes
- **AND** the response MUST preserve Java-compatible `code/data/msg` semantics for supported requests

#### Scenario: Dataset field metadata remains parity-safe
- **WHEN** a client requests `POST /api/dataset/fields`
- **THEN** the system MUST return field metadata with stable type and aggregation semantics expected by migrated consumers
- **AND** the endpoint MUST NOT silently omit required field metadata while still reporting success

#### Scenario: Dataset preview is deterministic under failure
- **WHEN** a client requests `POST /api/dataset/preview` and the preview cannot be completed
- **THEN** the system MUST return deterministic timeout, validation, or execution failure semantics
- **AND** MUST NOT return placeholder success with empty rows for a failed preview execution

### Requirement: Dataset Permission and Dependency Stability
Dataset flows SHALL keep datasource dependency checks and permission failures explicit during migration.

#### Scenario: Dataset operation blocked by unauthorized datasource dependency
- **WHEN** a user can access a dataset shell but lacks required datasource visibility for a dependent operation
- **THEN** the system MUST return explicit authorization-denied or dependency-denied semantics
- **AND** operators MUST be able to distinguish this result from missing dataset resources during troubleshooting

### Requirement: Dataset Interactive Aggregate Parity
The system SHALL provide dataset tree data for the frontend interactive aggregate view with the same governed contract stability expected by the other BI discovery domains.

#### Scenario: Interactive aggregate can load dataset tree consistently
- **WHEN** the frontend interactive aggregate requests dataset discovery data
- **THEN** the system MUST return dataset tree nodes using the established `BusiTreeNode` contract
- **AND** the interactive loader MUST NOT require a dataset-specific fallback behavior that breaks aggregate consistency

#### Scenario: Dataset interactive nodes remain structurally valid
- **WHEN** dataset tree data is consumed through the interactive aggregate flow
- **THEN** returned nodes MUST preserve valid identifiers, parent relationships, `leaf` semantics, and required children structure
