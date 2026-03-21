# Backend Wiring Baseline

This document freezes the current backend wiring for the core RBAC and BI domains.

## Central wiring source

- File: `apps/backend-go/internal/transport/http/router.go`

Repository evidence shows the following backend handlers are still instantiated and wired:

- `AuthHandler`
- `UserHandler`
- `RoleHandler`
- `OrgHandler`
- `MenuHandler`
- `PermissionCompatHandler`
- `DatasourceHandler`
- `DatasetHandler`
- `VisualizationHandler`

## Core handler evidence

### Auth
- `apps/backend-go/internal/transport/http/handler/auth_handler.go`
- Current routes include local login/logout-related behavior via router wiring

### User
- `apps/backend-go/internal/transport/http/handler/user_handler.go`
- Evidence includes list, create, update, delete, reset password compatibility path

### Role
- `apps/backend-go/internal/transport/http/handler/role_handler.go`
- Evidence includes query, page, create, edit, delete, detail, membership operations

### Organization
- `apps/backend-go/internal/transport/http/handler/org_handler.go`
- Evidence includes create, update, delete, list, detail, tree, child-org operations

### Menu
- `apps/backend-go/internal/transport/http/handler/menu_handler.go`
- Evidence includes query, create, update, delete, sort, hidden, detail

### Permission
- `apps/backend-go/internal/transport/http/handler/permission_compat_handler.go`
- Evidence includes menu permission, business permission, role permission save, resource query, target permission compatibility endpoints

### Datasource
- `apps/backend-go/internal/transport/http/handler/datasource_handler.go`
- Evidence includes list and validate operations; compatibility tree behavior is additionally served by compatibility bridge logic

### Dataset
- `apps/backend-go/internal/transport/http/handler/dataset_handler.go`
- Evidence includes tree, fields, preview, previewWithPermission

### Visualization (dashboard / big-screen)
- `apps/backend-go/internal/transport/http/handler/visualization_handler.go`
- Evidence includes detail, list, tree, canvas save/update, move/copy/delete, publish status, name check

## Current interpretation

The backend baseline does **not** indicate that core RBAC or BI domains were removed from the application. Instead:

1. the handlers still exist
2. the central router still wires them
3. the more likely regression source is at the frontend access path, permission semantics, or frontend/backend contract boundary

## Recovery implication

Frontend total-gate recovery should be prioritized before assuming backend re-implementation is necessary. Backend work in the early recovery phases should focus on:

- confirming route reachability for the APIs needed by the frontend bootstrap chain
- identifying any permission compatibility endpoints that are still partial or intentionally unavailable
- aligning frontend assumptions with current backend contract reality
