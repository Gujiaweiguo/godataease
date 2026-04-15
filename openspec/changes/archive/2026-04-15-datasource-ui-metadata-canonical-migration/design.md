## Context

This is round 8, the final round of datasource compatibility-to-canonical migration. The first 7 rounds moved 26 routes. Five remain: `/types` (hardcoded type list), `/showFinishPage` (GET preference), `/setShowFinishPage` (POST dismiss), `/latestUse` (recent types), and `/listSyncRecord/:dsId/:page/:limit` (sync record pagination).

All five routes live in `compatibility_bridge_handler.go`. The canonical handler pattern is established in `datasource_handler.go` within `RegisterDatasourceRoutes()`. Frontend API calls are in `apps/frontend/src/api/datasource.ts`. Tests live in `router_test.go` (backend) and `tests/unit/datasource/api.test.ts` (frontend).

## Goals / Non-Goals

**Goals:**
- Add 5 canonical route handlers under `/api/ds/*` in `RegisterDatasourceRoutes()`
- Migrate 5 frontend API calls from legacy `/datasource/*` to canonical `/ds/*` paths
- Add backend router tests for all 5 new endpoints
- Add frontend unit tests for all 5 migrated API functions
- Preserve all compatibility bridge routes untouched

**Non-Goals:**
- Deleting or deprecating compatibility bridge routes
- Adding new service layer methods (all 5 reuse existing service calls)
- Changing response shapes or business logic
- Adding new dependencies

## Decisions

### Decision 1: `/types` changes from POST to GET
The compatibility bridge uses `POST /datasource/types` for what is a read-only static list. The canonical route uses `GET /api/ds/types` to correctly reflect the idempotent, side-effect-free nature of the operation. The frontend call switches from `request.post` to `request.get`.

**Alternative considered:** Keep POST for backward consistency. Rejected because GET is semantically correct for a read operation and matches the pattern of other read-only canonical routes.

### Decision 2: `/showFinishPage` uses GET, `/setShowFinishPage` merges to POST `/showFinishPage`
The compatibility bridge has two separate routes: `GET /datasource/showFinishPage` and `POST /datasource/setShowFinishPage`. The canonical routes map these to `GET /api/ds/showFinishPage` (read preference) and `POST /api/ds/showFinishPage` (write preference), sharing the same path segment but differing by HTTP method.

**Alternative considered:** Use separate path segments like `/api/ds/showFinishPage` and `/api/ds/setShowFinishPage`. Rejected because HTTP method differentiation is idiomatic REST and keeps the API surface smaller.

### Decision 3: Sync record route shortens from `listSyncRecord` to `syncRecord`
The canonical path `POST /api/ds/syncRecord/:dsId/:page/:limit` drops the `list` prefix from the compatibility route `POST /datasource/listSyncRecord/:dsId/:page/:limit`. This follows the established pattern of shorter, cleaner canonical paths.

### Decision 4: All handlers use `defer recoverDatasourceServicePanic(c)`
Consistent with all existing canonical datasource handlers, each new handler wraps its logic with the panic recovery middleware pattern.

### Decision 5: URL parameter validation in syncRecord handler
The syncRecord handler parses `dsId` as int64, `page` as int (minimum 1), and `limit` as int (minimum 1, default 10) directly from URL parameters. Invalid values return explicit error responses rather than silent defaults.

## Risks / Trade-offs

- **[HTTP method change for /types]** Frontend must switch from `request.post` to `request.get`. Mitigation: straightforward change, testable.
- **[Route name shortening for syncRecord]** Frontend URL path changes from `/datasource/listSyncRecord/` to `/ds/syncRecord/`. Mitigation: single string change in API file.
- **[No new service methods]** All handlers reuse existing service calls. Risk: if service method signatures change, both bridge and canonical handlers need updates. Mitigation: this is inherent to the dual-route period and acceptable for the migration window.
