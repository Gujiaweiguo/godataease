## ADDED Requirements

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
