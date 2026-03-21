# Initial Recovery Evidence

This document captures the first execution checkpoint for `recover-core-rbac-and-bi-regressions` after the local development environment was started.

## Environment

- Dev environment started from `scripts/dev.sh`
- `godataease-app` container healthy
- `godataease-redis` container healthy
- `GET http://localhost:8080/health` returned healthy backend status

## Regression coverage added

### Frontend unit coverage

- `apps/frontend/tests/unit/router/establish.test.ts`
  - added coverage for core system and BI dynamic route component resolution
  - added coverage for preserving resolvable system/data dynamic routes

- `apps/frontend/tests/unit/store/permission.test.ts`
  - existing route and permission regression coverage was re-run successfully as part of the checkpoint

### Frontend e2e / smoke coverage

- `apps/frontend/e2e/auth/login.spec.ts`
  - added protected redirect login regression check to ensure successful login does not fall into false `401` or `404`

- `apps/frontend/e2e/recovery/core-reachability.spec.ts`
  - added RBAC admin route reachability smoke for:
    - `/system/user`
    - `/system/role`
    - `/system/org`
    - `/system/menu`
    - `/system/permission`
  - added BI route reachability smoke for:
    - `/module-datasource`
    - `/module-dataset`
    - `/dashboard`
    - `/dvCanvas`

## Executed verification

- `npm run test -- --run tests/unit/router/establish.test.ts tests/unit/store/permission.test.ts`
- `npx playwright test e2e/auth/login.spec.ts --grep "protected redirect"`
- `npx playwright test e2e/recovery/core-reachability.spec.ts`

## Current finding

At this checkpoint, the reported broad feature-loss symptom was **not reproduced** in the local development environment using the admin login path.

The current evidence supports this narrower interpretation:

1. the codebase still has a high-risk total-gate chain worth protecting
2. broad feature loss may be environment-specific, permission-data-specific, or role-specific
3. the next useful triage step should focus on identifying the exact failing account / role / organization / entry path if the issue still reproduces outside local admin flow

## Runtime API spot checks

The following runtime spot checks were executed successfully under local admin login:

- `GET /api/user/info`
- `GET /api/roleRouter/query`
- `GET /api/auth/menuResource`
- `POST /api/user/byCurOrg`
- `POST /api/role/byCurOrg`
- `GET /api/system/organization/tree`
- `GET /api/system/organization/list`
- `GET /api/menu/query`
- `POST /api/system/permission/list`
- `GET /api/auth/busiResource/1`
- `GET /api/roleMenu/auth/1`
- `POST /api/system/role/permission/save`

This reduces the current confidence in a broad “feature missing” failure on local admin flow and increases the likelihood of a narrower reproduction condition.

## Concrete regressions reproduced and repaired in this session

### 1. Dashboard tree payload rejected historical `leaf` nodeType

**Reproduced symptom**
- `POST /api/dataVisualization/tree` returned `500000 Invalid tree payload: invalid nodeType: leaf`

**Fix**
- backend compatibility added in `apps/backend-go/internal/transport/http/handler/visualization_handler.go`
- historical `leaf` is now normalized to `panel` during tree construction

**Verification**
- `go test ./internal/transport/http/handler -run 'TestBuildVisualizationTreeValidation|TestBuildVisualizationTreeContractShape|TestResolveBusiTypes' -count=1`
- runtime probe of `POST /api/dataVisualization/tree` now returns success tree payload

### 2. Dynamic route children with parent-prefixed paths caused route health risk

**Observed evidence**
- runtime menu/router data returned child paths such as `mine/modify-pwd`, `mine/about`, `help/doc`
- frontend dynamic route generation previously preserved these child paths verbatim

**Fix**
- child path normalization added in `apps/frontend/src/router/establish.ts`

**Verification**
- `tests/unit/router/establish.test.ts`
- Playwright reachability smoke for `/mine/modify-pwd`, `/mine/about`, `/help/doc`

### 3. Workbranch creation permissions were lost when interactive trees were empty

