# Frozen Endpoint Compatibility Matrix

## Overview

This document defines the frozen endpoint compatibility matrix for the Java-to-Go migration. It tracks the status of critical API routes across both backends, ensuring contract parity during the migration phase.

**Matrix Version:** 1.0.0  
**Last Updated:** 2026-02-18  
**Change ID:** add-go-security-compatibility-readiness

## Status Definitions

| Status | Description |
|--------|-------------|
| **full** | Complete implementation with full business logic parity |
| **partial** | Core functionality exists, edge cases may differ |
| **stub** | Endpoint exists but returns placeholder/mock data |
| **missing** | No implementation in Go backend |
| **exists** | Implemented in Java backend |

## Auth Mode Definitions

| Mode | Description |
|------|-------------|
| **public** | No authentication required |
| **authenticated** | Requires valid user session/token |
| **admin** | Requires admin or super-admin privileges |

## Priority Definitions

| Priority | Description |
|----------|-------------|
| **P0** | Critical path, blocks migration if not ready |
| **P1** | High traffic, needed for core functionality |
| **P2** | Nice to have, can be deferred |

---

## 1. Template Management Routes

| Route Path | HTTP Method | Module | Java Status | Go Status | Auth Mode | Priority | Owner |
|------------|-------------|--------|-------------|-----------|-----------|----------|-------|
| `/templateManage/templateList` | POST | templateManage | exists | missing | authenticated | P0 | template-team |
| `/templateManage/save` | POST | templateManage | exists | partial | authenticated | P0 | template-team |
| `/templateManage/delete` | POST | templateManage | exists | stub | authenticated | P1 | template-team |
| `/templateManage/deleteCategory` | POST | templateManage | exists | missing | authenticated | P1 | template-team |
| `/templateManage/findOne` | GET | templateManage | exists | partial | authenticated | P1 | template-team |
| `/templateManage/find` | POST | templateManage | exists | missing | authenticated | P1 | template-team |
| `/templateManage/findCategories` | POST | templateManage | exists | missing | authenticated | P1 | template-team |
| `/templateManage/nameCheck` | POST | templateManage | exists | missing | authenticated | P2 | template-team |
| `/templateManage/categoryTemplateNameCheck` | POST | templateManage | exists | missing | authenticated | P2 | template-team |
| `/templateManage/batchUpdate` | POST | templateManage | exists | missing | authenticated | P2 | template-team |
| `/templateManage/batchDelete` | POST | templateManage | exists | missing | authenticated | P2 | template-team |
| `/templateMarket/searchTemplate` | GET | templateMarket | exists | missing | authenticated | P0 | template-team |
| `/templateMarket/searchTemplateRecommend` | GET | templateMarket | exists | missing | authenticated | P1 | template-team |
| `/templateMarket/searchTemplatePreview` | GET | templateMarket | exists | missing | authenticated | P2 | template-team |
| `/templateMarket/categories` | GET | templateMarket | exists | missing | authenticated | P1 | template-team |
| `/templateMarket/categoriesObject` | GET | templateMarket | exists | missing | authenticated | P2 | template-team |
| `/template/create` | POST | template | exists | full | authenticated | P1 | template-team |
| `/template/get/:id` | GET | template | exists | full | authenticated | P1 | template-team |
| `/template/list` | POST | template | exists | full | authenticated | P1 | template-team |
| `/template/update` | POST | template | exists | full | authenticated | P1 | template-team |
| `/template/delete/:id` | DELETE | template | exists | full | authenticated | P1 | template-team |

**Go Route Group:** `/template` (native), `/templateManage` (via compatibility bridge - NOT IMPLEMENTED), `/templateMarket` (via compatibility bridge - NOT IMPLEMENTED)

---

## 2. Datasource Routes

