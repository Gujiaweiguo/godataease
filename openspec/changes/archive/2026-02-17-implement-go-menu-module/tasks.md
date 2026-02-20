# Plan: Go Menu 模块实现

## 任务清单

### MENU-001 实体定义 [完成]
- [x] 创建 CoreMenu 实体（映射 core_menu 表）
- [x] 定义 MenuMeta 结构
- [x] 定义 MenuVO 响应结构（树形）

### MENU-002 Repository 层 [完成]
- [x] 实现 MenuRepository
- [x] 实现 GetAll 方法（按 menu_sort 排序）

### MENU-003 Service 层 [完成]
- [x] 实现 Query 方法
- [x] 实现树形结构构建（buildTree）
- [x] 实现实体转 VO（convert）

### MENU-004 Handler 层 [完成]
- [x] 实现 GET /api/menu/query
- [x] 实现统一的响应格式

### MENU-005 集成与验证 [完成]
- [x] 注册路由（RegisterMenuRoutes）
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
| `internal/domain/menu/menu.go` | 菜单实体定义 |
| `internal/repository/menu_repo.go` | GORM Repository 层 |
| `internal/service/menu_service.go` | 业务逻辑层 |
| `internal/transport/http/handler/menu_handler.go` | HTTP Handler |
| `internal/transport/http/router.go` | 路由注册（需更新） |

## 数据库表结构

### core_menu
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| pid | BIGINT | 父ID |
| type | INT | 类型 |
| name | VARCHAR(100) | 名称 |
| component | VARCHAR(255) | 组件 |
| menu_sort | INT | 排序 |
| icon | VARCHAR(100) | 图标 |
| path | VARCHAR(255) | 路径 |
| hidden | BOOLEAN | 隐藏 |
| in_layout | BOOLEAN | 是否内部 |
| auth | BOOLEAN | 参与授权 |
