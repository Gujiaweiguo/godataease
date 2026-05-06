## MODIFIED Requirements

### Requirement: Row and column permission workflows remain first-class within the unified center
The system SHALL expose row and column permission workflows from the same permission center for supported governed authorization targets only.

#### Scenario: Row/column permission tab manages supported dataset data-access rules
- **WHEN** a user opens the row/column permission tab
- **THEN** the page MUST expose governed dataset-level row-filter and column-control workflows for supported `user` and `role` targets
- **AND** saving those rules MUST keep row-permission and column-permission behavior reachable from the unified permission center

#### Scenario: Deferred system-variable targets do not appear governed
- **WHEN** a permission-center flow encounters deferred system-variable or `sysParams` target semantics
- **THEN** the UI and backend contracts MUST explicitly hide or reject those unsupported targets
- **AND** the system MUST NOT present those targets as completed governed permission-center behavior

#### Scenario: System variable management remains outside row/column permission assignment
- **WHEN** a user needs to manage system variable definitions or selectable values
- **THEN** the system MUST continue to route that work through system variable management capability endpoints and screens
- **AND** the permission center MUST NOT imply that variable management support also provides system-variable permission assignment semantics
