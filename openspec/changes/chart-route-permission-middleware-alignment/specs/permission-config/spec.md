## MODIFIED Requirements

### Requirement: Row and Column Permissions Must Enforce Governed Runtime Behavior
The system MUST treat row filters, disabled columns, and masked columns as runtime-enforced governed behavior, MUST enforce row-permission gating consistently at the middleware/runtime boundary for governed dataset and chart entry points, MUST keep chart runtime authorization dataset-governed instead of introducing a separate chart resource permission model, and MUST explicitly trace or defer whitelist and system-variable dimensions until they are implemented.

#### Scenario: Governed data-access rule is evaluated at runtime
- **WHEN** a governed dataset request is evaluated against row or column permissions
- **THEN** the system MUST apply the configured enforcement semantics instead of relying on placeholder middleware or UI-only interpretation
- **AND** whitelist and system-variable dimensions MUST be either traceable in the governed permission model or explicitly marked as deferred instead of being silently accepted

#### Scenario: Middleware-enforced governed route establishes row-permission context
- **WHEN** a request enters a governed runtime route that is protected by row-permission middleware
- **THEN** the middleware MUST resolve and validate the runtime row-permission context needed for downstream evaluation before the handler continues
- **AND** the request MUST NOT proceed as a warning-only no-op once middleware enforcement is enabled

#### Scenario: Middleware fails closed when governed context cannot be established
- **WHEN** row-permission middleware cannot safely determine the governed dataset/group context, authenticated user context, or permission lookup prerequisites for a protected route
- **THEN** the request MUST terminate with explicit permission/error semantics
- **AND** the system MUST NOT continue into a permissive service path that could bypass governed enforcement

#### Scenario: Service-layer rule application remains authoritative after middleware rollout
- **WHEN** a governed runtime route passes through row-permission middleware and reaches the dataset/chart service layer
- **THEN** the service layer MUST remain the source of truth for select-column shaping, WHERE-clause construction, disabled-column filtering, and masking behavior
- **AND** middleware MUST NOT duplicate SQL rule compilation logic that already belongs to the row/column permission services

#### Scenario: Middleware rollout does not implicitly govern non-governed routes
- **WHEN** maintainers enable real row-permission middleware enforcement
- **THEN** only explicitly governed runtime routes MUST adopt that middleware behavior in the rollout scope
- **AND** routes outside the governed permission-aware runtime surface MUST NOT gain row-permission enforcement implicitly as an accidental side effect

#### Scenario: Governed chart runtime route resolves dataset-governed permission context
- **WHEN** a governed chart runtime request enters a canonical or compatibility chart data route that only identifies the chart
- **THEN** the system MUST resolve the backing `datasetGroupID` before permission-aware execution continues
- **AND** dataset view permission and downstream row/column enforcement MUST evaluate against that resolved dataset-governed identity instead of a separate chart resource permission model

#### Scenario: Governed chart runtime route fails closed on chart context resolution errors
- **WHEN** a governed chart runtime route cannot resolve the backing dataset group from the provided chart identity, or the resolved dataset-governed context cannot be validated
- **THEN** the request MUST terminate with explicit authorization or error semantics before chart data execution proceeds
- **AND** the system MUST NOT fall back to permissive chart execution because chart context resolution failed

#### Scenario: Compatibility chart permission flow does not synthesize admin identity
- **WHEN** an in-scope compatibility chart permission flow is invoked without authenticated user context
- **THEN** the system MUST return fail-closed unauthorized or permission-denied semantics
- **AND** the flow MUST NOT recover by substituting a default admin user ID or username

#### Scenario: Governed chart field listing stays permission-aware on governed runtime routes
- **WHEN** a governed chart runtime route exposes chart field or dataset-field results through permission-aware chart listing behavior
- **THEN** disabled-column and masking behavior MUST remain consistent with dataset-governed row/column permission semantics
- **AND** the route MUST NOT silently downgrade to non-permission-aware listing behavior solely because it entered through a compatibility path

#### Scenario: Deferred whitelist write attempts fail with stable contract language
- **WHEN** a client submits a non-empty `whiteList` value through the governed row-permission save workflow
- **THEN** the backend MUST reject the request with a stable deferred-semantics error that does not reference internal milestone labels
- **AND** the response MUST make it clear that whitelist persistence/editing is deferred in the permission center rather than partially supported

#### Scenario: Deferred whitelist read contract stays compatibility-safe
- **WHEN** a client reads governed row-permission data from the unified permission-center workflow before whitelist persistence is implemented
- **THEN** any exposed whitelist-related fields MUST remain compatibility-safe and clearly treated as deferred contract surface rather than active persisted state
- **AND** the system MUST NOT populate those fields with misleading non-empty data that implies supported whitelist persistence

#### Scenario: Unified permission center does not offer governed whitelist editing
- **WHEN** an administrator uses the unified permission-center row/column permission workflow
- **THEN** the UI MUST NOT expose a governed whitelist editing path for row permissions
- **AND** any legacy or adjacent flows outside the unified permission center MUST be treated as explicit deferred boundaries rather than evidence that whitelist editing is supported there
