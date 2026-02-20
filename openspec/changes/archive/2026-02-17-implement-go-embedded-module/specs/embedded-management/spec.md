## ADDED Requirements

### Requirement: Embedded Application Management
The system SHALL provide embedded application management functionality for third-party integration.

#### Scenario: Create embedded application
- **WHEN** admin creates an embedded application with name and domain
- **THEN** the system generates unique appId and appSecret
- **AND** returns success with code 000000

#### Scenario: Query embedded applications with pagination
- **WHEN** admin queries embedded application list with page and pageSize
- **THEN** the system returns paginated results with masked appSecret
- **AND** supports keyword filtering by name

#### Scenario: Update embedded application
- **WHEN** admin updates an embedded application's name or domain
- **THEN** the system updates the record
- **AND** returns success with code 000000

#### Scenario: Delete embedded application
- **WHEN** admin deletes an embedded application by id
- **THEN** the system removes the record
- **AND** returns success with code 000000

#### Scenario: Batch delete embedded applications
- **WHEN** admin provides a list of ids to delete
- **THEN** the system removes all specified records
- **AND** returns success with code 000000

### Requirement: Embedded Secret Management
The system SHALL provide secure secret management for embedded applications.

#### Scenario: Reset application secret
- **WHEN** admin requests to reset an application's secret
- **THEN** the system generates a new secret with configured length
- **OR** uses the provided custom secret
- **AND** updates the record

#### Scenario: Mask secret display
- **WHEN** returning application details in list view
- **THEN** the system masks the secret showing only first 4 and last 4 characters
- **AND** shows ******** for secrets with length <= 8

### Requirement: Embedded Token Generation
The system SHALL provide JWT token generation for embedded authentication.

#### Scenario: Generate embedded token
- **WHEN** system generates a token for embedded access
- **THEN** the token contains uid, oid, appId, and exp claims
- **AND** uses HMAC-SHA256 algorithm
- **AND** expires after 24 hours

#### Scenario: Validate embedded token
- **WHEN** system validates an embedded token
- **THEN** the system verifies signature using appSecret
- **AND** extracts user and organization information
- **AND** returns validation result

### Requirement: Domain Whitelist Validation
The system SHALL validate embedded access origins against configured domains.

#### Scenario: Parse domain list
- **WHEN** system parses domain configuration
- **THEN** splits by comma, semicolon, or whitespace
- **AND** removes trailing slashes
- **AND** returns normalized domain list

#### Scenario: Validate origin against whitelist
- **WHEN** embedded access request comes from an origin
- **THEN** the system checks if origin matches any allowed domain
- **AND** supports both full URL and hostname matching
- **AND** rejects requests from unauthorized origins

### Requirement: Embedded API Endpoints
The system SHALL provide the following API endpoints for embedded management.

#### Scenario: Pager endpoint
- **WHEN** POST /api/embedded/pager/{goPage}/{pageSize} is called
- **THEN** returns paginated list with keyword filter support

#### Scenario: Create endpoint
- **WHEN** POST /api/embedded/create is called
- **THEN** creates new embedded application with generated credentials

#### Scenario: Edit endpoint
- **WHEN** POST /api/embedded/edit is called
- **THEN** updates existing embedded application

#### Scenario: Delete endpoint
- **WHEN** POST /api/embedded/delete/{id} is called
- **THEN** deletes specified embedded application

#### Scenario: Batch delete endpoint
- **WHEN** POST /api/embedded/batchDelete is called
- **THEN** deletes all specified embedded applications

#### Scenario: Reset secret endpoint
- **WHEN** POST /api/embedded/reset is called
- **THEN** resets application secret

#### Scenario: Domain list endpoint
- **WHEN** GET /api/embedded/domainList is called
- **THEN** returns distinct list of all configured domains

#### Scenario: Init iframe endpoint
- **WHEN** POST /api/embedded/initIframe is called with token
- **THEN** validates token and origin
- **AND** returns allowed domains for the application

#### Scenario: Get token args endpoint
- **WHEN** GET /api/embedded/getTokenArgs is called
- **THEN** returns current user's userId and orgId

#### Scenario: Limit count endpoint
- **WHEN** GET /api/embedded/limitCount is called
- **THEN** returns the maximum number of embedded applications allowed