| Route Path | HTTP Method | Module | Java Status | Go Status | Auth Mode | Priority | Owner |
|------------|-------------|--------|-------------|-----------|-----------|----------|-------|
| `/datasource/list` | POST | datasource | exists | full | authenticated | P0 | datasource-team |
| `/datasource/tree` | POST | datasource | exists | full | authenticated | P0 | datasource-team |
| `/datasource/validate` | POST | datasource | exists | full | authenticated | P0 | datasource-team |
| `/datasource/validate/:id` | GET | datasource | exists | full | authenticated | P0 | datasource-team |
| `/datasource/types` | POST | datasource | exists | stub | authenticated | P1 | datasource-team |
| `/datasource/getTables` | POST | datasource | exists | full | authenticated | P0 | datasource-team |
| `/datasource/getTableStatus` | POST | datasource | exists | stub | authenticated | P1 | datasource-team |
| `/datasource/getSchema` | POST | datasource | exists | stub | authenticated | P1 | datasource-team |
| `/datasource/getTableField` | POST | datasource | exists | stub | authenticated | P1 | datasource-team |
| `/datasource/previewData` | POST | datasource | exists | stub | authenticated | P0 | datasource-team |
| `/datasource/get/:id` | GET | datasource | exists | stub | authenticated | P0 | datasource-team |
| `/datasource/hidePw/:id` | GET | datasource | exists | missing | authenticated | P1 | datasource-team |
| `/datasource/getSimpleDs/:id` | GET | datasource | exists | missing | authenticated | P2 | datasource-team |
| `/datasource/showFinishPage` | GET | datasource | exists | missing | authenticated | P2 | datasource-team |
| `/datasource/setShowFinishPage` | POST | datasource | exists | missing | authenticated | P2 | datasource-team |
| `/datasource/latestUse` | POST | datasource | exists | missing | authenticated | P2 | datasource-team |
| `/datasource/save` | POST | datasource | exists | stub | authenticated | P0 | datasource-team |
| `/datasource/update` | POST | datasource | exists | stub | authenticated | P0 | datasource-team |
| `/datasource/move` | POST | datasource | exists | missing | authenticated | P1 | datasource-team |
| `/datasource/reName` | POST | datasource | exists | missing | authenticated | P1 | datasource-team |
| `/datasource/createFolder` | POST | datasource | exists | missing | authenticated | P1 | datasource-team |
| `/datasource/checkRepeat` | POST | datasource | exists | missing | authenticated | P1 | datasource-team |
| `/datasource/checkApiDatasource` | POST | datasource | exists | missing | authenticated | P2 | datasource-team |
| `/datasource/loadRemoteFile` | POST | datasource | exists | missing | authenticated | P2 | datasource-team |
| `/datasource/syncApiTable` | POST | datasource | exists | missing | authenticated | P2 | datasource-team |
| `/datasource/syncApiDs` | POST | datasource | exists | missing | authenticated | P2 | datasource-team |
| `/datasource/uploadFile` | POST | datasource | exists | missing | authenticated | P1 | datasource-team |
| `/datasource/listSyncRecord/:dsId/:page/:limit` | POST | datasource | exists | missing | authenticated | P2 | datasource-team |
| `/datasource/delete/:id` | GET | datasource | exists | stub | authenticated | P0 | datasource-team |
| `/datasource/perDelete/:id` | POST | datasource | exists | missing | authenticated | P1 | datasource-team |
| `/ds/list` | POST | datasource (native) | n/a | full | authenticated | P1 | datasource-team |
| `/ds/validate` | POST | datasource (native) | n/a | full | authenticated | P1 | datasource-team |

**Go Route Groups:** `/datasource` (compatibility bridge), `/ds` (native)

---

## 3. Dataset Routes

| Route Path | HTTP Method | Module | Java Status | Go Status | Auth Mode | Priority | Owner |
|------------|-------------|--------|-------------|-----------|-----------|----------|-------|
| `/datasetTree/tree` | POST | datasetTree | exists | full | authenticated | P0 | dataset-team |
| `/datasetTree/get/:id` | POST | datasetTree | exists | stub | authenticated | P0 | dataset-team |
| `/datasetTree/details/:id` | POST | datasetTree | exists | stub | authenticated | P0 | dataset-team |
| `/datasetTree/dsDetails` | POST | datasetTree | exists | stub | authenticated | P1 | dataset-team |
| `/datasetTree/detailWithPerm` | POST | datasetTree | exists | missing | authenticated | P1 | dataset-team |
| `/datasetTree/getSqlParams` | POST | datasetTree | exists | missing | authenticated | P2 | dataset-team |
| `/datasetTree/save` | POST | datasetTree | exists | stub | authenticated | P0 | dataset-team |
| `/datasetTree/create` | POST | datasetTree | exists | stub | authenticated | P0 | dataset-team |
| `/datasetTree/rename` | POST | datasetTree | exists | missing | authenticated | P1 | dataset-team |
| `/datasetTree/move` | POST | datasetTree | exists | missing | authenticated | P1 | dataset-team |
| `/datasetTree/delete/:id` | POST | datasetTree | exists | stub | authenticated | P0 | dataset-team |
| `/datasetTree/perDelete/:id` | POST | datasetTree | exists | missing | authenticated | P1 | dataset-team |
| `/datasetTree/barInfo/:id` | GET | datasetTree | exists | missing | authenticated | P2 | dataset-team |
| `/datasetTree/exportDataset` | POST | datasetTree | exists | missing | authenticated | P2 | dataset-team |
| `/datasetData/tableField` | POST | datasetData | exists | full | authenticated | P0 | dataset-team |
| `/datasetData/previewData` | POST | datasetData | exists | full | authenticated | P0 | dataset-team |
| `/datasetData/getDatasetTotal` | POST | datasetData | exists | missing | authenticated | P1 | dataset-team |
| `/datasetData/previewSql` | POST | datasetData | exists | missing | authenticated | P2 | dataset-team |
| `/datasetData/enumValueObj` | POST | datasetData | exists | missing | authenticated | P2 | dataset-team |
| `/datasetData/enumValueDs` | POST | datasetData | exists | missing | authenticated | P2 | dataset-team |
| `/datasetData/enumValue` | POST | datasetData | exists | missing | authenticated | P2 | dataset-team |
| `/dataset/tree` | POST | dataset (native) | n/a | full | authenticated | P1 | dataset-team |
| `/dataset/fields` | POST | dataset (native) | n/a | full | authenticated | P1 | dataset-team |
| `/dataset/preview` | POST | dataset (native) | n/a | full | authenticated | P1 | dataset-team |

