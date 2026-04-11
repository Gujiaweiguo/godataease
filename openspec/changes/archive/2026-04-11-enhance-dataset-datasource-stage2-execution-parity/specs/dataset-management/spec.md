## ADDED Requirements

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
