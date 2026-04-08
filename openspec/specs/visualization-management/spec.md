# visualization-management Specification

## Purpose
Define data visualization lifecycle requirements for listing, saving, updating, and loading visualization canvases.
## Requirements
### Requirement: Visualization Core CRUD
The system SHALL provide visualization core CRUD capability in Go backend, and in-scope visualization write operations MUST preserve governed authorization semantics before mutating dashboard or big-screen resources.

#### Scenario: Create or update visualization
- **WHEN** client submits visualization definition payload
- **THEN** the system persists visualization metadata and content
- **AND** returns success with Java-compatible response envelope

#### Scenario: Query visualization detail
- **WHEN** client requests visualization detail by identifier
- **THEN** the system returns complete visualization definition for rendering

#### Scenario: Governed visualization write succeeds with sufficient permission
- **WHEN** an authenticated user invokes an in-scope visualization write route with the required governed authorization on the target visualization resource or governed parent scope
- **THEN** the backend MUST allow the mutation to proceed through the normal visualization service flow
- **AND** the resulting write behavior MUST remain compatible with existing visualization CRUD contracts

#### Scenario: Governed visualization write remains explicit when permission is insufficient
- **WHEN** a user invokes an in-scope visualization write route without sufficient governed authorization on the target resource or parent scope
- **THEN** the backend MUST return explicit authorization-denied semantics
- **AND** the failure MUST remain distinguishable from malformed payload, missing visualization resource, or generic route failure

#### Scenario: Visualization write route cannot resolve a safe authorization target
- **WHEN** an in-scope visualization write request does not provide the governed resource or parent information required to authorize that write safely
- **THEN** the backend MUST fail closed before mutating visualization state
- **AND** the system MUST NOT continue into an Auth-only write path as a fallback

#### Scenario: Remaining root visualization mutation route succeeds with sufficient permission
- **WHEN** an authenticated user invokes `/dataVisualization/updateBase`, `/dataVisualization/move`, `/dataVisualization/updatePublishStatus`, or `/dataVisualization/recoverToPublished` with sufficient governed authorization on the target visualization resource
- **THEN** the backend MUST allow the mutation to proceed through the normal visualization service flow
- **AND** the resulting write behavior MUST remain compatible with existing visualization CRUD contracts

#### Scenario: Remaining root visualization mutation route remains explicit when permission is insufficient
- **WHEN** a user invokes one of the remaining in-scope root visualization mutation routes without sufficient governed authorization on the target visualization resource
- **THEN** the backend MUST return explicit authorization-denied semantics
- **AND** the failure MUST remain distinguishable from malformed payload, missing visualization resource, or generic route failure

#### Scenario: Remaining root visualization mutation route fails closed when the governed target cannot be resolved
- **WHEN** one of the remaining in-scope root visualization mutation routes cannot provide the governed visualization resource information required for safe authorization
- **THEN** the backend MUST fail closed before mutating visualization state
- **AND** the system MUST NOT continue into an Auth-only write path as a fallback

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

### Requirement: Visualization Entry-Chain Recovery
The system SHALL keep dashboard and big-screen entry chains recoverable as a governed stabilization surface.

#### Scenario: Visualization entry path reaches usable page state
- **WHEN** a user enters dashboard or big-screen flows from a governed menu or route entry
- **THEN** the page MUST reach a usable initialized state for in-scope list, tree, or detail workflows
- **AND** broken route or discovery behavior MUST be classified explicitly instead of appearing as generic feature absence

#### Scenario: Visualization recovery preserves discovery-path integrity
- **WHEN** a recovery fix is applied to dashboard or big-screen discovery flows
- **THEN** tree/detail/resource-discovery payloads MUST remain consumable by the frontend path that triggered the flow
- **AND** the recovered flow MUST have targeted regression or smoke coverage

### Requirement: Visualization Detail Hardening After Recovery
The system SHALL preserve explicit detail-path semantics for dashboard and big-screen flows after the primary recovery batch is complete.

#### Scenario: Dashboard detail missing-resource behavior stays explicit at the boundary
- **WHEN** a dashboard detail request targets a missing resource
- **THEN** the frontend-facing boundary MUST preserve an explicit missing-resource response
- **AND** the response MUST remain distinguishable from permission denial

#### Scenario: Big-screen deeper detail paths remain consumable after hardening
- **WHEN** a big-screen detail or edit path is exercised beyond preview-only coverage
- **THEN** the route and detail payload MUST remain consumable by the intended frontend path
- **AND** failures MUST remain explicit instead of degrading into generic feature absence

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
