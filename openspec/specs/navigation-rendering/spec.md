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

### Requirement: Event-type navigation nodes remain executable in governed menus
The system SHALL execute authorized `menu_type='event'` nodes through the same governed menu tree used for shell navigation instead of treating them as broken routes.

#### Scenario: Toolbox child executes export-center event
- **WHEN** an authorized user expands Toolbox and clicks the governed Data Export Center child menu
- **THEN** the frontend MUST execute the configured export-center action semantics for that node
- **AND** the click MUST NOT be treated as a missing route or invalid router push

#### Scenario: Event node remains authorization-consistent after refresh
- **WHEN** the frontend refreshes authorized menus after login, locale switch, or permission refresh
- **THEN** event-type menu nodes MUST remain in the same authorized position in the menu tree as route-type nodes
- **AND** their visibility MUST continue to be governed by backend menu authorization data

### Requirement: Header entry points remain aligned with governed menu policy
The system SHALL keep shell header actions aligned with the governed navigation policy after menu restructuring.

#### Scenario: Removed More menu does not leave orphaned navigation logic
- **WHEN** the shell renders after the restructure
- **THEN** removed More-menu entry points MUST NOT be rendered from hardcoded header logic
- **AND** governed navigation reachability for remaining capabilities MUST come from the authorized menu tree or explicitly retained account actions only
