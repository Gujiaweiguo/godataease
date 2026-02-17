# 审计日志 API 文档

## 概述

审计日志 API 提供完整的操作审计功能，用于记录和追踪系统中的所有用户操作、权限变更、数据访问和系统操作。

## 基础信息

- **Base URL**: `/api/audit`
- **认证**: 需要有效的 JWT Token
- **响应格式**: JSON

---

## 1. 查询审计日志

### 请求

```http
GET /api/audit/list
```

### 查询参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|--------|------|
| userId | Long | 否 | 用户 ID |
| username | String | 否 | 用户名（模糊匹配）|
| actionType | String | 否 | 操作类型：USER_ACTION, PERMISSION_CHANGE, DATA_ACCESS, SYSTEM_CONFIG |
| resourceType | String | 否 | 资源类型：USER, ORGANIZATION, ROLE, PERMISSION, DATASET, DASHBOARD |
| organizationId | Long | 否 | 组织 ID |
| startTime | DateTime | 否 | 开始时间（格式：YYYY-MM-DD HH:mm:ss）|
| endTime | DateTime | 否 | 结束时间（格式：YYYY-MM-DD HH:mm:ss）|
| page | Integer | 否 | 页码，默认 1 |
| pageSize | Integer | 否 | 每页数量，默认 10 |

### 响应

```json
{
  "code": "000000",
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "userId": 1,
        "username": "admin",
        "actionType": "USER_ACTION",
        "actionName": "CREATE_USER",
        "resourceType": "USER",
        "resourceId": 100,
        "resourceName": "testuser",
        "operation": "CREATE",
        "status": "SUCCESS",
        "failureReason": null,
        "ipAddress": "192.168.1.100",
        "userAgent": "Mozilla/5.0...",
        "beforeValue": null,
        "afterValue": null,
        "organizationId": 1,
        "createTime": "2025-01-30T10:00:00"
      }
    ],
    "total": 100
  }
}
```

---

## 2. 获取用户审计日志

### 请求

```http
GET /api/audit/user/{userId}
```

### 路径参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|--------|------|
| userId | Long | 是 | 用户 ID |

### 响应

```json
{
  "code": "000000",
  "msg": "success",
  "data": [
    {
      "id": 1,
      "username": "admin",
      "actionType": "USER_ACTION",
      "actionName": "CREATE_USER",
      "resourceType": "USER",
      "operation": "CREATE",
      "status": "SUCCESS",
      "createTime": "2025-01-30T10:00:00"
    }
  ]
}
```

---

## 3. 获取审计日志详情

### 请求

```http
GET /api/audit/{id}
```

### 路径参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|--------|------|
| id | Long | 是 | 审计日志 ID |

### 响应

```json
{
  "code": "000000",
  "msg": "success",
  "data": {
    "id": 1,
    "userId": 1,
    "username": "admin",
    "actionType": "USER_ACTION",
    "actionName": "CREATE_USER",
    "resourceType": "USER",
    "resourceId": 100,
    "resourceName": "testuser",
    "operation": "CREATE",
    "status": "SUCCESS",
    "failureReason": null,
    "ipAddress": "192.168.1.100",
    "userAgent": "Mozilla/5.0...",
    "beforeValue": null,
    "afterValue": null,
    "organizationId": 1,
    "createTime": "2025-01-30T10:00:00"
  }
}
```

---

## 4. 导出审计日志

### 请求

```http
POST /api/audit/export?format=csv
```

### 查询参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|--------|------|
| format | String | 否 | 导出格式：csv（默认）或 json |

### 请求体

```json
[1, 2, 3, 4, 5]
```

### 响应

```json
{
  "code": "000000",
  "msg": "success",
  "data": "Export completed"
}
```

### 导出格式

#### CSV 格式

```csv
id,username,actionType,actionName,resourceType,resourceName,operation,status,ipAddress,createTime
1,admin,USER_ACTION,CREATE_USER,USER,testuser,CREATE,SUCCESS,192.168.1.100,2025-01-30T10:00:00
```

