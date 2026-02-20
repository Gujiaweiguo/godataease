# Change: 实现 Go 版本 Map（地图区域）模块

## Why

作为 Java 到 Go 渐进式迁移的第九个业务模块，Map（地图区域）模块提供地图区域树形数据：
- 只有 1 个核心 API 端点
- 树形结构展示，与 Menu 模块模式类似
- 不依赖其他复杂业务模块

## What Changes

### 本次范围

- 实现 Go 版本的区域实体（Area、CoreAreaCustom）
- 实现 GORM Repository 层
- 实现 Service 层（树形结构构建）
- 实现 HTTP Handler 和 API 端点
- 与 Java 版本 API 保持兼容

### 不包含

- 前端代码修改
- 数据库 schema 变更
- GeoJSON 文件上传功能
- 自定义区域管理 API

## Impact

### 代码影响

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/map/area.go` | 新增 |
| `backend-go/internal/repository/area_repo.go` | 新增 |
| `backend-go/internal/service/map_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/map_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/map/worldTree | 获取世界区域树 |

### 数据库表

| 表名 | 说明 |
|------|------|
| area | 核心区域数据 |
| core_area_custom | 自定义区域数据 |

## Exit Criteria

- [ ] API 端点实现并通过测试
- [ ] 响应格式与 Java 版本兼容（code: 000000/500000）
- [ ] 树形结构正确构建
- [ ] 代码通过 lint 检查
- [ ] OpenSpec 验证通过
