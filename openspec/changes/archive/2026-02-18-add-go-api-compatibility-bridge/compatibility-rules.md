# API Compatibility Rules

This document defines the response format, error codes, and data semantics for the Go API compatibility bridge, ensuring consistency with the existing Java backend API.

## 1. Response Structure

### 1.1 Standard Response Format

All API responses MUST follow this JSON structure:

```json
{
  "code": "STRING_CODE",
  "msg": "Human readable message",
  "data": <any type or null>
}
```

### 1.2 Field Specifications

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `code` | `string` | Yes | Response status code (NOT integer) |
| `msg` | `string` | Yes | Human-readable message describing the result |
| `data` | `any` | No | Response payload; may be object, array, primitive, or omitted |

### 1.3 Success Response Example

```json
{
  "code": "000000",
  "msg": "success",
  "data": {
    "id": 1,
    "name": "example"
  }
}
```

### 1.4 Error Response Example

```json
{
  "code": "10001",
  "msg": "Invalid parameter: username cannot be empty"
}
```

**Note:** Error responses typically omit the `data` field or set it to `null`.

---

## 2. Error Code Mapping

### 2.1 Code Categories

| Prefix | Category | Description |
|--------|----------|-------------|
| `0xxxxx` | Client Errors | Request validation, parameters |
| `2xxxxx` | Authentication | Auth token, login issues |
| `4xxxxx` | Server Errors | Internal processing failures |
| `5xxxxx` | Resource Errors | Not found, already exists |
| `7xxxxx` | Authorization | Permission denied, forbidden |

### 2.2 Standard Error Codes

| Code | Meaning | HTTP Status | When to Use |
|------|---------|-------------|-------------|
| `000000` | Success | 200 | Request completed successfully |
| `10001` | Bad Request | 200 | Invalid request parameters |
| `10002` | Parameter Missing | 200 | Required parameter not provided |
| `10003` | Parameter Format Error | 200 | Parameter format validation failed |
| `20001` | Unauthorized | 401 | Missing or invalid authentication |
| `20002` | Token Expired | 401 | Authentication token has expired |
| `20003` | Token Invalid | 401 | Authentication token is malformed |
| `40001` | Internal Error | 200 | Unexpected server-side error |
| `40002` | Database Error | 200 | Database operation failed |
| `40003` | External Service Error | 200 | Third-party service unavailable |
| `50001` | Not Found | 200 | Requested resource does not exist |
| `50002` | Already Exists | 200 | Resource already exists (duplicate) |
| `70001` | Forbidden | 403 | User lacks permission for operation |
| `70002` | Role Insufficient | 403 | User role does not have required privilege |

### 2.3 HTTP Status vs Response Code

**IMPORTANT:** The API uses HTTP 200 for most responses, including errors. The actual error status is in the `code` field.

**Exceptions:**
- `401 Unauthorized`: Used for auth failures that should trigger re-login
- `403 Forbidden`: Used when access is denied and should abort request

**Go Implementation Reference:**
```go
// From backend-go/internal/pkg/response/response.go
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    "000000",
        Message: "success",
        Data:    data,
    })
}

func Unauthorized(c *gin.Context, message string) {
    c.JSON(http.StatusUnauthorized, Response{
        Code:    "20001",
        Message: message,
    })
    c.Abort()
}
```

---

## 3. Empty Data Semantics

### 3.1 Array/List Responses

**Rule:** Empty arrays MUST return `[]`, NOT `null`.

**Correct:**
```json
{
  "code": "000000",
  "msg": "success",
  "data": []
}
```

**Incorrect:**
```json
{
  "code": "000000",
  "msg": "success",
  "data": null
}
```

### 3.2 Object Responses

**Rule:** Empty objects should return `{}` when the resource exists but has no properties. Use `null` only when the resource does not exist.

**Resource exists but empty:**
```json
{
  "code": "000000",
  "msg": "success",
  "data": {}
}
```

**Resource not found:**
```json
{
  "code": "50001",
  "msg": "Resource not found"
}
```

### 3.3 Null Data Rules

| Scenario | `data` value | Notes |
|----------|--------------|-------|
| Success with no return value | Omit field or `null` | DELETE operations typically omit data |
| Empty list | `[]` | NEVER use null for lists |
| Empty object | `{}` | When resource exists but empty |
| Resource not found | Return error code | Do NOT return success with null data |

---

## 4. Pagination Fields

### 4.1 Standard Pagination Response

```json
{
  "code": "000000",
  "msg": "success",
  "data": {
    "list": [...],
    "total": 100,
    "current": 1,
    "size": 10
  }
}
```

### 4.2 Field Names

| Field | Type | Description | Alternative Names (DO NOT USE) |
|-------|------|-------------|-------------------------------|
| `list` | `array` | Array of items in current page | `records`, `items`, `data`, `rows` |
| `total` | `number` | Total number of items | `totalCount`, `totalElements` |
| `current` | `number` | Current page number (1-based) | `page`, `pageNum`, `currentPage` |
| `size` | `number` | Items per page | `pageSize`, `perPage`, `limit` |

### 4.3 Pagination Request Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `current` | `int` | 1 | Page number to retrieve |
| `size` | `int` | 10 | Number of items per page |

**Java Reference (UserController.java):**
```java
Integer current = params.containsKey("current") 
    ? Integer.parseInt(params.get("current").toString()) : 1;
Integer size = params.containsKey("size") 
    ? Integer.parseInt(params.get("size").toString()) : 10;
```

---

## 5. Go-Java Bridge Implementation

### 5.1 Response Structure Mapping

| Java (Spring) | Go (Gin) | JSON Key |
|---------------|----------|----------|
| `Map<String, Object>` with `"code"` | `Response.Code` | `code` |
| `Map<String, Object>` with `"msg"` | `Response.Message` | `msg` |
| `Map<String, Object>` with `"data"` | `Response.Data` | `data` |

### 5.2 Code String Type

**CRITICAL:** The `code` field MUST be a STRING type in JSON, not an integer.

**Correct:**
```json
{"code": "000000", "msg": "success"}
```

**Incorrect:**
```json
{"code": 0, "msg": "success"}
```

### 5.3 Error Handling Pattern

**Go Pattern:**
```go
// Business logic error - returns HTTP 200 with error code
func Error(c *gin.Context, code string, message string) {
    c.JSON(http.StatusOK, Response{
        Code:    code,
        Message: message,
    })
}

// Authentication error - returns HTTP 401
func Unauthorized(c *gin.Context, message string) {
    c.JSON(http.StatusUnauthorized, Response{
        Code:    "20001",
        Message: message,
    })
    c.Abort()
}
```

**Java Pattern:**
```java
// Business logic error - returns HTTP 200 with error code
Map<String, Object> result = new HashMap<>();
result.put("code", "500000");
result.put("msg", "Failed: " + e.getMessage());
return result;
```

---

## 6. Validation Checklist

When implementing new Go API endpoints, verify:

- [ ] Response uses `code` as STRING type
- [ ] Success code is `"000000"`
- [ ] Empty arrays return `[]`, not `null`
- [ ] Pagination uses `list`, `total`, `current`, `size`
- [ ] Error messages are in Chinese when user-facing
- [ ] HTTP 200 used for business errors (code in response)
- [ ] HTTP 401/403 only for auth/permission issues

---

## 7. Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-02-18 | Initial document creation |
