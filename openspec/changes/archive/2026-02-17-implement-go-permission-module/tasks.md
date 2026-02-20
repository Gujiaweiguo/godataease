# Plan: Go Permission 模块实现

## 任务清单

### PERM-001 实体定义 [完成]
- [x] 创建 SysPerm 实体（映射 sys_perm 表）
- [x] 定义权限类型常量（PermTypeMenu、PermTypeButton、PermTypeData）
- [x] 定义权限状态常量（StatusEnabled、StatusDisabled）
- [x] 定义删除标记常量（DelFlagNormal、DelFlagDeleted）
- [x] 创建请求/响应 DTO 结构

### PERM-002 Repository 层 [完成]
- [x] 实现 PermRepository（Create、Update、Delete、GetByID、GetByKey）
- [x] 实现 List 方法（获取所有权限）
- [x] 实现 CheckKeyExists（检查权限标识）
- [x] 实现 GetByType（按类型查询权限）

### PERM-003 Service 层 [完成]
- [x] 实现 CreatePerm
- [x] 实现 UpdatePerm
- [x] 实现 DeletePerm
- [x] 实现 GetPermByID
- [x] 实现 ListPerms（含分页）
- [x] 实现 CheckPermKeyExists

### PERM-004 Handler 层 [完成]
- [x] 实现 POST /api/system/permission/list
- [x] 实现 POST /api/system/permission/create
- [x] 实现 POST /api/system/permission/update
- [x] 实现 POST /api/system/permission/delete/:permId

### PERM-005 集成与验证 [完成]
- [x] 注册路由（RegisterPermRoutes）
- [x] 在 router.go 中初始化依赖
- [x] 验证 API 响应格式与 Java 版本兼容
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
| `internal/domain/permission/permission.go` | 权限实体定义 |
| `internal/repository/perm_repo.go` | GORM Repository 层 |
| `internal/service/perm_service.go` | 业务逻辑层 |
| `internal/transport/http/handler/perm_handler.go` | HTTP Handler |
| `internal/transport/http/router.go` | 路由注册（需更新） |

## 数据库表结构

### sys_perm
| 字段 | 类型 | 说明 |
|------|------|------|
| perm_id | BIGINT | 主键 |
| perm_name | VARCHAR(100) | 权限名称 |
| perm_key | VARCHAR(100) | 权限标识 |
| perm_type | VARCHAR(20) | 权限类型：menu/button/data |
| perm_desc | VARCHAR(500) | 权限描述 |
| status | INT | 状态：0-禁用，1-启用 |
| create_by | VARCHAR(100) | 创建人 |
| create_time | DATETIME | 创建时间 |
| update_by | VARCHAR(100) | 更新人 |
| update_time | DATETIME | 更新时间 |
| del_flag | INT | 删除标记：0-未删除，1-已删除 |

## 依赖关系

```
implement-go-permission-module
├── 依赖: implement-go-org-module ✅
├── 依赖: implement-go-user-module ✅
├── 依赖: implement-go-audit-module ✅
└── 依赖: refactor-backend-to-go ✅
```
