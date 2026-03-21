# Operational Recovery Matrix

This matrix freezes the currently governed operational recovery surface for `recover-export-audit-operational-flows`.

The active recovery families in this follow-up batch are:
- Export-center
- Audit

## Classification vocabulary

- **route/access regression**: route or auth-path wiring allows an operational page or API path to degrade into HTML fallback, missing route, or unauthorized mismatch
- **API mismatch**: frontend caller and backend route/response contract no longer align
- **page-init failure**: page shell loads but the first governed query/init path fails to produce explicit usable state
- **state-sync failure**: frontend state or query polling no longer reflects the current backend result correctly
- **real implementation gap**: required operational behavior is genuinely missing rather than drifted

## Matrix

| Flow family | Governed flow slice | Frontend caller / entry path | Backend owner | Current classification | Verification surface | Missing verification target | Current recovery state |
|---|---|---|---|---|---|---|---|
| Export-center | Export task records query, paged task query, retry entry, and drawer initialization | `apps/frontend/src/views/visualized/data/dataset/ExportExcel.vue`; callers in `apps/frontend/src/api/dataset.ts` for `/api/exportCenter/exportTasks/records`, `/api/exportCenter/exportTasks/:status/:page/:size`, `/api/exportCenter/retry/:id` | `apps/backend-go/internal/transport/http/handler/export_handler.go` → `apps/backend-go/internal/service/export_service.go` → export repositories and permission service | route/access regression, API mismatch, page-init failure | `apps/frontend/tests/unit/dataset/exportExcel.test.ts`, `apps/frontend/tests/unit/dataset/api.test.ts`, `apps/backend-go/internal/transport/http/router_test.go` (`TestRegisterRoutes_ExportCenterRoutesRequireAuthentication`) | Add direct download-path verification and a route-level operational smoke beyond the frontend unit/page-init slice | Query/retry callers repaired to `/api/exportCenter/*`; unauthenticated query paths now return explicit `401 + 20001` |
| Audit | Audit page mount query, audit list query, audit detail-read, and route auth semantics | `apps/frontend/src/views/audit/index.vue`; callers in `apps/frontend/src/api/audit.ts`; governed routes `/api/audit/list` and `/api/audit/:id` | `apps/backend-go/internal/transport/http/handler/audit_handler.go` → `apps/backend-go/internal/service/audit_service.go` → audit repositories | route/access regression, page-init failure | `apps/frontend/tests/unit/audit/index.test.ts`, `apps/backend-go/internal/transport/http/handler/audit_handler_test.go`, `apps/backend-go/internal/transport/http/router_test.go` (`TestRegisterRoutes_AuditRoutesRequireAuthentication`) | Add route-level smoke or higher-level UI path verification beyond current unit/handler coverage | Audit list/detail routes now require auth; detail-read preserves explicit not-found semantics; page mount query is covered |

## Failing-or-missing verification targets

- **Export-center**
  - verify download path semantics and explicit non-success handling after auth
  - add one route-level operational smoke beyond current unit coverage
- **Audit**
  - add route-level or smoke verification for page entry plus detail-read from the UI path
  - extend explicit non-success coverage beyond current query-init and detail-not-found checks if later work broadens scope

## Current interpretation

This follow-up batch has already repaired the smallest real operational defects in export-center and audit. The remaining gaps are higher-level stability or route-smoke hardening tasks, not unresolved route/init blockers in the current slice.
