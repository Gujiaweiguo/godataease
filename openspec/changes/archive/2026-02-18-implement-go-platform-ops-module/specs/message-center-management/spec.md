## ADDED Requirements

### Requirement: Message Center Query
The system SHALL provide message center query capability in Go backend.

#### Scenario: Query message list
- **WHEN** user requests message list with pagination and filters
- **THEN** the system returns matched messages ordered by create time
- **AND** includes read/unread status for each message

### Requirement: Message Status Update
The system SHALL provide message status update capability.

#### Scenario: Mark message as read
- **WHEN** user marks a message as read
- **THEN** the system updates message status idempotently
- **AND** subsequent reads reflect updated status consistently
