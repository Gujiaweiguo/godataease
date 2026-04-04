## ADDED Requirements

### Requirement: System variable management remains separate from permission assignment semantics
The system SHALL keep system variable definition/value management separate from permission-center authorization semantics.

#### Scenario: Supported system variable CRUD remains available
- **WHEN** a client calls supported `/sysVariable/*` definition or value management APIs for create, edit, query, detail, selection, or delete flows
- **THEN** the backend MUST continue to provide deterministic system-variable management behavior compatible with management and editor workflows
- **AND** those APIs MUST NOT by themselves imply that system-variable permission assignment is supported in the permission center

#### Scenario: Unsupported permission-assignment semantics fail explicitly
- **WHEN** a client or UI path attempts to use system-variable semantics as a permission-center authorization target
- **THEN** the system MUST explicitly reject or hide that unsupported behavior
- **AND** the system MUST NOT silently accept, no-op, or present the path as partially implemented permission assignment support
