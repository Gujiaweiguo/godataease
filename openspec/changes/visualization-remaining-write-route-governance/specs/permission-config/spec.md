## MODIFIED Requirements

### Requirement: Visualization Write Routes Must Enforce Governed Resource Permission
The system MUST enforce governed resource-level permission checks on in-scope visualization write routes, MUST resolve dashboard and big-screen writes against the correct governed resource type, and MUST fail closed when the authorization target for a selected write route cannot be established safely.

#### Scenario: Remaining root visualization mutation routes require governed edit permission
- **WHEN** an authenticated user invokes `/dataVisualization/updateBase`, `/dataVisualization/move`, `/dataVisualization/updatePublishStatus`, or `/dataVisualization/recoverToPublished`
- **THEN** the system MUST resolve the governed visualization target before the handler mutates state
- **AND** the route MUST deny the request if the user lacks governed edit permission on that dashboard or big-screen resource

#### Scenario: Remaining root visualization mutation routes do not stay behind Auth-only protection
- **WHEN** one of the remaining in-scope root visualization mutation routes is reachable through the root legacy path only
- **THEN** that route MUST still enforce governed authorization semantics rather than treating successful authentication as sufficient
- **AND** adjacent visualization list, tree, and helper routes MUST NOT be implicitly broadened by this rollout
