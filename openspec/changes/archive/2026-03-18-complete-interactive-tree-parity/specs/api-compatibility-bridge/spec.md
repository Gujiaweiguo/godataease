## ADDED Requirements

### Requirement: Interactive Tree Compatibility Status Promotion
The compatibility bridge SHALL treat `dataVisualization/interactiveTree` as a governed parity endpoint once real resource-tree behavior is implemented.

#### Scenario: Promote interactive tree from partial to full
- **WHEN** interactive tree behavior is backed by implementation and regression evidence
- **THEN** governed whitelist or matrix metadata MUST be updated from `partial` to `full`
- **AND** the metadata update MUST reference the evidence proving parity scope completion

#### Scenario: Block governance promotion without parity evidence
- **WHEN** interactive tree still depends on synthetic placeholders or lacks regression evidence
- **THEN** governance metadata MUST remain non-full
- **AND** release documentation MUST NOT claim complete parity for the endpoint