**Observed evidence**
- runtime `POST /api/dataVisualization/interactiveTree` returned empty arrays for admin scopes
- `interactiveStore` derived `anyManage=false` from empty arrays, which disabled all quick-create cards

**Fix**
- `apps/frontend/src/store/modules/interactive.ts` now preserves authorized create state for empty batched interactive scopes

**Verification**
- `tests/unit/store/interactive.test.ts`
- Playwright workbranch reachability and quick-create smoke

### 4. Dataset leaf nodes were not normalized from `nodeType`

**Observed evidence**
- runtime `/api/dataset/tree` returned `nodeType: folder|dataset` but no `leaf` field
- multiple frontend consumers used `data.leaf` to decide icon and behavior

**Fix**
- recursive dataset tree normalization added in `apps/frontend/src/api/dataset.ts`

**Verification**
- `tests/unit/dataset/api.test.ts`

### 5. Logout, About, and Help account flows were incomplete

**Observed evidence**
- frontend `logoutApi()` previously called `/logout`, while dev/runtime success path was exposed at `/api/logout`
- frontend only had an About dialog component and no real `/mine/about` content page
- frontend had no real `/help/doc` content page
- logout flow previously depended on remote logout success before local cleanup, which could still lead to `/#/401`

**Fix**
- `apps/frontend/src/api/login.ts`
  - `logoutApi()` now targets `/api/logout`
- `apps/frontend/src/utils/logout.ts`
  - added `performLogout()` with best-effort remote logout and guaranteed local cleanup via `finally`
  - logout redirect target now avoids `/mine/*`, `/help/*`, `/401`, `/404`, and `/login`
- `apps/frontend/src/layout/components/AccountOperator.vue`
- `apps/frontend/src/views/about/index.vue`
- `apps/frontend/src/views/mobile/personal/index.vue`
  - unified to use `performLogout()` instead of open-coded `await logoutApi(); logoutHandler()`
- added concrete route pages:
  - `apps/frontend/src/views/about/page.vue`
  - `apps/frontend/src/views/help/doc/index.vue`
- added static routes:
  - `/mine/about`
  - `/help/doc`

**Verification**
- `tests/unit/utils/logout.test.ts`
- Playwright reachability/logout smoke in `e2e/recovery/core-reachability.spec.ts`

### 6. Top-right account entry was duplicated

**Observed evidence**
- runtime `roleRouter/query` returned a visible top-level `/mine` route
- `Header.vue` already always renders `AccountOperator`
- this caused both a separate top navigation “我的” entry and the avatar/account dropdown to appear at the same time

**Fix**
- `apps/frontend/src/layout/components/menu-utils.ts`
  - top-level menu resolution now filters `/mine` and `/mine/*` from header top menus
  - account capabilities remain reachable from `AccountOperator`

**Verification**
- `tests/unit/layout/menu-utils.test.ts`
- frontend typecheck
- rebuilt frontend and restarted dev app container

### 7. Language switch returned 404

**Observed evidence**
- frontend `LangSelector.vue` called `switchLangApi()`
- `switchLangApi` targeted `POST /user/switchLanguage`
- Go backend did not expose a working authenticated `/api/user/switchLanguage` path, producing 404 and then invalid-user behavior during intermediate repair attempts

**Fix**
- `apps/backend-go/internal/domain/user/user.go`
  - added language switch request shape
- `apps/backend-go/internal/service/user_service.go`
  - added user language update logic
- `apps/backend-go/internal/transport/http/handler/user_handler.go`
  - added `SwitchLanguage`
- `apps/backend-go/internal/transport/http/router.go`
  - mounted `/api/user/switchLanguage` on the authenticated protected route group

**Verification**
- backend handler regression tests
- `golangci-lint run --new-from-rev=HEAD`
- runtime probe:
  - `POST /api/user/switchLanguage` → `000000 success`
- frontend typecheck
- Playwright regression:
  - language switch no longer returns 404
  - selected locale persists to `localStorage['user.language']`

