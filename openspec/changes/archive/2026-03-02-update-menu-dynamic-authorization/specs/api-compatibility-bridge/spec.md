## ADDED Requirements

### Requirement: Dynamic Menu Output for Compatibility Endpoints
The compatibility bridge SHALL return dynamically assembled menu data instead of hardcoded values.

#### Scenario: roleRouter returns dynamic menu-based routes
- **WHEN** client calls GET /api/roleRouter/query
- **THEN** response is generated from database menu records filtered by user role authorization
- **AND** route structure is compatible with frontend route parser

#### Scenario: menuResource returns dynamic menu tree
- **WHEN** client calls GET /api/auth/menuResource
- **THEN** response is generated from database menu records filtered by user role authorization
- **AND** menu tree structure matches frontend expectations

### Requirement: Compatibility and Canonical Parity
Menu output from compatibility endpoints SHALL match canonical menu endpoints for same user.

#### Scenario: Same menu visibility across endpoints
- **WHEN** same authenticated user requests both canonical and compatibility menu endpoints
- **THEN** visible menu set is identical
- **AND** differences are limited to response structure fields

### Requirement: Hardcoded Menu Fallback Toggle
The system SHALL provide configuration toggle to revert to hardcoded menu behavior for emergency.

#### Scenario: Enable fallback mode
- **WHEN** configuration sets `menu.hardcoded_fallback=true`
- **THEN** compatibility endpoints return hardcoded menu data
- **AND** database-driven menu assembly is bypassed

#### Scenario: Disable fallback mode (default)
- **WHEN** configuration sets `menu.hardcoded_fallback=false` or unset
- **THEN** compatibility endpoints return dynamic menu data from database
