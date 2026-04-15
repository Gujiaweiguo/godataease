## Why

Seven previous migration rounds have moved 26 datasource routes from the compatibility bridge to canonical `/api/ds/*` paths. Five routes remain: UI metadata endpoints (`/types`, `/showFinishPage`, `/setShowFinishPage`, `/latestUse`) and the sync record listing endpoint (`/listSyncRecord`). These are the last compatibility-only datasource routes. Completing this migration gives the frontend a single, consistent API namespace and removes the last dependency on the legacy bridge for datasource operations.

## What Changes

- Add 5 new canonical handlers in `RegisterDatasourceRoutes()` under `/api/ds/*`:
  - `GET /api/ds/types` returns the hardcoded datasource type list (MySQL, PostgreSQL, SQL Server, Oracle, Excel)
  - `GET /api/ds/showFinishPage` calls `service.ShowFinishPage(userID)` from JWT context
  - `POST /api/ds/showFinishPage` calls `service.SetShowFinishPage(userID)` from JWT context
  - `POST /api/ds/latestUse` calls `service.LatestTypes(username)` from JWT context
  - `POST /api/ds/syncRecord/:dsId/:page/:limit` calls `service.ListSyncRecord(dsID, page, limit)` with URL param parsing
- Migrate 5 frontend API calls in `apps/frontend/src/api/datasource.ts` from legacy `/datasource/*` paths to canonical `/ds/*` paths
- Add backend router tests for all 5 new canonical endpoints
- Add frontend unit tests for all 5 migrated API functions
- Compatibility bridge routes are preserved (not deleted)

## Capabilities

### New Capabilities
- `datasource-ui-metadata-canonical`: Canonical API handlers for datasource type listing, finish-page preferences, recent-use tracking, and sync record pagination

### Modified Capabilities
- `datasource-management`: Frontend API paths migrate from compatibility bridge to canonical `/ds/*` paths

## Impact

- **Backend**: `apps/backend-go/internal/transport/http/handler/datasource_handler.go` adds 5 route registrations
- **Backend tests**: `apps/backend-go/internal/transport/http/router_test.go` adds 5 test cases
- **Frontend API**: `apps/frontend/src/api/datasource.ts` updates 5 function URL paths
- **Frontend tests**: `apps/frontend/tests/unit/datasource/api.test.ts` adds 5 test cases
- **No breaking changes**: Compatibility bridge routes remain active
- **Rollback**: Revert frontend URL paths to legacy routes; canonical handlers are additive
