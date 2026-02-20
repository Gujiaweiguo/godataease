# Plan: Go Role 模块实现

## 任务清单

### ROLE-001 实体定义 [完成]
- [x] 创建 SysRole 实体（映射 sys_role 表）
- [x] 定义 RoleCreator 请求 DTO
- [x] 定义 RoleEditor 请求 DTO
- [x] 定义 RoleVO 响应 VO
- [x] 定义 RoleDetailVO 响应 VO
- [x] 定义查询请求结构

### ROLE-002 Repository 层 [完成]
- [x] 实现 RoleRepository（Create、Update、Delete、GetByID）
- [x] 实现 Query 方法（支持 keyword 过滤）
- [x] 实现 CountByRoleCode（检查角色编码是否存在）

### ROLE-003 Service 层 [完成]
- [x] 实现 CreateRole
- [x] 实现 EditRole
- [x] 实现 DeleteRole（软删除）
- [x] 实现 GetRoleByID
- [x] 实现 QueryRoles（多条件查询）

### ROLE-004 Handler 层 [完成]
- [x] 实现 POST /api/role/query
- [x] 实现 POST /api/role/create
- [x] 实现 POST /api/role/edit
- [x] 实现 POST /api/role/delete/:id
- [x] 实现 GET /api/role/detail/:id
- [x] 实现统一的响应格式（code、data、msg）
- [x] 实现错误处理（返回 code: 500000）

### ROLE-005 集成与验证 [完成]
- [x] 注册路由（RegisterRoleRoutes）
- [x] 在 router.go 中初始化依赖
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
| `internal/domain/role/role.go` | 角色实体定义 |
| `internal/repository/role_repo.go` | GORM Repository 层 |
| `internal/service/role_service.go` | 业务逻辑层 |
| `internal/transport/http/handler/role_handler.go` | HTTP Handler |
| `internal/transport/http/router.go` | 路由注册（需更新） |

## 数据库表结构

### sys_role
| 字段 | 类型 | 说明 |
|------|------|------|
| role_id | BIGINT | 主键 |
| role_name | VARCHAR(100) | 角色名称 |
| role_code | VARCHAR(100) | 角色编码 |
| role_desc | VARCHAR(255) | 角色描述 |
| parent_id | BIGINT | 父角色ID |
| level | INT | 层级 |
| data_scope | VARCHAR(50) | 数据范围 |
| status | INT | 状态：0-禁用，1-启用 |
| create_by | VARCHAR(100) | 创建人 |
| create_time | DATETIME | 创建时间 |
| update_by | VARCHAR(100) | 更新人 |
| update_time | DATETIME | 更新时间 |

## 依赖关系

```
implement-go-role-module
├── 依赖: implement-go-user-module ✅
├── 依赖: implement-go-org-module ✅
├── 依赖: implement-go-permission-module ✅
└── 依赖: refactor-backend-to-go ✅
```

## 数据范围常量

| 值 | 说明 |
|------|------|
| all | 全部数据 |
| custom | 自定义 |
| dept | 本部门及以下 |
| dept_and_child | 本部门及子部门 |
| self | 仅本人 |
