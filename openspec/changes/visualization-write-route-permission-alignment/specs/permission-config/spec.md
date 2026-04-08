## ADDED Requirements

### Requirement: Visualization Write Routes Must Enforce Governed Resource Permission
The system MUST enforce governed resource-level permission checks on in-scope visualization write routes, MUST resolve dashboard and big-screen writes against the correct governed resource type, and MUST fail closed when the authorization target for a selected write route cannot be established safely.

#### Scenario: Existing visualization write route requires edit permission on the governed resource
- **WHEN** an authenticated user invokes an in-scope visualization write route that targets an existing visualization resource such as `updateBase`, `move`, `updatePublishStatus`, or `recoverToPublished`
- **THEN** the system MUST resolve the target visualization resource before the handler mutates state
- **AND** the route MUST deny the request if the user lacks governed edit permission on that dashboard or big-screen resource

#### Scenario: Parent-scoped visualization creation requires permission on the governed parent
- **WHEN** an authenticated user invokes an in-scope parent-scoped visualization creation route such as `saveCanvas` or its legacy-compatible alias with a positive governed `pid`
- **THEN** the system MUST authorize the request against the parent governed resource scope before creating the new visualization
- **AND** the route MUST fail closed instead of treating Auth-only access as sufficient

#### Scenario: Visualization write gate respects dashboard and big-screen resource types
- **WHEN** an in-scope visualization write request identifies a `dashboard` target or a `dataV`/big-screen target
- **THEN** the system MUST evaluate permission against `dashboard` resources for dashboard writes
- **AND** the system MUST evaluate permission against `screen` resources for big-screen writes rather than collapsing both into one resource type

#### Scenario: Visualization write aliases do not remain weaker than canonical routes
- **WHEN** the same in-scope visualization write operation is reachable through both canonical `/api/dataVisualization/*` and legacy-compatible `/dataVisualization/*` paths
- **THEN** both entry points MUST enforce equivalent governed authorization semantics
- **AND** the system MUST NOT leave the legacy-compatible write path behind Auth-only protection once the canonical route is governed
