# go-compat-endpoint-coverage Specification

## Purpose
This capability governs endpoint coverage parity between frontend critical flows and Go backend compatibility routes, ensuring no regression during Java-to-Go migration.

## Requirements
### Requirement: Critical Frontend Flow Endpoint Coverage Matrix
The system SHALL maintain a governed matrix for critical frontend flows that maps each required API endpoint to Go runtime implementation status.

#### Scenario: Build matrix from role/menu and dashboard flows
- **WHEN** the change defines critical flows for role-menu permission management and dashboard resource-tree operations
- **THEN** the matrix includes endpoint path, method, owning module, and implementation status (`full/partial/stub/missing`)

#### Scenario: Reject untracked endpoint gaps
- **WHEN** a frontend critical-flow endpoint is missing from the matrix
- **THEN** the compatibility validation process MUST fail before release

### Requirement: Endpoint Coverage Regression Verification
The system SHALL provide repeatable verification that critical-flow endpoints are reachable and contract-compatible.

#### Scenario: Verify non-404 and envelope contract
- **WHEN** regression verification runs for matrix endpoints
- **THEN** each endpoint MUST return non-404 behavior
- **AND** responses MUST follow `code/data/msg` envelope for API routes
