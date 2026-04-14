# datasource-preview-sync-canonical Specification

## Purpose
Define the governed canonical datasource preview and synchronization routes and the compatibility-safe migration boundary for frontend callers.

## Requirements
### Requirement: Datasource Preview Data Must Have A Canonical Route
The system SHALL provide datasource preview data through a canonical route at `POST /api/ds/previewData`.

#### Scenario: Query datasource preview data through canonical route
- **WHEN** a client calls `POST /api/ds/previewData` with a valid datasource preview request
- **THEN** the system MUST return preview rows using the established compatibility-safe response envelope
- **AND** the canonical route MUST preserve preview semantics currently expected by datasource preview flows

### Requirement: Datasource Table Sync Must Have A Canonical Route
The system SHALL provide datasource table synchronization through a canonical route at `POST /api/ds/syncApiTable`.

#### Scenario: Trigger table sync through canonical route
- **WHEN** a client calls `POST /api/ds/syncApiTable` with a valid datasource table sync request
- **THEN** the system MUST execute table synchronization with behavior equivalent to the compatibility route
- **AND** the response MUST preserve explicit success/failure semantics without static placeholder fallback

### Requirement: Datasource Sync Discovery Must Have A Canonical Route
The system SHALL provide datasource synchronization discovery through a canonical route at `POST /api/ds/syncApiDs`.

#### Scenario: Query sync discovery through canonical route
- **WHEN** a client calls `POST /api/ds/syncApiDs` with a valid sync discovery request
- **THEN** the system MUST return synchronization discovery payloads using the established compatibility-safe response envelope
- **AND** the canonical route MUST remain compatible with existing datasource sync workflows

### Requirement: Canonical Preview And Sync Migration Must Preserve Compatibility Safety
The system SHALL allow frontend datasource preview and sync callers to migrate to canonical routes without removing compatibility aliases in the same change.

#### Scenario: Frontend preview and sync callers switch safely to canonical routes
- **WHEN** the frontend datasource API layer migrates preview and sync callers from `/datasource/*` to `/api/ds/*`
- **THEN** the corresponding compatibility routes MUST remain available during the migration window
- **AND** rollback MUST remain possible by reverting frontend route selection without requiring emergency backend route restoration
