## ADDED Requirements

### Requirement: Datasource File Upload Must Have A Canonical Route
The system SHALL provide datasource file upload through a canonical route at `POST /api/ds/uploadFile`.

#### Scenario: Upload datasource file through canonical route
- **WHEN** a client calls `POST /api/ds/uploadFile` with a valid multipart datasource file upload request
- **THEN** the system MUST return upload results using the established compatibility-safe response envelope
- **AND** the canonical route MUST preserve the upload semantics currently expected by datasource ingest flows

### Requirement: Datasource Remote File Load Must Have A Canonical Route
The system SHALL provide datasource remote file loading through a canonical route at `POST /api/ds/loadRemoteFile`.

#### Scenario: Load remote datasource file through canonical route
- **WHEN** a client calls `POST /api/ds/loadRemoteFile` with a valid remote file ingest request
- **THEN** the system MUST return remote file loading results using the established compatibility-safe response envelope
- **AND** the canonical route MUST remain compatible with existing datasource file ingest workflows

### Requirement: Canonical Datasource File Ingest Migration Must Preserve Compatibility Safety
The system SHALL allow frontend datasource file ingest callers to migrate to canonical routes without removing compatibility aliases in the same change.

#### Scenario: Frontend file ingest callers switch safely to canonical routes
- **WHEN** the frontend datasource API layer migrates upload and remote file callers from `/datasource/*` to `/api/ds/*`
- **THEN** the corresponding compatibility routes MUST remain available during the migration window
- **AND** rollback MUST remain possible by reverting frontend route selection without requiring emergency backend route restoration
