## ADDED Requirements

### Requirement: Share Editable Link Controls
The system SHALL allow authorized users to edit the externally visible share link state for an existing shared resource.

#### Scenario: Update share UUID with deterministic validation
- **WHEN** a client calls `/share/editUuid` with a resource identifier and a candidate UUID
- **THEN** the backend validates the UUID format and uniqueness against existing share links
- **AND** applies the UUID update only when validation succeeds
- **AND** returns a deterministic validation-compatible response that the current frontend can interpret as success or error

#### Scenario: Reject invalid share UUID changes
- **WHEN** a client calls `/share/editUuid` with an invalid or conflicting UUID value
- **THEN** the backend leaves the existing share UUID unchanged
- **AND** returns a deterministic validation error message

### Requirement: Share Expiration Editing
The system SHALL allow authorized users to update expiration state for an existing shared resource.

#### Scenario: Enable or update share expiration
- **WHEN** a client calls `/share/editExp` with a valid expiration timestamp for a shared resource
- **THEN** the backend persists the new expiration state for that share
- **AND** subsequent `/share/detail/:resourceId` responses reflect the updated expiration value

#### Scenario: Clear or reject invalid expiration updates
- **WHEN** a client calls `/share/editExp` with a clearable value or an invalid timestamp
- **THEN** the backend either clears expiration or rejects the update deterministically according to validation rules
- **AND** the persisted share state remains internally consistent

### Requirement: Share Password Editing
The system SHALL allow authorized users to update password protection state for an existing shared resource.

#### Scenario: Update share password with generated or custom value
- **WHEN** a client calls `/share/editPwd` with a resource identifier and either an auto-generated or custom password value
- **THEN** the backend persists the updated password-protection state for that share
- **AND** subsequent `/share/detail/:resourceId` responses reflect the updated password state needed by the frontend share dialog

#### Scenario: Disable or replace invalid password state deterministically
- **WHEN** a client calls `/share/editPwd` with a disabled password state or an invalid password payload
- **THEN** the backend updates or rejects password protection deterministically according to validation rules
- **AND** the resulting share state remains compatible with existing share validation behavior
