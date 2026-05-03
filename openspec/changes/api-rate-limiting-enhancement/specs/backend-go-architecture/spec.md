## ADDED Requirements

### Requirement: Backend Configuration SHALL Include Rate Limit Policy Settings
The Go backend configuration model SHALL include first-class rate-limit settings so operators can manage backend selection, global defaults, and route-specific overrides through configuration and environment binding.

#### Scenario: RateLimitConfig is loaded from application configuration
- **WHEN** the backend loads `config.yaml` and bound environment variables at startup
- **THEN** the system MUST populate a `RateLimitConfig` structure in the application configuration model
- **AND** that structure MUST include enablement, default request budget, default window, Redis backend selection, and route override settings

#### Scenario: Existing route protections migrate without code-only policy values
- **WHEN** the backend registers the existing login, datasource validation, and audit export/download rate limits
- **THEN** the system MUST source their limit values from `RateLimitConfig.RouteOverrides`
- **AND** the configuration model MUST preserve the current route protections without requiring hardcoded numeric values in handlers