**Go Route Groups:** `/datasetTree` (compatibility bridge), `/datasetData` (compatibility bridge), `/dataset` (native)

---

## 4. Chart Routes

| Route Path | HTTP Method | Module | Java Status | Go Status | Auth Mode | Priority | Owner |
|------------|-------------|--------|-------------|-----------|-----------|----------|-------|
| `/chart/getData` | POST | chart | exists | partial | authenticated | P0 | chart-team |
| `/chart/save` | POST | chart | exists | missing | authenticated | P0 | chart-team |
| `/chart/getDetail/:id` | POST | chart | exists | stub | authenticated | P0 | chart-team |
| `/chart/checkSameDataSet/:viewIdSource/:viewIdTarget` | GET | chart | exists | missing | authenticated | P1 | chart-team |
| `/chart/listByDQ/:id/:chartId` | POST | chart | exists | missing | authenticated | P1 | chart-team |
| `/chart/copyField/:id/:chartId` | POST | chart | exists | missing | authenticated | P2 | chart-team |
| `/chart/deleteField/:id` | POST | chart | exists | missing | authenticated | P2 | chart-team |
| `/chart/deleteFieldByChart/:chartId` | POST | chart | exists | missing | authenticated | P2 | chart-team |
| `/chartData/getData` | POST | chartData | exists | full | authenticated | P0 | chart-team |
| `/chartData/getFieldData/:fieldId/:fieldType` | POST | chartData | exists | missing | authenticated | P1 | chart-team |
| `/chartData/getDrillFieldData/:fieldId` | POST | chartData | exists | missing | authenticated | P1 | chart-team |
| `/chartData/innerExportDetails` | POST | chartData | exists | missing | authenticated | P2 | chart-team |
| `/chartData/innerExportDataSetDetails` | POST | chartData | exists | missing | authenticated | P2 | chart-team |
| `/chart/query` | POST | chart (native) | n/a | full | authenticated | P1 | chart-team |
| `/chart/data` | POST | chart (native) | n/a | full | authenticated | P1 | chart-team |

**Go Route Groups:** `/chart` (compatibility bridge), `/chartData` (compatibility bridge), `/chart` (native)

---

## 5. Share Routes

| Route Path | HTTP Method | Module | Java Status | Go Status | Auth Mode | Priority | Owner |
|------------|-------------|--------|-------------|-----------|-----------|----------|-------|
| `/share/create` | POST | share | exists | full | authenticated | P0 | share-team |
| `/share/validate` | POST | share | exists | full | authenticated | P0 | share-team |
| `/share/revoke/:id` | DELETE | share | exists | full | authenticated | P0 | share-team |
| `/share/status/:resourceId` | GET | share | exists | full | authenticated | P1 | share-team |
| `/share/detail/:resourceId` | GET | share | exists | full | authenticated | P1 | share-team |
| `/share/switcher/:resourceId` | POST | share | exists | full | authenticated | P1 | share-team |
| `/share/ticket/create` | POST | share | exists | full | authenticated | P1 | share-team |
| `/share/ticket/validate` | POST | share | exists | full | authenticated | P1 | share-team |
| `/share/editExp` | POST | share | exists | missing | authenticated | P2 | share-team |
| `/share/editPwd` | POST | share | exists | missing | authenticated | P2 | share-team |
| `/share/query` | POST | share | exists | missing | authenticated | P2 | share-team |
| `/share/proxyInfo` | POST | share | exists | missing | authenticated | P2 | share-team |
| `/share/validatePwd` | POST | share | exists | missing | authenticated | P2 | share-team |
| `/share/queryRelationByUserId/:uid` | GET | share | exists | missing | authenticated | P2 | share-team |
| `/share/editUuid` | POST | share | exists | missing | authenticated | P2 | share-team |

