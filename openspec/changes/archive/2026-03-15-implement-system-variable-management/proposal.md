## Why

The frontend already depends on a full `/sysVariable/*` API surface for variable and variable-value management, but the Go mainline has no corresponding module. This leaves the system variable UI unusable on Go mainline and blocks parity for a distinct configuration domain that is not covered by existing system parameter APIs.

## What Changes

- Add a dedicated Go system variable management module for variable CRUD, variable detail/query, and variable-value management.
- Define parity-safe request and response behavior for variable paging, detail lookup, and value selection flows used by the frontend.
- Separate system variable management from existing `sysParameter` platform configuration semantics.
- Add backend verification coverage for variable and variable-value management paths.

## Capabilities

### New Capabilities
- `system-variable-management`: Manage system variables and their selectable values through the `/sysVariable/*` API surface required by the frontend.

### Modified Capabilities

## Impact

- Affected code likely includes:
  - `apps/backend-go/internal/domain/*`
  - `apps/backend-go/internal/service/*`
  - `apps/backend-go/internal/repository/*`
  - `apps/backend-go/internal/transport/http/handler/*`
- Affected APIs:
  - `/sysVariable/create`
  - `/sysVariable/edit`
  - `/sysVariable/detail/:id`
  - `/sysVariable/delete/:id`
  - `/sysVariable/query`
  - `/sysVariable/value/selected/:page/:limit`
  - `/sysVariable/value/selected/:id`
  - `/sysVariable/value/create`
  - `/sysVariable/value/edit`
  - `/sysVariable/value/delete/:id`
  - `/sysVariable/value/batchDel`
- Affected frontend call site:
  - `apps/frontend/src/api/variable.ts`
