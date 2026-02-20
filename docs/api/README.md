# API 文档索引

本文档包含 DataEase 主要功能模块的 API 使用指南。

## 后端实现说明

- 主线后端：Go + Gin（`apps/backend-go`）
- 历史后端：Spring Boot（`legacy/backend-java`，只读备份）

## 文档列表

1. [嵌入式 BI (Embedded BI)](./embedded-bi.md)
   - 多维度嵌入（Dashboard、Screen、模块页面、图表）
   - Token 管理和自动刷新
   - Iframe 和 DIV 两种嵌入模式
   - 参数传递和事件通信

2. [权限系统](./permission-system.md)
   - 用户管理（CRUD、批量操作）
   - 组织管理（树形结构、成员管理）
   - 角色管理（创建、分配权限、成员）
   - 权限配置（菜单、资源、导出权限）
   - 数据权限（行级过滤、列级脱敏）

3. [快速入门](../quick-start.md)
   - 环境搭建
   - 项目结构
   - 开发规范
   - 常用命令
   - 故障排查

## 快速导航

### 按功能查找

| 功能 | 文档 |
|------|------|
| 嵌入第三方系统 | [嵌入式 BI](./embedded-bi.md) |
| 用户和组织管理 | [权限系统](./permission-system.md) |
| 数据权限控制 | [权限系统](./permission-system.md) |

### 按开发阶段查找

| 阶段 | 文档 |
|------|------|
| 快速开始 | [快速入门](../quick-start.md) |
| API 参考 | 本目录所有文档 |
| OpenSpec 规范 | [OpenSpec 指南](../../openspec/AGENTS.md) |

## API 端点汇总

### 认证相关

```
POST   /api/auth/login
GET    /api/auth/menuResource
POST    /api/auth/menuPermission
```

### 嵌入相关

```
POST   /api/embedded/iframe/tokenInfo
GET    /embedded/iframe/dashboard/{id}/designer
GET    /embedded/iframe/dashboard/{id}/view
GET    /embedded/iframe/screen/{id}/view
GET    /embedded/iframe/module/{type}/view
GET    /embedded/iframe/chart/{id}/view
```

### 组织管理

```
POST   /api/org/create
GET    /api/org/tree
POST   /api/org/members/add
DELETE /api/org/{id}
```

### 角色管理

```
POST   /api/role/create
GET    /api/role/list
POST   /api/role/permissions/assign
POST   /api/role/members/assign
```

### 权限管理

```
GET    /api/permission/list
POST   /api/permission/assign
POST   /api/rowPermission/create
POST   /api/columnPermission/create
```

### 用户管理

```
GET    /api/user/list
POST   /api/user/create
PUT    /api/user/{id}
DELETE /api/user/{id}
POST   /api/user/batch/assign-org
```

## 联系方式

如有问题或建议，请：

1. 查看 [快速入门](../quick-start.md) 故障排查部分
2. 查看 [贡献指南](../CONTRIBUTING.md)
3. 提交 Issue 或 Pull Request

## 更新日志

| 日期 | 版本 | 更新内容 |
|--------|--------|----------|
| 2026-01-28 | 2.10.19 | 添加嵌入 BI 和权限系统 API 文档 |
