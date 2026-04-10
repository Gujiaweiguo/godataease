## Delta: api-compatibility-bridge

### Added Scenario: Compatibility bridge exposes watermark identity endpoints
- **WHEN** compatibility bridge registers Java-era `/user/*` routes
- **THEN** it MUST include `GET /user/personInfo` and `GET /user/ipInfo`
- **AND** these endpoints MUST resolve through Go user handler implementations
