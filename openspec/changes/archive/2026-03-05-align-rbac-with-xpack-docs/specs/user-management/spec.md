## ADDED Requirements

### Requirement: Excel User Import with Partial Success
The system SHALL provide Excel-based user bulk import with template validation, partial-success processing, and downloadable error reports.

#### Scenario: Import file contains valid and invalid rows
- **WHEN** an administrator uploads a valid template file containing both compliant and non-compliant rows
- **THEN** the system MUST import all compliant rows
- **AND** the system MUST reject non-compliant rows without rolling back compliant rows
- **AND** the system MUST return an error report download reference

#### Scenario: Import file exceeds size limit
- **WHEN** an administrator uploads a file larger than 10 MB
- **THEN** the system MUST reject the upload with a validation error
- **AND** no user record MUST be created

### Requirement: User Password Reset Flow
The system SHALL allow authorized administrators to reset a user's password to the configured initial password policy.

#### Scenario: Reset password for active user
- **WHEN** an administrator triggers password reset on an enabled user
- **THEN** the system MUST update the user's password hash
- **AND** the system MUST return a success response compatible with existing frontend handling

#### Scenario: Unauthorized user attempts reset
- **WHEN** a non-authorized operator calls password reset endpoint
- **THEN** the system MUST deny the request
- **AND** the system MUST not mutate credential data
