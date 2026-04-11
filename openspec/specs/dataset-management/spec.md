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

### Requirement: Dataset Entry and Initialization Recovery
The system SHALL keep dataset entry paths, initialization flows, and governed preview/browse operations recoverable as a broken-feature surface.

#### Scenario: Dataset page initializes from governed entry paths
- **WHEN** a user enters dataset management through an in-scope menu, route, or governed compatibility path
- **THEN** the page MUST complete its required initialization sequence for browsing workflows
- **AND** initialization failure MUST NOT be disguised as a successful but empty business state

#### Scenario: Dataset recovery preserves deterministic failure semantics
- **WHEN** a dataset browse, field, or preview operation fails during stabilization
- **THEN** the system MUST preserve distinguishable semantics for authorization failure, dependency failure, missing route, and business execution failure
- **AND** targeted regression evidence MUST exist for the recovered path

### Requirement: Historical Datasets Can Be Backfilled Into Governed Resource Identity
The system SHALL support backfilling historical dataset records into the governed resource identity model.

#### Scenario: Backfill a historical dataset with valid governing scope
- **WHEN** a historical dataset can be mapped to a valid organization boundary and governed parent scope
- **THEN** the system MUST register a stable governed resource identity for that dataset
- **AND** repeated backfill execution MUST remain idempotent

#### Scenario: Historical dataset reuses existing governed identity on repeated backfill
- **WHEN** the backfill process runs again for a dataset that has already been governed
- **THEN** the system MUST reuse the existing governed resource identity
- **AND** the process MUST NOT create duplicate governed entries or duplicate inherited grants

#### Scenario: Historical dataset cannot be mapped safely
- **WHEN** a historical dataset depends on missing, orphaned, or invalid upstream governance context
- **THEN** the system MUST skip automatic backfill
- **AND** the skip outcome MUST be traceable for later remediation
- **AND** the dataset MUST NOT be represented as fully governed when required scope information is absent

### Requirement: Backfilled Dataset Authorization Must Remain Consistent Across Permission and Dependency Flows
The system SHALL ensure that historical datasets backfilled into the governed model resolve the same effective authorization state across permission-center views and dataset runtime flows.

#### Scenario: Backfilled dataset preserves permission and dependency consistency
- **WHEN** a historical dataset is governed after backfill
- **THEN** dataset browse, field, preview, and permission-center views MUST remain consistent with the same effective authorization outcome
- **AND** operators MUST be able to distinguish authorization failure from missing dependency or missing resource

#### Scenario: Backfilled dataset inherits governed parent permission
- **WHEN** a historical dataset is backfilled beneath a governed parent group that already has effective grants
- **THEN** the dataset MUST inherit those grants through the governed authorization model
- **AND** permission queries MUST return the inherited result without requiring manual re-authorization

#### Scenario: Backfilled dataset respects datasource dependency boundary
- **WHEN** a user accesses a backfilled dataset whose dependent datasource is not authorized or not governable
- **THEN** the system MUST preserve explicit dependency-denied or authorization-denied semantics
- **AND** the result MUST remain distinguishable from a missing dataset resource

### Requirement: Dataset Export Compatibility Must Be Executable
The system SHALL provide an executable dataset export capability for governed compatibility paths instead of returning a placeholder unsupported response.

#### Scenario: Export dataset through compatibility route
- **WHEN** a client calls `POST /datasetTree/exportDataset` with a valid dataset identifier and supported export parameters
- **THEN** the system MUST create or delegate to a real export workflow compatible with the existing export-center semantics
- **AND** the response MUST use deterministic `code/data/msg` semantics rather than a placeholder unsupported result

#### Scenario: Dataset export failure remains diagnosable
- **WHEN** dataset export cannot be started because of invalid parameters, missing dataset, authorization failure, or export service failure
- **THEN** the system MUST return an explicit non-success outcome that distinguishes business failure from route absence
- **AND** operators MUST be able to trace the failure through logs or export task records

### Requirement: Dataset Field Delete Compatibility Must Use Real Business Semantics
The system SHALL provide executable field deletion behavior for compatibility routes covering both single-field deletion and chart-scoped batch deletion.

#### Scenario: Delete a dataset field by field identifier
- **WHEN** a client calls `POST /datasetField/delete/{id}` for a field that is eligible for deletion
- **THEN** the system MUST execute real delete logic through the dataset service and repository layers
- **AND** the endpoint MUST return deterministic success or failure semantics instead of a placeholder compatibility response

#### Scenario: Delete dataset fields by chart identifier
- **WHEN** a client calls `POST /datasetField/deleteByChartId/{id}` for a chart-scoped field set
- **THEN** the system MUST delete only the fields governed by that chart-scoped association
- **AND** the operation MUST NOT remove unrelated dataset fields outside the requested chart scope

#### Scenario: Field deletion rejects invalid dependency state
- **WHEN** a field deletion request targets a missing field, an unauthorized field, or a field whose dependency constraints forbid removal
- **THEN** the system MUST return explicit failure semantics
- **AND** the failure MUST remain distinguishable from a missing route or generic placeholder success

### Requirement: Dataset Export Compatibility Must Provide A Consumable Task Lifecycle
The system SHALL provide a consumable export task lifecycle for dataset export compatibility flows, not just task creation semantics.

#### Scenario: Dataset export creates a consumable export-center task
- **WHEN** a client calls `POST /datasetTree/exportDataset` with a valid dataset export request under the non-blob compatibility flow
- **THEN** the system MUST create or reuse an export-center task that can later be queried and consumed by the frontend export workflow
- **AND** the response MUST include enough deterministic task identity information for later status or download handling

#### Scenario: Dataset export task can be diagnosed after failure
- **WHEN** the dataset export task fails during generation, download preparation, or later export-center processing
- **THEN** operators and clients MUST be able to distinguish task-creation success from downstream export execution failure
- **AND** failure details MUST remain traceable through task status, logs, or export metadata

#### Scenario: Historical compatibility dataset identity is resolved predictably
- **WHEN** a governed compatibility path references a historical or compatibility dataset identifier during export handling
- **THEN** the system MUST resolve that identifier through a documented and deterministic compatibility strategy
- **AND** the result MUST remain distinguishable from a true missing dataset resource

### Requirement: Dataset Field Deletion Must Enforce Dependency-Aware Blocking
The system SHALL block dataset field deletion when the target field is still referenced by protected downstream consumers.

#### Scenario: Delete field blocked by chart or calculation dependency
- **WHEN** a dataset field targeted by `POST /datasetField/delete/{id}` is still referenced by a chart, calculation field, or equivalent protected runtime dependency
- **THEN** the system MUST reject the delete request
- **AND** the response MUST identify that the failure is dependency-driven rather than route absence or generic execution failure

#### Scenario: Delete field blocked by downstream configuration dependency
- **WHEN** a dataset field targeted for deletion is still referenced by governed downstream configuration that would become invalid after removal
- **THEN** the system MUST reject the delete request with explicit non-success semantics
- **AND** the request MUST NOT silently succeed while leaving inconsistent downstream state

#### Scenario: Delete field succeeds only after dependency-free validation
- **WHEN** a dataset field delete request passes the dependency scan and ownership validation
- **THEN** the system MUST execute the delete operation
- **AND** the success response MUST imply that the protected dependency checks were satisfied
