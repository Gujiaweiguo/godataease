## ADDED Requirements

### Requirement: Admin-Domain Compatibility Policy Must Distinguish Canonical and Legacy Contracts
The system MUST explicitly govern which admin-domain routes are canonical, which are compatibility aliases, and which are scheduled for retirement.

#### Scenario: Maintainer reviews an admin-domain route family
- **WHEN** user, role, organization, or permission endpoints are evaluated for alignment
- **THEN** the workflow MUST classify each route family as canonical, compatibility alias, or deprecated
- **AND** the classification MUST be reviewable before semantic parity is declared complete

### Requirement: Missing Compatibility Paths Must Be Classified Before Release
The system MUST classify missing or unimplemented compatibility paths as governed contract gaps.

#### Scenario: Frontend references legacy admin-domain path
- **WHEN** a governed frontend flow still references a legacy path with no working backend behavior
- **THEN** the system MUST classify the issue as a compatibility gap
- **AND** release readiness MUST depend on either a working alias, a frontend migration, or an approved retirement decision