#### JSON 格式

```json
[
  {
    "id": 1,
    "username": "admin",
    "actionType": "USER_ACTION",
    "actionName": "CREATE_USER",
    "resourceType": "USER",
    "resourceName": "testuser",
    "operation": "CREATE",
    "status": "SUCCESS",
    "ipAddress": "192.168.1.100",
    "createTime": "2025-01-30T10:00:00"
  }
]
```

---

## 5. 清理过期审计日志

### 请求

```http
DELETE /api/audit/retention?days=90
```

### 查询参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|--------|------|
| days | Integer | 否 | 清理多少天前的日志，默认 90 |

### 响应

```json
{
  "code": "000000",
  "msg": "success",
  "data": "Deleted logs older than 90 days"
}
```

---

## 6. 创建审计日志

### 请求

```http
POST /api/audit/log
```

### 请求体

```json
{
  "userId": 1,
  "username": "admin",
  "actionType": "USER_ACTION",
  "actionName": "CUSTOM_ACTION",
  "resourceType": "DATASET",
  "resourceId": 100,
  "resourceName": "sales_data",
  "operation": "EXPORT",
  "status": "SUCCESS",
  "failureReason": null,
  "ipAddress": "192.168.1.100",
  "userAgent": "Mozilla/5.0...",
  "organizationId": 1,
  "beforeValue": null,
  "afterValue": "1000 records exported"
}
```

### 响应

```json
{
  "code": "000000",
  "msg": "success",
  "data": {
    "id": 1,
    "createTime": "2025-01-30T10:00:00"
  }
}
```

---

## 常量定义

### 操作类型 (actionType)

| 值 | 说明 |
|------|------|
| USER_ACTION | 用户操作 |
| PERMISSION_CHANGE | 权限变更 |
| DATA_ACCESS | 数据访问 |
| SYSTEM_CONFIG | 系统配置 |

### 资源类型 (resourceType)

| 值 | 说明 |
|------|------|
| USER | 用户 |
| ORGANIZATION | 组织 |
| ROLE | 角色 |
| PERMISSION | 权限 |
| DATASET | 数据集 |
| DASHBOARD | 仪表板 |
| EMBEDDED_APP | 嵌入式应用 |

### 操作类型 (operation)

| 值 | 说明 |
|------|------|
| CREATE | 创建 |
| UPDATE | 更新 |
| DELETE | 删除 |
| EXPORT | 导出 |
| LOGIN | 登录 |
| LOGOUT | 登出 |
| VIEW | 查看 |

### 状态 (status)

| 值 | 说明 |
|------|------|
| SUCCESS | 成功 |
| FAILED | 失败 |

---

## 错误码

| 错误码 | 说明 |
|----------|------|
| 000000 | 成功 |
| 40001 | 参数错误 |
| 40002 | 未找到审计日志 |
| 50001 | 服务器内部错误 |

---

## 使用示例

### 查询最近失败的操作

```bash
curl -X GET "http://localhost:8100/api/audit/list?status=FAILED&page=1&pageSize=20" \
  -H "Authorization: Bearer <token>"
```

### 查询特定用户的操作

```bash
curl -X GET "http://localhost:8100/api/audit/list?userId=1&actionType=USER_ACTION" \
  -H "Authorization: Bearer <token>"
```

### 查询日期范围内的操作

```bash
curl -X GET "http://localhost:8100/api/audit/list?startTime=2025-01-01 00:00:00&endTime=2025-01-31 23:59:59" \
  -H "Authorization: Bearer <token>"
```

### 导出审计日志

```bash
curl -X POST "http://localhost:8100/api/audit/export?format=csv" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d "[1,2,3,4,5]"
```

### 清理90天前的日志

```bash
curl -X DELETE "http://localhost:8100/api/audit/retention?days=90" \
  -H "Authorization: Bearer <token>"
```
