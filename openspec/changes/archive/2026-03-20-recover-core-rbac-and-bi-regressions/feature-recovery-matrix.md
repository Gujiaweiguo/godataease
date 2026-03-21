# Feature Recovery Matrix

This matrix freezes the current repository-observed state for the nine feature families reported as “lost”.

## Classification vocabulary

- **route loss**: feature code exists, but frontend route registration or route matching likely blocks reachability
- **menu loss**: feature code exists, but authorized menu generation/visibility likely hides the entry
- **permission mismatch**: frontend/backend authorization semantics likely block a healthy feature path
- **API mismatch**: frontend page/API contract and backend route/response no longer line up
- **page init failure**: feature page exists but likely fails during initial data loading/bootstrap
- **real implementation gap**: the repository shows a concrete missing or intentionally unavailable sub-capability

## Matrix

| Feature family | Frontend evidence | Backend evidence | Current classification | Why this is the most likely current symptom |
|---|---|---|---|---|
| User | `apps/frontend/src/views/system/user/index.vue`, `apps/frontend/src/api/user.ts` | `apps/backend-go/internal/transport/http/handler/user_handler.go` | route loss / menu loss | User page and API wrapper both exist; handler is wired in `router.go`, so a missing feature symptom is more likely caused by menu/route generation than missing backend CRUD |
| Role | `apps/frontend/src/views/system/role/index.vue`, `apps/frontend/src/api/user.ts` role APIs | `apps/backend-go/internal/transport/http/handler/role_handler.go` | route loss / menu loss | Role page and handler exist; recent dynamic route and permission refresh changes are more likely to hide it than remove it |
| Organization | `apps/frontend/src/views/system/org/index.vue`, user/org-related APIs under `src/api/user.ts` | `apps/backend-go/internal/transport/http/handler/org_handler.go` | route loss / page init failure | Organization page exists and handler exists; organization tree/detail pages are especially sensitive to route entry and init API timing |
| Menu | `apps/frontend/src/views/system/menu/index.vue` | `apps/backend-go/internal/transport/http/handler/menu_handler.go` | menu loss / route loss | Menu administration depends directly on authorized route generation and runtime menu wiring, both currently high-risk |
| Permission | `apps/frontend/src/views/system/permission/index.vue`, permission flow in `src/permission.ts` | `apps/backend-go/internal/transport/http/handler/permission_compat_handler.go` | permission mismatch / real implementation gap | Core permission endpoints exist, but target-permission compatibility paths have known explicit non-success behavior, so this domain may contain both access-path issues and true sub-feature gaps |
| Datasource | `apps/frontend/src/views/visualized/data/datasource/index.vue`, `apps/frontend/src/api/datasource.ts` | `apps/backend-go/internal/transport/http/handler/datasource_handler.go` | menu loss / page init failure | Datasource page and handler exist; current symptom is more consistent with entry discovery and initialization breakage than missing backend implementation |
| Dataset | `apps/frontend/src/views/visualized/data/dataset/index.vue`, `apps/frontend/src/api/dataset.ts` | `apps/backend-go/internal/transport/http/handler/dataset_handler.go` | menu loss / page init failure | Dataset view and handler exist; likely blocked by route/menu/interactive/bootstrap issues |
| Dashboard | `apps/frontend/src/views/dashboard/index.vue`, `apps/frontend/src/api/visualization/dataVisualization.ts` | `apps/backend-go/internal/transport/http/handler/visualization_handler.go` | menu loss / permission mismatch / page init failure | Dashboard depends on authorized menu visibility, dynamic route validity, and visualization discovery APIs, so it is especially vulnerable to total-gate regressions |
| Big-screen | `apps/frontend/src/views/data-visualization/index.vue`, `apps/frontend/src/views/visualized/view/screen/index.vue`, `src/api/visualization/dataVisualization.ts` | `apps/backend-go/internal/transport/http/handler/visualization_handler.go` | menu loss / permission mismatch / page init failure | Big-screen routes and handlers exist; current symptom aligns with broken entry/discovery chain rather than deleted implementation |

## Immediate interpretation

The current codebase does **not** support the hypothesis that these nine capabilities were broadly deleted from backend implementation. The repository evidence supports a narrower hypothesis:

1. the frontend total-gate chain is unstable or misaligned
2. dynamic routes and authorized menus are likely not being regenerated/reconciled correctly
3. some permission compatibility paths may contain true remaining gaps, especially in admin-side permission detail workflows

This matrix should be treated as the baseline for the first implementation wave of `recover-core-rbac-and-bi-regressions`.
