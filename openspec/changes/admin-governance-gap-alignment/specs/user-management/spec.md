## ADDED Requirements

### Requirement: User Lifecycle Must Be Organization-Scoped by Default
The system MUST treat organization selection as a first-class prerequisite for governed user administration.

#### Scenario: Administrator creates or edits a governed user
- **WHEN** an administrator creates or edits a user in the governed workflow
- **THEN** the workflow MUST bind the operation to an explicit organization context
- **AND** account identity fields that are defined as immutable by baseline policy MUST remain immutable after creation

### Requirement: User Enabled State Must Govern Login Eligibility
The system MUST align user enabled-state semantics with the frozen official baseline.

#### Scenario: Disabled user attempts authentication
- **WHEN** a disabled user attempts to authenticate or reuse a governed session
- **THEN** the system MUST deny access according to the governed error semantics
- **AND** user-management and authentication flows MUST observe the same enabled-state contract

### Requirement: User Import and Source Metadata Must Have Explicit Governance Boundaries
The system MUST treat import, error-report output, and third-party source metadata as governed user-management concerns with explicit rollout boundaries.

#### Scenario: Team evaluates user import parity
- **WHEN** maintainers align user import against the frozen official baseline
- **THEN** the change MUST preserve partial-success and error-report behavior
- **AND** any unsupported third-party source metadata MUST be recorded as deferred or intentionally bounded instead of being silently ignored
