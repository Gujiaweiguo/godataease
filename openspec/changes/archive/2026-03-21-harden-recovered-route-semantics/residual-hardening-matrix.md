# Residual Hardening Matrix

This matrix freezes the remaining non-blocking hardening work surfaced by the archived recovery batches.

## Scope

Active hardening families:
- Datasource
- Dashboard
- Big-screen
- Export-center
- Audit

Excluded from this change:
- Dataset hardening, unless a new concrete blocker emerges
- Broad refactors or product-level redesign

## Classification vocabulary

- **route/access hardening**: path is already recovered but still lacks a stricter auth or route-boundary guarantee
- **detail semantic hardening**: resource detail or missing-resource behavior is only partially frozen and should be verified closer to the frontend-facing boundary
- **route-smoke hardening**: unit/handler coverage exists, but route-level or smoke-level proof is still missing
- **optional deeper coverage**: non-blocking verification that improves confidence without changing current recovered behavior

## Matrix

| Flow family | Residual hardening slice | Frontend caller / entry path | Backend owner | Current verification surface | Missing hardening target | Current status |
|---|---|---|---|---|---|---|
| Datasource | Forbidden semantics for datasource list aliases | `apps/frontend/src/api/datasource.ts` callers for governed datasource list paths | `apps/backend-go/internal/transport/http/handler/datasource_handler.go` → `datasource_service.go` → `datasource_repo.go` | unauthenticated alias proof in router tests; middleware-level datasource view 401/403 proof; datasource entry/init Playwright smoke | add direct forbidden regression coverage for datasource list aliases so forbidden stays distinguishable from unauthenticated and missing-resource outcomes | residual hardening only; recovered route/init behavior already closed |
| Dashboard | Missing-resource semantics for detail paths at frontend-facing boundary | `apps/frontend/src/api/visualization/dataVisualization.ts` `findById()` and dashboard route/edit path callers | `apps/backend-go/internal/transport/http/handler/visualization_handler.go` → `visualization_service.go` → `visualization_repo.go` | dashboard edit/preview smoke; `findById` payload-consumption proof; service-level not-found coverage | add handler- or route-level proof that missing-resource detail responses stay distinguishable from permission denial | residual hardening only; payload-consumption already verified |
| Big-screen | Deeper detail/edit semantics beyond preview/discovery smoke | `apps/frontend/src/api/visualization/dataVisualization.ts` shared visualization callers; big-screen editor/preview entry paths | `apps/backend-go/internal/transport/http/handler/visualization_handler.go` → `visualization_service.go` → `visualization_repo.go` | preview/discovery smoke; shared `findById` payload-consumption assertions; screen-view auth proof | add deeper detail/edit semantic coverage so big-screen behavior is not inferred only from preview/discovery and shared visualization paths | residual hardening only; preview/discovery already verified |
| Export-center | Download-path and route-level hardening | `apps/frontend/src/views/visualized/data/dataset/ExportExcel.vue`; `apps/frontend/src/api/dataset.ts` export-center callers | `apps/backend-go/internal/transport/http/handler/export_handler.go` → `export_service.go` and export permission service | unit coverage for query/retry callers and page-init; router auth proof for `/api/exportCenter/exportTasks/*`; live 401 probes | add explicit download-path verification and one route-level smoke beyond unit coverage | residual hardening only; query/retry recovery already closed |
| Audit | Route-level page-entry/detail smoke beyond current unit/handler coverage | `apps/frontend/src/views/audit/index.vue`; `apps/frontend/src/api/audit.ts` | `apps/backend-go/internal/transport/http/handler/audit_handler.go` → `audit_service.go` → audit repositories | unit coverage for page mount query-init; handler coverage for create/list/detail/not-found; router auth proof for `/api/audit/*`; live 401 probes | add route-level smoke or browser-level proof for audit page entry plus detail read | residual hardening only; route/auth semantics already repaired |

## Current interpretation

All entries in this matrix describe post-recovery hardening work. None of them are required to restore baseline usability; they exist to tighten semantics and route-level confidence around already-recovered flows.
