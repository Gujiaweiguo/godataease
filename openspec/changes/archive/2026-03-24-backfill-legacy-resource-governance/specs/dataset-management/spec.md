## ADDED Requirements

### Requirement: Historical Datasets Can Be Backfilled Into Governed Resource Identity
The system SHALL support backfilling historical dataset records into the governed resource identity model.

#### Scenario: Backfill a historical dataset with valid governing scope
- **WHEN** a historical dataset can be mapped to a valid organization boundary and governed parent scope
- **THEN** the system MUST register a stable governed resource identity for that dataset
- **AND** repeated backfill execution MUST remain idempotent

#### Scenario: Historical dataset reuses existing governed identity on repeated backfill
- **WHEN** the backfill process runs again for a dataset that has already been governed
- **THEN** the system MUST reuse the existing governed resource identity
- **AND** the process MUST NOT create duplicate governed entries or duplicate inherited grants

#### Scenario: Historical dataset cannot be mapped safely
- **WHEN** a historical dataset depends on missing, orphaned, or invalid upstream governance context
- **THEN** the system MUST skip automatic backfill
- **AND** the skip outcome MUST be traceable for later remediation
- **AND** the dataset MUST NOT be represented as fully governed when required scope information is absent

### Requirement: Backfilled Dataset Authorization Must Remain Consistent Across Permission and Dependency Flows
The system SHALL ensure that historical datasets backfilled into the governed model resolve the same effective authorization state across permission-center views and dataset runtime flows.

#### Scenario: Backfilled dataset preserves permission and dependency consistency
- **WHEN** a historical dataset is governed after backfill
- **THEN** dataset browse, field, preview, and permission-center views MUST remain consistent with the same effective authorization outcome
- **AND** operators MUST be able to distinguish authorization failure from missing dependency or missing resource

#### Scenario: Backfilled dataset inherits governed parent permission
- **WHEN** a historical dataset is backfilled beneath a governed parent group that already has effective grants
- **THEN** the dataset MUST inherit those grants through the governed authorization model
- **AND** permission queries MUST return the inherited result without requiring manual re-authorization

#### Scenario: Backfilled dataset respects datasource dependency boundary
- **WHEN** a user accesses a backfilled dataset whose dependent datasource is not authorized or not governable
- **THEN** the system MUST preserve explicit dependency-denied or authorization-denied semantics
- **AND** the result MUST remain distinguishable from a missing dataset resource
