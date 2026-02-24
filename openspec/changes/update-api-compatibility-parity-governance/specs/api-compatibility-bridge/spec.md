## ADDED Requirements

### Requirement: Placeholder Success Prohibition for Compatibility Endpoints
Migration-scoped compatibility endpoints SHALL NOT return placeholder success responses when core behavior is not implemented.

#### Scenario: Reject placeholder success semantics
- **WHEN** a compatibility endpoint lacks required business implementation
- **THEN** the endpoint MUST return explicit non-success semantics (for example deterministic unavailable/error status)
- **AND** MUST NOT return `code=000000` with empty placeholder payload to simulate parity

### Requirement: Compatibility Runtime-to-Matrix Status Consistency
Compatibility route status metadata SHALL remain consistent with observed runtime behavior.

#### Scenario: Keep matrix status synchronized with runtime behavior
- **WHEN** endpoint behavior changes between `full/partial/stub/missing`
- **THEN** migration matrix and whitelist metadata MUST be updated in the same change set
- **AND** changes MUST include evidence references to tests or contract-diff outputs

#### Scenario: Block merge on status drift
- **WHEN** CI detects status drift between route behavior and governed metadata
- **THEN** the pipeline MUST fail with actionable drift diagnostics
- **AND** merge MUST be blocked until metadata and behavior are aligned

### Requirement: Time-Bounded Stub Waiver Governance
Temporary compatibility stubs SHALL require explicit, time-bounded waiver governance.

#### Scenario: Use approved waiver for temporary stubs
- **WHEN** a team keeps an endpoint in stub status for a release window
- **THEN** the stub MUST reference approved owner, reason, and expiry metadata
- **AND** expired waiver MUST stop unblocking release gates automatically
