## MODIFIED Requirements

### Requirement: Java Route Compatibility Bridge
The Go backend SHALL provide compatibility route mappings for Java-era API prefixes during migration.

#### Scenario: `/user/org/option` keeps user-option semantics regardless of optional org handler availability
- **WHEN** the compatibility bridge registers `/user/org/option` under the `/user` route group
- **THEN** the endpoint MUST resolve to user-option behavior compatible with `user.GetUserOptions`
- **AND** it MUST NOT switch to organization-list semantics based on whether an org handler instance is injected
