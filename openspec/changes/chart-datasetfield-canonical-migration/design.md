## Context

The Go backend uses a compatibility bridge (`compatibility_bridge_handler.go`) to map Java-era API paths to Go handler logic. Currently, 17 chart/chartData/datasetField routes are implemented as inline closures within the bridge file, each directly calling service methods and constructing responses inline. The rest of the backend follows a canonical pattern: a handler struct owns service references, named methods handle requests, and `Register*Routes` functions wire them to a Gin router group.

The existing canonical routes `/api/chart/query` and `/api/chart/data` already demonstrate this pattern on `ChartHandler`. The `DatasetHandler` already has 20+ canonical routes registered under `/api/dataset/`. This change extends the same pattern to the remaining 17 routes.

A key constraint: some chartData routes call `DatasetService` methods (e.g., `GetFieldEnum`, `GetFieldEnumDs`), and some datasetField routes call `ChartService` methods (e.g., `ListByDQ`, `ListByDQWithPermission`). This cross-domain dependency requires extending both handler structs with an additional service reference.

## Goals / Non-Goals

**Goals:**
- Add 10 canonical handler methods to `ChartHandler` covering chart (6) and chartData (4) routes
- Add 7 canonical handler methods to `DatasetHandler` covering datasetField (7) routes
- Extend `ChartHandler` with `datasetService *service.DatasetService` field
- Extend `DatasetHandler` with `chartService *service.ChartService` field
- Register all 17 canonical routes in `router.go` under `/api/chart/`, `/api/chartData/`, `/api/datasetField/`
- Update frontend `chart.ts` (12 functions) and `dataset.ts` (9 functions) to call canonical paths
- Retain compat bridge routes as aliases for backward compatibility

**Non-Goals:**
- Removing or refactoring the compatibility bridge inline handlers (they stay as aliases)
- Changing service method signatures or business logic
- Adding new API endpoints beyond the 17 routes being migrated
- Modifying frontend component logic beyond API path changes

## Decisions

### Decision 1: Cross-dependency via struct fields rather than constructor injection only

Extend handler struct definitions with the cross-domain service as a new field. Constructor functions accept the additional service and assign it. The alternative (method-level injection or context-based service lookup) was rejected because it breaks the existing handler pattern and makes dependency flow harder to trace.

**Rationale:** Both `ChartHandler` and `DatasetHandler` already follow the "struct owns services" pattern. Adding one more field is consistent and keeps the constructor as the single point of wiring.

### Decision 2: Canonical routes on separate route groups

Register `/api/chart/*` (extended), `/api/chartData/*`, and `/api/datasetField/*` as three distinct route groups in `router.go`. The `chartData` and `datasetField` groups are separate from `chart` and `dataset` to match the Java-era path convention the frontend already uses.

**Rationale:** The frontend already distinguishes these three path prefixes. Matching them avoids confusion and keeps the migration straightforward: change the prefix from `/chart/*` to `/api/chart/*` etc.

### Decision 3: Helper functions remain package-level

Functions like `flattenChartFieldList` and `parseMultFieldValuesRequest` are currently package-level in the handler package. They remain package-level; canonical handler methods call them directly, same as the compat bridge closures do today.

**Rationale:** Moving these to struct methods would require refactoring the compat bridge as well. Keeping them package-level is minimal change and both old and new code can share them.

### Decision 4: Frontend path migration is a simple prefix swap

Frontend `chart.ts` changes paths from `/chart/*` and `/chartData/*` to `/api/chart/*` and `/api/chartData/*`. Frontend `dataset.ts` changes paths from `/datasetField/*` to `/api/datasetField/*`. The Vite proxy already forwards `/api/*` to the Go backend.

**Rationale:** No request body or response shape changes. The only change is the URL path prefix.

## Risks / Trade-offs

- **Circular dependency risk** between `ChartHandler` and `DatasetHandler`: mitigated by passing service interfaces (not handlers) and initializing them in the router constructor in the correct order. The services themselves don't depend on each other's handlers. → Mitigation: constructor order is `datasetService` first, then `chartService`, then handlers.

- **Compat bridge and canonical routes could diverge**: both call the same service methods, but if one is updated and the other isn't, behavior splits. → Mitigation: the compat bridge aliases should eventually call the canonical handler methods instead of duplicating service calls. This is a follow-up task.

- **Frontend migration is all-or-nothing per file**: switching `chart.ts` halfway would mix old and new paths. → Mitigation: switch each file completely in one commit.

- **Router constructor signature changes**: adding service parameters to `NewChartHandler` and `NewDatasetHandler` is a breaking change for any code that constructs them. → Mitigation: both are only constructed in `router.go`, so the impact is contained.
