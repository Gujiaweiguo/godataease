# Security Fixtures

This directory contains sample fixtures for negative security contract testing.

## Directory Structure

```
security-fixtures/
├── 401-unauthorized/      # Anonymous/expired token scenarios
├── 403-forbidden/         # Insufficient privilege scenarios
├── row-level-bypass/      # Cross-tenant access scenarios
├── column-masking/        # Sensitive column exposure scenarios
└── export-auth-bypass/    # Export/download auth bypass scenarios
```

## Fixture Mapping

| Scenario ID | Category | Fixture File |
|-------------|----------|--------------|
| NS-001 | 401_UNAUTHORIZED | `401-unauthorized/ns-001-anonymous-access.json` |
| NS-002 | 401_UNAUTHORIZED | `401-unauthorized/ns-002-expired-token.json` |
| NS-003 | 403_FORBIDDEN | `403-forbidden/ns-003-low-privilege-admin.json` |
| NS-004 | 403_FORBIDDEN | `403-forbidden/ns-004-delete-no-permission.json` |
| NS-005 | ROW_LEVEL_BYPASS | `row-level-bypass/ns-005-cross-tenant-dashboard.json` |
| NS-006 | ROW_LEVEL_BYPASS | `row-level-bypass/ns-006-cross-tenant-dataset.json` |
| NS-007 | COLUMN_MASKING_FAILURE | `column-masking/ns-007-email-masking.json` |
| NS-008 | COLUMN_MASKING_FAILURE | `column-masking/ns-008-phone-masking.json` |
| NS-009 | EXPORT_AUTH_BYPASS | `export-auth-bypass/ns-009-export-no-auth.json` |
| NS-010 | EXPORT_AUTH_BYPASS | `export-auth-bypass/ns-010-cross-tenant-download.json` |

## Identity Contexts

| Context | Description | Token Required |
|---------|-------------|----------------|
| `anonymous` | No authentication token | No |
| `low_privilege` | Viewer/read-only role | Yes (viewer token) |
| `cross_tenant` | User from different org | Yes (other-org token) |

## Fixture Schema

Each fixture follows the baseline fixture schema with additional security fields:

```json
{
  "scenario_id": "NS-001",
  "apiIdentity": { ... },
  "identityContext": {
    "type": "anonymous",
    "token": null
  },
  "expectedResponse": {
    "status": 401,
    "code": "401001",
    "msg": "Unauthorized"
  },
  "securityAssertions": [
    { "rule": "SEC-401-001", "expected": true }
  ]
}
```

## Usage

1. Load fixtures by scenario ID from the appropriate directory
2. Apply identity context to request headers
3. Execute request against both Java and Go backends
4. Compare responses against expected values
5. Apply security assertion rules to validate results
