## Why

Go backend migration has completed the login and core menu bootstrap path, but several legacy frontend permission-management and dashboard-tree APIs still return `404`, causing role/menu configuration pages and dashboard resource operations to fail. This change is needed now to restore operational parity for critical admin and visualization workflows before full cutover hardening.

## What Changes

- Add missing compatibility endpoints for auth/permission operations used by role and menu management pages.
- Add system-role legacy path aliases so frontend calls under `/system/role/*` remain compatible with Go handlers.
- Add missing visualization compatibility endpoints required by dashboard/screen resource-tree operations.
- Define a governed endpoint matrix and acceptance checks to prevent future drift between frontend calls and backend route coverage.

## Capabilities

### New Capabilities
- `go-compat-endpoint-coverage`: Govern and enforce endpoint coverage parity between frontend critical flows and Go backend compatibility routes.

### Modified Capabilities
- `api-compatibility-bridge`: Expand compatibility bridge requirements to include permission and visualization legacy endpoints.
- `permission-config`: Require role/menu permission APIs used by frontend permission management pages to be available in Go runtime.
- `visualization-management`: Require visualization resource-tree compatibility APIs for dashboard/screen operations.
- `role-management`: Require legacy `/system/role/*` compatibility mapping to canonical Go role operations.

## Impact

- Affected backend code:
  - `apps/backend-go/internal/transport/http/router.go`
  - `apps/backend-go/internal/transport/http/handler/frontend_compat_handler.go`
  - `apps/backend-go/internal/transport/http/handler/role_handler.go`
  - `apps/backend-go/internal/transport/http/handler/visualization_handler.go`
  - new compatibility handlers for auth/permission endpoint family
- Affected frontend integrations (verification target):
  - `apps/frontend/src/api/auth.ts`
  - `apps/frontend/src/api/visualization/dataVisualization.ts`
  - `apps/frontend/src/views/system/role/index.vue`
  - `apps/frontend/src/views/common/DeResourceTree.vue`
- Operational impact: reduce runtime `404` in permission management and dashboard workflows; unblock migration parity validation.
