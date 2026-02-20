# Plan: Go User 模块实现

## 任务清单

### USER-001 实体定义 [完成]
- [x] 创建 SysUser 实体（映射 sys_user 表）
- [x] 创建 SysUserRole 实体（映射 sys_user_role 表）
- [x] 创建 SysUserPerm 实体（映射 sys_user_perm 表）
- [x] 定义用户状态常量（StatusEnabled、StatusDisabled）
- [x] 定义删除标记常量（DelFlagNormal、DelFlagDeleted）
- [x] 定义用户来源常量（FromLocal、FromThirdParty）
- [x] 创建请求/响应 DTO 结构

### USER-002 Repository 层 [完成]
- [x] 实现 UserRepository（Create、Update、Delete、GetByID、GetByUsername）
- [x] 实现 Query 方法（支持 keyword、orgId、status 过滤）
- [x] 实现 CountByUsername（检查用户名是否存在）
- [x] 实现 CheckEmailExists（检查邮箱是否存在）
- [x] 实现 ListUsersByIds（按 ID 列表查询）
- [x] 实现 UserRoleRepository（关联表操作）
- [x] 实现 UserPermRepository（权限关联表操作）

### USER-003 Service 层 [完成]
- [x] 实现 CreateUser（含密码加密）
- [x] 实现 UpdateUser
- [x] 实现 DeleteUser（软删除）
- [x] 实现 GetUserByID
- [x] 实现 GetUserByUsername
- [x] 实现 SearchUsers（多条件查询）
- [x] 实现 ResetPassword
- [x] 实现 UpdateUserStatus
- [x] 实现分页逻辑

### USER-004 Handler 层 [完成]
- [x] 实现 POST /api/system/user/list
- [x] 实现 POST /api/system/user/create
- [x] 实现 POST /api/system/user/update
- [x] 实现 POST /api/system/user/delete/:id
- [x] 实现 GET /api/system/user/options
- [x] 实现统一的响应格式（code、data、msg）
- [x] 实现错误处理（返回 code: 500000）

### USER-005 集成与验证 [完成]
- [x] 注册路由（RegisterUserRoutes）
- [x] 在 router.go 中初始化依赖
- [x] 审计日志集成（通过现有中间件）
- [x] API 响应格式与 Java 版本兼容
- [x] OpenSpec 验证通过
- [x] 代码通过编译检查

## 里程碑

- [x] M1: 实体定义完成
- [x] M2: Repository 层完成
- [x] M3: Service 层完成
- [x] M4: Handler 层完成
- [x] M5: 集成验证通过

## 文件清单

| 文件 | 说明 |
|------|------|
| `internal/domain/user/user.go` | 用户实体定义 |
| `internal/repository/user_repo.go` | GORM Repository 层 |
| `internal/service/user_service.go` | 业务逻辑层 |
| `internal/transport/http/handler/user_handler.go` | HTTP Handler |
| `internal/transport/http/router.go` | 路由注册（需更新） |
| `cmd/api/main.go` | 入口（需更新） |

## 数据库表结构

### sys_user
| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | BIGINT | 主键 |
| username | VARCHAR(100) | 用户名 |
| password | VARCHAR(255) | 密码（加密） |
| nick_name | VARCHAR(100) | 昵称 |
| email | VARCHAR(100) | 邮箱 |
| phone | VARCHAR(50) | 手机号 |
| from | INT | 来源：0-本系统，1-第三方 |
| sub | VARCHAR(255) | 第三方系统标识 |
| avatar | VARCHAR(500) | 头像URL |
| dept_id | BIGINT | 部门ID |
| status | INT | 状态：0-禁用，1-启用 |
| del_flag | INT | 删除标记：0-正常，1-已删除 |
| create_by | VARCHAR(100) | 创建人 |
| create_time | DATETIME | 创建时间 |
| update_by | VARCHAR(100) | 更新人 |
| update_time | DATETIME | 更新时间 |
| language | VARCHAR(20) | 语言 |

## 依赖关系

```
implement-go-user-module
├── 依赖: implement-go-audit-module ✅
└── 依赖: refactor-backend-to-go ✅
```
