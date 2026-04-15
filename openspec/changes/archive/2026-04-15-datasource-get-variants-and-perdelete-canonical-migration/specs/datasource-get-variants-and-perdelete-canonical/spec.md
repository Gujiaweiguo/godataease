## ADDED Requirements

### Requirement: Datasource HidePw Must Have A Canonical Route
The system SHALL provide datasource details with password hidden through a canonical route at `GET /api/ds/hidePw/:id`.

#### Scenario: Get datasource with password hidden through canonical route
- **WHEN** a client calls `GET /api/ds/hidePw/:id` with a valid datasource ID
- **THEN** the system MUST return datasource details with password fields masked/hidden using the established compatibility-safe response envelope
- **AND** the canonical route MUST preserve the hidePw semantics currently expected by datasource editing flows

#### Scenario: Get datasource with password hidden for non-existent ID
- **WHEN** a client calls `GET /api/ds/hidePw/:id` with a non-existent datasource ID
- **THEN** the system MUST return an explicit non-success response with actionable error detail
- **AND** the response MUST NOT degrade into a generic success or empty payload

### Requirement: Datasource Simple Must Have A Canonical Route
The system SHALL provide simplified datasource info through a canonical route at `GET /api/ds/simple/:id`.

#### Scenario: Get simplified datasource info through canonical route
- **WHEN** a client calls `GET /api/ds/simple/:id` with a valid datasource ID
- **THEN** the system MUST return simplified datasource information using the established compatibility-safe response envelope
- **AND** the canonical route MUST preserve the simplified response structure currently expected by datasource selection flows

#### Scenario: Get simplified datasource info for non-existent ID
- **WHEN** a client calls `GET /api/ds/simple/:id` with a non-existent datasource ID
- **THEN** the system MUST return an explicit non-success response with actionable error detail
- **AND** the response MUST NOT silently degrade into an empty success payload

### Requirement: Datasource PerDelete Must Have A Canonical Route
The system SHALL provide permanent datasource deletion through a canonical route at `POST /api/ds/perDelete/:id`.

#### Scenario: Permanently delete datasource through canonical route
- **WHEN** a client calls `POST /api/ds/perDelete/:id` with a valid datasource ID
- **THEN** the system MUST permanently delete the datasource and return results using the established compatibility-safe response envelope
- **AND** the canonical route MUST preserve the permanent deletion semantics currently expected by datasource management flows

#### Scenario: Permanently delete non-existent datasource
- **WHEN** a client calls `POST /api/ds/perDelete/:id` with a non-existent datasource ID
- **THEN** the system MUST return an explicit non-success response with actionable error detail
- **AND** the response MUST NOT silently degrade into an empty success payload

#### Scenario: Permanently delete datasource with unmet preconditions
- **WHEN** a client calls `POST /api/ds/perDelete/:id` for a datasource that has blocking dependencies
- **THEN** the system MUST return an explicit non-success response indicating the blocking condition
- **AND** the response MUST NOT perform a partial deletion or return a misleading success

### Requirement: Canonical Datasource Get Variants And PerDelete Migration Must Preserve Compatibility Safety
The system SHALL allow frontend datasource callers to migrate to canonical routes without removing compatibility aliases in the same change.

#### Scenario: Frontend callers switch safely to canonical routes
- **WHEN** the frontend datasource API layer migrates getHidePwById, getSimpleDs, and perDelete callers from `/datasource/*` to `/api/ds/*`
- **THEN** the corresponding compatibility routes MUST remain available during the migration window
- **AND** rollback MUST remain possible by reverting frontend route selection without requiring emergency backend route restoration

### Requirement: Frontend Datasource HidePw Call Must Use Canonical Route
The system SHALL ensure the frontend datasource hidePw call uses the canonical route.

#### Scenario: Datasource hidePw uses canonical route
- **WHEN** the frontend calls datasource hidePw (getHidePwById)
- **THEN** the API call MUST target the canonical route `/api/ds/hidePw/:id` instead of the compatibility path `/datasource/hidePw/:id`

### Requirement: Frontend Datasource Simple Call Must Use Canonical Route
The system SHALL ensure the frontend datasource simple call uses the canonical route.

#### Scenario: Datasource getSimpleDs uses canonical route
- **WHEN** the frontend calls datasource getSimpleDs
- **THEN** the API call MUST target the canonical route `/api/ds/simple/:id` instead of the compatibility path `/datasource/getSimpleDs/:id`

### Requirement: Frontend Datasource PerDelete Call Must Use Canonical Route
The system SHALL ensure the frontend datasource permanent delete call uses the canonical route.

#### Scenario: Datasource perDelete uses canonical route
- **WHEN** the frontend calls datasource perDelete
- **THEN** the API call MUST target the canonical route `/api/ds/perDelete/:id` instead of the compatibility path `/datasource/perDelete/:id`
