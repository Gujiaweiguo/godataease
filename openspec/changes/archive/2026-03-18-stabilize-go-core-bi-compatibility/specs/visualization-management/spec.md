## ADDED Requirements

### Requirement: Dashboard and Big-Screen Critical Flow Stability
The system SHALL treat dashboard and big-screen load, tree, and resource-management APIs as release-blocking surfaces for the migrated core BI flow.

#### Scenario: Visualization resource tree stays routable for managed flows
- **WHEN** frontend requests dashboard or big-screen resource trees through governed compatibility or canonical paths
- **THEN** all in-scope routes MUST be reachable without `404`
- **AND** each response MUST contain the node structure required for downstream folder, leaf, and authorization-aware operations

#### Scenario: Visualization detail remains consumable by migrated clients
- **WHEN** a dashboard or big-screen detail request succeeds
- **THEN** the system MUST return a complete visualization definition consumable by existing rendering clients
- **AND** MUST NOT replace missing business data with placeholder success semantics

### Requirement: Resource Operation Precheck Stability
Visualization resource-tree payloads SHALL remain valid for copy, move, delete, and selection preparation flows.

#### Scenario: Managed resource operation receives valid tree payload
- **WHEN** frontend prepares a dashboard or big-screen resource operation from returned tree data
- **THEN** required identifiers, node typing, and pathing fields MUST be present and internally consistent
- **AND** malformed payloads MUST return explicit non-success semantics instead of causing downstream silent failure
