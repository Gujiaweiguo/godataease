# Operational Semantic Regression Matrix

This matrix freezes the current failure-semantics baseline for the operational follow-up batch in `recover-export-audit-operational-flows`.

## Semantic targets

| Condition | Expected semantic | Must not degrade into |
|---|---|---|
| Unauthenticated | explicit `401 + 20001` style non-success at the governed HTTP boundary | silent success, HTML fallback, or business-shaped not-found responses |
| Forbidden | explicit forbidden non-success where permission-gated operational paths deny access | missing route/resource, success, or generic internal error |
| Missing route/resource | explicit not-found or unavailable response at the operational boundary | authorization failure or placeholder success |
| Business/query failure | explicit non-success with diagnosable failure message | misleading blank-success state or swallowed empty result |

## Current operational semantic baseline

| Flow family | Governed path | Unauthenticated proof | Missing-resource proof | Business/query failure proof | Current semantic status |
|---|---|---|---|---|---|
| Export-center | `/api/exportCenter/exportTasks/records`, `/api/exportCenter/exportTasks/:status/:goPage/:pageSize` | `apps/backend-go/internal/transport/http/router_test.go` (`TestRegisterRoutes_ExportCenterRoutesRequireAuthentication`) and live probes returning `401 + 20001` | not yet directly frozen for download/delete paths | frontend unit tests and API wrapper tests confirm query/retry contract usage; unauthenticated success regression is removed | query/retry semantics are explicit; deeper download/delete semantics remain future hardening |
| Audit | `/api/audit/list`, `/api/audit/:id` | `apps/backend-go/internal/transport/http/router_test.go` (`TestRegisterRoutes_AuditRoutesRequireAuthentication`) and live probes returning `401 + 20001` | `apps/backend-go/internal/transport/http/handler/audit_handler_test.go` (`TestAuditHandler_GetAuditLogByID_NotFound`) proves `50001 Audit log not found` | `apps/frontend/tests/unit/audit/index.test.ts` proves failed query init surfaces explicit error message | auth, not-found, and query non-success semantics are all explicit for the current slice |

## Current interpretation

For the recovered export-center/audit slice, the key semantic regressions have been frozen at the current operational boundary. What remains is broader route-smoke and optional hardening, not ambiguity about whether the current paths fail explicitly.
