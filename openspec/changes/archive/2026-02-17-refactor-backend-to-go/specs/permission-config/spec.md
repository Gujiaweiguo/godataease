# permission-config Specification Delta

## MODIFIED Requirements

### Requirement: Row-Level Permission Filtering

系统 SHALL 在 Go 实现中保持与 Java 版本完全一致的行级权限过滤逻辑。

#### Scenario: Permission SQL generation
- **WHEN** 生成行级权限过滤 SQL
- **THEN** Go 实现 SHALL 生成与 Java 版本语法等效的 WHERE 子句

#### Scenario: Permission filter result
- **WHEN** 执行带权限过滤的查询
- **THEN** Go 实现 SHALL 返回与 Java 版本完全相同的过滤结果

### Requirement: Column Permission Control

系统 SHALL 在 Go 实现中保持与 Java 版本完全一致的列级权限控制。

#### Scenario: Column is disabled for a user
- **WHEN** a user queries a dataset with a disabled column
- **THEN** Go 实现 SHALL 以与 Java 版本相同的方式从结果中隐藏该列

#### Scenario: Column is masked for a user
- **WHEN** a user queries a dataset with a masked column
- **THEN** Go 实现 SHALL 以与 Java 版本相同的方式进行脱敏处理

## ADDED Requirements

### Requirement: Permission Cache Consistency

系统 SHALL 在 Go 实现中保持与 Java 版本相同的权限缓存行为。

#### Scenario: Cache invalidation
- **WHEN** 权限配置变更
- **THEN** Go 实现 SHALL 以与 Java 版本相同的方式失效相关缓存
