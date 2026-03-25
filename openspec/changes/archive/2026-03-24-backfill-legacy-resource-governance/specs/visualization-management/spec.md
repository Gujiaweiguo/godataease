## ADDED Requirements

### Requirement: Historical Visualization Resources Can Be Backfilled Into Governed Resource Identity
The system SHALL support backfilling historical dashboard and big-screen resources into the governed resource identity model.

#### Scenario: Backfill a historical dashboard or big-screen resource with valid governing scope
- **WHEN** a historical dashboard or big-screen resource can be mapped to a valid organization boundary and governed parent scope
- **THEN** the system MUST register a stable governed resource identity for that resource
- **AND** repeated backfill execution MUST NOT create duplicate governed entries

#### Scenario: Historical visualization resource reuses existing governed identity on repeated backfill
- **WHEN** the backfill process runs again for a visualization resource that has already been governed
- **THEN** the system MUST reuse the existing governed resource identity
- **AND** the process MUST NOT create duplicate inherited grants or duplicate authorization mappings

#### Scenario: Historical visualization resource lacks safe governing scope
- **WHEN** a historical dashboard or big-screen resource cannot be mapped safely to a governed parent or organization scope
- **THEN** the system MUST skip automatic backfill
- **AND** the skip outcome MUST be auditable for follow-up remediation
- **AND** the resource MUST NOT be represented as fully governed in unified permission workflows

### Requirement: Backfilled Visualization Authorization Must Align With Permission-Center Results
The system SHALL ensure that backfilled dashboard and big-screen resources share the same effective authorization state across runtime access and unified permission-center views.

#### Scenario: Backfilled dashboard authorization matches governed resource view
- **WHEN** a historical dashboard has been backfilled into the governed model
- **THEN** by-resource and by-user permission views MUST return the same effective grants seen by runtime authorization checks

#### Scenario: Backfilled big-screen authorization matches governed resource view
- **WHEN** a historical big-screen resource has been backfilled into the governed model
- **THEN** by-resource and by-user permission views MUST return the same effective grants seen by runtime authorization checks

#### Scenario: Backfilled visualization inherits governed parent permission
- **WHEN** a historical dashboard or big-screen resource is backfilled beneath a governed parent group that already has effective grants
- **THEN** the resource MUST inherit those grants through the governed authorization model
- **AND** tree queries, detail queries, and permission-center resource queries MUST expose the inherited effective result consistently

#### Scenario: Backfilled visualization denial remains distinguishable from missing resource
- **WHEN** a user accesses a backfilled dashboard or big-screen resource without sufficient permission
- **THEN** the system MUST return authorization-denied semantics
- **AND** the result MUST remain distinguishable from missing visualization resource or missing route behavior
