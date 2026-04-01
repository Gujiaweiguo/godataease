# T7 Permission Scope Note

## Completed in T7
- Replaced menu target compatibility 501 responses with canonical alias behavior backed by existing role-menu authorization storage.
- Added query/save round-trip coverage for menu target compatibility endpoints.
- Replaced `RequireMenuAuth` hard deny behavior with real menu-path lookup plus persisted role-menu authorization checks.
- Populated `role_ids` in authenticated middleware context through repository-backed resolution.
- Applied menu authorization middleware to the governed permission compatibility API group for `/system/permission` semantics.
- Preserved existing resource permission dual-view behavior and verified the menu/resource compatibility paths together.

## Intentionally Not Expanded in T7
- No broad router-wide rollout of menu-auth middleware beyond the governed permission-center API group.
- No new permission-center UI redesign or new menu-target UI flow was introduced.
- No row/column permission runtime work was started here.

## Why This Boundary Is Acceptable
- The highest-value T7 gaps were incomplete compat endpoints and dead/stubbed menu authorization behavior.
- Those gaps are now closed without introducing a new permission model.
- Wider rollout of menu-auth to more route families would materially increase risk and overlaps with later waves.

## Deferred Items
- broader menu-auth rollout outside permission-center APIs
- row/column/whitelist/system-variable semantics (T8)
