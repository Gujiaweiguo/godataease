## 1. Route Registration

- [ ] 1.1 Create `RegisterChartDataRoutes(r *gin.RouterGroup, h *ChartHandler)` function that registers the 4 chartData endpoints (`getFieldData/:fieldId/:fieldType`, `getDrillFieldData/:fieldId`, `innerExportDetails`, `innerExportDataSetDetails`) wiring directly to `h.GetFieldData`, `h.GetDrillFieldData`, `h.InnerExportDetails`, `h.InnerExportDataSetDetails`
- [ ] 1.2 Call `RegisterChartDataRoutes` in `router.go` setup alongside existing compat routes
- [ ] 1.3 Refactor `RegisterChartDataCompatRoutes` chartData inline closures to delegate to `chartHandler.GetFieldData` etc. instead of duplicating service-call logic

## 2. Testing

- [ ] 2.1 Add unit tests in `chart_handler_test.go` for untested canonical methods: `SaveFromMap`, `ListByDQ`, `CopyField`, `DeleteField`, `DeleteFieldByChart`, `GetFieldData`, `GetDrillFieldData`, `InnerExportDetails`, `InnerExportDataSetDetails`
- [ ] 2.2 Verify `router_test.go` covers the new chartData canonical routes

## 3. Verification

- [ ] 3.1 Run `make test` — all tests pass
- [ ] 3.2 Run `make drift-check` — no contract drift introduced
