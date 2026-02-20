# Change: 实现 Go 版本 Audit 模块

## Why

作为 Java 到 Go 渐进式迁移的第一个业务模块，Audit（审计日志）模块具有以下优势：
- 独立模块，无复杂业务依赖
- 数据库表结构已完整定义
- API 契约清晰，易于验证
- 可作为后续模块迁移的模板

## What Changes

### 本次范围

- 实现 Go 版本的审计日志实体（AuditLog、LoginFailure、AuditLogDetail）
- 实现 GORM Repository 层
- 实现 Service 层业务逻辑
- 实现 HTTP Handler 和 API 端点
- 实现审计中间件（替代 Java AOP）
- 与 Java 版本 API 保持兼容

### 不包含

- 前端代码修改
- 数据库 schema 变更
- 其他业务模块迁移

## Impact

### 代码影响

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/audit/` | 新增 |
| `backend-go/internal/repository/audit_repo.go` | 新增 |
| `backend-go/internal/service/audit_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/audit_handler.go` | 新增 |
| `backend-go/internal/transport/http/middleware/audit.go` | 更新 |

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/audit/log | 创建审计日志 |
| GET | /api/audit/list | 分页查询 |
| GET | /api/audit/:id | 查询单条 |
| GET | /api/audit/user/:userId | 按用户查询 |
| POST | /api/audit/export | 导出日志 |
| DELETE | /api/audit/retention | 清理过期日志 |
| POST | /api/audit/login-failure | 记录登录失败 |

### 风险评估

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| API 兼容性 | 中 | 保持与 Java 相同的请求/响应格式 |
| 数据一致性 | 低 | 使用相同的数据库表 |
| 性能差异 | 低 | Go GORM 性能优于 MyBatis |

## Exit Criteria

- [ ] 所有 API 端点实现并通过测试
- [ ] 响应格式与 Java 版本兼容
- [ ] 代码通过 lint 检查
