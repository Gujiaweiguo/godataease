# menu-access-governance Specification

## Purpose
Define governance rules for menu accessibility so that visible menus, routable pages, page initialization APIs, and permission decisions remain consistent. This capability prevents users from encountering broken navigation, ambiguous permission failures, or generic 404 responses when accessing authorized menus.

## Requirements

### Requirement: Visible Menu Must Resolve to Reachable Page
The system SHALL ensure that every visible runtime menu resolves to a registered and reachable frontend page.

#### Scenario: User clicks a visible menu
- **WHEN** a user clicks a menu item that is visible in runtime navigation
- **THEN** the frontend router MUST resolve the target route successfully
- **AND** the application MUST load the expected page component
- **AND** the user MUST NOT see a generic 404 page caused by missing route registration or invalid component mapping

#### Scenario: Menu points to invalid route configuration
- **WHEN** a menu item references a route path or component that is not registered correctly
- **THEN** the system MUST classify the issue as a menu-routing inconsistency
- **AND** the menu item MUST be corrected or removed from authorized runtime output

### Requirement: Authorized Menu Access Must Not Depend on Missing APIs
The system SHALL ensure that pages reachable from authorized menus do not fail due to missing backend APIs required for initial rendering.

#### Scenario: Authorized user opens menu page with initialization requests
- **WHEN** an authorized user opens a page from a visible menu
- **THEN** all required initialization APIs for that page MUST resolve to implemented backend routes or compatible handlers
- **AND** the page MUST NOT fail with `Request failed with status code 404` due to missing backend endpoints

#### Scenario: Frontend references deprecated backend endpoint
- **WHEN** a page initialization flow calls an endpoint that has been removed, renamed, or migrated
- **THEN** the inconsistency MUST be identified as an API mapping issue
- **AND** the frontend or backend contract MUST be aligned before the menu is considered healthy

### Requirement: Unauthorized Access Must Be Distinguished From Missing Resources
The system SHALL distinguish permission denial from resource absence in both routing behavior and API responses.

#### Scenario: User directly accesses unauthorized menu route
- **WHEN** a user without permission directly visits a protected menu route
- **THEN** the system MUST deny access through the authorization path
- **AND** the user MUST receive an insufficient-permission response or redirect behavior
- **AND** the system MUST NOT misclassify the access as a generic 404 caused by missing resource

#### Scenario: Requested resource truly does not exist
- **WHEN** a page or API target is not implemented or no longer exists
- **THEN** the system MAY return 404
- **AND** the result MUST remain distinguishable from authorization failure for troubleshooting and user feedback

### Requirement: Menu Visibility Must Match Effective Authorization
The system SHALL use effective authorization state as the source of truth for runtime menu visibility and direct route access decisions.

#### Scenario: Menu granted to role
- **WHEN** an administrator grants a menu to a role
- **THEN** users with that role MUST receive the menu in authorized navigation results
- **AND** direct access to the corresponding route MUST follow the same effective authorization decision

#### Scenario: Menu revoked from role
- **WHEN** an administrator revokes a menu from a role
- **THEN** users with that role MUST lose menu visibility on the next authorization refresh
- **AND** direct access to the corresponding route MUST be denied consistently

### Requirement: Menu Access Anomalies Must Be Classifiable
The system SHALL classify menu access failures into actionable categories for remediation and regression tracking.

#### Scenario: Menu access failure occurs
- **WHEN** a menu access failure is detected during testing or runtime verification
- **THEN** the issue MUST be classifiable into at least one of the following categories:
  - menu configuration error
  - frontend route mismatch
  - component loading error
  - frontend API path mismatch
  - backend endpoint missing or changed
  - permission mapping inconsistency
  - authorization / 404 semantic confusion

#### Scenario: Team reviews menu remediation status
- **WHEN** maintainers review menu access health
- **THEN** the system or verification process MUST provide a role × menu × result matrix
- **AND** each known anomaly MUST be traceable to a concrete classification and remediation state

### Requirement: Core Menu Paths Must Be Regression-Testable
The system SHALL provide a repeatable verification baseline for core menus across key roles.

#### Scenario: Core menu regression verification
- **WHEN** the team executes menu access regression checks
- **THEN** the verification MUST cover key roles and core menus
- **AND** it MUST confirm menu visibility, route reachability, page initialization success, and correct permission behavior
- **AND** failures MUST be reportable without relying on ad hoc manual memory

### Requirement: Core Feature Recovery Matrix
The system and verification process SHALL maintain a recovery matrix for core RBAC and BI feature domains during regression remediation.

#### Scenario: Classify feature-loss symptom
- **WHEN** a core feature is reported as missing or inaccessible
- **THEN** the recovery process MUST classify the issue as route loss, menu loss, permission mismatch, API mismatch, page initialization failure, or real implementation gap
- **AND** the classification MUST be traceable to concrete file or endpoint evidence

### Requirement: Menu Authorization Enforcement Must Use Persisted Effective Grants
The system MUST enforce governed menu access using persisted effective authorization state rather than placeholder or hardcoded denial behavior.

#### Scenario: Authorized or unauthorized user accesses a governed menu route
- **WHEN** a user loads a route protected by governed menu access
- **THEN** the system MUST evaluate the user's effective role-menu grants
- **AND** the result MUST be classified as allowed, forbidden, or truly missing without relying on a stubbed enforcement path
