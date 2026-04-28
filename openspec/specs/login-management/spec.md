# login-management Specification

## Purpose
Define login and authentication requirements for credential validation, token issuance, and session safety.
## Requirements
### Requirement: Local Login
The system SHALL provide local login functionality for user authentication.

#### Scenario: Successful login
- **WHEN** admin user logs in with correct credentials
- **THEN** the system generates a JWT token
- **AND** returns token and expiration time

#### Scenario: Failed login - wrong username
- **WHEN** user logs in with non-admin username
- **THEN** the system returns error "仅admin账号可用"

#### Scenario: Failed login - wrong password
- **WHEN** admin user logs in with wrong password
- **THEN** the system returns error "用户名或密码错误"

### Requirement: JWT Token Generation
The system SHALL generate JWT tokens compatible with Java version.

#### Scenario: Token format
- **WHEN** generating token
- **THEN** uses HMAC-SHA256 algorithm
- **AND** includes uid and oid claims
- **AND** uses MD5(password) as secret key

### Requirement: Logout
The system SHALL provide logout functionality.

#### Scenario: Logout
- **WHEN** user calls logout endpoint
- **THEN** returns success (no actual token invalidation in this version)

### Requirement: Login API Endpoints
The system SHALL provide the following API endpoints.

#### Scenario: Local login endpoint
- **WHEN** POST /login/localLogin is called
- **THEN** validates credentials
- **AND** returns TokenVO on success

#### Scenario: Logout endpoint
- **WHEN** GET /logout is called
- **THEN** returns success

### Requirement: Local Login Must Enforce Burst-Safe Request Limits
The system SHALL enforce request-throttling on local login endpoints so repeated login attempts cannot be performed at unbounded frequency.

#### Scenario: Local login remains available within request budget
- **WHEN** a client calls `POST /login/localLogin` or `POST /api/login/localLogin` within the configured login request budget
- **THEN** the system MUST continue to validate credentials using the existing login semantics
- **AND** successful and failed credential outcomes MUST preserve their current response behavior

#### Scenario: Local login exceeds request budget
- **WHEN** a client exceeds the configured request budget for local login attempts within the active rate-limit window
- **THEN** the system MUST reject the login attempt with an explicit throttling response
- **AND** the system MUST NOT continue into credential validation for the throttled request

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

### Requirement: Login Bootstrap Must Restore Authorized Runtime State
The system SHALL restore current-user, authorized menu, and dynamic route state after successful login so core feature domains remain reachable.

#### Scenario: Successful login restores authorized routes
- **WHEN** a user logs in successfully
- **THEN** the frontend MUST initialize current-user state, authorized menu state, and dynamic routes before protected feature navigation is considered healthy
- **AND** the user MUST NOT land in a false missing-feature state caused by incomplete bootstrap

#### Scenario: Login redirect targets protected page
- **WHEN** login completes with a protected redirect target
- **THEN** the system MUST resolve the redirect only after authorization-driven route setup is ready
- **AND** the user MUST NOT be sent to a misleading `401` or `404` caused by bootstrap race or missing dynamic route registration
