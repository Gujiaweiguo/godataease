## ADDED Requirements

### Requirement: Datasource Type List Must Have A Canonical Route
The system SHALL provide the hardcoded datasource type list through a canonical route at `GET /api/ds/types`.

#### Scenario: Get datasource type list through canonical route
- **WHEN** a client calls `GET /api/ds/types`
- **THEN** the system MUST return the hardcoded list of supported datasource types (MySQL, PostgreSQL, SQL Server, Oracle, Excel)
- **AND** the canonical route MUST return the same response shape as the compatibility route `POST /datasource/types`

#### Scenario: Datasource type list requires no service call
- **WHEN** the canonical types endpoint is invoked
- **THEN** the handler MUST return the static type list without invoking any service layer method
- **AND** the response MUST use the established compatibility-safe response envelope

### Requirement: Datasource ShowFinishPage Must Have A Canonical GET Route
The system SHALL provide the finish-page visibility status through a canonical route at `GET /api/ds/showFinishPage`.

#### Scenario: Get finish-page status through canonical route
- **WHEN** a client calls `GET /api/ds/showFinishPage` with a valid authenticated user
- **THEN** the system MUST extract the user ID from JWT context and call `service.ShowFinishPage(userID)`
- **AND** the response MUST use the established compatibility-safe response envelope

#### Scenario: Get finish-page status without authentication
- **WHEN** a client calls `GET /api/ds/showFinishPage` without a valid JWT token
- **THEN** the system MUST return an authentication-required error response
- **AND** the response MUST NOT reveal internal user ID extraction details

### Requirement: Datasource SetShowFinishPage Must Have A Canonical POST Route
The system SHALL provide the finish-page dismissal action through a canonical route at `POST /api/ds/showFinishPage`.

#### Scenario: Dismiss finish page through canonical route
- **WHEN** a client calls `POST /api/ds/showFinishPage` with a valid authenticated user
- **THEN** the system MUST extract the user ID from JWT context and call `service.SetShowFinishPage(userID)`
- **AND** the response MUST use the established compatibility-safe response envelope

#### Scenario: Dismiss finish page without authentication
- **WHEN** a client calls `POST /api/ds/showFinishPage` without a valid JWT token
- **THEN** the system MUST return an authentication-required error response

### Requirement: Datasource LatestUse Must Have A Canonical Route
The system SHALL provide recently used datasource types through a canonical route at `POST /api/ds/latestUse`.

#### Scenario: Get recently used datasource types through canonical route
- **WHEN** a client calls `POST /api/ds/latestUse` with a valid authenticated user
- **THEN** the system MUST extract the username from JWT context and call `service.LatestTypes(username)`
- **AND** the response MUST use the established compatibility-safe response envelope

#### Scenario: Get recently used types without authentication
- **WHEN** a client calls `POST /api/ds/latestUse` without a valid JWT token
- **THEN** the system MUST return an authentication-required error response

### Requirement: Datasource SyncRecord List Must Have A Canonical Route
The system SHALL provide sync record pagination through a canonical route at `POST /api/ds/syncRecord/:dsId/:page/:limit`.

#### Scenario: List sync records through canonical route
- **WHEN** a client calls `POST /api/ds/syncRecord/:dsId/:page/:limit` with valid parameters
- **THEN** the system MUST parse dsId (int64), page (int, minimum 1), and limit (int, minimum 1, default 10) from URL params
- **AND** call `service.ListSyncRecord(dsID, page, limit)`
- **AND** return results using the established compatibility-safe response envelope

#### Scenario: List sync records with invalid dsId
- **WHEN** a client calls `POST /api/ds/syncRecord/invalid/1/10`
- **THEN** the system MUST return an explicit error response indicating invalid datasource ID format
- **AND** the response MUST NOT silently return an empty success payload

#### Scenario: List sync records with page below minimum
- **WHEN** a client calls `POST /api/ds/syncRecord/1/0/10`
- **THEN** the system MUST return an explicit error response indicating invalid page parameter
- **AND** the system MUST NOT accept page values less than 1

### Requirement: Canonical Datasource UI Metadata Migration Must Preserve Compatibility Safety
The system SHALL allow frontend datasource callers to migrate to canonical routes without removing compatibility aliases in the same change.

#### Scenario: Frontend callers switch safely to canonical routes
- **WHEN** the frontend datasource API layer migrates types, showFinishPage, setShowFinishPage, latestUse, and syncRecord callers from `/datasource/*` to `/api/ds/*`
- **THEN** the corresponding compatibility routes MUST remain available during the migration window
- **AND** rollback MUST remain possible by reverting frontend route selection without requiring emergency backend route restoration

### Requirement: Frontend Datasource Types Call Must Use Canonical Route
The system SHALL ensure the frontend datasource types call uses the canonical route.

#### Scenario: Datasource types uses canonical route
- **WHEN** the frontend calls datasource types
- **THEN** the API call MUST target the canonical route `GET /api/ds/types` instead of the compatibility path `POST /datasource/types`

### Requirement: Frontend Datasource ShowFinishPage Call Must Use Canonical Route
The system SHALL ensure the frontend datasource showFinishPage call uses the canonical route.

#### Scenario: Datasource showFinishPage uses canonical route
- **WHEN** the frontend calls datasource showFinishPage
- **THEN** the API call MUST target the canonical route `GET /ds/showFinishPage` instead of the compatibility path `/datasource/showFinishPage`

### Requirement: Frontend Datasource SetShowFinishPage Call Must Use Canonical Route
The system SHALL ensure the frontend datasource setShowFinishPage call uses the canonical route.

#### Scenario: Datasource setShowFinishPage uses canonical route
- **WHEN** the frontend calls datasource setShowFinishPage
- **THEN** the API call MUST target the canonical route `POST /ds/showFinishPage` instead of the compatibility path `/datasource/setShowFinishPage`

### Requirement: Frontend Datasource LatestUse Call Must Use Canonical Route
The system SHALL ensure the frontend datasource latestUse call uses the canonical route.

#### Scenario: Datasource latestUse uses canonical route
- **WHEN** the frontend calls datasource latestUse
- **THEN** the API call MUST target the canonical route `POST /ds/latestUse` instead of the compatibility path `/datasource/latestUse`

### Requirement: Frontend Datasource SyncRecord Call Must Use Canonical Route
The system SHALL ensure the frontend datasource syncRecord call uses the canonical route.

#### Scenario: Datasource syncRecord uses canonical route
- **WHEN** the frontend calls datasource syncRecord
- **THEN** the API call MUST target the canonical route `POST /ds/syncRecord/:dsId/:page/:limit` instead of the compatibility path `/datasource/listSyncRecord/:dsId/:page/:limit`
