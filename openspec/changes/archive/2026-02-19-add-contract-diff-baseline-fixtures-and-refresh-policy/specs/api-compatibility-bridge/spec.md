## ADDED Requirements

### Requirement: Interface-Level Baseline Fixture Governance
The contract diff system SHALL manage baseline fixtures at interface level with deterministic structure, naming, and metadata.

#### Scenario: Resolve baseline fixture by API identity
- **WHEN** a contract diff run targets a specific API (`path + method`)
- **THEN** the system MUST resolve exactly one baseline fixture by deterministic path and naming rules
- **AND** the resolved fixture MUST contain required metadata for version and ownership traceability

#### Scenario: Reject malformed baseline fixtures
- **WHEN** a baseline fixture is missing required fields or has invalid metadata
- **THEN** the validation step MUST fail with deterministic configuration errors
- **AND** the gate MUST stop before performing parity comparison

### Requirement: Incremental Baseline Refresh Workflow
The system SHALL support incremental baseline refresh with dry-run preview and controlled apply mode.

#### Scenario: Preview incremental changes in dry-run mode
- **WHEN** refresh is executed in dry-run mode
- **THEN** no baseline files MUST be modified
- **AND** a diff preview report MUST be generated for review

#### Scenario: Apply incremental refresh for scoped APIs only
- **WHEN** refresh is executed in apply mode with a scoped API set
- **THEN** only targeted baseline fixtures MUST be updated
- **AND** non-target baseline fixtures MUST remain unchanged

### Requirement: Review, Rollback, and Drift Control Policy
The baseline mechanism SHALL enforce review gates, rollback capability, and drift safeguards to reduce false positives/negatives.

#### Scenario: Enforce review before baseline adoption
- **WHEN** baseline fixture changes are proposed
- **THEN** required reviewers (owner/tech lead/QA) MUST approve before adoption
- **AND** unapproved baseline updates MUST NOT be used to bypass gate failures

#### Scenario: Roll back to last stable baseline
- **WHEN** baseline updates cause regression or unacceptable drift
- **THEN** the system MUST support rollback to the last stable baseline version
- **AND** rollback evidence (trigger, operator, result) MUST be archived for audit
