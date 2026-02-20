# Change: 实现 Go 版本 Embedded（嵌入式）模块

## Why

作为 Java 到 Go 渐进式迁移的第六个业务模块，Embedded（嵌入式应用管理）模块具有以下特点：
- 审计日志、用户、组织、权限模块已实现
- 嵌入式功能允许第三方应用安全嵌入 DataEase 仪表板
- 支持基于 JWT 的 Token 认证和域名白名单验证
- 10 个核心 API 端点，功能完整

## What Changes

### 本次范围

- 实现 Go 版本的嵌入式应用实体（CoreEmbedded）
- 实现 GORM Repository 层
- 实现 Service 层业务逻辑（包含 Token 生成和验证）
- 实现 HTTP Handler 和 API 端点
- 与 Java 版本 API 保持兼容
- 集成审计日志中间件

### 不包含

- 前端代码修改
- 数据库 schema 变更
- 实际的 iframe 渲染逻辑
- License 限制功能

## Impact

### 代码影响

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/embedded/embedded.go` | 更新 |
| `backend-go/internal/repository/embedded_repo.go` | 新增 |
| `backend-go/internal/service/embedded_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/embedded_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/embedded/pager/{goPage}/{pageSize} | 分页查询嵌入式应用 |
| POST | /api/embedded/create | 创建嵌入式应用 |
| POST | /api/embedded/edit | 编辑嵌入式应用 |
| POST | /api/embedded/delete/{id} | 删除嵌入式应用 |
| POST | /api/embedded/batchDelete | 批量删除 |
| POST | /api/embedded/reset | 重置密钥 |
| GET | /api/embedded/domainList | 获取域名列表 |
| POST | /api/embedded/initIframe | 初始化 iframe |
| GET | /api/embedded/getTokenArgs | 获取 Token 参数 |
| GET | /api/embedded/limitCount | 获取限制数量 |

### 数据库表

| 表名 | 说明 |
|------|------|
| core_embedded | 嵌入式应用主表 |

### 风险评估

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| JWT Token 兼容性 | 中 | 使用相同的 HMAC-SHA256 算法 |
| 密钥生成兼容性 | 低 | 使用相同的随机字符串生成方式 |
| 域名验证逻辑 | 中 | 保持与 Java 相同的解析和匹配逻辑 |

## Exit Criteria

- [ ] 所有 API 端点实现并通过测试
- [ ] 响应格式与 Java 版本兼容（code: 000000/500000）
- [ ] JWT Token 生成和验证与 Java 版本一致
- [ ] 密钥脱敏显示与 Java 版本一致
- [ ] 代码通过 lint 检查
- [ ] OpenSpec 验证通过
