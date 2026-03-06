# Menu Management API Documentation

## Overview
本文档描述菜单管理与角色菜单授权相关的 API 端点。

---

## Menu Management Endpoints

### 1. Query Menu Tree
**Endpoint**: `GET /api/menu/query`  
**Description**: 获取层级菜单树（全部菜单，不按角色过滤）  
**Auth**: Required  

**Response**:
```json
{
  "code": "000000",
  "data": [
    {
      "id": 1,
      "pid": 0,
      "name": "系统管理",
      "path": "/system",
      "component": "Layout",
      "icon": "system",
      "menuSort": 1,
      "hidden": false,
      "auth": true,
      "children": [...]
    }
  ]
}
```

---

### 2. Create Menu
**Endpoint**: `POST /api/menu/create`  
**Description**: 创建新菜单项  
**Auth**: Required  

**Request Body**:
```json
{
  "pid": 0,
  "type": 0,
  "name": "菜单名称",
  "component": "views/example/index",
  "menuSort": 1,
  "icon": "example",
  "path": "/example",
  "hidden": false,
  "inLayout": true,
  "auth": true
}
```

**Response**:
```json
{
  "code": "000000",
  "data": 123  // 新创建的菜单 ID
}
```

---

### 3. Update Menu
**Endpoint**: `POST /api/menu/edit`  
**Description**: 更新菜单项  
**Auth**: Required  

**Request Body**:
```json
{
  "id": 123,
  "pid": 0,
  "type": 0,
  "name": "更新后的名称",
  "component": "views/example/index",
  "menuSort": 2,
  "icon": "example",
  "path": "/example",
  "hidden": false,
  "inLayout": true,
  "auth": true
}
```

**Response**:
```json
{
  "code": "000000",
  "data": null
}
```

---

### 4. Delete Menu
**Endpoint**: `POST /api/menu/delete`  
**Description**: 删除菜单项（如有子菜单则拒绝）  
**Auth**: Required  

**Request Body**:
```json
{
  "id": 123
}
```

**Response**:
```json
{
  "code": "000000",
  "data": null
}
```

**Error Cases**:
- 有子菜单时返回错误：`该菜单下存在子菜单，无法删除`

---

### 5. Update Menu Sort
**Endpoint**: `POST /api/menu/updateSort`  
**Description**: 更新菜单排序  
**Auth**: Required  

---

### 6. Update Menu Visibility
**Endpoint**: `POST /api/menu/updateHidden`  
**Description**: 更新菜单可见性  
**Auth**: Required  

---

## Role-Menu Authorization Endpoints

### 1. Get Role Menu Authorization
**Endpoint**: `GET /api/roleMenu/auth/:roleId`  
**Description**: 获取指定角色的菜单授权列表  
**Auth**: Required  

**Response**:
```json
{
  "code": "000000",
  "data": [1, 2, 3, 5, 8]  // 菜单 ID 列表
}
```

---

### 2. Save Role Menu Authorization
**Endpoint**: `POST /api/roleMenu/auth`  
**Description**: 保存角色菜单授权（幂等操作）  
**Auth**: Required  

**Request Body**:
```json
{
  "roleId": 2,
  "menuIds": [1, 2, 3, 5, 8]
}
```

**Response**:
```json
{
  "code": "000000",
  "data": null
}
```

**Error Cases**:
- 角色不存在：`角色不存在`
- 菜单不存在：`菜单不存在`

---

## Compatibility Endpoints

### 1. Get Role Routers (Legacy)
**Endpoint**: `GET /api/roleRouter/query`  
**Description**: 兼容旧版前端，返回角色路由配置  
**Auth**: Required  

**Note**: 内部使用 `MenuService.Query()`，返回格式与旧版兼容。

---

### 2. Get Menu Resource (Legacy)
**Endpoint**: `GET /api/auth/menuResource`  
**Description**: 兼容旧版前端，返回菜单资源  
**Auth**: Required  

**Note**: 内部使用 `MenuService.Query()`，返回格式与旧版兼容。

---

## Authorization Middleware

### Menu Authorization Check
所有菜单相关端点都会通过 `MenuAuthMiddleware` 检查用户是否有权访问该菜单路由。

**Error Response**:
```json
{
  "code": "403000",
  "message": "无权访问该菜单"
}
```

---

## Configuration

### Fallback Mode
在 `configs/config.yaml` 中配置：
```yaml
menu:
  hardcoded_fallback: false
```

- `false`（默认）: 使用数据库动态菜单
- `true`: 使用硬编码菜单（仅用于紧急回退）

---

## Changes from Previous Version

### New Endpoints
1. `GET /api/menu/query` - 查询层级菜单树
2. `POST /api/menu/create` - 创建菜单
3. `POST /api/menu/edit` - 更新菜单
4. `POST /api/menu/delete` - 删除菜单
5. `GET /api/roleMenu/auth/:roleId` - 获取角色菜单授权
6. `POST /api/roleMenu/auth` - 保存角色菜单授权

### Modified Endpoints
1. `GET /api/roleRouter/query` - 重构为使用 `MenuAssemblyService`
2. `GET /api/auth/menuResource` - 重构为使用 `MenuAssemblyService`

### Removed
- 硬编码菜单数组（移除，改为从数据库动态加载）

---

## Migration Notes
- 所有旧端点保持兼容，前端无需修改
- 新增端点用于更精细的菜单管理
- 建议逐步迁移到新端点

---

**Document Version**: 1.0  
**Last Updated**: 2026-03-06  
**Author**: Gujiaweiguo
