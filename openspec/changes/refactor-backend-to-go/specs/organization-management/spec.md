# organization-management Specification Delta

## MODIFIED Requirements

### Requirement: Organization Tree Structure

系统 SHALL 在 Go 实现中保持与 Java 版本相同的组织树形结构。

#### Scenario: Tree query compatibility
- **WHEN** 查询组织树
- **THEN** Go 实现 SHALL 返回与 Java 版本相同的树形结构数据格式

#### Scenario: Recursive CTE query
- **WHEN** 查询子组织
- **THEN** Go 实现 SHALL 使用与 Java 版本等效的递归 CTE SQL

### Requirement: Organization Member Management

系统 SHALL 在 Go 实现中保持与 Java 版本相同的成员管理行为。

#### Scenario: Member assignment
- **WHEN** 分配用户到组织
- **THEN** Go 实现 SHALL 以相同方式更新组织成员关系表
