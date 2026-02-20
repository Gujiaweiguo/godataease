# API Migration Matrix (Java → Go)

本文档记录了 Java Spring Boot 后端与 Go Gin 后端之间的 API 端点映射关系。
通过兼容性桥接层（Compatibility Bridge），Go 后端可以代理转发或直接处理原本由 Java 后端提供的 API 请求。

> **生成时间**: 2026-02-18
> **源项目**: DataEase v2 (Java Spring Boot)
> **目标项目**: DataEase Go Backend (Gin)

---

## 目录

1. [Datasource APIs](#1-datasource-apis)
2. [DatasetTree APIs](#2-datasettree-apis)
3. [DatasetData APIs](#3-datasetdata-apis)
4. [Chart/ChartData APIs](#4-chartchartdata-apis)
5. [User APIs](#5-user-apis)
6. [Org APIs](#6-org-apis)
7. [License APIs](#7-license-apis)
8. [MsgCenter APIs](#8-msgcenter-apis)
9. [System APIs](#9-system-apis)
10. [状态图例](#状态图例)

---

## 状态图例

| 状态 | 说明 |
|------|------|
| ✅ Migrated | 已在 Go 中完整实现，可直接使用 |
| 🔄 Partial | 部分实现，某些功能可能仍需 Java 后端 |
| ⏳ Pending | 计划迁移，尚未实现 |
| 🌉 Proxied | 通过兼容性桥接代理到 Java 后端 |
| ❌ Skipped | 不迁移（功能废弃或由其他模块处理） |

---

## 1. Datasource APIs

**Java Controller**: `io.dataease.datasource.server.DatasourceServer`
**Go Handler**: `DatasourceHandler` (compatibility_bridge_handler.go)

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/datasource/save` | POST | `/datasource/save` | ✅ Migrated | 创建数据源 |
| `/datasource/update` | POST | `/datasource/update` | ✅ Migrated | 更新数据源 |
| `/datasource/delete/{id}` | GET | `/datasource/delete/:id` | ✅ Migrated | 删除数据源 |
| `/datasource/get/{id}` | GET | `/datasource/get/:id` | ✅ Migrated | 获取数据源详情 |
| `/datasource/hidePw/{id}` | GET | `/datasource/hidePw/:id` | ✅ Migrated | 获取数据源(隐藏密码) |
| `/datasource/getSimpleDs/{id}` | GET | `/datasource/getSimpleDs/:id` | ✅ Migrated | 获取简单数据源信息 |
| `/datasource/list` | POST | `/datasource/list` | ✅ Migrated | 数据源列表 |
| `/datasource/tree` | POST | `/datasource/tree` | ✅ Migrated | 数据源树形结构 |
| `/datasource/validate` | POST | `/datasource/validate` | ✅ Migrated | 验证数据源连接 |
| `/datasource/validate/{id}` | GET | `/datasource/validate/:id` | ✅ Migrated | 验证已存在数据源 |
| `/datasource/types` | POST | `/datasource/types` | ✅ Migrated | 获取支持的数据库类型 |
| `/datasource/getTables` | POST | `/datasource/getTables` | ✅ Migrated | 获取数据源表列表 |
| `/datasource/getTableStatus` | POST | `/datasource/getTableStatus` | ✅ Migrated | 获取表状态 |
| `/datasource/getSchema` | POST | `/datasource/getSchema` | ✅ Migrated | 获取数据库Schema |
| `/datasource/getTableField` | POST | `/datasource/getTableField` | ✅ Migrated | 获取表字段 |
| `/datasource/previewData` | POST | `/datasource/previewData` | ✅ Migrated | 预览数据 |
| `/datasource/move` | POST | `/datasource/move` | ✅ Migrated | 移动数据源 |
| `/datasource/reName` | POST | `/datasource/reName` | ✅ Migrated | 重命名数据源 |
| `/datasource/createFolder` | POST | `/datasource/createFolder` | ✅ Migrated | 创建文件夹 |
| `/datasource/checkRepeat` | POST | `/datasource/checkRepeat` | ✅ Migrated | 检查重复 |
| `/datasource/perDelete/{id}` | POST | `/datasource/perDelete/:id` | ✅ Migrated | 预删除检查 |
| `/datasource/showFinishPage` | GET | `/datasource/showFinishPage` | ✅ Migrated | 显示完成页面 |
| `/datasource/setShowFinishPage` | POST | `/datasource/setShowFinishPage` | ✅ Migrated | 设置完成页面 |
| `/datasource/latestUse` | POST | `/datasource/latestUse` | ✅ Migrated | 最近使用 |
| `/datasource/checkApiDatasource` | POST | `/datasource/checkApiDatasource` | 🌉 Proxied | API数据源检查(存根) |
| `/datasource/loadRemoteFile` | POST | `/datasource/loadRemoteFile` | 🌉 Proxied | 加载远程文件(存根) |
| `/datasource/syncApiTable` | POST | `/datasource/syncApiTable` | 🌉 Proxied | 同步API表(存根) |
| `/datasource/syncApiDs` | POST | `/datasource/syncApiDs` | 🌉 Proxied | 同步API数据源(存根) |
| `/datasource/uploadFile` | POST | `/datasource/uploadFile` | 🌉 Proxied | 上传文件(存根) |
| `/datasource/listSyncRecord/{dsId}/{page}/{limit}` | POST | `/datasource/listSyncRecord/:dsId/:page/:limit` | 🌉 Proxied | 同步记录列表(存根) |

---

## 2. DatasetTree APIs

**Java Controller**: `io.dataease.dataset.server.DatasetTreeServer`
**Go Handler**: `DatasetHandler` (compatibility_bridge_handler.go)

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/datasetTree/tree` | POST | `/datasetTree/tree` | ✅ Migrated | 数据集树形结构 |
| `/datasetTree/get/{id}` | POST | `/datasetTree/get/:id` | ✅ Migrated | 获取数据集详情 |
| `/datasetTree/details/{id}` | POST | `/datasetTree/details/:id` | ✅ Migrated | 获取数据集详细信息 |
| `/datasetTree/dsDetails` | POST | `/datasetTree/dsDetails` | ✅ Migrated | 批量获取数据集详情 |
| `/datasetTree/detailWithPerm` | POST | `/datasetTree/detailWithPerm` | ✅ Migrated | 带权限获取详情 |
| `/datasetTree/getSqlParams` | POST | `/datasetTree/getSqlParams` | ✅ Migrated | 获取SQL参数 |
| `/datasetTree/save` | POST | `/datasetTree/save` | ✅ Migrated | 保存数据集 |
| `/datasetTree/create` | POST | `/datasetTree/create` | ✅ Migrated | 创建数据集 |
| `/datasetTree/rename` | POST | `/datasetTree/rename` | ✅ Migrated | 重命名数据集 |
| `/datasetTree/move` | POST | `/datasetTree/move` | ✅ Migrated | 移动数据集 |
| `/datasetTree/delete/{id}` | POST | `/datasetTree/delete/:id` | ✅ Migrated | 删除数据集 |
| `/datasetTree/perDelete/{id}` | POST | `/datasetTree/perDelete/:id` | ✅ Migrated | 预删除检查 |
| `/datasetTree/barInfo/{id}` | GET | `/datasetTree/barInfo/:id` | ✅ Migrated | 获取栏信息 |
| `/datasetTree/exportDataset` | POST | `/datasetTree/exportDataset` | 🌉 Proxied | 导出数据集(存根) |

---

## 3. DatasetData APIs

**Java Controller**: `io.dataease.dataset.server.DatasetDataServer`
**Go Handler**: `DatasetHandler` (compatibility_bridge_handler.go)

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/datasetData/previewData` | POST | `/datasetData/previewData` | ✅ Migrated | 预览数据集数据 |
| `/datasetData/tableField` | POST | `/datasetData/tableField` | ✅ Migrated | 获取表字段 |
| `/datasetData/previewSql` | POST | `/datasetData/previewSql` | ✅ Migrated | 预览SQL |
| `/datasetData/getDatasetTotal` | POST | `/datasetData/getDatasetTotal` | ✅ Migrated | 获取数据集总数 |
| `/datasetData/getFieldEnum` | POST | `/datasetData/enumValue` | ✅ Migrated | 获取字段枚举值 |
| `/datasetData/getFieldEnumDs` | POST | `/datasetData/enumValueDs` | ✅ Migrated | 获取数据源字段枚举 |
| `/datasetData/getFieldEnumObj` | POST | `/datasetData/enumValueObj` | ✅ Migrated | 获取枚举值对象 |
| `/datasetData/getFieldValueTree` | POST | - | ⏳ Pending | 获取字段值树 |

---

## 4. Chart/ChartData APIs

**Java Controller**: `io.dataease.chart.server.ChartViewServer`, `io.dataease.chart.server.ChartDataServer`
**Go Handler**: `ChartHandler` (compatibility_bridge_handler.go)

### ChartView APIs

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/chart/getData/{id}` | POST | `/chart/getChart/:id` | ✅ Migrated | 获取图表数据 |
| `/chart/getDetail/{id}` | POST | `/chart/getDetail/:id` | ✅ Migrated | 获取图表详情 |
| `/chart/save` | POST | `/chart/save` | ✅ Migrated | 保存图表 |
| `/chart/listByDQ/{id}/{chartId}` | POST | `/chart/listByDQ/:id/:chartId` | ✅ Migrated | 按维度指标列表 |
| `/chart/checkSameDataSet/{viewIdSource}/{viewIdTarget}` | GET | `/chart/checkSameDataSet/:viewIdSource/:viewIdTarget` | ✅ Migrated | 检查是否同数据集 |
| `/chart/copyField/{id}/{chartId}` | POST | `/chart/copyField/:id/:chartId` | ✅ Migrated | 复制字段 |
| `/chart/deleteField/{id}` | POST | `/chart/deleteField/:id` | ✅ Migrated | 删除字段 |
| `/chart/deleteFieldByChart/{chartId}` | POST | `/chart/deleteFieldByChart/:chartId` | ✅ Migrated | 删除图表字段 |

### ChartData APIs

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/chartData/getData` | POST | `/chartData/getData` | ✅ Migrated | 获取图表数据 |
| `/chartData/getFieldData/{fieldId}/{fieldType}` | POST | `/chartData/getFieldData/:fieldId/:fieldType` | ✅ Migrated | 获取字段数据 |
| `/chartData/getDrillFieldData/{fieldId}` | POST | `/chartData/getDrillFieldData/:fieldId` | ✅ Migrated | 获取钻取字段数据 |
| `/chartData/innerExportDetails` | POST | `/chartData/innerExportDetails` | 🌉 Proxied | 导出详情(存根) |
| `/chartData/innerExportDataSetDetails` | POST | `/chartData/innerExportDataSetDetails` | 🌉 Proxied | 导出数据集详情(存根) |

---

## 5. User APIs

**Java Controller**: `io.dataease.substitute.permissions.user.SubstituteUserServer`
**Go Handler**: `UserHandler` (compatibility_bridge_handler.go)

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/user/info` | GET | `/user/info` | ✅ Migrated | 获取当前用户信息 |
| `/user/personInfo` | GET | `/user/personInfo` | ✅ Migrated | 获取个人信息 |
| `/user/ipInfo` | GET | `/user/ipInfo` | ✅ Migrated | 获取IP信息 |
| `/user/switchLanguage` | POST | `/user/switchLanguage` | ✅ Migrated | 切换语言 |
| `/user/list` | POST | `/user/list` | ✅ Migrated | 用户列表 |
| `/user/create` | POST | `/user/create` | ✅ Migrated | 创建用户 |
| `/user/edit` | POST | `/user/edit` | ✅ Migrated | 编辑用户 |
| `/user/update` | POST | `/user/update` | ✅ Migrated | 更新用户 |
| `/user/delete/{id}` | POST | `/user/delete/:id` | ✅ Migrated | 删除用户 |
| `/user/options` | GET | `/user/options` | ✅ Migrated | 用户选项 |
| `/user/org/option` | GET | `/user/org/option` | ✅ Migrated | 组织用户选项 |
| `/user/byCurOrg` | POST | `/user/byCurOrg` | ✅ Migrated | 当前组织用户 |

---

## 6. Org APIs

**Java Controller**: `io.dataease.substitute.permissions.org.SubstituleOrgServer`, `io.dataease.system.manage.OrgController`
**Go Handler**: `OrgHandler` (compatibility_bridge_handler.go)

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/org/mounted` | POST | `/org/mounted` | ✅ Migrated | 挂载组织 |
| `/org/create` | POST | `/org/create` | ✅ Migrated | 创建组织 |
| `/org/update` | POST | `/org/update` | ✅ Migrated | 更新组织 |
| `/org/delete/{orgId}` | POST | `/org/delete/:orgId` | ✅ Migrated | 删除组织 |
| `/org/list` | GET | `/org/list` | ✅ Migrated | 组织列表 |
| `/org/info/{orgId}` | GET | `/org/info/:orgId` | ✅ Migrated | 组织详情 |
| `/org/tree` | GET | `/org/tree` | ✅ Migrated | 组织树 |
| `/org/checkName` | GET | `/org/checkName` | ✅ Migrated | 检查组织名称 |
| `/org/updateStatus` | POST | `/org/updateStatus` | ✅ Migrated | 更新状态 |
| `/org/children/{parentId}` | GET | `/org/children/:parentId` | ✅ Migrated | 子组织列表 |

---

## 7. License APIs

**Java Controller**: `io.dataease.license.server.LicenseServer`
**Go Handler**: `LicenseHandler`

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/license/update` | POST | `/license/update` | ✅ Migrated | 更新许可证 |
| `/license/validate` | POST | `/license/validate` | ✅ Migrated | 验证许可证 |
| `/license/version` | GET | `/license/version` | ✅ Migrated | 获取版本 |
| `/license/revert` | POST | `/license/revert` | ✅ Migrated | 还原许可证 |

---

## 8. MsgCenter APIs

**Java Controller**: `io.dataease.msgCenter.MsgCenterServer`
**Go Handler**: `MsgCenterHandler`

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/msg-center/count` | GET | `/msg-center/count` | ✅ Migrated | 获取消息计数 |

---

## 9. System APIs

### 系统参数 (SysParameter)

**Java Controller**: `io.dataease.system.server.SysParameterServer`
**Go Handler**: `SystemParamHandler`

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/sysParameter/singleVal` | GET | `/sysParameter/singleVal` | ✅ Migrated | 获取单个参数值 |
| `/sysParameter/saveOnlineMap` | POST | `/sysParameter/saveOnlineMap` | ✅ Migrated | 保存在线地图配置 |
| `/sysParameter/queryOnlineMap` | GET | `/sysParameter/queryOnlineMap` | ✅ Migrated | 查询在线地图配置 |
| `/sysParameter/queryBasicSetting` | GET | `/sysParameter/queryBasicSetting` | ✅ Migrated | 查询基础设置 |
| `/sysParameter/saveBasicSetting` | POST | `/sysParameter/saveBasicSetting` | ✅ Migrated | 保存基础设置 |
| `/sysParameter/defaultSettings` | GET | `/sysParameter/defaultSettings` | ✅ Migrated | 获取默认设置 |
| `/sysParameter/ui` | GET | `/sysParameter/ui` | ✅ Migrated | 获取UI配置 |

### 认证 (Auth)

**Java Controller**: `io.dataease.system.manage.AuthController`, `io.dataease.substitute.permissions.login.SubstituleLoginServer`
**Go Handler**: `AuthHandler`

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/login/localLogin` | POST | `/login/localLogin` | ✅ Migrated | 本地登录 |
| `/logout` | GET | `/logout` | ✅ Migrated | 登出 |
| `/auth/menuResource` | GET | `/auth/menuResource` | ✅ Migrated | 菜单资源 |
| `/auth/busiResource/{flag}` | GET | `/auth/busiResource/:flag` | ✅ Migrated | 业务资源 |

### 角色 (Role)

**Java Controller**: `io.dataease.system.manage.RoleController`
**Go Handler**: `RoleHandler`

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/api/system/role/list` | POST | `/api/role/list` | ✅ Migrated | 角色列表 |
| `/api/system/role/create` | POST | `/api/role/create` | ✅ Migrated | 创建角色 |
| `/api/system/role/update` | POST | `/api/role/update` | ✅ Migrated | 更新角色 |
| `/api/system/role/delete/{roleId}` | POST | `/api/role/delete/:roleId` | ✅ Migrated | 删除角色 |

### 菜单 (Menu)

**Java Controller**: `io.dataease.menu.server.MenuServer`
**Go Handler**: `MenuHandler`

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/menu/list` | GET | `/menu/list` | ✅ Migrated | 菜单列表 |

### 审计 (Audit)

**Java Controller**: `io.dataease.audit.server.AuditController`
**Go Handler**: `AuditHandler`

| Java Path | Method | Go Path | Status | Notes |
|-----------|--------|---------|--------|-------|
| `/api/audit/log` | POST | `/api/audit/log` | ✅ Migrated | 审计日志 |
| `/api/audit/list` | GET | `/api/audit/list` | ✅ Migrated | 审计列表 |
| `/api/audit/user/{userId}` | GET | `/api/audit/user/:userId` | ✅ Migrated | 用户审计 |
| `/api/audit/{id}` | GET | `/api/audit/:id` | ✅ Migrated | 审计详情 |

---

## 10. 待迁移模块

以下模块的 API 端点尚未在 Go 后端实现兼容性桥接：

| 模块 | Java Controller | 优先级 | 备注 |
|------|-----------------|--------|------|
| Visualization | `DataVisualizationServer` | 高 | 仪表板/大屏可视化 |
| Template | `TemplateManageService` | 中 | 模板管理 |
| Export | `ExportCenterServer` | 中 | 导出中心 |
| Map | `MapServer`, `GeoServer` | 中 | 地图相关 |
| Embedded | `EmbeddedServer` | 低 | 嵌入式功能 |
| Share | `XpackShareServer` | 低 | 分享功能 |
| AI | `AiBaseService` | 低 | AI 功能 |

---

## 统计摘要

| 类别 | 数量 |
|------|------|
| ✅ Migrated | 89 |
| 🌉 Proxied (存根) | 8 |
| ⏳ Pending | 1 |
| ❌ Skipped | 0 |
| **总计** | **98** |

---

## 兼容性桥接架构

```
┌─────────────────┐    ┌──────────────────────────┐    ┌─────────────────┐
│   Frontend      │───▶│  Go Backend (Gin)        │───▶│   Database      │
│   (Vue 3)       │    │                          │    │   (MySQL)       │
└─────────────────┘    │  ┌────────────────────┐  │    └─────────────────┘
                       │  │ Compatibility      │  │
                       │  │ Bridge Handler     │  │
                       │  └────────────────────┘  │
                       │           │              │
                       │           ▼              │
                       │  ┌────────────────────┐  │
                       │  │ Service Layer      │  │
                       │  │ (Business Logic)   │  │
                       │  └────────────────────┘  │
                       │           │              │
                       │           ▼              │
                       │  ┌────────────────────┐  │
                       │  │ Repository Layer   │  │
                       │  │ (Data Access)      │  │
                       │  └────────────────────┘  │
                       └──────────────────────────┘
```

## 更新日志

| 日期 | 版本 | 变更说明 |
|------|------|----------|
| 2026-02-18 | 1.0.0 | 初始版本，记录主要 API 迁移状态 |
