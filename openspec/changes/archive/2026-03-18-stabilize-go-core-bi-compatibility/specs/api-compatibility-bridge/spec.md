## ADDED Requirements

### Requirement: Core BI Compatibility Gate Scope
The compatibility bridge SHALL define datasource, dataset, dashboard, and big-screen endpoints as a governed critical-flow scope for migration release readiness.

#### Scenario: Govern core BI routes as required verification scope
- **WHEN** compatibility gate scope is prepared for release or merge validation
- **THEN** the scope MUST identify the canonical route, compatibility alias, owner, and blocking level for each in-scope BI endpoint family
- **AND** the governed scope MUST be reviewable in version control with evidence references

#### Scenario: Block release on missing core BI verification
- **WHEN** any governed datasource, dataset, dashboard, or big-screen endpoint is not evaluated or fails required parity checks
- **THEN** release readiness MUST be treated as failed
- **AND** the system MUST provide actionable evidence for the missing or failing route family

### Requirement: Core BI Compatibility Endpoints Must Not Be Stub-Success
Compatibility endpoints that participate in the critical BI flow SHALL return implemented behavior or explicit non-success semantics.

#### Scenario: Reject placeholder success for critical BI compatibility route
- **WHEN** an in-scope BI compatibility endpoint lacks required business implementation
- **THEN** the endpoint MUST return deterministic non-success behavior
- **AND** MUST NOT return `code=000000` with placeholder payload to simulate parity

#### Scenario: Keep alias and canonical route semantics aligned for critical BI flows
- **WHEN** both canonical and compatibility forms of an in-scope BI route are invoked with equivalent inputs
- **THEN** status, envelope semantics, and business result shape MUST remain aligned
- **AND** divergence MUST be treated as a governed compatibility regression
