## ADDED Requirements

### Requirement: Dynamic Route Recovery Must Preserve Core Menu Reachability
The system SHALL regenerate runtime routes in a way that keeps authorized core menus reachable after login and permission refresh.

#### Scenario: Permission refresh after login
- **WHEN** the frontend refreshes authorized routes after login or focus-based permission refresh
- **THEN** core RBAC and BI menus MUST remain aligned with generated runtime routes
- **AND** authorized pages MUST remain reachable without false `404` classification

#### Scenario: Revoked or invalid path remains distinguishable
- **WHEN** a route is genuinely invalid or no longer authorized
- **THEN** the frontend MUST classify it consistently through the authorization or missing-route path
- **AND** the result MUST remain distinguishable for remediation and debugging
