## ADDED Requirements

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
