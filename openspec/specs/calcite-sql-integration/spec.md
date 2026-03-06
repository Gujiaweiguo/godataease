# calcite-sql-integration Specification

## Purpose
Define Calcite-based SQL validation integration requirements for dataset preview and error handling.
## Requirements
### Requirement: Calcite SQL Parse and Validate Integration
The Go backend SHALL implement Calcite-backed SQL parsing and validation for migration-critical SQL workflows.

#### Scenario: Parse valid SQL through Calcite
- **WHEN** a client submits syntactically valid SQL to a Calcite-integrated endpoint
- **THEN** the backend invokes Calcite parse RPC successfully
- **AND** returns normalized parse output in contract-compatible response format

#### Scenario: Reject invalid SQL with deterministic error
- **WHEN** a client submits invalid SQL
- **THEN** the backend returns a deterministic validation failure
- **AND** includes structured error details compatible with Java client handling

### Requirement: Calcite RPC Reliability Policy
The backend SHALL apply deterministic timeout and retry behavior for Calcite RPC calls.

#### Scenario: Retry transient Calcite failures
- **WHEN** Calcite RPC fails with transient timeout or connection errors
- **THEN** the backend retries according to configured retry policy
- **AND** records the final failure reason after retry exhaustion

#### Scenario: Fail fast on non-retriable errors
- **WHEN** Calcite returns non-retriable request/validation errors
- **THEN** the backend MUST NOT retry
- **AND** returns mapped error semantics immediately

### Requirement: Dataset SQL Workflow Parity
Dataset SQL preview and related compatibility workflows SHALL enforce Calcite validation before execution.

#### Scenario: Validate before SQL preview execution
- **WHEN** a dataset SQL preview request is received
- **THEN** the backend validates SQL via Calcite before executing preview logic
- **AND** blocks execution when validation fails
