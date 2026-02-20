# Failure Taxonomy

## Purpose

Categorize failure types for actionable feedback in contract diff reports.

## Failure Categories

| Category | Description |
|----------|-------------|
| `STATUS_DIFF` | HTTP status code mismatch (e.g., 200 vs 500, 404 vs 200) |
| `CODE_DIFF` | Response code field mismatch (e.g., "000000" vs "500000") |
| `MSG_DIFF` | Message field semantic difference (e.g., "Success" vs "Error") |
| `PAYLOAD_SCHEMA_DIFF` | Data structure difference (missing/extra fields, type changes) |
| `PAYLOAD_VALUE_DIFF` | Field value difference (same schema, different values) |
| `TIMEOUT` | Request exceeded timeout threshold |
| `CONNECTION_ERROR` | Could not reach backend service |

## Severity Levels

| Level | Meaning | CI Behavior |
|-------|---------|-------------|
| `critical` | Must block | Fails the gate immediately |
| `high` | Should block | Fails unless explicitly waived |
| `normal` | Warning | Logs warning, does not fail |

## Category-to-Severity Mapping

| Category | Default Severity |
|----------|------------------|
| `STATUS_DIFF` | critical |
| `CODE_DIFF` | critical |
| `MSG_DIFF` | high |
| `PAYLOAD_SCHEMA_DIFF` | critical |
| `PAYLOAD_VALUE_DIFF` | normal |
| `TIMEOUT` | high |
| `CONNECTION_ERROR` | critical |

## Example Failures

### STATUS_DIFF
```
Java: 200 OK
Go: 500 Internal Server Error
```
**Action**: Backend error handling differs; investigate Go implementation.

### CODE_DIFF
```
Java: {"code": "000000", "msg": "success"}
Go: {"code": "500000", "msg": "Internal error"}
```
**Action**: Business logic diverges; align error handling.

### PAYLOAD_SCHEMA_DIFF
```
Java: {"data": {"id": 1, "name": "test"}}
Go: {"data": {"id": 1}}  // missing "name" field
```
**Action**: Response structure incomplete; add missing fields.

### PAYLOAD_VALUE_DIFF
```
Java: {"data": {"count": 10}}
Go: {"data": {"count": 11}}
```
**Action**: Non-deterministic or logic difference; review if intentional.
