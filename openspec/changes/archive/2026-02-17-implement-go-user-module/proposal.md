# Change: 实现 Go 版本 User 模块

## Why

作为 Java 到 Go 渐进式迁移的第二个业务模块，User（用户管理）模块具有以下优势：
- 审计日志模块已实现，用户操作可被审计
- 用户管理是核心基础模块，其他模块（组织、权限）依赖用户
- 数据库表结构已完整定义（sys_user 及关联表）
- API 契约清晰，5 个核心端点易于验证

## What Changes

### 本次范围

- 实现 Go 版本的用户实体（SysUser、SysUserRole、SysUserPerm）
- 实现 GORM Repository 层
- 实现 Service 层业务逻辑
- 实现 HTTP Handler 和 API 端点
- 与 Java 版本 API 保持兼容
- 集成审计日志中间件

### 不包含

- 前端代码修改
- 数据库 schema 变更
- 认证/登录逻辑（将在后续模块实现）
- 其他业务模块迁移

## Impact

### 代码影响

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/user/user.go` | 更新 |
| `backend-go/internal/repository/user_repo.go` | 新增 |
| `backend-go/internal/service/user_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/user_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/system/user/list | 分页查询用户列表 |
| POST | /api/system/user/create | 创建用户 |
| POST | /api/system/user/update | 更新用户 |
| POST | /api/system/user/delete/:id | 删除用户 |
| GET | /api/system/user/options | 获取用户选项列表 |

### 数据库表

| 表名 | 说明 |
|------|------|
| sys_user | 用户主表 |
| sys_user_role | 用户角色关联表 |
| sys_user_perm | 用户权限关联表 |

### 风险评估

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| API 兼容性 | 中 | 保持与 Java 相同的请求/响应格式 |
| 密码加密兼容性 | 中 | 使用与 Java 相同的 bcrypt 算法 |
| 数据一致性 | 低 | 使用相同的数据库表 |

## Exit Criteria

- [ ] 所有 API 端点实现并通过测试
- [ ] 响应格式与 Java 版本兼容（code: 000000/500000）
- [ ] 密码加密方式与 Java 版本一致
- [ ] 代码通过 lint 检查
- [ ] OpenSpec 验证通过
