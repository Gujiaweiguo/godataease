## 1. Recovery matrix and baseline freeze

- [x] 1.1 Build a feature recovery matrix for user, role, organization, menu, permission, datasource, dataset, dashboard, and big-screen.
- [x] 1.2 For each feature family, classify the current symptom as route loss, menu loss, permission mismatch, API mismatch, page init failure, or real implementation gap.
- [x] 1.3 Freeze the current frontend total-gate chain: login page, user bootstrap, permission refresh, route generation, and path validation.
- [x] 1.4 Freeze the current backend wiring for auth, menu, user, role, organization, permission, datasource, dataset, and visualization handlers.

## 2. Total gate recovery

- [x] 2.1 Add failing frontend tests that expose login-success but route-unready behavior.
- [x] 2.2 Add failing tests for current-user bootstrap and permission refresh ordering.
- [x] 2.3 Add failing tests for false `401` / `404` behavior caused by missing dynamic route registration.
- [x] 2.4 Repair login redirect sequencing so protected redirects wait for authorized route readiness.
- [x] 2.5 Repair current-user and authorized menu bootstrap so dynamic route generation is reliable.
- [x] 2.6 Repair permission refresh and `pathValid()` semantics so unauthorized and missing-route outcomes stay distinguishable.

## 3. RBAC administration recovery

- [x] 3.1 Verify and recover user management page reachability and initialization APIs.
- [x] 3.2 Verify and recover role management page reachability, pagination, and detail flows.
- [x] 3.3 Verify and recover organization tree and detail access paths.
- [x] 3.4 Verify and recover menu administration visibility and CRUD entry paths.
- [x] 3.5 Verify and recover permission administration entry paths and compatibility APIs used by the frontend.
- [x] 3.6 Add regression tests for the recovered RBAC admin routes and key page-init APIs.

### 3.A Scope-tightening tasks for the system-management closure

- [x] 3.A.1 Audit frontend system-management API usage and classify each call as aligned, compatibility-aliased, mismatched, or missing.
- [x] 3.A.2 Reconcile duplicate frontend system-management API modules so each user/role/org/menu/permission flow has one canonical client path.
- [x] 3.A.3 Verify the menu visibility, route reachability, and page-init closure for user, role, organization, menu, and permission pages after login.
- [x] 3.A.4 Verify that unauthorized outcomes and missing-route outcomes remain distinguishable across system-management entry paths.
- [x] 3.A.5 Document any remaining real implementation gaps separately from route/bootstrap/compatibility regressions.

## 4. Deferred scope

Datasource, dataset, dashboard, big-screen, and other broad broken-feature recovery work are intentionally moved out of this change. They should be tracked in a dedicated stabilization change so this change can close on system-management/RBAC/menu recovery only.

## 5. Regression gates and release readiness

- [x] 5.1 Add or update automated checks for login → menu → route → page-init health.
- [x] 5.2 Add or update smoke coverage for RBAC admin pages.
- [x] 5.3 Add or update smoke coverage for first-level and second-level menu reachability within the recovered system-management scope.
- [x] 5.4 Document remaining real implementation gaps separately from recovered access-path regressions.
- [x] 5.5 Run frontend lint/typecheck/tests and backend tests required for the recovered scope.
