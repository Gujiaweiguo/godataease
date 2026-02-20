## ADDED Requirements

### Requirement: Export Async State and Authorization Compatibility
The Go backend SHALL align export task state transitions and download authorization semantics with migration baseline.

#### Scenario: Query export task lifecycle state
- **WHEN** client polls export task status during execution
- **THEN** task states follow deterministic transition rules compatible with Java client expectations
- **AND** terminal states include sufficient failure reason for client handling

#### Scenario: Download exported artifact without permission
- **WHEN** an unauthorized caller requests export artifact download
- **THEN** backend denies request using mapped Java-compatible error semantics
- **AND** no artifact content or signed link is exposed
