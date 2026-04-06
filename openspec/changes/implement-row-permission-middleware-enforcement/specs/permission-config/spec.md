## MODIFIED Requirements

### Requirement: Row and Column Permissions Must Enforce Governed Runtime Behavior
The system MUST treat row filters, disabled columns, and masked columns as runtime-enforced governed behavior, MUST enforce row-permission gating consistently at the middleware/runtime boundary for governed entry points, and MUST explicitly trace or defer whitelist and system-variable dimensions until they are implemented.

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