**Go Route Group:** `/share`

---

## 6. Export Center Routes

| Route Path | HTTP Method | Module | Java Status | Go Status | Auth Mode | Priority | Owner |
|------------|-------------|--------|-------------|-----------|-----------|----------|-------|
| `/exportCenter/exportTasks` | GET | exportCenter | exists | full | authenticated | P0 | export-team |
| `/exportCenter/pager/:goPage/:pageSize` | GET | exportCenter | exists | partial | authenticated | P0 | export-team |
| `/exportCenter/delete/:id` | GET | exportCenter | exists | full | authenticated | P0 | export-team |
| `/exportCenter/delete` | POST | exportCenter | exists | full | authenticated | P0 | export-team |
| `/exportCenter/deleteAll/:type` | POST | exportCenter | exists | full | authenticated | P0 | export-team |
| `/exportCenter/download/:id` | GET | exportCenter | exists | full | authenticated | P0 | export-team |
| `/exportCenter/generateDownloadUri/:id` | GET | exportCenter | exists | full | authenticated | P0 | export-team |
| `/exportCenter/retry/:id` | POST | exportCenter | exists | full | authenticated | P0 | export-team |
| `/exportCenter/exportLimit` | GET | exportCenter | exists | full | authenticated | P1 | export-team |

**Go Route Group:** `/exportTasks` (note: different path prefix in Go)

---

## 7. Permission Routes

| Route Path | HTTP Method | Module | Java Status | Go Status | Auth Mode | Priority | Owner |
|------------|-------------|--------|-------------|-----------|-----------|----------|-------|
| `/api/system/permission/list` | POST | permission | exists | full | admin | P0 | permission-team |
| `/api/system/permission/create` | POST | permission | exists | full | admin | P0 | permission-team |
| `/api/system/permission/update` | POST | permission | exists | full | admin | P0 | permission-team |
| `/api/system/permission/delete/:permId` | POST | permission | exists | full | admin | P0 | permission-team |

**Go Route Group:** `/api/system/permission`

---

## 8. User Routes

| Route Path | HTTP Method | Module | Java Status | Go Status | Auth Mode | Priority | Owner |
|------------|-------------|--------|-------------|-----------|-----------|----------|-------|
| `/user/list` | POST | user | exists | full | authenticated | P0 | user-team |
| `/user/create` | POST | user | exists | full | admin | P0 | user-team |
| `/user/edit` | POST | user | exists | full | admin | P0 | user-team |
| `/user/update` | POST | user | exists | full | admin | P0 | user-team |
| `/user/delete/:id` | POST | user | exists | full | admin | P0 | user-team |
| `/user/options` | GET | user | exists | full | authenticated | P1 | user-team |
| `/user/org/option` | GET | user | exists | full | authenticated | P1 | user-team |
| `/user/byCurOrg` | POST | user | exists | full | authenticated | P1 | user-team |

**Go Route Groups:** `/user` (compatibility bridge), `/api/user` (native)

---

## 9. Organization Routes

| Route Path | HTTP Method | Module | Java Status | Go Status | Auth Mode | Priority | Owner |
|------------|-------------|--------|-------------|-----------|-----------|----------|-------|
| `/org/create` | POST | org | exists | full | admin | P0 | org-team |
| `/org/update` | POST | org | exists | full | admin | P0 | org-team |
| `/org/delete/:orgId` | POST | org | exists | full | admin | P0 | org-team |
| `/org/list` | GET | org | exists | full | authenticated | P0 | org-team |
| `/org/info/:orgId` | GET | org | exists | full | authenticated | P1 | org-team |
| `/org/tree` | GET | org | exists | full | authenticated | P0 | org-team |
| `/org/checkName` | GET | org | exists | full | admin | P1 | org-team |
| `/org/updateStatus` | POST | org | exists | full | admin | P1 | org-team |
| `/org/children/:parentId` | GET | org | exists | full | authenticated | P1 | org-team |
| `/org/mounted` | POST | org | exists | stub | authenticated | P2 | org-team |
| `/api/system/organization/list` | GET | organization | exists | full | authenticated | P1 | org-team |

**Go Route Groups:** `/org` (compatibility bridge), `/api/org` (native)

---

