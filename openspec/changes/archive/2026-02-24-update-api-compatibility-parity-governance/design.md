# Design: API Compatibility Parity Governance

## Overview

This document defines governance rules for migration-scoped compatibility endpoints to eliminate placeholder success semantics and ensure runtime-metadata consistency.

---

## 1. Compatibility Endpoint Status Taxonomy (Task 1.1)

### Status Definitions

| Status | Code Behavior Criteria | Migration Matrix Representation |
|--------|------------------------|--------------------------------|
| `full` | Complete business logic implementation. All request paths return correct data with proper authorization and validation. No delegation to legacy or stub behavior. | `"status": "full"` |
| `partial` | Core business logic implemented but with known gaps (e.g., missing optional fields, limited filter support, pagination edge cases). Gaps must be documented with tickets. | `"status": "partial"` with `gaps: [...]` |
| `stub` | Returns deterministic non-success response indicating feature is not yet available. Must include error code and message explaining unavailability. | `"status": "stub"` with `waiver: {...}` if temporary |
| `missing` | No route registered. Request returns 404 or is not routed through compatibility bridge. | `"status": "missing"` or absent from matrix |

### Code-Behavior Evidence Requirements

For each status claim, the endpoint MUST provide verifiable evidence:

- **`full`**: Integration test covering happy path + edge cases + authorization checks
- **`partial`**: Integration test + documented gap list with linked tickets
- **`stub`**: Unit test verifying deterministic error response
- **`missing`**: Absence from route registration OR explicit 404 test

### Status Transition Rules

```
missing → stub → partial → full
   ↑__________|__________|___________|
              (regression allowed with documented reason)
```

- Transition to `full` requires passing all contract-diff tests for that endpoint
- Transition from `full` to lower status requires approval and rollback plan

---

## 2. Placeholder Success Prohibition Rules (Task 1.2)

### Definition

**Placeholder Success**: A response that returns HTTP 200 with `code: "000000"` (success) but contains:
- Empty or null data when business logic should return meaningful results
- Static/hardcoded values that do not reflect actual system state
- Partial data without indicating incompleteness

### Prohibited Patterns

| Pattern | Example | Why Prohibited |
|---------|---------|----------------|
| Empty success on unimplemented feature | `{"code":"000000","data":null}` for list endpoint | Creates false parity signal |
| Static placeholder data | `{"code":"000000","data":{"items":[]}}` when items exist | Misleads consumers about data availability |
| Success without business logic | Handler returns `response.Success(c, nil)` without calling service | Bypasses actual implementation |

### Required Behavior for Unimplemented Features

When core behavior is not implemented:

1. **Return explicit non-success response**:
   ```go
   // CORRECT
   response.Error(c, "501001", "Feature not implemented: detailed description")
   
   // WRONG
   response.Success(c, nil)  // or empty slice, or static data
   ```

2. **Use appropriate error codes**:
   - `501xxx`: Feature not implemented
   - `503xxx`: Feature temporarily unavailable
   - `500xxx`: Internal error (when implementation exists but failed)

3. **Include actionable message**: Consumer should understand what's missing and expected timeline if known.

### Detection Mechanism

CI gate MUST detect placeholder success patterns:

1. **Static analysis**: Flag handlers that return `Success(c, nil)` or `Success(c, []struct{}{})` without service call
2. **Contract diff**: Compare Go response structure with Java baseline - empty data when Java returns populated data is a drift signal
3. **Runtime assertion**: For `stub` status, verify response code is not `000000`

---

## 3. Error Semantics for Unavailable Features (Task 1.3)

### Error Response Structure

All compatibility endpoints MUST use the standard error contract:

```json
{
  "code": "501001",
  "msg": "Feature not implemented: SQL validation via Calcite",
  "data": null
}
```

### Error Code Taxonomy

| Code Range | Category | Example Codes |
|------------|----------|---------------|
| `501xxx` | Not Implemented | `501001` - Feature not implemented |
| `503xxx` | Temporarily Unavailable | `503001` - Service degraded, retry later |
| `500xxx` | Internal Error | `500000` - Generic internal error |
| `400xxx` | Client Error | `400001` - Invalid request |
| `401xxx` | Authentication Error | `401001` - Unauthorized |
| `403xxx` | Authorization Error | `403001` - Forbidden |

### Required Error Semantics by Status

| Endpoint Status | Required Response | HTTP Status |
|-----------------|-------------------|-------------|
| `stub` | Error with `501xxx` or `503xxx` code | 200 (per Java contract) or 501 |
| `partial` with gap | Success for available paths, error for unavailable | 200 |
| Authorization failure | Error with `401xxx` or `403xxx` | 200 (per Java contract) |

### Example Implementations

```go
// Stub endpoint - feature not implemented
func (h *Handler) CalciteValidate(c *gin.Context) {
    response.Error(c, "501001", "Feature not implemented: Calcite SQL validation")
}

// Partial endpoint - some paths available
func (h *Handler) SyncAPITable(c *gin.Context) {
    if !h.seatunnelClient.Available() {
        response.Error(c, "503001", "Sync service temporarily unavailable")
        return
    }
    // ... actual implementation
}

// Correct authorization error
func (h *Handler) SensitiveOperation(c *gin.Context) {
    if !h.authService.HasPermission(userID, "sensitive:op") {
        response.Error(c, "403001", "Insufficient permissions for this operation")
        return
    }
    // ... actual implementation
}
```

---

## 4. Migration Matrix Schema Extension

The migration matrix MUST be extended to include:

```yaml
endpoints:
  - path: /api/datasource/syncApiTable
    method: POST
    status: stub          # full | partial | stub | missing
    owner: team-datasource
    implemented_at: null  # or commit hash
    waiver:               # only for stub with temporary exemption
      reason: "Awaiting SeaTunnel integration"
      approved_by: tech-lead
      expires: 2026-03-15
    gaps: []              # only for partial
    evidence:
      test_file: test/integration/datasource_sync_test.go
      contract_diff: testdata/contract-diff/reports/2026-02-24-syncApiTable.json
```

---

## 5. CI Gate Integration

### Pre-Merge Checks

1. **Status Drift Detection**: Compare declared status in matrix vs. actual code behavior
2. **Placeholder Success Scanner**: Flag suspicious `Success(c, nil)` patterns
3. **Contract Diff**: For `full` status, verify Go response matches Java baseline

### Blocking Conditions

- PR changes endpoint status without matrix update
- PR introduces placeholder success for previously working endpoint
- PR marks endpoint as `full` without passing contract diff

---

## 6. Waiver Process (Task 3.3)

### When Waivers Are Allowed

- Temporary `stub` status during active development (< 30 days)
- Planned deprecation with consumer migration period
- Emergency hotfix requiring temporary feature disable

### Waiver Requirements

1. **Owner**: Named individual responsible for resolution
2. **Reason**: Clear justification for temporary exemption
3. **Expiry**: Date after which waiver auto-expires
4. **Approval**: Tech lead sign-off in code review

### Expiry Handling

- CI MUST warn 7 days before expiry
- Expired waivers MUST fail CI gate
- Renewal requires new approval with updated justification
