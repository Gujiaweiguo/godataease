## ADDED Requirements

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
