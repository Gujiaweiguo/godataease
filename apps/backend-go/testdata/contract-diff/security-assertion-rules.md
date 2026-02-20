# Security Assertion Rules

## Overview

This document defines assertion rules for security-related contract diff testing. These rules validate authentication, authorization, row-level security, column masking, and export permissions.

---

## 401/403 Assertion Rules

### SEC-401-001: Anonymous Access Must Return 401

| Field | Value |
|-------|-------|
| ID | `SEC-401-001` |
| Description | Requests without valid credentials must return HTTP 401 |
| Expected Status | `401` |
| Expected Code | `UNAUTHORIZED` |
| Expected Msg Pattern | `unauthorized\|token.*required\|authentication.*failed` |

### SEC-401-002: Expired Token Must Return 401

| Field | Value |
|-------|-------|
| ID | `SEC-401-002` |
| Description | Requests with expired JWT tokens must return HTTP 401 |
| Expected Status | `401` |
| Expected Code | `TOKEN_EXPIRED` |
| Expected Msg Pattern | `expired\|token.*invalid` |

### SEC-403-001: Insufficient Privilege Must Return 403

| Field | Value |
|-------|-------|
| ID | `SEC-403-001` |
| Description | Authenticated users without required permissions must return HTTP 403 |
| Expected Status | `403` |
| Expected Code | `FORBIDDEN` |
| Expected Msg Pattern | `permission.*denied\|forbidden\|no.*privilege` |

### Assertion Format

```yaml
assertion:
  expected_status: 401 | 403
  expected_code: string
  expected_msg_pattern: regex
```

---

## Row-Level Assertion Rules

### SEC-ROW-001: Cross-Tenant Data Must Not Be Visible

| Field | Value |
|-------|-------|
| ID | `SEC-ROW-001` |
| Description | API responses must not include data belonging to other tenants |
| Row Count Threshold | `response.rows <= expected_max` |
| Forbidden ID List | `response must not contain ids from forbidden_id_list` |

### SEC-ROW-002: Row Count Must Match Expected Filtered Count

| Field | Value |
|-------|-------|
| ID | `SEC-ROW-002` |
| Description | Row counts must match permission-based filtering rules |
| Assertion | `actual_count == expected_count` |

### Assertion Format

```yaml
assertion:
  row_count_threshold: integer
  forbidden_id_list: [string]
  expected_count: integer
```

---

## Column Masking Assertion Rules

### SEC-MASK-001: Email Columns Must Be Masked

| Field | Value |
|-------|-------|
| ID | `SEC-MASK-001` |
| Description | Email fields must be partially masked (e.g., `a***@example.com`) |
| Masked Column Patterns | `*email*`, `*Email*` |
| Forbidden Patterns | Full email regex match |

### SEC-MASK-002: Phone Columns Must Be Masked

| Field | Value |
|-------|-------|
| ID | `SEC-MASK-002` |
| Description | Phone fields must be partially masked (e.g., `138****5678`) |
| Masked Column Patterns | `*phone*`, `*mobile*`, `*Phone*` |
| Forbidden Patterns | Full phone number regex match |

### Assertion Format

```yaml
assertion:
  masked_column_patterns: [glob]
  forbidden_patterns: [regex]
  mask_type: partial | full
```

---

## Export Auth Assertion Rules

### SEC-EXP-001: Export Without Auth Must Fail

| Field | Value |
|-------|-------|
| ID | `SEC-EXP-001` |
| Description | Export operations require authentication |
| Auth Required | `true` |
| Expected Failure Status | `401` |

### SEC-EXP-002: Cross-User Download Must Fail

| Field | Value |
|-------|-------|
| ID | `SEC-EXP-002` |
| Description | Users cannot download exports created by other users |
| Cross User Blocked | `true` |
| Expected Failure Status | `403` |

### Assertion Format

```yaml
assertion:
  auth_required: true | false
  cross_user_blocked: true | false
  expected_failure_status: 401 | 403
```

---

## Failure Classification

### Blocking Levels

| Level | Description | Action |
|-------|-------------|--------|
| `CRITICAL` | Authentication bypass, data leak | Block deployment immediately |
| `HIGH` | Authorization bypass, missing masking | Block deployment, require fix |
| `MEDIUM` | Inconsistent error codes, partial masking | Warning, review required |
| `LOW` | Minor inconsistency in error messages | Log only, no blocking |

### Classification Rules

```yaml
classification:
  SEC-401-*: CRITICAL
  SEC-403-*: HIGH
  SEC-ROW-*: CRITICAL
  SEC-MASK-*: HIGH
  SEC-EXP-*: HIGH
```

### Failure Output Format

```yaml
failure:
  rule_id: string
  severity: CRITICAL | HIGH | MEDIUM | LOW
  endpoint: string
  message: string
  actual_value: any
  expected_value: any
```