## 10. Authentication Routes

| Route Path | HTTP Method | Module | Java Status | Go Status | Auth Mode | Priority | Owner |
|------------|-------------|--------|-------------|-----------|-----------|----------|-------|
| `/login/localLogin` | POST | auth | exists | full | public | P0 | auth-team |
| `/logout` | GET | auth | exists | full | authenticated | P0 | auth-team |

**Go Route Group:** Root level

---

## Summary Statistics

| Module | Total Routes | Java Exists | Go Full | Go Partial | Go Stub | Go Missing |
|--------|--------------|-------------|---------|------------|---------|------------|
| Template Management | 21 | 21 | 5 | 1 | 1 | 14 |
| Datasource | 32 | 30 | 4 | 8 | 8 | 12 |
| Dataset | 23 | 21 | 5 | 8 | 3 | 7 |
| Chart | 14 | 14 | 2 | 1 | 2 | 9 |
| Share | 15 | 15 | 8 | 0 | 0 | 7 |
| Export Center | 9 | 9 | 7 | 1 | 0 | 1 |
| Permission | 4 | 4 | 4 | 0 | 0 | 0 |
| User | 8 | 8 | 8 | 0 | 0 | 0 |
| Organization | 11 | 11 | 9 | 0 | 1 | 1 |
| Authentication | 2 | 2 | 2 | 0 | 0 | 0 |
| **TOTAL** | **139** | **135** | **54** | **11** | **15** | **51** |

---

## Critical Path Routes (P0)

These routes MUST be fully implemented before migration cutover:

### Template Management (P0 Blockers)
- [ ] `/templateManage/templateList` - List templates
- [ ] `/templateManage/save` - Save template (partial, needs full implementation)
- [ ] `/templateMarket/searchTemplate` - Search marketplace templates

### Datasource (P0 Blockers)
- [x] `/datasource/list` - List datasources
- [x] `/datasource/tree` - Datasource tree
- [x] `/datasource/validate` - Validate connection
- [ ] `/datasource/previewData` - Preview data (stub)
- [ ] `/datasource/get/:id` - Get datasource (stub)
- [ ] `/datasource/save` - Save datasource (stub)
- [ ] `/datasource/update` - Update datasource (stub)
- [ ] `/datasource/delete/:id` - Delete datasource (stub)

### Dataset (P0 Blockers)
- [x] `/datasetTree/tree` - Dataset tree
- [ ] `/datasetTree/get/:id` - Get dataset (stub)
- [ ] `/datasetTree/details/:id` - Dataset details (stub)
- [ ] `/datasetTree/save` - Save dataset (stub)
- [ ] `/datasetTree/create` - Create dataset (stub)
- [ ] `/datasetTree/delete/:id` - Delete dataset (stub)
- [x] `/datasetData/tableField` - Table fields
- [x] `/datasetData/previewData` - Preview data

### Chart (P0 Blockers)
- [ ] `/chart/getData` - Get chart data (partial)
- [ ] `/chart/save` - Save chart
- [ ] `/chart/getDetail/:id` - Chart detail (stub)
- [x] `/chartData/getData` - Chart data query

### Export Center (P0 Blockers)
- [x] `/exportCenter/exportTasks` - Export tasks list
- [ ] `/exportCenter/pager` - Paginated list (partial)
- [x] `/exportCenter/delete/*` - Delete operations
- [x] `/exportCenter/download/:id` - Download
- [x] `/exportCenter/retry/:id` - Retry export

---

## Response Contract Semantics

### Standard Response Format

Both Java and Go backends MUST return responses in the following format:

```json
{
  "code": "000000",
  "msg": "success",
  "data": { ... }
}
```

### Error Response Format

```json
{
  "code": "500000",
  "msg": "Error description",
  "data": null
}
```

### Compatibility Stub Response

For unimplemented compatibility endpoints, the Go backend MUST return:

```json
{
  "code": "501000",
  "msg": "Endpoint not yet implemented in Go backend",
  "data": null
}
```

HTTP Status: `501 Not Implemented`

**CRITICAL:** Compatibility endpoints MUST NOT return silent success with empty placeholder payload.

---

## Migration Readiness Checklist

- [ ] All P0 routes have Go status of `full`
- [ ] All compatibility bridge routes return correct response codes
- [ ] Contract diff tests pass for all critical routes
- [ ] Shadow traffic validation passes
- [ ] Row-level permission parity verified
- [ ] Column-level permission parity verified
- [ ] Export task async behavior matches Java implementation

---

## Changelog

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-02-18 | atlas | Initial frozen matrix for SEC-COMP-001 |
