## Delta: audit-go

### Added Scenario: UserService audit dependency is wired at startup

- **WHEN** the application starts and initializes the UserService
- **THEN** the UserService MUST have its audit service dependency injected via SetAuditService
- **AND** password reset operations MUST produce audit log entries in de_audit_log
