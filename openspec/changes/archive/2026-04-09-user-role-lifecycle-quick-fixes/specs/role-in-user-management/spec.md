## Delta: role-in-user-management

### Added Scenario: Role member management APIs use correct canonical paths

- **WHEN** the RoleTab member management dialog queries available roles or selected roles for a user
- **THEN** the frontend MUST call `/role/user/option` and `/role/user/selected` respectively
- **AND** the path segments MUST match the backend canonical route registration order (`/role/user/...` not `/user/role/...`)
