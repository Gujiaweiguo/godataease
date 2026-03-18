## ADDED Requirements

### Requirement: BI Permission Failure Semantic Stability
The system SHALL keep permission outcomes for core BI flows semantically distinguishable across datasource, dataset, dashboard, and big-screen routes during migration.

#### Scenario: Existing BI resource denied by permission
- **WHEN** an authenticated user accesses an existing BI resource or API without sufficient permission
- **THEN** the system MUST return authorization-denied semantics consistent with the governed permission model
- **AND** MUST NOT convert the result into a generic `404` or placeholder success response

#### Scenario: Missing BI resource remains distinguishable from denied access
- **WHEN** a requested BI resource is absent or not registered
- **THEN** the system MAY return missing-resource semantics
- **AND** operators and clients MUST be able to distinguish that result from permission denial during debugging and support

### Requirement: Permission-Aware Resource Tree Responses
Permission-filtered BI resource trees SHALL remain structurally valid for migrated frontend consumers.

#### Scenario: Filtered tree preserves required operation fields
- **WHEN** authorization filtering removes unauthorized dashboards, screens, datasets, or datasource-linked nodes from a returned tree
- **THEN** the remaining tree MUST preserve required identifiers, hierarchy, and node type fields for supported operations
- **AND** permission filtering MUST NOT produce malformed tree payloads that later fail in copy, move, selection, or preview preparation flows
