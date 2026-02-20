## ADDED Requirements

### Requirement: CI Contract Diff Gate for Critical Java/Go APIs
The migration pipeline SHALL run a CI contract diff gate for critical Java/Go APIs before merge and release.

#### Scenario: Run gate on pull request
- **WHEN** a pull request changes Go backend code or compatibility routing behavior
- **THEN** CI MUST execute the Java/Go contract diff job for whitelisted critical APIs
- **AND** the pull request MUST be blocked when the gate result is failed

#### Scenario: Run gate on protected branch
- **WHEN** code is merged into protected branches
- **THEN** CI SHALL rerun the contract diff gate and persist the result for audit

### Requirement: Whitelisted Critical API Set Governance
The system SHALL maintain a versioned whitelist of critical APIs included in the contract diff gate.

#### Scenario: Maintain whitelist metadata
- **WHEN** an API is added to or removed from the whitelist
- **THEN** the change MUST include endpoint, HTTP method, owner, and blocking level metadata
- **AND** the change MUST be reviewable in version control

#### Scenario: Limit gate scope to governed APIs
- **WHEN** contract diff gate executes
- **THEN** only APIs in the approved whitelist are considered for pass/fail evaluation

### Requirement: Failure Threshold Policy and Report Archival
The contract diff gate SHALL enforce explicit failure thresholds and archive diff reports for each run.

#### Scenario: Enforce threshold-based failure
- **WHEN** diff results violate configured thresholds (overall parity, required APIs, or blocking-level differences)
- **THEN** CI MUST return a failed status and provide categorized failure reasons

#### Scenario: Archive gate reports
- **WHEN** the contract diff gate completes
- **THEN** CI SHALL archive machine-readable and human-readable reports as build artifacts
- **AND** archived reports MUST allow tracing by commit or pull request identifier

#### Scenario: Update threshold policy
- **WHEN** threshold policy is updated for migration phase changes
- **THEN** the policy change MUST be versioned and reviewable in repository history
- **AND** gate evaluation MUST use the latest approved threshold version
