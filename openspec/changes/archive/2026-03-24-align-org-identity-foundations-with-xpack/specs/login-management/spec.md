## ADDED Requirements

### Requirement: Login Bootstrap Returns Organization-Aware Identity Context
The system SHALL return an organization-aware identity bootstrap that allows the frontend to initialize user, organization, and authorized runtime state without reconstructing identity semantics client-side.

#### Scenario: Successful login returns identity bootstrap
- **WHEN** a user completes a successful login flow
- **THEN** the bootstrap response MUST include stable current-user identity fields required by frontend state initialization
- **AND** the bootstrap response MUST include current organization context or the data needed to resolve it immediately

#### Scenario: Current user info refresh preserves organization context
- **WHEN** the client refreshes current-user runtime information after login
- **THEN** the system MUST return the same organization-aware identity semantics as the login bootstrap contract
- **AND** the frontend MUST NOT need ad hoc fallback requests to reconstruct organization scope
