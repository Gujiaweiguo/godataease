## ADDED Requirements

### Requirement: User Management Must Expose Role-Workflow Entry Consistent With Organization Context
The system SHALL expose role-workflow entry from user management using the same organization context that governs user administration.

#### Scenario: Administrator transitions from user list to role tab
- **WHEN** an administrator opens role workflows from the user-management surface
- **THEN** the system MUST carry forward the active organization context already established for user administration
- **AND** role member discovery MUST remain consistent with that carried context