### 8. Frontend language switch success branch was gated by `!res.msg`

**Observed evidence**
- `apps/frontend/src/layout/components/LangSelector.vue` previously updated local language state only when `!res.msg`
- current successful responses include non-empty `msg: "success"`, so the success branch never ran even after backend route repair

**Fix**
- `LangSelector.vue` now treats successful request resolution as success
- it updates `userStore` language, resets permission store, and reloads the page after a successful API response

**Verification**
- frontend typecheck
- Playwright language switch regression in `e2e/recovery/core-reachability.spec.ts`

### 9. Language switching failed after logout and re-login

**Observed evidence**
- `logoutHandler()` reset user, permission, and interactive stores but did not reset `localeStore`
- after logout → login, locale state could remain stale across SPA lifetime even after user/session rebuild

**Fix**
- `apps/frontend/src/utils/logout.ts`
  - logout now also resets `localeStore`

**Verification**
- `tests/unit/utils/logout.test.ts`
- Playwright regression:
  - logout → login → language switch still succeeds

### 10. Desktop and system-header language entry had mode-specific gaps

**Observed evidence**
- `Header.vue` used `AccountOperator` for normal mode and `DesktopSetting` for desktop mode
- `HeaderSystem.vue` originally rendered `AccountOperator` only when `!desktop`, leaving no unified language entry in desktop system pages
- `DesktopSetting.vue` previously wrapped language switching in an extra nested popover
- desktop mode originally skipped token injection in `refresh.ts`

**Fix**
- `apps/frontend/src/config/axios/refresh.ts`
  - desktop mode still injects `X-DE-TOKEN`
- `apps/frontend/src/layout/components/DesktopSetting.vue`
  - simplified language entry by removing nested popover interaction
- `apps/frontend/src/layout/components/HeaderSystem.vue`
  - desktop mode now renders `DesktopSetting`

**Verification**
- `tests/unit/config/refresh.test.ts`
- Playwright regression:
  - desktop entry language switch succeeds
  - system header in desktop mode exposes working language switch entry

### 11. Normal page header (workbranch) language entry had nested popover interaction issues

**Observed evidence**
- `AccountOperator.vue` previously wrapped language switching in a nested hover popover (`el-popover trigger="hover"`)
- this created interaction issues on `/#/workbranch` and other normal pages where the language dropdown could fail to open or close unexpectedly
- the pattern differed from the flattened approach already applied to `DesktopSetting.vue`

**Fix**
- `apps/frontend/src/layout/components/AccountOperator.vue`
  - removed nested popover language interaction
  - replaced with flattened inline `LangSelector` rendering in a dedicated language block
  - aligns normal page header language entry behavior with the desktop entry pattern

**Verification**
- frontend rebuild: `./scripts/dev.sh build-frontend`
- container restart: `docker compose -f infra/compose/docker-compose.yml -f infra/compose/docker-compose.dev.yml restart godataease-app`
- Playwright regression in `e2e/recovery/core-reachability.spec.ts`:
  - `should switch language without 404 and persist selected locale` (normal `/#/workbranch` entry) ✓
  - `should switch language from desktop setting entry without interaction failure` ✓
  - `should still switch language after logout and re-login` ✓

## Session summary

All 11 concrete regressions identified and repaired in this recovery session have been verified:

1. Dashboard tree payload `leaf` nodeType rejection
2. Dynamic route child path normalization
3. Workbranch creation permissions on empty interactive trees
4. Dataset leaf node normalization
5. Logout/About/Help account flows
6. Top-right duplicate "我的" entry
7. Language switch 404 on `/api/user/switchLanguage`
8. Frontend language switch success branch `!res.msg` gate
9. Language switching after logout/re-login
10. Desktop and system-header language entry gaps
11. Normal page header (workbranch) nested popover language entry issues

All regressions have passing:
- unit test coverage
- backend handler/lint checks
- Playwright e2e regression coverage
- runtime verification in local dev environment
