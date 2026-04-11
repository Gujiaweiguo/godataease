## ADDED Requirements

### Requirement: Datasource Table Status Must Expose Stable Execution-State Semantics
The system SHALL expose datasource table status using a stable business-facing execution-state model rather than a minimally aggregated transport value.

#### Scenario: Datasource table status maps runtime evidence into a stable state set
- **WHEN** the system returns table status for a datasource table with available sync or execution evidence
- **THEN** it MUST map that evidence into a stable documented state set suitable for frontend rendering
- **AND** the frontend contract MUST NOT depend on raw task-system enum values

#### Scenario: Datasource table status includes authoritative update time
- **WHEN** table status is derived from multiple possible runtime timestamps such as create, start, or end time
- **THEN** the system MUST choose a deterministic authoritative update time according to documented precedence
- **AND** the chosen time MUST be suitable for stable frontend display and regression testing

#### Scenario: Datasource table status preserves unknown-state semantics
- **WHEN** runtime evidence is incomplete, missing, or insufficient to derive a meaningful execution state
- **THEN** the system MUST return an explicit unknown, warning, or equivalent non-success-assuming state
- **AND** it MUST NOT infer a completed state solely from route reachability or table existence
