# embedded-bi Specification Delta

## MODIFIED Requirements

### Requirement: Token-Based Embedding Initialization

系统 SHALL 支持使用 Go 实现的 JWT 算法初始化嵌入式内容，生成与 Java 版本兼容的 embedded token。

#### Scenario: Generating an embedded token
- **WHEN** a caller provides app id, app secret, and a valid user account
- **THEN** the system returns an embedded token compatible with the Java implementation

#### Scenario: Token format compatibility
- **WHEN** generating an embedded token in Go
- **THEN** the token format and claims SHALL be identical to the Java implementation

## ADDED Requirements

### Requirement: Embedded API Performance

系统 SHALL 在 Go 实现中保持或提升嵌入式 API 的性能指标。

#### Scenario: Token generation latency
- **WHEN** generating an embedded token
- **THEN** the P95 latency SHALL be less than or equal to the Java implementation

#### Scenario: Embedding initialization latency
- **WHEN** initializing an embedded session
- **THEN** the P95 latency SHALL be less than or equal to the Java implementation
