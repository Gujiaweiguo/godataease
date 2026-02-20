# Change: 实现 Go 版本 Login（登录）模块

## Why

作为 Java 到 Go 渐进式迁移的第十个业务模块，Login（登录）模块是系统认证的核心：
- 基于 substitute 实现简化版，支持 admin 登录
- JWT Token 生成和验证
- 与 Java 版本 API 保持兼容

## What Changes

### 本次范围

- 实现登录请求/响应结构
- 实现 JWT Token 生成（HMAC-SHA256）
- 实现 HTTP Handler 和 API 端点
- 支持简化版登录（不做 RSA 加解密）

### 不包含

- RSA 加解密（前端直接传明文，或后续迭代）
- MFA 多因素认证
- 第三方登录
- Token 刷新机制

## Impact

### 代码影响

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/auth/auth.go` | 新增 |
| `backend-go/internal/service/auth_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/auth_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /login/localLogin | 本地登录 |
| GET | /logout | 登出 |

## Exit Criteria

- [ ] API 端点实现并通过测试
- [ ] JWT Token 格式与 Java 兼容
- [ ] 响应格式与 Java 版本兼容
- [ ] 代码通过编译检查
- [ ] OpenSpec 验证通过
