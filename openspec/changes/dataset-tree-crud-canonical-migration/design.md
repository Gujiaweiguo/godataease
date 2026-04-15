## Context

This is round 9 of the compatibility bridge migration. Rounds 1-8 covered 31 datasource module routes. The backend uses a layered architecture: handler (transport/http/handler) → service → repository → domain. Canonical routes live under `/api/*` in `router.go`, registered via dedicated `registerXxxRoutes()` methods.

The 6 datasetTree CRUD routes currently sit inside `compatibility_bridge_handler.go` as anonymous closures under a `datasetTreeGroup` Gin group. Each closure parses the request, calls a method on `datasetHandler.service`, and wraps the result in the standard `response.Success`/`response.Error` envelope.

The canonical `DatasetHandler` at `handler/dataset_handler.go` already has methods: `Tree`, `Fields`, `Preview`, `PreviewWithPermission`. It holds a `*service.DatasetService` field. The 6 new methods follow the same pattern.

`parseDatasetWriteRequest` is a package-level function in `compatibility_bridge_handler.go`. Since both files share the `handler` package, the canonical handler can call it directly without moving or duplicating it.

## Goals / Non-Goals

**Goals:**
- Add 6 canonical handler methods on `DatasetHandler`: `Save`, `Create`, `Rename`, `Move`, `Delete`, `PerDelete`
- Register 6 routes in `registerDatasetRoutes()` alongside existing tree/fields/preview routes
- Update 6 frontend API URLs in `dataset.ts` from `/datasetTree/*` to `/dataset/*`
- Keep all compatibility bridge routes untouched

**Non-Goals:**
- Removing or modifying existing compatibility bridge routes
- Changing request/response formats or adding new fields
- Adding permission middleware (not present in bridge version, can be a separate change)
- Refactoring `parseDatasetWriteRequest` out of the bridge file

## Decisions

### 1. Handler methods on existing DatasetHandler

**Decision**: Add new methods (`Save`, `Create`, `Rename`, `Move`, `Delete`, `PerDelete`) to the existing `DatasetHandler` struct in `dataset_handler.go`.

**Rationale**: The struct already holds `*service.DatasetService` which provides all the service methods. Keeping all dataset handlers together is cleaner than creating a second handler struct. This matches how previous migrations added canonical handlers.

### 2. Reuse `parseDatasetWriteRequest` as-is

**Decision**: Call the existing `parseDatasetWriteRequest` function from the canonical handler methods.

**Rationale**: Both files are in the `handler` package, so the function is accessible. It handles the JSON-to-`WriteRequest` mapping correctly. Duplicating it would drift over time. Moving it would require touching the bridge file (violating the "preserve bridge" principle).

### 3. Route registration alongside existing routes

**Decision**: Add the 6 routes in `registerDatasetRoutes()` inside the existing `datasetGroup` block.

**Rationale**: Keeps all dataset routes in one place. The method already creates `api.Group("/dataset")`, so new routes like `datasetGroup.POST("/save", ...)` land at `/api/dataset/save` automatically.

### 4. Frontend URL prefix change

**Decision**: Change `/datasetTree/save` → `/dataset/save`, etc. (not `/api/dataset/save`).

**Rationale**: The Vite dev proxy adds the `/api` prefix automatically. The existing canonical frontend calls (like `/dataset/tree`) confirm this pattern.

## Risks / Trade-offs

- **parseDatasetWriteRequest coupling**: The canonical handler depends on a function in the bridge file. If the bridge file is ever removed, this function must be relocated first. → Acceptable: bridge removal is a later milestone, and the function can be moved then.
- **No permission middleware**: The bridge routes lack permission checks, and the canonical routes match this behavior. A future change can add middleware. → Intentionally out of scope.
- **Dual-path traffic during rollout**: Both old and new paths work simultaneously. No migration risk since the bridge is never removed.
