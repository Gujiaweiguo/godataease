# Security Audit Checklist

## Overview
本文档提供菜单动态授权功能的安全审计检查清单。

---

## Authentication & Authorization

### A1. Authentication Requirements
- [x] 所有菜单端点需要认证
  - **Evidence**: 所有端点通过 JWT 中间件验证
  - **File**: `internal/transport/http/middleware/auth.go`

- [x] Token 验证失败返回 401
  - **Evidence**: `auth.go` - `Unauthorized` response
  - **Test**: `middleware/auth_test.go`

### A2. Authorization Enforcement
- [x] 菜单访问授权检查
  - **Evidence**: `MenuAuthMiddleware` 检查用户是否有权访问菜单路由
  - **File**: `internal/transport/http/middleware/menu_auth.go`

- [x] 未授权访问返回 403
  - **Evidence**: `menu_auth.go` - `Forbidden()` response
  - **Test**: `middleware/menu_auth_test.go`

- [x] 管理员绕过策略安全
  - **Evidence**: 仅 `role_id == 1` 绕过检查
  - **File**: `internal/service/menu_service.go` - `isAdminRole()`

---

## Input Validation

### I1. Request Validation
- [x] 所有输入参数有类型和格式验证
  - **Evidence**: `binding:"required"` 标签验证
  - **Files**: `menu_handler.go`, `role_menu_handler.go`

- [x] 无效输入返回 400
  - **Evidence**: `response.Error(c, "500000", "Invalid request: "+err.Error())`
  - **Test**: Handler 测试覆盖

### I2. SQL Injection Prevention
- [x] 使用参数化查询
  - **Evidence**: GORM ORM 自动参数化
  - **Files**: `internal/repository/*.go`

- [x] 无字符串拼接 SQL
  - **Evidence**: 代码审查确认无拼接
  - **Verified**: 2026-03-06

### I3. ID Validation
- [x] 角色/菜单 ID 存在性检查
  - **Evidence**: `role_menu_service.go` - 验证角色和菜单 ID
  - **Error**: `ErrRoleNotFound`, `ErrMenuNotFound`

---

## Data Integrity

### D1. Database Constraints
- [x] 唯一约束防止重复授权
  - **Constraint**: `uk_role_menu(role_id, menu_id)`
  - **File**: `migrations/mysql/20260222_create_sys_role_menu.sql`

- [x] 外键约束保证引用完整性
  - **FK**: `fk_role_menu_role`, `fk_role_menu_menu`
  - **File**: 同上

- [ ] **注意**: 当前数据库缺少唯一约束和外键
  - **Status**: 需要重新执行迁移或手动添加
  - **Risk**: 中等 - 可能出现重复授权记录

### D2. Transaction Safety
- [x] 角色菜单保存在事务中执行
  - **Evidence**: `role_menu_service.go` - `SaveRoleMenus` 在事务中执行
  - **File**: `internal/service/role_menu_service.go`

---

## Information Disclosure

### ID1. Error Messages
- [x] 错误消息不泄露敏感信息
  - **Evidence**: 通用错误消息，不包含堆栈或内部路径
  - **Files**: `response/` 包

- [x] 日志不包含敏感数据
  - **Evidence**: 密码等敏感字段未记录
  - **Verified**: 2026-03-06

### ID2. API Response
- [x] API 响应不暴露内部实现细节
  - **Evidence**: 标准响应格式 `{"code": "...", "data": ...}`
  - **File**: `response/response.go`

---

## Configuration Security

### C1. Sensitive Configuration
- [x] JWT 密钥从环境变量或配置文件读取
  - **Evidence**: `configs/config.yaml` - `jwt.secret`
  - **Warning**: 不要提交生产密钥到版本控制

- [x] 数据库密码从环境变量读取
  - **Evidence**: `config.go` - 支持 `DATABASE_PASSWORD` 环境变量

### C2. Fallback Mode
- [x] 硬编码回退模式默认关闭
  - **Default**: `menu.hardcoded_fallback: false`
  - **File**: `configs/config.yaml`

- [x] 回退模式仅用于紧急情况
  - **Documentation**: API 文档中有说明
  - **Risk**: 低 - 需要显式启用

---

## Denial of Service

### DO1. Rate Limiting
- [ ] API 速率限制
  - **Status**: 未实现
  - **Risk**: 低 - 可在反向代理层实现
  - **Recommendation**: 在 Nginx/Gateway 层添加速率限制

### DO2. Query Performance
- [x] 菜单查询有索引优化
  - **Evidence**: `idx_role_id`, `idx_menu_id` 索引
  - **File**: `migrations/mysql/20260222_create_sys_role_menu.sql`

- [x] 无 N+1 查询问题
  - **Evidence**: 使用 IN 查询批量获取菜单
  - **File**: `menu_service.go` - `QueryByRoleIDs`

---

## Audit Trail

### AU1. Operation Logging
- [ ] 菜单变更操作日志
  - **Status**: 未实现
  - **Risk**: 低 - 审计追踪功能
  - **Recommendation**: 添加审计日志记录菜单变更

---

## Security Checklist Summary

| Category | Item | Status | Risk Level |
|----------|------|--------|------------|
| Authentication | Token validation | ✅ Pass | N/A |
| Authorization | Menu access check | ✅ Pass | N/A |
| Authorization | Admin bypass secure | ✅ Pass | N/A |
| Input Validation | Parameter validation | ✅ Pass | N/A |
| Input Validation | SQL injection prevention | ✅ Pass | N/A |
| Input Validation | ID validation | ✅ Pass | N/A |
| Data Integrity | Unique constraint | ⚠️ Missing | Medium |
| Data Integrity | Foreign keys | ⚠️ Missing | Medium |
| Data Integrity | Transaction safety | ✅ Pass | N/A |
| Info Disclosure | Error messages | ✅ Pass | N/A |
| Info Disclosure | Log security | ✅ Pass | N/A |
| Configuration | JWT secret | ✅ Pass | N/A |
| Configuration | Fallback default | ✅ Pass | N/A |
| DoS | Rate limiting | ⚠️ Not implemented | Low |
| DoS | Query performance | ✅ Pass | N/A |
| Audit Trail | Operation logging | ⚠️ Not implemented | Low |

---

## Recommendations

### High Priority
1. **添加数据库约束**: 重新执行迁移或手动添加唯一约束和外键
   ```sql
   ALTER TABLE sys_role_menu ADD CONSTRAINT uk_role_menu UNIQUE (role_id, menu_id);
   ALTER TABLE sys_role_menu ADD CONSTRAINT fk_role_menu_role 
     FOREIGN KEY (role_id) REFERENCES sys_role(id) ON DELETE CASCADE;
   ALTER TABLE sys_role_menu ADD CONSTRAINT fk_role_menu_menu 
     FOREIGN KEY (menu_id) REFERENCES core_menu(id) ON DELETE CASCADE;
   ```

### Medium Priority
2. **添加速率限制**: 在 Nginx/Gateway 层实现

### Low Priority
3. **添加审计日志**: 记录菜单变更操作

---

## Security Audit Sign-off

**Auditor**: ________________________  
**Date**: ________________________  
**Result**: [ ] Approved [ ] Conditional [ ] Rejected  

**Conditions (if applicable)**:
1. ________________________
2. ________________________

---

**Document Version**: 1.0  
**Last Updated**: 2026-03-06  
**Author**: Gujiaweiguo
