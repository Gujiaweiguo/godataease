# navigation-rendering Specification

## Purpose
Define unified, backend-driven navigation rendering rules so top and side menus stay authorization-consistent and avoid frontend hardcoded visibility logic.
## Requirements
### Requirement: Unified Dynamic Navigation Source
The system SHALL render top navigation and side navigation from the same authorized menu tree response, and that response SHALL provide locale-consistent menu metadata for the effective locale.

#### Scenario: Render top and side menus after login
- **WHEN** a user logs in and authorized menu tree is fetched
- **THEN** top navigation MUST be derived from root-level authorized menus
- **AND** side navigation MUST be derived from currently active top menu subtree
- **AND** menu titles consumed by both navigations MUST be localized using the same effective locale

#### Scenario: Refresh navigation after locale switch
- **WHEN** the frontend switches locale and refetches the authorized menu tree
- **THEN** top and side navigation MUST continue to use the same authorized menu nodes as before the switch
- **AND** user-visible menu titles MUST update to the newly effective locale without requiring frontend hardcoded title overrides

### Requirement: Remove Frontend Hardcoded Filters
The system SHALL NOT rely on frontend hardcoded menu whitelist or static exclusion rules for runtime menu visibility.

#### Scenario: Menu visibility controlled by backend flags
- **WHEN** backend marks a menu node hidden or revokes authorization for the role
- **THEN** the node MUST disappear from both top and side navigation after refresh
- **AND** direct route access MUST remain blocked by authorization checks

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
