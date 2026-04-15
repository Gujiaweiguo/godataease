# datasource-validation-checking-canonical Specification

## Purpose
Define canonical datasource validation and checking route requirements.

## Requirements

### Requirement: Datasource Validate By ID Must Have A Canonical Route
The system SHALL provide datasource validation by ID through a canonical route at `GET /api/ds/validate/:id`.

#### Scenario: Validate datasource by ID through canonical route
- **WHEN** a client calls `GET /api/ds/validate/:id` with a valid datasource ID
- **THEN** the system MUST return validation results using the established compatibility-safe response envelope
- **AND** the canonical route MUST preserve the validation semantics currently expected by datasource flows

#### Scenario: Validate datasource by ID with invalid ID
- **WHEN** a client calls `GET /api/ds/validate/:id` with an invalid or non-existent datasource ID
- **THEN** the system MUST return an explicit non-success response with actionable error detail
- **AND** the response MUST NOT degrade into a generic success or empty payload

### Requirement: Datasource Name/Type Repeat Check Must Have A Canonical Route
The system SHALL provide datasource name/type repeat checking through a canonical route at `POST /api/ds/checkRepeat`.

#### Scenario: Check datasource repeat through canonical route
- **WHEN** a client calls `POST /api/ds/checkRepeat` with a valid datasource write request
- **THEN** the system MUST return repeat check results using the established compatibility-safe response envelope
- **AND** the canonical route MUST remain compatible with existing datasource name/type checking workflows

#### Scenario: Check datasource repeat with invalid request
- **WHEN** a client calls `POST /api/ds/checkRepeat` with an invalid or malformed request body
- **THEN** the system MUST return an explicit non-success response
- **AND** the response MUST NOT silently degrade into an empty success payload

### Requirement: API Datasource Validity Check Must Have A Canonical Route
The system SHALL provide API datasource validity checking through a canonical route at `POST /api/ds/checkApiDatasource`.

#### Scenario: Check API datasource validity through canonical route
- **WHEN** a client calls `POST /api/ds/checkApiDatasource` with a valid API datasource check request
- **THEN** the system MUST return API datasource check results using the established compatibility-safe response envelope
- **AND** the canonical route MUST remain compatible with existing API datasource validation workflows

#### Scenario: Check API datasource validity with invalid request
- **WHEN** a client calls `POST /api/ds/checkApiDatasource` with an invalid request body
- **THEN** the system MUST return an explicit non-success response with actionable error detail
- **AND** the response MUST NOT silently degrade into an empty success payload

### Requirement: Canonical Datasource Validation/Checking Migration Must Preserve Compatibility Safety
The system SHALL allow frontend datasource validation/checking callers to migrate to canonical routes without removing compatibility aliases in the same change.

#### Scenario: Frontend validation/checking callers switch safely to canonical routes
- **WHEN** the frontend datasource API layer migrates validateById, checkRepeat, and checkApiDatasource callers from `/datasource/*` to `/api/ds/*`
- **THEN** the corresponding compatibility routes MUST remain available during the migration window
- **AND** rollback MUST remain possible by reverting frontend route selection without requiring emergency backend route restoration

#### Scenario: Frontend cancel token mechanism remains functional after canonical cutover
- **WHEN** the frontend switches checkApiDatasource to the canonical route
- **THEN** the cancel token map key MUST be updated to match the canonical path
- **AND** in-flight API datasource check requests MUST remain cancellable through the updated cancel mechanism
