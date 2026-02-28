## ADDED Requirements

### Requirement: Visualization Resource Tree Compatibility API
The system SHALL provide visualization tree API compatibility for dashboard and screen resource management workflows.

#### Scenario: Query visualization resource tree
- **WHEN** frontend requests `POST /dataVisualization/tree` through compatibility base path
- **THEN** backend MUST return tree data compatible with resource-tree consumers
- **AND** endpoint MUST be reachable without `404`

### Requirement: Tree Response Contract Consistency
Visualization tree responses SHALL remain structurally consistent with frontend consumers that perform folder/leaf operations.

#### Scenario: Use tree payload in dashboard resource operations
- **WHEN** frontend performs copy/move/delete preparation with tree response
- **THEN** response nodes MUST include required fields for pathing and authorization decisions
- **AND** malformed or missing required structure MUST return explicit non-success response
