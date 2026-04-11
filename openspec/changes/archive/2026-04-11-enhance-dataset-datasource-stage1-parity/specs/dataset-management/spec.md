## ADDED Requirements

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
