# Compatibility Endpoint Inventory (Task 2.1)

**Generated**: 2026-02-24
**Source**: `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
**Reference**: `apps/backend-go/testdata/contract-diff/critical-whitelist.yaml`

---

## Summary

| Status | Count | Notes |
|--------|-------|-------|
| `full` | 58 | Complete implementation with service layer calls |
| `partial` | 3 | Implemented but depends on optional integrations (SeaTunnel/Calcite) |
| `stub` | 0 | No placeholder success patterns detected |
| `missing` | 0 | All routes in whitelist are registered |

---

## Endpoint Status Matrix

### Datasource Endpoints (`/datasource/*`)

| Path | Method | Status | Evidence | Notes |
|------|--------|--------|----------|-------|
| `/datasource/list` | POST | full | `datasourceHandler.List` | P0 in whitelist |
| `/datasource/tree` | POST | full | `datasourceHandler.service.Tree` | P0 in whitelist |
| `/datasource/validate` | POST | full | `datasourceHandler.Validate` | P0 in whitelist |
| `/datasource/validate/:id` | GET | full | `datasourceHandler.service.ValidateByID` | |
| `/datasource/types` | POST | full | Static type list | P1 in whitelist |
| `/datasource/getTables` | POST | full | `datasourceHandler.service.GetTables` | P0 in whitelist |
| `/datasource/getTableStatus` | POST | full | `datasourceHandler.service.GetTableStatus` | |
| `/datasource/getSchema` | POST | full | `datasourceHandler.service.GetSchema` | |
| `/datasource/getTableField` | POST | full | `datasourceHandler.service.GetTableField` | |
| `/datasource/previewData` | POST | full | `datasourceHandler.service.PreviewData` | P0 in whitelist |
| `/datasource/get/:id` | GET | full | `datasourceHandler.service.GetByID` | |
| `/datasource/hidePw/:id` | GET | full | `datasourceHandler.service.GetByID` | |
| `/datasource/getSimpleDs/:id` | GET | full | `datasourceHandler.service.GetByID` | |
| `/datasource/showFinishPage` | GET | full | `datasourceHandler.service.ShowFinishPage` | |
| `/datasource/setShowFinishPage` | POST | full | `datasourceHandler.service.SetShowFinishPage` | |
| `/datasource/latestUse` | POST | full | `datasourceHandler.service.LatestTypes` | |
| `/datasource/save` | POST | full | `datasourceHandler.service.Save` | P0 in whitelist |
| `/datasource/update` | POST | full | `datasourceHandler.service.Update` | |
| `/datasource/move` | POST | full | `datasourceHandler.service.Move` | |
| `/datasource/reName` | POST | full | `datasourceHandler.service.Rename` | |
| `/datasource/createFolder` | POST | full | `datasourceHandler.service.CreateFolder` | |
| `/datasource/checkRepeat` | POST | full | `datasourceHandler.service.CheckRepeat` | |
| `/datasource/checkApiDatasource` | POST | full | `datasourceHandler.service.CheckAPIDatasource` | |
| `/datasource/loadRemoteFile` | POST | full | `datasourceHandler.service.LoadRemoteFile` | |
| `/datasource/syncApiTable` | POST | **partial** | `datasourceHandler.service.SyncAPITable` | Requires SeaTunnel; returns error if unavailable |
| `/datasource/syncApiDs` | POST | **partial** | `datasourceHandler.service.SyncAPIDs` | Requires SeaTunnel; returns error if unavailable |
| `/datasource/listSyncRecord/:dsId/:page/:limit` | POST | full | `datasourceHandler.service.ListSyncRecord` | Persisted records |
| `/datasource/uploadFile` | POST | full | `datasourceHandler.service.UploadFile` | |
| `/datasource/delete/:id` | GET | full | `datasourceHandler.service.Delete` | P0 in whitelist |
| `/datasource/perDelete/:id` | POST | full | `datasourceHandler.service.PerDelete` | |

### Dataset Tree Endpoints (`/datasetTree/*`)

| Path | Method | Status | Evidence | Notes |
|------|--------|--------|----------|-------|
| `/datasetTree/tree` | POST | full | `datasetHandler.Tree` | P0 in whitelist |
| `/datasetTree/get/:id` | POST | full | `buildDatasetDetail` | P1 in whitelist |
| `/datasetTree/details/:id` | POST | full | `buildDatasetDetail` | |
| `/datasetTree/dsDetails` | POST | full | `buildDatasetDetail` loop | |
| `/datasetTree/detailWithPerm` | POST | full | `buildDatasetDetail` loop | |
| `/datasetTree/getSqlParams` | POST | full | `datasetHandler.service.GetSQLParams` | |
| `/datasetTree/save` | POST | full | `datasetHandler.service.Save` | |
| `/datasetTree/create` | POST | full | `datasetHandler.service.Create` | |
| `/datasetTree/rename` | POST | full | `datasetHandler.service.Rename` | |
| `/datasetTree/move` | POST | full | `datasetHandler.service.Move` | |
| `/datasetTree/delete/:id` | POST | full | `datasetHandler.service.Delete` | |
| `/datasetTree/perDelete/:id` | POST | full | `datasetHandler.service.PerDelete` | |
| `/datasetTree/barInfo/:id` | GET | full | `datasetHandler.service.GetGroupByID` | |
| `/datasetTree/exportDataset` | POST | full | `chartHandler.exportService.InnerExportDetails` | |

### Dataset Data Endpoints (`/datasetData/*`)

| Path | Method | Status | Evidence | Notes |
|------|--------|--------|----------|-------|
| `/datasetData/tableField` | POST | full | `datasetHandler.Fields` | P0 in whitelist |
| `/datasetData/previewData` | POST | full | `datasetHandler.Preview` | P0 in whitelist |
| `/datasetData/getDatasetTotal` | POST | full | `datasetHandler.service.Preview` | |
| `/datasetData/previewSql` | POST | **partial** | `datasetHandler.service.PreviewSQL` | Optional Calcite validation |
| `/datasetData/enumValueObj` | POST | full | `datasetHandler.service.GetFieldEnumObj` | |
| `/datasetData/enumValueDs` | POST | full | `datasetHandler.service.GetFieldEnumDs` | |
| `/datasetData/enumValue` | POST | full | `datasetHandler.service.GetFieldEnum` | |

### Chart Data Endpoints (`/chartData/*`)

| Path | Method | Status | Evidence | Notes |
|------|--------|--------|----------|-------|
| `/chartData/getData` | POST | full | `chartHandler.Data` | P0 in whitelist |
| `/chartData/getFieldData/:fieldId/:fieldType` | POST | full | `datasetHandler.service.GetFieldEnum` | |
| `/chartData/getDrillFieldData/:fieldId` | POST | full | `datasetHandler.service.GetFieldEnumDs` | |
| `/chartData/innerExportDetails` | POST | full | `chartHandler.exportService.InnerExportDetails` | |
| `/chartData/innerExportDataSetDetails` | POST | full | `chartHandler.exportService.InnerExportDetails` | |

### Chart Endpoints (`/chart/*`)

| Path | Method | Status | Evidence | Notes |
|------|--------|--------|----------|-------|
| `/chart/getData` | POST | full | `chartHandler.Data` | P0 in whitelist |
| `/chart/getChart/:id` | POST | full | `chartHandler.service.Query` | |
| `/chart/getDetail/:id` | POST | full | `chartHandler.service.Query` | |
| `/chart/checkSameDataSet/:viewIdSource/:viewIdTarget` | GET | full | `chartHandler.service.Query` x2 | |
| `/chart/save` | POST | full | `chartHandler.service.SaveFromMap` | |
| `/chart/listByDQ/:id/:chartId` | POST | full | `chartHandler.service.ListByDQ` | |
| `/chart/copyField/:id/:chartId` | POST | full | `chartHandler.service.CopyField` | |
| `/chart/deleteField/:id` | POST | full | `chartHandler.service.DeleteField` | |
| `/chart/deleteFieldByChart/:chartId` | POST | full | `chartHandler.service.DeleteFieldByChart` | |

### User Endpoints (`/user/*`)

| Path | Method | Status | Evidence | Notes |
|------|--------|--------|----------|-------|
| `/user/list` | POST | full | `user.ListUsers` | P0 in whitelist |
| `/user/create` | POST | full | `user.CreateUser` | P0 in whitelist |
| `/user/edit` | POST | full | `user.UpdateUser` | |
| `/user/update` | POST | full | `user.UpdateUser` | |
| `/user/delete/:id` | POST | full | `user.DeleteUser` | |
| `/user/options` | GET | full | `user.GetUserOptions` | |
| `/user/org/option` | GET | full | `user.GetUserOptions` | |
| `/user/byCurOrg` | POST | full | `user.ListUsers` | |

### Organization Endpoints (`/org/*`)

| Path | Method | Status | Evidence | Notes |
|------|--------|--------|----------|-------|
| `/org/create` | POST | full | `org.CreateOrg` | |
| `/org/update` | POST | full | `org.UpdateOrg` | |
| `/org/delete/:orgId` | POST | full | `org.DeleteOrg` | |
| `/org/list` | GET | full | `org.ListOrgs` | P0 in whitelist |
| `/org/info/:orgId` | GET | full | `org.GetOrgByID` | |
| `/org/tree` | GET | full | `org.GetOrgTree` | P0 in whitelist |
| `/org/checkName` | GET | full | `org.CheckOrgName` | |
| `/org/updateStatus` | POST | full | `org.UpdateOrgStatus` | |
| `/org/children/:parentId` | GET | full | `org.GetChildOrgs` | |
| `/org/mounted` | POST | full | `org.orgService.ListOrgs` | |

---

## Integration-Dependent Endpoints

### SeaTunnel Sync (partial status)

| Endpoint | Behavior When Unavailable | Error Code |
|----------|---------------------------|------------|
| `/datasource/syncApiTable` | Returns error from `ensureSeatunnelClient()` | Service-layer error |
| `/datasource/syncApiDs` | Returns error from `ensureSeatunnelClient()` | Service-layer error |

**Recommendation**: These endpoints should return `503xxx` error code when SeaTunnel is not configured, per Task 1.3 error semantics.

### Calcite SQL Validation (partial status)

| Endpoint | Behavior When Calcite Disabled | Error Code |
|----------|-------------------------------|------------|
| `/datasetData/previewSql` | Falls back to local SQL validation | N/A (still functional) |

**Note**: Calcite is optional enhancement, not required for basic functionality. Status remains `partial` until Calcite integration is complete per `add-calcite-sql-integration` change.

---

## Placeholder Success Patterns Detected

The following patterns were found in `compatibility_bridge_handler.go` but are **NOT placeholder success** issues:

| Line | Pattern | Context | Verdict |
|------|---------|---------|---------|
| 81 | `Success(c, []map[string]string{...})` | `/datasource/types` - static type list | OK: Intentional static data |
| 199, 422, 582, 916, 928, 940 | `Success(c, nil)` | Delete operations | OK: No return data expected |
| 739, 756 | `Success(c, []string{})` | Empty enum fallback | OK: Valid empty result |
| 1215 | `Success(c, []gin.H{})` | Empty enum request | OK: Invalid query ID returns empty |

**No true placeholder success patterns detected** - all `Success(c, nil)` calls are for operations that legitimately return no data (delete, update status) or have valid empty result cases.

---

## Whitelist Coverage Analysis

### P0 Critical APIs (from whitelist)

| Path | Whitelist Status | Verified Status | Match |
|------|------------------|-----------------|-------|
| `/templateManage/templateList` | full | N/A (not in bridge) | - |
| `/templateManage/save` | full | N/A (not in bridge) | - |
| `/datasource/list` | full | full | ✓ |
| `/datasource/tree` | full | full | ✓ |
| `/datasource/validate` | full | full | ✓ |
| `/datasource/getTables` | full | full | ✓ |
| `/datasource/previewData` | full | full | ✓ |
| `/datasource/save` | full | full | ✓ |
| `/datasource/delete/:id` | full | full | ✓ |
| `/datasetTree/tree` | full | full | ✓ |
| `/datasetData/tableField` | full | full | ✓ |
| `/datasetData/previewData` | full | full | ✓ |
| `/chartData/getData` | full | full | ✓ |
| `/chart/getData` | full | full | ✓ |
| `/user/list` | full | full | ✓ |
| `/user/create` | full | full | ✓ |
| `/org/list` | full | full | ✓ |
| `/org/tree` | full | full | ✓ |

### P1 High Priority APIs with Status Mismatch

| Path | Whitelist Status | Verified Status | Action Required |
|------|------------------|-----------------|-----------------|
| `/templateMarket/categories` | stub | N/A (not in bridge) | Update whitelist |

---

## Next Steps (Task 2.2)

1. **Update whitelist** to reflect actual SeaTunnel-dependent endpoints as `partial`
2. **Add gap documentation** for `/datasource/syncApiTable` and `/datasource/syncApiDs`
3. **Sync matrix** with this inventory for CI drift detection
