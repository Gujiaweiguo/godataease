## ADDED Requirements

### Requirement: Dataset and Datasource Interactive Governance Consistency
The compatibility and governance model SHALL describe dataset and datasource interactive discovery in a way that is consistent with the interactive aggregate view used by the frontend.

#### Scenario: Interactive aggregate governance covers dataset and datasource discovery
- **WHEN** interactive discovery behavior is evaluated for release readiness
- **THEN** dataset and datasource discovery paths used by the aggregate interactive loader MUST be documented with implementation evidence and governed status
- **AND** the documented status MUST match the actual runtime loading path used by the frontend
