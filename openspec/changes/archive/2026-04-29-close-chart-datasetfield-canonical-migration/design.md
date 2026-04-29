## Context

The `2026-04-16-chart-datasetfield-canonical-migration` archived change defined 24 tasks to create canonical handler methods and route registrations for chart, chartData, and datasetField endpoints. Code analysis shows sections 1-4 (handler methods) are fully implemented, but section 5 (route registration) is missing Task 5.2 — the `RegisterChartDataRoutes` function for 4 chartData endpoints.

Currently:
- `/api/chart/*` — 10 canonical routes registered in `registerChartRoutes()` ✅
- `/api/datasetField/*` — 7 canonical routes registered inline in `registerChartRoutes()` ✅
- `/api/chartData/*` — 4 endpoints served ONLY via compat bridge inline closures ❌ (no canonical registration)

The 4 chartData handler methods exist on `ChartHandler` but are not wired to routes.

## Goals / Non-Goals

**Goals:**
- Create `RegisterChartDataRoutes` to wire chartData canonical handlers to routes
- Refactor compat bridge to delegate to canonical handlers instead of duplicating logic
- Add unit tests for untested canonical handlers

**Non-Goals:**
- Frontend path migration (already works via axios baseURL)
- New API endpoints or behavior changes
- Contract-diff baseline fixtures (can be added separately)

## Decisions

### D1: RegisterChartDataRoutes follows existing pattern
Follow the exact same pattern as `RegisterChartDataCompatRoutes` — a standalone function accepting a `*gin.RouterGroup` and `*ChartHandler`.

### D2: Compat bridge delegates to canonical handlers
Replace inline closures in `RegisterChartDataCompatRoutes` with direct calls to `chartHandler.GetFieldData`, etc. This eliminates ~200 lines of duplicated service-call logic.

### D3: Register canonical routes alongside compat
Both canonical and compat routes should coexist for backward compatibility. The canonical routes go under the same `/api/chartData/*` group. Since compat already serves these paths, the canonical registration ensures the handler methods are properly wired even if compat is later removed.

## Risks / Trade-offs

- **[Risk] Double registration**: If both canonical and compat register the same path, Gin will panic. → **Mitigation**: Register canonical routes first, then have compat delegate to them. OR: have compat simply call the canonical handler functions directly without re-registering routes.
