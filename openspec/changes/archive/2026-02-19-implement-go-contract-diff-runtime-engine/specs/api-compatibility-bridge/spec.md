## ADDED Requirements

### Requirement: Contract Diff Runtime Engine Execution
The system SHALL provide an executable contract diff runtime engine that compares Java and Go API responses based on a governed whitelist.

#### Scenario: Load and validate whitelist before execution
- **WHEN** the runtime engine starts with a whitelist file
- **THEN** it MUST validate required fields (`path`, `method`, `owner`, `priority`, `blockingLevel`) for every API entry
- **AND** it MUST stop with a deterministic configuration error if any required field is missing or invalid

#### Scenario: Execute comparisons in bounded parallel mode
- **WHEN** the runtime engine processes the whitelist API set
- **THEN** it MUST execute Java/Go requests with configurable bounded concurrency
- **AND** it MUST produce deterministic aggregated results independent of request completion order

### Requirement: Stable Timeout and Retry Semantics
The runtime engine SHALL apply stable timeout and retry behavior for transient request failures.

#### Scenario: Retry transient failures
- **WHEN** an API request fails due to transient timeout or connection issues
- **THEN** the engine MUST retry according to configured retry count and timeout policy
- **AND** each exhausted retry MUST be recorded with failure category and final error detail

#### Scenario: Preserve deterministic final status
- **WHEN** retries are exhausted for an API entry
- **THEN** the engine MUST emit a deterministic final failure result for that entry
- **AND** the gate decision MUST use the post-retry final result only

### Requirement: Structured Diff and Failure Classification Output
The runtime engine SHALL output structured diffs and classified failures for CI gate consumption.

#### Scenario: Emit multi-dimensional diff
- **WHEN** Java and Go responses are both available
- **THEN** the output MUST include diff dimensions for `status`, `code`, `msg`, `payload schema`, and `payload value`
- **AND** each API result MUST include `passed` status and categorized failure reason when failed

#### Scenario: Produce gate-ready reports and exit semantics
- **WHEN** diff execution finishes
- **THEN** the engine MUST write both machine-readable JSON and human-readable Markdown reports
- **AND** it MUST return non-zero exit code if blocking-level failures violate threshold policy, otherwise return zero
