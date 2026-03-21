## ADDED Requirements

### Requirement: Login Bootstrap Must Restore Authorized Runtime State
The system SHALL restore current-user, authorized menu, and dynamic route state after successful login so core feature domains remain reachable.

#### Scenario: Successful login restores authorized routes
- **WHEN** a user logs in successfully
- **THEN** the frontend MUST initialize current-user state, authorized menu state, and dynamic routes before protected feature navigation is considered healthy
- **AND** the user MUST NOT land in a false missing-feature state caused by incomplete bootstrap

#### Scenario: Login redirect targets protected page
- **WHEN** login completes with a protected redirect target
- **THEN** the system MUST resolve the redirect only after authorization-driven route setup is ready
- **AND** the user MUST NOT be sent to a misleading `401` or `404` caused by bootstrap race or missing dynamic route registration
