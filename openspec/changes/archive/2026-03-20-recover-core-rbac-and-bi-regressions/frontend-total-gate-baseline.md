# Frontend Total-Gate Baseline

This document freezes the current frontend “total gate” chain that determines whether core features appear reachable at all.

## Gate components

### 1. Login page and post-login redirect

- File: `apps/frontend/src/views/login/index.vue`
- Current behavior:
  - `handleLogin()` calls `loginApi(param)`
  - success path sets token/exp/time in `userStore`
  - then immediately executes `router.push({ path: queryRedirectPath })`
- Risk:
  - protected redirect may execute before authorized menus and dynamic routes are fully ready

### 2. Current-user bootstrap

- File: `apps/frontend/src/store/modules/user.ts`
- Current behavior:
  - `setUser()` calls `userInfo()` from `src/api/user.ts`
  - maps returned payload into `uid`, `name`, `oid`, `language`, `token`, `exp`, `time`
- Risk:
  - if `userInfo()` shape changes or arrives later than route generation, downstream permission/bootstrap becomes unstable

### 3. Authorized menu and route bootstrap

- File: `apps/frontend/src/permission.ts`
- Current behavior:
  - `loadAuthorizedRoutes()` calls `getRoleRouters()` from `src/api/common.ts`
  - response is passed to `permissionStore.generateRoutes()`
  - generated routes are mounted with `router.addRoute()`
  - `interactiveStore.initInteractive(true)` is triggered after route generation
- Risk:
  - this file combines login state, menu loading, route generation, permission refresh, unauthorized-vs-404 decisions, and interactive bootstrap in one chain

### 4. Dynamic route store

- File: `apps/frontend/src/store/modules/permission.ts`
- Current behavior:
  - `generateRoutes()` converts backend menus into runtime routes using `generateRoutesFn2`
  - adds catch-all dynamic not-found route
  - `pathValid()` recursively checks current route presence inside generated routers
- Risk:
  - if backend role routers are incomplete or generation order is wrong, `pathValid()` can misclassify healthy features as missing

### 5. Static shell router

- File: `apps/frontend/src/router/index.ts`
- Current behavior:
  - only a limited set of static base routes is always present
  - most system-management and BI business routes depend on dynamic generation
- Risk:
  - if dynamic route generation fails, whole feature domains disappear even though page files still exist

## Current gate hypothesis

The current highest-probability explanation for the “功能都丢了” symptom is:

1. login succeeds
2. redirect occurs
3. current-user / role-router / dynamic-route bootstrap is incomplete or misordered
4. `pathValid()` and permission refresh then classify the route as unauthorized or missing
5. users experience this as lost features across many modules at once

## Phase-1 repair target

The first repair wave should therefore stabilize this order:

1. login success
2. current-user state
3. authorized menu fetch
4. dynamic route generation
5. path validation
6. protected redirect

No module-specific feature recovery should be trusted before this chain is healthy.
