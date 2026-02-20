# Negative Security Contract Matrix

## Overview

This document defines security scenarios for negative testing of the Go backend implementation. Negative security testing validates that unauthorized access attempts are properly blocked and that data isolation mechanisms work correctly.

The goal is to ensure parity between Java and Go backends for security-critical behaviors.

## Scenario Categories

| Category | Description |
|----------|-------------|
| `401_UNAUTHORIZED` | Requests without valid authentication token |
| `403_FORBIDDEN` | Authenticated requests with insufficient privileges |
| `ROW_LEVEL_BYPASS` | Attempts to access rows outside user's scope |
| `COLUMN_MASKING_FAILURE` | Exposure of sensitive columns without masking |
| `EXPORT_AUTH_BYPASS` | Unauthorized export or download operations |

## Scenario Matrix

| scenario_id | category | endpoint | method | identity_context | expected_status | expected_code | blocking_level | notes |
|-------------|----------|----------|--------|------------------|-----------------|---------------|----------------|-------|
| NS-001 | 401_UNAUTHORIZED | `/api/dashboard/list` | GET | anonymous | 401 | `401001` | critical | No token provided |
| NS-002 | 401_UNAUTHORIZED | `/api/dataset/data` | POST | anonymous | 401 | `401001` | critical | Expired token |
| NS-003 | 403_FORBIDDEN | `/api/sys/user/list` | GET | low_privilege | 403 | `403001` | critical | Non-admin accessing admin endpoint |
| NS-004 | 403_FORBIDDEN | `/api/org/delete/{id}` | DELETE | low_privilege | 403 | `403001` | critical | Insufficient delete permission |
| NS-005 | ROW_LEVEL_BYPASS | `/api/dashboard/detail/{id}` | GET | cross_tenant | 403 | `403002` | critical | Access dashboard from another org |
| NS-006 | ROW_LEVEL_BYPASS | `/api/dataset/detail/{id}` | GET | cross_tenant | 403 | `403002` | critical | Access dataset from another org |
| NS-007 | COLUMN_MASKING_FAILURE | `/api/dataset/data` | POST | low_privilege | 200 | `000000` | high | Email column should be masked |
| NS-008 | COLUMN_MASKING_FAILURE | `/api/user/profile` | GET | low_privilege | 200 | `000000` | high | Phone column should be masked |
| NS-009 | EXPORT_AUTH_BYPASS | `/api/dashboard/export/{id}` | GET | anonymous | 401 | `401001` | critical | Export without auth |
| NS-010 | EXPORT_AUTH_BYPASS | `/api/report/download/{id}` | GET | cross_tenant | 403 | `403002` | critical | Download report from another org |

## Critical Scenarios Detail

### NS-001: Anonymous Access to Protected Endpoint

- **Category**: `401_UNAUTHORIZED`
- **Endpoint**: `GET /api/dashboard/list`
- **Expected**: 401 with code `401001`
- **Test**: Request without `Authorization` header

### NS-002: Low-Privilege User Accessing Admin Endpoint

- **Category**: `403_FORBIDDEN`
- **Endpoint**: `GET /api/sys/user/list`
- **Expected**: 403 with code `403001`
- **Test**: Authenticate as viewer role, access admin-only endpoint

### NS-003: Cross-Tenant Data Access

- **Category**: `ROW_LEVEL_BYPASS`
- **Endpoint**: `GET /api/dashboard/detail/{id}`
- **Expected**: 403 with code `403002`
- **Test**: User from org A attempts to access dashboard belonging to org B

### NS-004: Sensitive Column Exposure

- **Category**: `COLUMN_MASKING_FAILURE`
- **Endpoint**: `POST /api/dataset/data`
- **Expected**: 200 with masked email values (e.g., `j***@example.com`)
- **Test**: Request dataset containing PII, verify masking applied

### NS-005: Export Download with Expired Token

- **Category**: `EXPORT_AUTH_BYPASS`
- **Endpoint**: `GET /api/dashboard/export/{id}`
- **Expected**: 401 with code `401001`
- **Test**: Use expired JWT token for export request

## Machine-Readable Format

### JSON Representation

```json
{
  "version": "1.0",
  "scenarios": [
    {
      "scenario_id": "NS-001",
      "category": "401_UNAUTHORIZED",
      "endpoint": "/api/dashboard/list",
      "method": "GET",
      "identity_context": "anonymous",
      "expected_status": 401,
      "expected_code": "401001",
      "blocking_level": "critical"
    },
    {
      "scenario_id": "NS-005",
      "category": "ROW_LEVEL_BYPASS",
      "endpoint": "/api/dashboard/detail/{id}",
      "method": "GET",
      "identity_context": "cross_tenant",
      "expected_status": 403,
      "expected_code": "403002",
      "blocking_level": "critical"
    }
  ]
}
```

## Blocking Levels

| Level | CI Behavior |
|-------|-------------|
| `critical` | Fails gate immediately |
| `high` | Fails unless explicitly waived |

## Related Documents

- [Failure Taxonomy](./failure-taxonomy.md) - Category/severity mapping
- [Gate Thresholds](./gate-thresholds.yaml) - CI pass/fail criteria
