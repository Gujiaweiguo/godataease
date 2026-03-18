## ADDED Requirements

### Requirement: Datasource Interactive Aggregate Parity
The system SHALL provide datasource tree data for the frontend interactive aggregate view with the same governed contract stability expected by the other BI discovery domains.

#### Scenario: Interactive aggregate can load datasource tree consistently
- **WHEN** the frontend interactive aggregate requests datasource discovery data
- **THEN** the system MUST return datasource tree nodes using the established `BusiTreeNode` contract
- **AND** the interactive loader MUST NOT require a datasource-specific fallback behavior that breaks aggregate consistency

#### Scenario: Datasource interactive nodes remain structurally valid
- **WHEN** datasource tree data is consumed through the interactive aggregate flow
- **THEN** returned nodes MUST preserve valid identifiers, parent relationships, `leaf` semantics, and required children structure
