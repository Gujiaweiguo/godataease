## Delta: user-management

### Added Scenario: Canonical default-password endpoint
- **WHEN** a client requests the user-management default password endpoint
- **THEN** the backend MUST provide `GET /system/user/defaultPwd` with the same response semantics as the compatibility alias

### Added Scenario: Canonical reset-password endpoint with audit middleware
- **WHEN** an authorized operator triggers reset password under user management
- **THEN** the backend MUST provide `POST /system/user/resetPwd/:id`
- **AND** the endpoint MUST execute audit middleware with user-action configuration

### Added Scenario: Frontend uses canonical reset-password routes
- **WHEN** frontend user-management code requests default password or triggers reset password
- **THEN** it MUST call `/system/user/defaultPwd` and `/system/user/resetPwd/:uid`
- **AND** it MUST NOT depend on `/user/defaultPwd` or `/user/resetPwd/:uid` as primary paths
