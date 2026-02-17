# Plan: Go Audit 模块实现

## 任务清单

### AUDIT-001 实体定义 [完成]
- [x] 创建 AuditLog 实体
- [x] 创建 LoginFailure 实体
- [x] 创建 AuditLogDetail 实体
- [x] 定义常量和枚举（ActionType、Operation、ResourceType、Status）

### AUDIT-002 Repository 层 [完成]
- [x] 实现 AuditLogRepository（Create、GetByID、GetByUserID、Query、DeleteBeforeDate）
- [x] 实现 LoginFailureRepository（Create、GetByUsername、CountSinceTime）
- [x] 实现 AuditLogDetailRepository（Create、GetByAuditLogID）

### AUDIT-003 Service 层 [完成]
- [x] 实现 CreateAuditLog
- [x] 实现 QueryAuditLogs（多条件分页）
- [x] 实现 GetAuditLogByID
- [x] 实现 GetAuditLogsByUserID
- [x] 实现 RecordLoginFailure
- [x] 实现 DeleteAuditLogsBeforeDate
- [x] 实现 ExportAuditLogs（CSV/JSON）

### AUDIT-004 Handler 层 [完成]
- [x] 实现 POST /api/audit/log
- [x] 实现 GET /api/audit/list
- [x] 实现 GET /api/audit/:id
- [x] 实现 GET /api/audit/user/:userId
- [x] 实现 POST /api/audit/export
- [x] 实现 DELETE /api/audit/retention
- [x] 实现 POST /api/audit/login-failure
- [x] 实现 GET /api/audit/download

### AUDIT-005 中间件 [完成]
- [x] 实现审计中间件框架（AuditLog）
- [x] 支持自动记录请求信息（user_id、ip_address、user_agent）
- [x] 支持异步日志写入

### AUDIT-006 验收 [完成]
- [x] 路由注册（RegisterAuditRoutes）
- [x] OpenSpec 验证通过
- [x] 代码结构完整

## 里程碑

- [x] M1: 实体定义完成
- [x] M2: Repository 层完成
- [x] M3: Service 层完成
- [x] M4: Handler 层完成
- [x] M5: 中间件完成
- [x] M6: 集成验证通过

## 文件清单

| 文件 | 说明 |
|------|------|
| `internal/domain/audit/audit.go` | 实体定义和常量 |
| `internal/repository/audit_repo.go` | GORM Repository 层 |
| `internal/service/audit_service.go` | 业务逻辑层 |
| `internal/transport/http/handler/audit_handler.go` | HTTP Handler |
| `internal/transport/http/middleware/auth.go` | 审计中间件 |
| `internal/transport/http/router.go` | 路由注册（已更新） |
| `cmd/api/main.go` | 入口（已更新） |

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/audit/log | 创建审计日志 |
| GET | /api/audit/list | 分页查询（支持多条件过滤） |
| GET | /api/audit/:id | 根据 ID 查询 |
| GET | /api/audit/user/:userId | 按用户查询 |
| POST | /api/audit/export | 导出日志（CSV/JSON） |
| DELETE | /api/audit/retention | 清理过期日志 |
| POST | /api/audit/login-failure | 记录登录失败 |
| GET | /api/audit/download | 下载导出文件 |
