## ADDED Requirements
### Requirement: Dynamic Compatibility Menu Responses
Compatibility endpoints for frontend menu loading SHALL be generated from the same authorization-backed menu service as canonical menu APIs.

#### Scenario: Compatibility menuResource uses dynamic authorization source
- **WHEN** a client calls `GET /api/auth/menuResource`
- **THEN** the response menu tree is generated from persisted menu data filtered by role authorization
- **AND** the endpoint does not return a hardcoded static menu list

#### Scenario: Compatibility roleRouter uses dynamic authorization source
- **WHEN** a client calls `GET /api/roleRouter/query`
- **THEN** router records are generated from authorized persisted menus
- **AND** route metadata remains contract-compatible with existing frontend route parser

### Requirement: Compatibility and Canonical Parity for Menu Authorization
The system SHALL keep menu authorization semantics equivalent between compatibility and canonical endpoints.

#### Scenario: Same user receives consistent menu visibility
- **WHEN** the same authenticated user requests both canonical and compatibility menu endpoints
- **THEN** the visible menu scope is consistent across endpoints
- **AND** differences are limited to contract shape fields required by each endpoint
