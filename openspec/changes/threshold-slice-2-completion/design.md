## Context

Slice 1 (PR #231) delivered 11 threshold CRUD endpoints, a pure-function evaluator engine (`FilterRows`, `GeneratePreviewHTML`, `ConvertRulesToText`), and 50+ passing tests. Two integration gaps remain:

1. **Preview is a stub**: `ThresholdService.Preview()` at line 235 returns a hard-coded error. The evaluator can generate HTML, but the service can't fetch chart data to feed it.
2. **No deletion lifecycle**: When a visualization is soft-deleted via `VisualizationService.DeleteLogic()`, its charts' threshold records are left orphaned. The `DeleteWithChart` method exists on `ThresholdService` but nothing calls it.

The existing codebase uses setter injection for cross-service dependencies (`SetResourcePermissionService`, `SetTemplateService`, etc.) and defines small interfaces for isolated concerns (e.g. `ThresholdRepo`).

## Goals / Non-Goals

**Goals:**
- Wire the Preview endpoint so `POST /threshold/preview` returns rendered HTML
- Wire threshold cleanup into visualization deletion so no orphaned records remain
- Add template style normalization to strip blue highlight backgrounds from preview output
- Create the frontend `api/threshold.ts` module for UI integration
- Maintain the existing test baseline (50+ tests continue passing)

**Non-Goals:**
- Scheduled threshold execution (future slice)
- Threshold notification delivery (future slice)
- Frontend UI components for threshold management (separate work)
- Snapshot (`resourceTable="snapshot"`) preview behavior (out of scope per slice 1)
- Changes to the evaluator engine's matching logic

## Decisions

### Decision 1: Chart data accessor via interface, not direct service injection

**Choice**: Define a `ThresholdChartDataAccessor` interface with a single method:
```go
type ThresholdChartDataAccessor interface {
    GetChartDataForThreshold(ctx context.Context, chartID int64, resourceTable string) ([]map[string]any, []FieldDTO, error)
}
```

**Rationale**: `ThresholdService` needs chart data (rows + field metadata) to run preview. Direct `ChartService` injection would pull in chart rendering, permissions, and dataset logic. A narrow interface follows the same pattern as `ThresholdRepo` and respects Go's interface segregation principle. The concrete implementation lives in `ChartService` via a new method.

**Alternatives considered**:
- **A) Inject ChartService directly**: Heavy coupling. Threshold would depend on chart's full API surface. Rejected.
- **B) Add a repository-level query**: Chart data assembly involves dataset fields, permissions, and view logic that belongs in the service layer, not the repository. Rejected.

### Decision 2: Setter injection for ThresholdService into VisualizationService

**Choice**: Add `SetThresholdService(ts *ThresholdService)` to `VisualizationService`, matching the existing setter pattern at lines 52-70 of `visualization_service.go`.

**Rationale**: The codebase already uses this pattern for `SetResourcePermissionService`, `SetTemplateService`, `SetDatasetRepository`, and `SetAuditService`. Threshold cleanup is a cross-cutting side effect, same as audit logging. Setter injection avoids circular constructor dependencies.

**Alternatives considered**:
- **A) Event/callback system**: Over-engineering for a single call site. The codebase has no event bus. Rejected.
- **B) Middleware approach**: Deletion cleanup is a business rule, not an HTTP concern. Rejected.

### Decision 3: Visualization deletion enumerates charts then cleans thresholds

**Choice**: In `DeleteLogic`, after the soft-delete succeeds, parse `ComponentData` JSON to extract chart IDs, then call `ThresholdService.DeleteWithChart()` for each.

**Rationale**: The visualization's `ComponentData` contains the chart component list. Walking the JSON to find chart IDs is straightforward. Calling the existing `DeleteWithChart` method reuses tested logic and handles the `resourceTable` guard.

**Alternatives considered**:
- **A) Bulk delete by visualization ID**: Would require a new repository method (`DeleteByVisualizationID`). Adds schema coupling. Rejected.
- **B) Database cascade**: The schema uses logical soft-deletes, not FK cascades. Rejected.

### Decision 4: normalizeTemplateStyles as a pure function in the evaluator

**Choice**: Add `normalizeTemplateStyles(html string) string` to `threshold_evaluator.go`. It strips `background-color` styles from `<span>` elements that match the blue highlight pattern used in templates.

**Rationale**: The evaluator is already the home for HTML generation functions (`GeneratePreviewHTML`, `buildThresholdTableHTML`). Style normalization is a post-processing step on the generated HTML. Keeping it as a pure function makes it trivially testable.

## Risks / Trade-offs

**[Risk] Chart data accessor implementation complexity** → The `GetChartDataForThreshold` method on `ChartService` must assemble rows + fields. If chart data assembly is tightly coupled to HTTP request context, the interface method may need a lightweight DTO. Mitigation: check existing chart data paths before implementing. If too complex, fall back to a narrower accessor that returns pre-computed data.

**[Risk] ComponentData parsing fragility** → Chart component JSON structure could vary between visualization versions. Mitigation: parse defensively with `json.RawMessage`, log warnings for unparseable components, and never fail the deletion if threshold cleanup errors.

**[Risk] Setter injection nil guard** → If `ThresholdService` is not wired before `DeleteLogic` is called, the cleanup silently skips. Mitigation: the setter is called in `router.go` immediately after service construction, same as all other setters. Add a nil check in `DeleteLogic` with a log warning.

**[Risk] normalizeTemplateStyles regex fragility** → Blue highlight patterns could vary across templates. Mitigation: use a broad regex that strips any `background-color` from `<span>` elements with `id="changeText-*"`, matching the pattern already used in `GeneratePreviewHTML`.
