# visualization-management Specification

## Purpose
Define data visualization lifecycle requirements for listing, saving, updating, and loading visualization canvases.
## Requirements
### Requirement: Visualization Core CRUD
The system SHALL provide visualization core CRUD capability in Go backend.

#### Scenario: Create or update visualization
- **WHEN** client submits visualization definition payload
- **THEN** the system persists visualization metadata and content
- **AND** returns success with Java-compatible response envelope

#### Scenario: Query visualization detail
- **WHEN** client requests visualization detail by identifier
- **THEN** the system returns complete visualization definition for rendering

### Requirement: Visualization Listing
The system SHALL provide visualization list query capability.

#### Scenario: List visualizations
- **WHEN** client requests visualization list with workspace and keyword filters
- **THEN** the system returns paginated visualization summaries
- **AND** list ordering and pagination semantics remain stable across Java and Go implementations



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

### Requirement: Interactive Visualization Tree Parity
The system SHALL provide `dataVisualization/interactiveTree` as an authorization-filtered visualization resource tree for dashboard and big-screen discovery flows.

#### Scenario: Interactive tree returns real dashboard resources
- **WHEN** frontend requests `POST /api/dataVisualization/interactiveTree` with `dashboard` scope
- **THEN** the system MUST return real dashboard resource nodes derived from visualization data rather than synthetic authorization placeholders
- **AND** each returned node MUST preserve the frontend tree contract required by interactive consumers

#### Scenario: Interactive tree returns real big-screen resources
- **WHEN** frontend requests `POST /api/dataVisualization/interactiveTree` with `dataV` scope
- **THEN** the system MUST return real big-screen resource nodes derived from visualization data rather than synthetic authorization placeholders
- **AND** the node hierarchy MUST remain usable for downstream selection and navigation flows

### Requirement: Interactive Tree Authorization Filtering
Interactive visualization trees SHALL filter unauthorized visualization resources without breaking the remaining tree structure.

#### Scenario: Unauthorized visualization nodes are filtered safely
- **WHEN** a user lacks access to part of the dashboard or big-screen resource tree
- **THEN** unauthorized nodes MUST be excluded from the response
- **AND** remaining nodes MUST keep valid identifiers, parent relationships, and leaf semantics
