# navigation-rendering Specification

## Purpose
Define unified, backend-driven navigation rendering rules so top and side menus stay authorization-consistent and avoid frontend hardcoded visibility logic.
## Requirements
### Requirement: Unified Dynamic Navigation Source
The system SHALL render top navigation and side navigation from the same authorized menu tree response.

#### Scenario: Render top and side menus after login
- **WHEN** a user logs in and authorized menu tree is fetched
- **THEN** top navigation MUST be derived from root-level authorized menus
- **AND** side navigation MUST be derived from currently active top menu subtree

### Requirement: Remove Frontend Hardcoded Filters
The system SHALL NOT rely on frontend hardcoded menu whitelist or static exclusion rules for runtime menu visibility.

#### Scenario: Menu visibility controlled by backend flags
- **WHEN** backend marks a menu node hidden or revokes authorization for the role
- **THEN** the node MUST disappear from both top and side navigation after refresh
- **AND** direct route access MUST remain blocked by authorization checks
