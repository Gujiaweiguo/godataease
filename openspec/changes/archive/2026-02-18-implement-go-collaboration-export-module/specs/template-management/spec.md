## ADDED Requirements

### Requirement: Template Management Core
The system SHALL provide template management core capability in Go backend.

#### Scenario: Manage template lifecycle
- **WHEN** client creates, updates, or deletes a template
- **THEN** the system persists template changes with permission checks
- **AND** returns Java-compatible response envelope

### Requirement: Template Market Query
The system SHALL provide template market query capability.

#### Scenario: Query template market
- **WHEN** client requests template market list with filters
- **THEN** the system returns matched templates with pagination
- **AND** result includes required metadata for preview and import
