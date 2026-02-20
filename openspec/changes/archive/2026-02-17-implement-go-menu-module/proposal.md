# Change: 实现 Go 版本 Menu（菜单）模块

## Why

作为 Java 到 Go 渐进式迁移的第八个业务模块，Menu（菜单）模块提供系统导航功能：
- 最简单的模块，只有 1 个 API 端点
- 树形结构展示，是典型的递归数据结构
- 不依赖其他复杂业务模块

## What Changes

### 本次范围

- 实现 Go 版本的菜单实体（CoreMenu）
- 实现 GORM Repository 层
- 实现 Service 层（树形结构构建）
- 实现 HTTP Handler 和 API 端点
- 与 Java 版本 API 保持兼容

### 不包含

- 前端代码修改
- 数据库 schema 变更
- 国际化（i18n）功能
- 企业版菜单过滤

## Impact

### 代码影响

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/menu/menu.go` | 新增 |
| `backend-go/internal/repository/menu_repo.go` | 新增 |
| `backend-go/internal/service/menu_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/menu_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/menu/query | 查询菜单树 |

### 数据库表

| 表名 | 说明 |
|------|------|
| core_menu | 菜单主表 |

## Exit Criteria

- [ ] API 端点实现并通过测试
- [ ] 响应格式与 Java 版本兼容（code: 000000/500000）
- [ ] 树形结构正确构建
- [ ] 代码通过 lint 检查
- [ ] OpenSpec 验证通过
