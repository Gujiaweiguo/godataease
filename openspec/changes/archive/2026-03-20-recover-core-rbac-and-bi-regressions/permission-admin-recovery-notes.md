## Permission administration recovery notes

### Verified recovered paths

- `MenuPermission` now completes save → reload → echo via role-menu authorization APIs.
- `ResourcePermission` was aligned back to the existing role-based permission contract and now completes save → reload → echo.
- `DataPermission` now has backend pager/save/delete exposure for the frontend page's row/column permission endpoints, and the page successfully triggers both row and column pager APIs for a selected dataset.

### Current DataPermission compatibility subset

The recovered `DataPermission` path is an explicit compatibility subset, not a full legacy-equivalent implementation.

- Row permissions currently support the frontend's simple form as **single-condition user/role rules**.
- `filterType = variable` is still outside the supported subset.
- Row-permission `whiteList` is not persisted because the current Go-side data model available in this repo does not expose matching persisted columns in the active schema.
- Column permissions are recovered as **dataset-level** disable/mask rules using the current Go-side `data_perm_column` model.

### Why this is recorded separately

This change recovered permission-administration entry paths and the frontend APIs they depend on, but it intentionally did not claim full parity with the broader legacy data-permission feature set. Any later work that needs system-variable row rules, persisted whitelist semantics, or more complete legacy parity should be tracked as a separate follow-up implementation task.

## Additional remaining implementation gaps inside recovered system-management scope

- `RoleTab.vue` now correctly loads current role permissions before opening the permission dialog, but it still uses a fixed `size: 100` request and does not expose a real pagination control. This is no longer a reachability problem; it is a UX/feature-completeness gap.
- The recovered system-management pages are now reachable both by direct route and by menu-driven navigation after login. Remaining issues in this scope should therefore be treated as page-level feature gaps, not login/bootstrap/route-loss regressions, unless new evidence contradicts that.
