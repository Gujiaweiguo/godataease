## Why

The compatibility bridge in `compatibility_bridge_handler.go` currently routes 6 datasetTree CRUD operations (`save`, `create`, `rename`, `move`, `delete`, `perDelete`) through legacy `/datasetTree/*` paths. As part of the ongoing migration (round 9), these routes need canonical `/api/dataset/*` equivalents so the frontend can call proper REST endpoints while the bridge routes remain intact for backward compatibility.

## What Changes

- Add 6 new canonical route handlers in the backend Go API under `/api/dataset/*`:
  - `POST /api/dataset/save` — save/update a dataset
  - `POST /api/dataset/create` — create a new dataset
  - `POST /api/dataset/rename` — rename an existing dataset
  - `POST /api/dataset/move` — move a dataset to a new parent
  - `POST /api/dataset/delete/:id` — soft-delete a dataset
  - `POST /api/dataset/perDelete/:id` — permanently delete a dataset
- Register these routes in `router.go` inside `registerDatasetRoutes()`, alongside existing canonical dataset routes
- Update 6 frontend API calls in `dataset.ts` to point at canonical paths (`/dataset/save`, `/dataset/create`, etc.)
- All 6 compatibility bridge routes in `compatibility_bridge_handler.go` remain untouched

## Capabilities

### New Capabilities

- `dataset-tree-crud-canonical`: Canonical CRUD endpoints for dataset tree operations (save, create, rename, move, delete, perDelete) under `/api/dataset/*`

### Modified Capabilities

- `dataset-management`: Frontend API paths updated from legacy `/datasetTree/*` to canonical `/dataset/*`

## Impact

- **Backend**: `apps/backend-go/internal/handler/dataset/` — new handler methods or extension of existing canonical handler
- **Backend**: `apps/backend-go/internal/router/router.go` — route registration in `registerDatasetRoutes()`
- **Frontend**: `apps/frontend/src/api/dataset.ts` — 6 URL path changes
- **No breaking changes**: Compatibility bridge routes are preserved; this is purely additive
- **Rollback**: Revert frontend URL changes; canonical backend routes are inert until called
