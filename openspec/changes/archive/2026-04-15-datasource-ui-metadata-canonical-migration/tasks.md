## 1. Backend Canonical Route Handlers

- [ ] 1.1 Add `GET /api/ds/types` handler in `RegisterDatasourceRoutes()` that returns the hardcoded type list (MySQL, PostgreSQL, SQL Server, Oracle, Excel) with `defer recoverDatasourceServicePanic(c)` and `response.Success(c, typesList)` envelope
- [ ] 1.2 Add `GET /api/ds/showFinishPage` handler that extracts userID from JWT via `getCurrentUserID(c)`, calls `h.service.ShowFinishPage(userID)`, and returns result with panic recovery
- [ ] 1.3 Add `POST /api/ds/showFinishPage` handler that extracts userID from JWT via `getCurrentUserID(c)`, calls `h.service.SetShowFinishPage(userID)`, and returns result with panic recovery
- [ ] 1.4 Add `POST /api/ds/latestUse` handler that extracts username from JWT via `getCurrentUsername(c)`, calls `h.service.LatestTypes(username)`, and returns result with panic recovery
- [ ] 1.5 Add `POST /api/ds/syncRecord/:dsId/:page/:limit` handler that parses dsId (int64), page (int, min 1), limit (int, min 1, default 10) from URL params, calls `h.service.ListSyncRecord(dsID, page, limit)`, and returns result with panic recovery and explicit error responses for invalid params

## 2. Backend Router Tests

- [ ] 2.1 Add router test for `GET /api/ds/types` verifying 200 status and response contains the expected type list
- [ ] 2.2 Add router test for `GET /api/ds/showFinishPage` with valid auth verifying 200 status
- [ ] 2.3 Add router test for `POST /api/ds/showFinishPage` with valid auth verifying 200 status
- [ ] 2.4 Add router test for `POST /api/ds/latestUse` with valid auth verifying 200 status
- [ ] 2.5 Add router tests for `POST /api/ds/syncRecord/:dsId/:page/:limit` covering valid params (200), invalid dsId, page below 1, and limit below 1

## 3. Frontend API Migration

- [ ] 3.1 Update `datasourceTypes` function in `apps/frontend/src/api/datasource.ts` from `request.post({ url: '/datasource/types', data })` to `request.get({ url: '/ds/types' })`
- [ ] 3.2 Update `showFinishPage` function from `request.get({ url: '/datasource/showFinishPage' })` to `request.get({ url: '/ds/showFinishPage' })`
- [ ] 3.3 Update `setShowFinishPage` function from `request.post({ url: '/datasource/setShowFinishPage', data })` to `request.post({ url: '/ds/showFinishPage' })`
- [ ] 3.4 Update `latestUse` function from `request.post({ url: '/datasource/latestUse', data })` to `request.post({ url: '/ds/latestUse' })`
- [ ] 3.5 Update `listSyncRecord` function from `request.post({ url: '/datasource/listSyncRecord/' + dsId + '/' + page + '/' + limit })` to `request.post({ url: '/ds/syncRecord/' + dsId + '/' + page + '/' + limit })`

## 4. Frontend Unit Tests

- [ ] 4.1 Add test for `datasourceTypes` verifying `GET /ds/types` is called and response returned correctly
- [ ] 4.2 Add test for `showFinishPage` verifying `GET /ds/showFinishPage` is called and response returned correctly
- [ ] 4.3 Add test for `setShowFinishPage` verifying `POST /ds/showFinishPage` is called and response returned correctly
- [ ] 4.4 Add test for `latestUse` verifying `POST /ds/latestUse` is called and response returned correctly
- [ ] 4.5 Add test for `listSyncRecord` verifying `POST /ds/syncRecord/:dsId/:page/:limit` is called with correct URL construction and response returned correctly

## 5. Verification

- [ ] 5.1 Run backend tests: `cd apps/backend-go && make test` confirming all pass
- [ ] 5.2 Run frontend lint and type check: `cd apps/frontend && npm run lint && npm run ts:check` confirming clean
- [ ] 5.3 Run frontend unit tests: `cd apps/frontend && npm run test:core -- --run` confirming all pass
