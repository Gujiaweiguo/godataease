## ADDED Requirements

### Requirement: Menu Authorization Enforcement Must Use Persisted Effective Grants
The system MUST enforce governed menu access using persisted effective authorization state rather than placeholder or hardcoded denial behavior.

#### Scenario: Authorized or unauthorized user accesses a governed menu route
- **WHEN** a user loads a route protected by governed menu access
- **THEN** the system MUST evaluate the user's effective role-menu grants
- **AND** the result MUST be classified as allowed, forbidden, or truly missing without relying on a stubbed enforcement path
