## Delta: watermark-management

### Added Scenario: Watermark settings page loads user person info
- **WHEN** watermark settings page requests user identity fields for watermark preview
- **THEN** backend MUST provide `/user/personInfo` with `id`, `account`, `name`, and `ip` fields

### Added Scenario: Runtime watermark resolver loads current IP info
- **WHEN** runtime watermark logic requests current user/ip identity
- **THEN** backend MUST provide `/user/ipInfo` with `account`, `name`, and `ip` fields
