## 1. Chart Data Accessor Interface

- [ ] 1.1 Define `ThresholdChartDataAccessor` interface in `threshold_service.go` with method `GetChartDataForThreshold(ctx context.Context, chartID int64, resourceTable string) ([]map[string]any, []FieldDTO, error)`
- [ ] 1.2 Add `chartDataAccessor` field to `ThresholdService` struct and update `NewThresholdService` or add setter `SetChartDataAccessor`
- [ ] 1.3 Implement `GetChartDataForThreshold` on `ChartService` in a new file `chart_threshold_accessor.go` that assembles rows and fields for a chart view

## 2. Preview Implementation

- [ ] 2.1 Replace the Preview stub in `ThresholdService.Preview()` with real logic: call `chartDataAccessor.GetChartDataForThreshold`, parse threshold rules into a filter tree, build the field map, delegate to `GeneratePreviewHTML`, and apply `normalizeTemplateStyles`
- [ ] 2.2 Add `normalizeTemplateStyles(html string) string` to `threshold_evaluator.go` that strips `background-color` from `<span>` elements with `id="changeText-*"` using regex
- [ ] 2.3 Add unit tests for `Preview()` covering: matching data returns HTML, no matching rows returns empty string, missing chart returns error, malformed rules returns error
- [ ] 2.4 Add unit tests for `normalizeTemplateStyles` covering: strips blue highlight backgrounds, preserves other styles, handles HTML without background-color, handles empty input

## 3. Visualization-Threshold Deletion Wiring

- [ ] 3.1 Add `thresholdService` field and `SetThresholdService(ts *ThresholdService)` setter to `VisualizationService` in `visualization_service.go`
- [ ] 3.2 Implement `extractChartIDsFromComponentData(componentData json.RawMessage) []int64` helper in `visualization_service.go` that parses ComponentData JSON and returns chart component IDs
- [ ] 3.3 Update `VisualizationService.DeleteLogic()` to call threshold cleanup after soft-delete: iterate extracted chart IDs and call `ThresholdService.DeleteWithChart(ctx, chartID, "core")` for each, with nil guard and error logging
- [ ] 3.4 Add unit tests for `extractChartIDsFromComponentData` covering: valid component data with multiple charts, empty component data, malformed JSON, single chart
- [ ] 3.5 Add unit tests for deletion-threshold lifecycle: verify threshold cleanup is called for each chart when visualization is deleted, verify deletion succeeds when threshold service is nil

## 4. Router Wiring

- [ ] 4.1 Update `router.go` to wire `ThresholdService` into `VisualizationService` via `visualService.SetThresholdService(thresholdService)` after both services are constructed

## 5. Frontend API Module

- [ ] 5.1 Create `apps/frontend/src/api/threshold.ts` with typed API functions for all 11 threshold endpoints (save, edit, formInfo, delete, deleteWithChart, switch, batchReci, pager, preview, anyThreshold, instancePager) using existing axios config patterns

## 6. Verification

- [ ] 6.1 Run `make test` in `apps/backend-go` and confirm all tests pass (existing 50+ plus new tests)
- [ ] 6.2 Run `npm run lint && npm run ts:check` in `apps/frontend` and confirm clean
