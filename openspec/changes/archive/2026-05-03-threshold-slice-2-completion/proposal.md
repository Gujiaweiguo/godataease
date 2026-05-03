## Why

The Go threshold alerting module (slice 1, PR #231) landed 11 CRUD endpoints, a pure-function evaluator engine, and 50+ tests — but two critical integration gaps remain: (1) the **Preview** endpoint is a stub returning an error, and (2) **visualization/chart deletion** does not trigger threshold cleanup, leaving orphaned threshold records. Without these, the threshold alerting feature is non-functional end-to-end.

## What Changes

- **Wire Preview endpoint**: Replace the stub in `ThresholdService.Preview()` with a real implementation that fetches chart data (rows + fields), builds the field map, and delegates to the already-implemented `GeneratePreviewHTML()` evaluator function.
- **Inject chart data access into ThresholdService**: Add a chart data accessor interface (or direct repository dependency) so `Preview()` can retrieve chart view data without coupling to the full ChartService.
- **Wire threshold cleanup into visualization deletion**: Add `ThresholdService` dependency to `VisualizationService` and call `DeleteWithChart()` during `DeleteLogic()` for all charts in the deleted visualization.
- **Add `convertStyle` parity**: Normalize template styles in preview output (strip blue highlight backgrounds) to match Java behavior.
- **Create frontend threshold API module**: Add `api/threshold.ts` with typed calls to all threshold endpoints so the UI layer can connect.

## Capabilities

### New Capabilities

_None_ — all endpoints already exist from slice 1.

### Modified Capabilities

- `threshold-management`: Preview endpoint changes from stub to functional; adds chart data dependency and style normalization.
- `visualization-management`: Visualization deletion now triggers threshold cleanup for associated charts.

## Impact

- **Backend services**: `ThresholdService` gains a chart data accessor dependency; `VisualizationService` gains a `ThresholdService` dependency via setter injection (existing pattern).
- **Router/wiring**: `router.go` must pass `ThresholdService` reference to `VisualizationService` during initialization.
- **Frontend**: New `api/threshold.ts` file (no existing code modified).
- **Tests**: New unit tests for `Preview()` and deletion lifecycle; existing tests unaffected.
- **Database**: No schema changes.
- **API contract**: No new endpoints; existing `POST /threshold/preview` returns HTML instead of error.
