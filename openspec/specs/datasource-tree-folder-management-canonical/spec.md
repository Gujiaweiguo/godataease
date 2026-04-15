# datasource-tree-folder-management-canonical Specification

## Purpose
Define canonical datasource tree and folder management route requirements.

## Requirements

### Requirement: Datasource Move Must Have A Canonical Route
The system SHALL provide datasource or folder move through a canonical route at `POST /api/ds/move`.

#### Scenario: Move datasource or folder through canonical route
- **WHEN** a client calls `POST /api/ds/move` with a valid move request specifying the target datasource/folder and destination
- **THEN** the system MUST return move results using the established compatibility-safe response envelope
- **AND** the canonical route MUST preserve the move semantics currently expected by datasource tree management flows

#### Scenario: Move datasource or folder with invalid target or destination
- **WHEN** a client calls `POST /api/ds/move` with an invalid or non-existent target/destination
- **THEN** the system MUST return an explicit non-success response with actionable error detail
- **AND** the response MUST NOT degrade into a generic success or empty payload

### Requirement: Datasource Rename Must Have A Canonical Route
The system SHALL provide datasource or folder rename through a canonical route at `POST /api/ds/reName`.

#### Scenario: Rename datasource or folder through canonical route
- **WHEN** a client calls `POST /api/ds/reName` with a valid rename request specifying the target and new name
- **THEN** the system MUST return rename results using the established compatibility-safe response envelope
- **AND** the canonical route MUST remain compatible with existing datasource tree management workflows

#### Scenario: Rename datasource or folder with invalid request
- **WHEN** a client calls `POST /api/ds/reName` with an invalid or malformed request body
- **THEN** the system MUST return an explicit non-success response
- **AND** the response MUST NOT silently degrade into an empty success payload

### Requirement: Datasource Folder Creation Must Have A Canonical Route
The system SHALL provide datasource folder creation through a canonical route at `POST /api/ds/createFolder`.

#### Scenario: Create datasource folder through canonical route
- **WHEN** a client calls `POST /api/ds/createFolder` with a valid folder creation request
- **THEN** the system MUST return folder creation results using the established compatibility-safe response envelope
- **AND** the canonical route MUST remain compatible with existing datasource folder management workflows

#### Scenario: Create datasource folder with invalid or duplicate name
- **WHEN** a client calls `POST /api/ds/createFolder` with an invalid request or duplicate folder name
- **THEN** the system MUST return an explicit non-success response with actionable error detail
- **AND** the response MUST NOT silently degrade into an empty success payload

### Requirement: Canonical Datasource Tree/Folder Management Migration Must Preserve Compatibility Safety
The system SHALL allow frontend datasource tree/folder management callers to migrate to canonical routes without removing compatibility aliases in the same change.

#### Scenario: Frontend tree/folder management callers switch safely to canonical routes
- **WHEN** the frontend datasource API layer migrates move, reName, and createFolder callers from `/datasource/*` to `/api/ds/*`
- **THEN** the corresponding compatibility routes MUST remain available during the migration window
- **AND** rollback MUST remain possible by reverting frontend route selection without requiring emergency backend route restoration

### Requirement: Frontend Dataset Datasource Tree And Tables Calls Must Use Canonical Routes
The system SHALL ensure frontend dataset module calls for datasource tree and tables use existing canonical routes.

#### Scenario: Dataset tree query uses canonical route
- **WHEN** the frontend dataset module requests the datasource tree
- **THEN** the API call MUST target the canonical route `/ds/tree` instead of the compatibility path `/datasource/tree`

#### Scenario: Dataset getTables query uses canonical route
- **WHEN** the frontend dataset module requests datasource tables
- **THEN** the API call MUST target the canonical route `/ds/tables` instead of the compatibility path `/datasource/getTables`

### Requirement: Frontend Datasource Validate POST Must Use Canonical Route
The system SHALL ensure the frontend datasource validate POST call uses the existing canonical route.

#### Scenario: Datasource validate POST uses canonical route
- **WHEN** the frontend calls datasource validate with POST method
- **THEN** the API call MUST target the canonical route `/ds/validate` instead of the compatibility path `/datasource/validate`
