## MODIFIED Requirements

### Requirement: Java Route Compatibility Bridge
The Go backend SHALL provide compatibility route mappings for Java-era API prefixes during migration.

#### Scenario: Call Java-style route prefix
- **WHEN** a client calls a Java-style route that is listed in the migration compatibility matrix
- **THEN** the Go backend routes the request to the equivalent Go capability
- **AND** the business behavior matches the canonical Go route behavior
- **AND** the route MUST belong to an explicit governance bucket: permanent shim, frontend migration target, or dual-support transition path

#### Scenario: `/user/org/option` keeps user-option semantics regardless of optional org handler availability
- **WHEN** the compatibility bridge registers `/user/org/option` under the `/user` route group
- **THEN** the endpoint MUST resolve to user-option behavior compatible with `user.GetUserOptions`
- **AND** it MUST NOT switch to organization-list semantics based on whether an org handler instance is injected

### Requirement: Alias Consistency Verification
The migration process SHALL verify functional parity between canonical and compatibility routes.

#### Scenario: Verify alias parity for governed IAM routes
- **WHEN** regression suite executes for both canonical Go route and compatibility alias in the governed IAM surface
- **THEN** both routes MUST return equivalent status, code, and payload semantics
- **AND** parity verification MUST respect the route-family governance bucket so that temporary dual-support aliases remain traceable until retirement
