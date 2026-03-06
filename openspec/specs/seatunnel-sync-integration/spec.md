# seatunnel-sync-integration Specification

## Purpose
Define SeaTunnel synchronization integration requirements for task submission, status tracking, and retry control.
## Requirements
### Requirement: SeaTunnel Sync Task Orchestration
The Go backend SHALL orchestrate datasource sync tasks through SeaTunnel for migration-critical sync workflows.

#### Scenario: Submit sync task successfully
- **WHEN** a client requests datasource or table sync
- **THEN** the backend submits a SeaTunnel task and returns task identity and initial status
- **AND** task metadata is persisted for later query

#### Scenario: Return deterministic submit failure
- **WHEN** SeaTunnel submission fails
- **THEN** the backend returns deterministic failure semantics
- **AND** failure reason is recorded in sync task metadata

### Requirement: Sync Task Lifecycle and Status Parity
The backend SHALL provide deterministic lifecycle status mapping compatible with Java client expectations.

#### Scenario: Query task status
- **WHEN** a client queries sync status
- **THEN** the backend returns mapped lifecycle status (`pending/running/success/failed/cancelled`)
- **AND** includes progress and last update timestamp when available

#### Scenario: Cancel running task
- **WHEN** a client requests task cancellation for a cancellable running task
- **THEN** the backend invokes SeaTunnel cancel operation
- **AND** returns final cancellation status deterministically

### Requirement: Sync Record Pagination
The backend SHALL provide paginated sync records for compatibility route consumers.

#### Scenario: List sync records by datasource
- **WHEN** a client requests `listSyncRecord` with datasource ID and paging parameters
- **THEN** the backend returns persisted sync records with `total/current/size/records`
- **AND** records are ordered deterministically by creation or update time
