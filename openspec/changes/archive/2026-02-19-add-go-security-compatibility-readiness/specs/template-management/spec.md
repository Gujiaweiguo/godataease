## ADDED Requirements

### Requirement: Template Migration Route Alias Integrity
The Go backend SHALL preserve migration-safe route alias integrity for template capabilities.

#### Scenario: Canonical and Java alias routes return equivalent behavior
- **WHEN** client invokes canonical Go template routes and Java-compatible alias routes for the same operation
- **THEN** both routes return equivalent status, `code`, payload semantics, and permission outcomes
- **AND** no alias route bypasses authorization checks
