## MODIFIED Requirements

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
