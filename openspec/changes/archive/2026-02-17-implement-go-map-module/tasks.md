# Plan: Go Map 模块实现

## 任务清单

### MAP-001 实体定义 [完成]
- [x] 创建 Area 实体（映射 area 表）
- [x] 创建 CoreAreaCustom 实体（映射 core_area_custom 表）
- [x] 定义 AreaNode 响应结构（树形）

### MAP-002 Repository 层 [完成]
- [x] 实现 AreaRepository
- [x] 实现 GetAll 方法
- [x] 实现 CoreAreaCustomRepository

### MAP-003 Service 层 [完成]
- [x] 实现 GetWorldTree 方法
- [x] 实现树形结构构建

### MAP-004 Handler 层 [完成]
- [x] 实现 GET /api/map/worldTree
- [x] 实现统一的响应格式

### MAP-005 集成与验证 [完成]
- [x] 注册路由（RegisterMapRoutes）
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
| `internal/domain/map/area.go` | 区域实体定义 |
| `internal/repository/area_repo.go` | GORM Repository 层 |
| `internal/service/map_service.go` | 业务逻辑层 |
| `internal/transport/http/handler/map_handler.go` | HTTP Handler |
| `internal/transport/http/router.go` | 路由注册（需更新） |

## 数据库表结构

### area
| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(50) | 区域ID |
| level | VARCHAR(50) | 级别 |
| name | VARCHAR(100) | 名称 |
| pid | VARCHAR(50) | 父ID |

### core_area_custom
| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(50) | 区域ID |
| level | VARCHAR(50) | 级别 |
| name | VARCHAR(100) | 名称 |
| pid | VARCHAR(50) | 父ID |
