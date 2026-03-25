## ADDED Requirements

### Requirement: Permission-Center Authorization Denial Must Remain Distinguishable From Missing Route Or Resource Errors
The system SHALL preserve the semantic distinction between authorization denial and missing route/resource failures after permission workflows are consolidated into one center.

#### Scenario: User opens route denied by unified permission center
- **WHEN** a protected page or API exists but the current user lacks the required permission granted through the unified center
- **THEN** the system MUST produce authorization-denied behavior
- **AND** the result MUST remain distinguishable from an actual missing route or missing backend endpoint

#### Scenario: User opens route missing from implementation
- **WHEN** the requested route or API truly does not exist
- **THEN** the system MAY return 404
- **AND** operators and regression checks MUST still be able to classify the result separately from authorization denial
