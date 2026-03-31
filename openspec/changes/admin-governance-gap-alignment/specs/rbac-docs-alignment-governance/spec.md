## ADDED Requirements

### Requirement: Governance Baseline Freeze for Admin-Domain Alignment
The change MUST freeze the official admin-domain baseline before any implementation task can be considered in scope.

#### Scenario: Execution starts for admin governance alignment
- **WHEN** maintainers begin execution for organization, user, role, or permission alignment
- **THEN** the change MUST record the exact official document URLs and baseline date
- **AND** later gap decisions MUST trace back to that frozen baseline instead of ad hoc interpretations

### Requirement: Gap Classification Matrix Must Be the Source of Truth
The change MUST classify every audited item as implemented-but-inconsistent, partially-implemented, or not-implemented with supporting evidence.

#### Scenario: Team reviews an admin-domain gap
- **WHEN** a capability gap is reviewed during implementation planning or execution
- **THEN** the change MUST provide manual reference, current evidence, risk level, and planned action for that gap
- **AND** unclassified items MUST NOT bypass the governance workflow

### Requirement: Semantic-First Change Decomposition
The change MUST decompose execution around shared semantics rather than page adjacency.

#### Scenario: Team decides execution order
- **WHEN** maintainers sequence work across organization, user, role, and permission domains
- **THEN** the workflow MUST prioritize foundation semantics before lifecycle alignment and permission-center alignment
- **AND** row/column permission work MUST NOT become a first-wave change ahead of shared IAM semantics
