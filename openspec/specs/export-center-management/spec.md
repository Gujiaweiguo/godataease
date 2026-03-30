# export-center-management Specification

## Purpose
Define export center task lifecycle requirements for querying, retrying, deleting, and downloading exports.
## Requirements
### Requirement: Export Task Lifecycle
The system SHALL provide export task lifecycle management in Go backend.

#### Scenario: Create export task
- **WHEN** client submits export request with resource and format parameters
- **THEN** the system creates an asynchronous export task
- **AND** returns task identifier for tracking

#### Scenario: Query export task status
- **WHEN** client queries export task status
- **THEN** the system returns progress and final status
- **AND** status values are consistent with Java migration baseline

### Requirement: Export File Retrieval
The system SHALL provide controlled export file download capability.

#### Scenario: Download exported file
- **WHEN** export task is completed and caller is authorized
- **THEN** the system returns downloadable file stream or signed URL
- **AND** applies expiration and access control checks

### Requirement: Export Center Stability Recovery
The system SHALL treat export-center list, retry, and download flows as a governed broken-feature recovery surface.

#### Scenario: Export task management remains reachable and explicit
- **WHEN** a user opens export-center task management and queries or retries export tasks
- **THEN** the relevant route and page-init flow MUST remain reachable and explicit about success or failure
- **AND** recovery work MUST NOT normalize broken behavior into silent empty success

#### Scenario: Export download failure stays diagnosable during recovery
- **WHEN** export download fails because of authorization, task state, or backend execution problems
- **THEN** the system MUST return or surface deterministic failure semantics
- **AND** verification evidence MUST show the recovered flow is distinguishable from missing-route behavior

### Requirement: Export Center Route-Level Hardening
The system SHALL preserve explicit export-center route and download semantics after the operational recovery batch is complete.

#### Scenario: Export-center download path remains explicit after auth
- **WHEN** a caller exercises export-center download behavior through a governed route
- **THEN** authorization, not-found, and business-state failures MUST remain explicit
- **AND** the path MUST not degrade into HTML fallback or silent success

### Requirement: Export center remains reachable through governed toolbox navigation
The system SHALL keep export-center functionality reachable through the governed Toolbox menu after shell navigation restructuring.

#### Scenario: Authorized user opens export center from toolbox child menu
- **WHEN** an authorized user expands Toolbox and selects Data Export Center
- **THEN** the system MUST open the export-center workflow using the same business semantics as before the restructure
- **AND** the workflow MUST remain reachable without depending on a header More menu entry

#### Scenario: Export-center access respects menu authorization after restructure
- **WHEN** a role loses authorization to the export-center child menu or its toolbox parent
- **THEN** the export-center entry MUST disappear from governed navigation on next authorization refresh
- **AND** retained users with authorization MUST continue to reach the workflow through Toolbox
