## ADDED Requirements

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
