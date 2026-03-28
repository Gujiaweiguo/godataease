## ADDED Requirements

### Requirement: Historical Datasources Can Be Backfilled Into Governed Resource Identity
The system SHALL support backfilling historical datasource records into the governed resource identity model used by unified permission management.

#### Scenario: Backfill a historical datasource with valid governing scope
- **WHEN** a historical datasource has a valid organization boundary and parent governing scope
- **THEN** the system MUST register a stable governed resource identity for that datasource
- **AND** repeated backfill execution MUST NOT create duplicate governed resource records

#### Scenario: Historical datasource keeps stable identity across repeated backfill runs
- **WHEN** the backfill process is re-executed for a datasource that has already been governed
- **THEN** the system MUST reuse the existing governed resource identity
- **AND** the process MUST NOT create duplicate inheritance state or duplicate authorization mappings

#### Scenario: Historical datasource lacks safe governing scope
- **WHEN** a historical datasource cannot be mapped safely to a governed parent or organization scope
- **THEN** the system MUST skip automatic registration
- **AND** the result MUST be recorded for remediation
- **AND** the datasource MUST NOT be represented as fully governed in unified permission workflows

### Requirement: Backfilled Datasource Authorization Must Match Runtime Semantics
The system SHALL ensure that a backfilled historical datasource uses the same effective authorization semantics as newly governed datasource resources.

#### Scenario: Backfilled datasource appears consistently in permission and runtime flows
- **WHEN** a historical datasource has been backfilled successfully
- **THEN** unified permission queries, user/resource perspectives, and datasource runtime authorization checks MUST resolve against the same governed authorization state

#### Scenario: Backfilled datasource inherits governed parent permission
- **WHEN** a historical datasource is backfilled beneath a parent resource group that already has effective grants
- **THEN** the datasource MUST inherit those grants through the governed authorization model
- **AND** permission queries MUST return the inherited result without requiring manual re-authorization

#### Scenario: Backfilled datasource denial remains distinguishable
- **WHEN** a user accesses a backfilled datasource without sufficient permission
- **THEN** the system MUST return authorization-denied semantics
- **AND** the result MUST remain distinguishable from missing datasource or missing route behavior
