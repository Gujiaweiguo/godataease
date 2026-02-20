# Plan: Go Organization 模块实现

## 任务清单

### ORG-001 实体定义 [完成]
- [x] 创建 SysOrg 实体（映射 sys_org 表）
- [x] 定义组织状态常量（StatusEnabled、StatusDisabled）
- [x] 定义删除标记常量（DelFlagNormal、DelFlagDeleted）
- [x] 创建请求/响应 DTO 结构
- [x] 创建树形结构响应 DTO（OrgTreeNode）

### ORG-002 Repository 层 [完成]
- [x] 实现 OrgRepository（Create、Update、Delete、GetByID、GetByName）
- [x] 实现 List 方法（获取所有组织）
- [x] 实现 ListByParentID（获取子组织）
- [x] 实现 CheckNameExists（检查组织名称）
- [x] 实现 GetOrgTree（递归获取组织树）
- [x] 实现 CountChildren（统计子组织数量）

### ORG-003 Service 层 [完成]
- [x] 实现 CreateOrg（含层级计算）
- [x] 实现 UpdateOrg
- [x] 实现 DeleteOrg（级联检查）
- [x] 实现 GetOrgByID
- [x] 实现 ListOrgs
- [x] 实现 GetOrgTree
- [x] 实现 UpdateOrgStatus
- [x] 实现 CheckOrgNameExists

### ORG-004 Handler 层 [完成]
- [x] 实现 POST /api/system/organization/create
- [x] 实现 POST /api/system/organization/update
- [x] 实现 POST /api/system/organization/delete/:orgId
- [x] 实现 GET /api/system/organization/list
- [x] 实现 GET /api/system/organization/info/:orgId
- [x] 实现 GET /api/system/organization/tree
- [x] 实现 GET /api/system/organization/checkName
- [x] 实现 POST /api/system/organization/updateStatus
- [x] 实现 GET /api/system/organization/children/:parentId

### ORG-005 集成与验证 [完成]
- [x] 注册路由（RegisterOrgRoutes）
- [x] 在 router.go 中初始化依赖
- [x] 验证 API 响应格式与 Java 版本兼容
- [x] 验证树形结构查询正确
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
| `internal/domain/org/org.go` | 组织实体定义 |
| `internal/repository/org_repo.go` | GORM Repository 层 |
| `internal/service/org_service.go` | 业务逻辑层 |
| `internal/transport/http/handler/org_handler.go` | HTTP Handler |
| `internal/transport/http/router.go` | 路由注册（需更新） |

## 数据库表结构

### sys_org
| 字段 | 类型 | 说明 |
|------|------|------|
| org_id | BIGINT | 主键 |
| org_name | VARCHAR(100) | 组织名称 |
| org_desc | VARCHAR(500) | 组织描述 |
| parent_id | BIGINT | 父组织ID（0为顶级） |
| level | INT | 层级：1-顶级，2-二级... |
| dept_id | BIGINT | 关联部门ID（兼容） |
| status | INT | 状态：0-禁用，1-启用 |
| del_flag | INT | 删除标记：0-未删除，1-已删除 |
| create_by | VARCHAR(100) | 创建人 |
| create_time | DATETIME | 创建时间 |
| update_by | VARCHAR(100) | 更新人 |
| update_time | DATETIME | 更新时间 |

## 依赖关系

```
implement-go-org-module
├── 依赖: implement-go-user-module ✅
├── 依赖: implement-go-audit-module ✅
└── 依赖: refactor-backend-to-go ✅
```
