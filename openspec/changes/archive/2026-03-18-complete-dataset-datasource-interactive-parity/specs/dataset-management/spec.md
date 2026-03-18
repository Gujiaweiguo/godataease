## ADDED Requirements

### Requirement: Dataset Interactive Aggregate Parity
The system SHALL provide dataset tree data for the frontend interactive aggregate view with the same governed contract stability expected by the other BI discovery domains.

#### Scenario: Interactive aggregate can load dataset tree consistently
- **WHEN** the frontend interactive aggregate requests dataset discovery data
- **THEN** the system MUST return dataset tree nodes using the established `BusiTreeNode` contract
- **AND** the interactive loader MUST NOT require a dataset-specific fallback behavior that breaks aggregate consistency

#### Scenario: Dataset interactive nodes remain structurally valid
- **WHEN** dataset tree data is consumed through the interactive aggregate flow
- **THEN** returned nodes MUST preserve valid identifiers, parent relationships, `leaf` semantics, and required children structure
