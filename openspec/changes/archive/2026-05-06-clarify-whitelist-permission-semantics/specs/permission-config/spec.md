## MODIFIED Requirements

### Requirement: Row and Column Permissions Must Enforce Governed Runtime Behavior
The system MUST treat row filters, disabled columns, and masked columns as runtime-enforced governed behavior, and MUST explicitly trace or defer whitelist and system-variable dimensions until they are implemented.

#### Scenario: Governed data-access rule is evaluated at runtime
- **WHEN** a governed dataset request is evaluated against row or column permissions
- **THEN** the system MUST apply the configured enforcement semantics instead of relying on placeholder middleware or UI-only interpretation
- **AND** whitelist and system-variable dimensions MUST be either traceable in the governed permission model or explicitly marked as deferred instead of being silently accepted

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
