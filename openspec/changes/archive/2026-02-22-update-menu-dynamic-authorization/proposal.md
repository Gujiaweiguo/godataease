# Change: Update menu system to dynamic authorization model

## Why
Current menu behavior in Go migration is incomplete: core navigation partially depends on hardcoded compatibility responses, and there is no complete menu-management capability for role-menu authorization lifecycle. This causes visible menu inconsistency and runtime 404/permission confusion.

## What Changes
- Replace hardcoded compatibility menu output with DB-driven menu composition.
- Introduce complete menu management capability (query/create/update/delete/sort/visibility) based on `core_menu`.
- Introduce role-menu binding model and APIs so menu visibility is determined by role authorization rather than frontend hardcoding.
- Define compatibility endpoint parity for `/api/roleRouter/query` and `/api/auth/menuResource` as dynamic outputs from the same authorization source.
- Define migration and fallback behavior to prevent lockout during rollout.

## Impact
- Affected specs:
  - `menu-management`
  - `permission-config`
  - `role-management`
  - `api-compatibility-bridge`
- Affected code (implementation stage):
  - `apps/backend-go/internal/transport/http/handler/frontend_compat_handler.go`
  - `apps/backend-go/internal/transport/http/handler/menu_handler.go`
  - `apps/backend-go/internal/service/menu_service.go`
  - `apps/backend-go/internal/service/role_service.go`
  - `apps/frontend/src/api/auth.ts`
  - `apps/frontend/src/views/system/*`
