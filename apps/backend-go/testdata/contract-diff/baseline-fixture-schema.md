# Baseline Fixture Schema

## Overview

Baseline fixtures are JSON files that capture the expected response shape and behavior of Java API endpoints. They serve as the ground truth for contract diffing during the Go migration, enabling automated comparison between Java responses and Go responses.

## Required Fields

### apiIdentity

Uniquely identifies the API endpoint.

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | API endpoint path (e.g., `/api/dataset/list`) |
| `method` | string | HTTP method: `GET`, `POST`, `PUT`, `DELETE` |
| `owner` | string | Team or module responsible for this API |
| `priority` | string | Migration priority: `P0`, `P1`, `P2` |
| `blockingLevel` | string | Impact level: `critical`, `high`, `medium`, `low` |

### requestContext

Optional but recommended for non-GET endpoints to ensure reproducibility.

| Field | Type | Description |
|-------|------|-------------|
| `body` | object | Request body for POST/PUT/PATCH |
| `params` | object | Query parameters |
| `pathParams` | object | Path variable values |

### standardResponse

The expected response structure following DataEase conventions.

| Field | Type | Description |
|-------|------|-------------|
| `status` | integer | HTTP status code (200, 400, 500, etc.) |
| `code` | string | Business code (e.g., `000000` for success) |
| `msg` | string | Response message |
| `data` | object \| array | Response payload structure |

### versionMetadata

Tracks when and how the baseline was captured.

| Field | Type | Description |
|-------|------|-------------|
| `captured_at` | string | ISO 8601 timestamp of capture |
| `java_version` | string | Java runtime version used |
| `baseline_version` | string | Semver of fixture schema (e.g., `1.0.0`) |
| `commit_sha` | string | Git commit SHA of Java codebase |

## Optional Fields

### headers

Relevant request/response headers that affect behavior.

| Field | Type | Description |
|-------|------|-------------|
| `request` | object | Headers sent to Java backend |
| `response` | object | Headers returned by Java backend |

### notes

Free-form context about this baseline.

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Human-readable description |
| `caveats` | string[] | Known limitations or edge cases |
| `relatedApis` | string[] | Related endpoint paths |

## Validation Rules

1. **All required fields must be present** - Missing top-level fields cause validation failure
2. **path and method must match whitelist entry** - Only documented APIs are valid
3. **baseline_version must be semver** - Format: `MAJOR.MINOR.PATCH`
4. **status must be valid HTTP code** - 100-599 range
5. **code must be string** - Even numeric codes serialized as strings
6. **priority must be valid enum** - One of: `P0`, `P1`, `P2`
7. **blockingLevel must be valid enum** - One of: `critical`, `high`, `medium`, `low`

## Example Fixture

```json
{
  "apiIdentity": {
    "path": "/api/dataset/list",
    "method": "POST",
    "owner": "dataset-team",
    "priority": "P0",
    "blockingLevel": "critical"
  },
  "requestContext": {
    "body": {
      "page": 1,
      "pageSize": 10
    }
  },
  "standardResponse": {
    "status": 200,
    "code": "000000",
    "msg": "success",
    "data": {
      "total": 0,
      "list": []
    }
  },
  "versionMetadata": {
    "captured_at": "2025-01-15T10:30:00Z",
    "java_version": "21.0.2",
    "baseline_version": "1.0.0",
    "commit_sha": "abc123def456"
  },
  "headers": {
    "request": {
      "Content-Type": "application/json"
    }
  },
  "notes": {
    "description": "Dataset list endpoint baseline",
    "caveats": ["Empty list is valid response"],
    "relatedApis": ["/api/dataset/detail"]
  }
}
```
