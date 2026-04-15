## 1. Backend Handler Struct Extensions

- [ ] 1.1 Add `datasetService *service.DatasetService` field to `ChartHandler` struct in `chart_handler.go`; update `NewChartHandler` to accept and assign it
- [ ] 1.2 Add `chartService *service.ChartService` field to `DatasetHandler` struct in `dataset_handler.go`; update `NewDatasetHandler` to accept and assign it
- [ ] 1.3 Update `router.go` constructor calls: pass `datasetService` to `NewChartHandler` and `chartService` to `NewDatasetHandler` in the correct initialization order

## 2. Chart Canonical Handler Methods

- [ ] 2.1 Add `CheckSameDataSet` method to `ChartHandler`: parse `:viewIdSource` and `:viewIdTarget` path params, call `h.service.Query()` for each, compare dataset IDs, return result
- [ ] 2.2 Add `SaveFromMap` method to `ChartHandler`: bind JSON body as `map[string]interface{}`, call `h.service.SaveFromMap(body)`, return result
- [ ] 2.3 Add `ListByDQ` method to `ChartHandler`: parse `:id` and `:chartId` path params, check user permission via middleware, call `ListByDQWithPermission` or `ListByDQ`, return result
- [ ] 2.4 Add `CopyField` method to `ChartHandler`: parse `:id` and `:chartId` path params, call `h.service.CopyField(id, chartID)`, return result
- [ ] 2.5 Add `DeleteField` method to `ChartHandler`: parse `:id` path param, call `h.service.DeleteField(id)`, return result
- [ ] 2.6 Add `DeleteFieldByChart` method to `ChartHandler`: parse `:chartId` path param, call `h.service.DeleteFieldByChart(chartID)`, return result

## 3. ChartData Canonical Handler Methods

- [ ] 3.1 Add `GetFieldData` method to `ChartHandler`: parse `:fieldId` and `:fieldType` path params, call `h.datasetService.GetFieldEnum()`, return result
- [ ] 3.2 Add `GetDrillFieldData` method to `ChartHandler`: parse `:fieldId` path param, call `h.datasetService.GetFieldEnumDs(fieldID)`, return result
- [ ] 3.3 Add `InnerExportDetails` method to `ChartHandler`: bind export request body, call `h.exportService.InnerExportDetails(&req)`, generate filename via `service.GenerateExcelFilename()`, set response header, return result
- [ ] 3.4 Add `InnerExportDataSetDetails` method to `ChartHandler`: same logic as `InnerExportDetails` but registered on the dataset details path

## 4. DatasetField Canonical Handler Methods

- [ ] 4.1 Add `ListByDatasetGroup` method to `DatasetHandler`: parse `:datasetId` path param, check user permission, call `h.chartService.ListByDQWithPermission` or `h.chartService.ListByDQ` with chartId=0, flatten result via `flattenChartFieldList()`, return flattened list
- [ ] 4.2 Add `ListWithPermissions` method to `DatasetHandler`: same logic as `ListByDatasetGroup` but for GET requests, parse `:datasetId` from path
- [ ] 4.3 Add `SaveField` method to `DatasetHandler`: bind JSON body to field struct, call `h.service.SaveField(&field)`, return result
- [ ] 4.4 Add `GetFieldFunctions` method to `DatasetHandler`: call `h.service.GetFieldFunctions()`, return result
- [ ] 4.5 Add `MultFieldValuesForPermissions` method to `DatasetHandler`: bind request body via `parseMultFieldValuesRequest(c)`, call `h.service.GetFieldEnum(req)`, return result
- [ ] 4.6 Add `CopilotFields` method to `DatasetHandler`: parse `:id` path param as datasetID, get userID from middleware, call `h.service.CopilotFields(datasetID, userID)`, return result
- [ ] 4.7 Add `ListFieldsByDsIds` method to `DatasetHandler`: bind JSON body to extract `DsIds` array, call `h.service.ListFieldsByDsIds(req.DsIds)`, return result

## 5. Route Registration

- [ ] 5.1 Add canonical route registrations in `RegisterChartRoutes` for the 6 new chart endpoints: `GET /checkSameDataSet/:viewIdSource/:viewIdTarget`, `POST /save`, `POST /listByDQ/:id/:chartId`, `POST /copyField/:id/:chartId`, `POST /deleteField/:id`, `POST /deleteFieldByChart/:chartId`
- [ ] 5.2 Create `RegisterChartDataRoutes` function with route group `/chartData` for the 4 chartData endpoints: `POST /getFieldData/:fieldId/:fieldType`, `POST /getDrillFieldData/:fieldId`, `POST /innerExportDetails`, `POST /innerExportDataSetDetails`
- [ ] 5.3 Create `RegisterDatasetFieldRoutes` function with route group `/datasetField` for the 7 datasetField endpoints: `POST /listByDatasetGroup/:datasetId`, `GET /listWithPermissions/:datasetId`, `POST /save`, `POST /getFunction`, `POST /multFieldValuesForPermissions`, `POST /copilotFields/:id`, `POST /listByDsIds`
- [ ] 5.4 Wire the new route registration functions in `router.go` `setupRoutes` method, passing the appropriate handlers

## 6. Testing

- [ ] 6.1 Add unit tests for the 10 new `ChartHandler` methods: verify path param parsing, service call delegation, and response formatting
- [ ] 6.2 Add unit tests for the 7 new `DatasetHandler` methods: verify path param parsing, cross-service call delegation, and response formatting
- [ ] 6.3 Extend `router_test.go` to verify all 17 new canonical routes are registered and respond correctly
- [ ] 6.4 Run `make test` in backend to confirm no regressions

## Notes

**No frontend changes needed.** The axios baseURL is already `/api` (via `VITE_API_BASEPATH=/api`), so frontend paths like `/chart/*` become `/api/chart/*` automatically. The new canonical backend routes registered under `/api/` will match these requests. Compat bridge routes at root level remain for backward compatibility.
