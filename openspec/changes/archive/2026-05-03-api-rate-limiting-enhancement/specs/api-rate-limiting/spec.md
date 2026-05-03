## ADDED Requirements

### Requirement: Rate Limiting SHALL Support Distributed State via Redis
The system SHALL support a Redis-backed rate-limiting backend so that request budgets can be shared across multiple Go backend instances while preserving an in-memory fallback when distributed state is unavailable.

#### Scenario: Multi-instance shared limits
- **WHEN** multiple backend instances enforce the same configured rate-limited route for the same identity key
- **THEN** the system MUST use shared Redis-backed rate-limit state when Redis-backed limiting is enabled
- **AND** the combined requests across instances MUST count against the same request budget

#### Scenario: Fallback to in-memory when Redis unavailable
- **WHEN** Redis-backed limiting is enabled in configuration but a Redis backend cannot be used
- **THEN** the system MUST continue enforcing rate limits with the in-memory backend
- **AND** the system MUST preserve route behavior instead of disabling request processing unexpectedly

### Requirement: Rate-Limited Responses SHALL Include Standard RateLimit Headers
The system SHALL include standard rate-limit metadata headers on both successful and rejected evaluations for a rate-limited request.

#### Scenario: Successful request includes headers
- **WHEN** a request is allowed by an active rate-limit policy
- **THEN** the response MUST include `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset` headers
- **AND** the system MUST preserve the existing success response body semantics

#### Scenario: Rejected request includes Retry-After
- **WHEN** a request is rejected because the active rate-limit budget has been exhausted
- **THEN** the response MUST use HTTP `429 Too Many Requests`
- **AND** the response MUST include `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, and `Retry-After` headers

### Requirement: Rate Limit Parameters SHALL Be Configurable
The system SHALL load rate-limit defaults, backend selection, enablement, and route overrides from application configuration rather than hardcoded handler arguments.

#### Scenario: Global defaults applied to all authenticated routes
- **WHEN** global rate limiting is enabled for authenticated API routes
- **THEN** the system MUST apply the configured default request budget and window to authenticated routes that do not define a route-specific override
- **AND** the system MUST key authenticated requests by user identity with client IP as the documented fallback

#### Scenario: Per-route overrides take precedence
- **WHEN** a route-specific override is configured for a protected route
- **THEN** the system MUST apply the override values instead of the global default values
- **AND** the system MUST preserve the route-specific budget semantics defined by that override

#### Scenario: Rate limiting can be disabled
- **WHEN** rate limiting is disabled by configuration
- **THEN** the system MUST stop enforcing the global default policy
- **AND** the system MUST allow operators to disable the capability without removing the middleware implementation from the codebase
