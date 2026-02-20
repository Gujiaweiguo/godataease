# 权限系统 API 文档

## 概述

DataEase 提供完整的用户、组织、角色和权限管理系统，支持菜单权限、资源权限、行级权限和列级权限。

当前 API 由 Go 主线后端（`apps/backend-go`）提供；Java 后端（`legacy/backend-java`）为只读备份。

## 认证

### 1. 用户登录

```
POST /api/auth/login
Content-Type: application/json

Request Body:
{
  "username": "admin",
  "pwd": "Admin168",
  "loginType": 0
}

Response:
{
  "code": "000000",
  "msg": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expireTime": 1706492800000
  }
}
```

### 2. 权限验证

所有需要认证的 API 都需要在请求头中携带 Token：

```http
GET /api/permission/list
Authorization: Bearer {token}
Content-Type: application/json
```

## 组织管理

### 1. 创建组织

```
POST /api/org/create
Authorization: Bearer {token}

Request Body:
{
  "name": "研发部门",
  "parentId": 0,
  "description": "负责产品研发"
}

Response:
{
  "code": "000000",
  "msg": "success",
  "data": {
    "id": 1,
    "name": "研发部门",
    "parentId": 0
    ...
  }
}
```

### 2. 获取组织树

```
GET /api/org/tree
Authorization: Bearer {token}

Response:
{
  "code": "000000",
  "data": [
    {
      "id": 1,
      "name": "研发部门",
      "children": [...]
    }
  ]
}
```

### 3. 添加成员到组织

```
POST /api/org/members/add
Authorization: Bearer {token}

Request Body:
{
  "orgId": 1,
  "userId": 123
}

Response:
{
  "code": "000000",
  "msg": "success"
}
```

## 角色管理

### 1. 创建角色

```
POST /api/role/create
Authorization: Bearer {token}

Request Body:
{
  "name": "数据分析师",
  "description": "可查看所有数据集和仪表板",
  "type": "custom"
}

Response:
{
  "code": "000000",
  "data": {
    "id": 1,
    "name": "数据分析师",
    ...
  }
}
```

### 2. 为角色分配权限

```
POST /api/role/permissions/assign
Authorization: Bearer {token}

Request Body:
{
  "roleId": 1,
  "permIds": [101, 102, 103]
}

Response:
{
  "code": "000000",
  "msg": "success"
}
```

### 3. 为角色分配成员

```
POST /api/role/members/assign
Authorization: Bearer {token}

Request Body:
{
  "roleId": 1,
  "userIds": [123, 124, 125]
}

Response:
{
  "code": "000000",
  "msg": "success"
}
```

## 权限配置

### 1. 获取权限列表

```
GET /api/permission/list
Authorization: Bearer {token}
?type=menu
```

| 权限类型 | type 参数 | 说明 |
|----------|------------|------|
| 菜单权限 | menu | 控制前端菜单可见性 |
| 资源权限 | resource | 控制对特定资源的访问 |
| 导出权限 | export | 控制数据导出功能 |

### 2. 为用户分配权限

```
POST /api/permission/assign
Authorization: Bearer {token}

Request Body:
{
  "userId": 123,
  "permIds": [101, 102, 103]
}

Response:
{
  "code": "000000",
  "msg": "success"
}
```

### 3. 权限继承

权限支持继承规则：

```
继承链：用户 → 组织 → 角色 → 权限

优先级：
1. 用户直接权限（最高）
2. 用户所在组织的权限
3. 用户所属角色的权限
```

## 数据权限

### 1. 行级权限

```
POST /api/rowPermission/create
Authorization: Bearer {token}

Request Body:
{
  "datasetId": 123,
  "name": "销售数据过滤",
  "expression": "${sys.userId} = region_id",
  "whiteList": [1, 2, 3]
}

Response:
{
  "code": "000000",
  "msg": "success"
}
```

### 2. 列级权限

```
POST /api/columnPermission/create
Authorization: Bearer {token}

Request Body:
{
  "datasetId": 123,
  "name": "隐藏敏感字段",
  "type": "mask",
  "expression": "email LIKE '%@%internal.com'",
  "fields": ["email", "phone"]
}

Response:
{
  "code": "000000",
  "msg": "success"
}
```

| 权限类型 | type 值 | 说明 |
|----------|----------|------|
| 禁用 | disable | 完全隐藏字段 |
| 脱敏 | mask | 显示部分内容（如：***@***.com） |
| 明文 | plain | 正常显示 |

### 3. 权限过滤

查询数据时会自动应用权限过滤：

```sql
-- 行级权限示例
SELECT * FROM sales_data
WHERE ${sys.userId} = region_id
AND (id NOT IN (SELECT white_list_id FROM white_lists))

-- 列级权限示例
SELECT
  CASE
    WHEN type = 'mask' THEN '***@***.com'
    ELSE email
  END AS email,
  phone,
  amount
FROM customer_data
```

## 用户管理

### 1. 获取用户列表

```
GET /api/user/list
Authorization: Bearer {token}
?page=1&pageSize=20&keyword=test

Response:
{
  "code": "000000",
  "data": {
    "items": [...],
    "total": 100
  }
}
```

### 2. 创建用户

```
POST /api/user/create
Authorization: Bearer {token}

Request Body:
{
  "username": "testuser",
  "nickname": "测试用户",
  "email": "test@example.com",
  "orgId": 1,
  "roleIds": [2]
}

Response:
{
  "code": "000000",
  "msg": "success",
  "data": {
    "id": 123,
    "username": "testuser",
    ...
  }
}
```

### 3. 批量操作

```
POST /api/user/batch/assign-org
Authorization: Bearer {token}

Request Body:
{
  "userIds": [1, 2, 3],
  "orgId": 1
}

Response:
{
  "code": "000000",
  "msg": "success"
}
```

## 权限检查

### 1. 检查菜单权限

```
POST /api/auth/menuPermission
Authorization: Bearer {token}

Request Body:
{
  "roleId": 1
}

Response:
{
  "code": "000000",
  "data": [101, 102, 103]
}
```

### 2. 检查资源权限

```
POST /api/auth/busPermission
Authorization: Bearer {token}

Request Body:
{
  "roleId": 1
}

Response:
{
  "code": "000000",
  "data": [201, 202, 203]
}
```

## 错误代码

| 代码 | 说明 | HTTP 状态 |
|------|------|----------|
| 000000 | 成功 | 200 |
| 401000 | Token 为空 | 401 |
| 403000 | 权限不足 | 403 |
| 500000 | 服务器错误 | 500 |

## 安全建议

1. **Token 存储**：
   - 前端存储 Token 应使用 localStorage（需加密）或 SessionStorage
   - Token 应该在 HTTPS 环境下传输

2. **权限最小化原则**：
   - 只授予必要的权限
   - 定期审查权限分配
   - 使用角色和组来简化管理

3. **数据隔离**：
   - 不同组织的数据应该严格隔离
   - 使用行级权限过滤敏感数据

4. **审计日志**：
   - 记录所有权限变更
   - 记录所有数据访问
   - 定期审查日志

## 最佳实践

1. **使用角色而非直接分配权限**
2. **利用继承机制减少重复配置**
3. **对于批量操作使用事务**
4. **实现权限缓存以提高性能**
5. **在 UI 中提供清晰的权限配置界面**
