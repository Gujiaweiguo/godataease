# dataset-management Specification

## Purpose
TBD - created by archiving change implement-go-dataset-datasource-module. Update Purpose after archive.
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

