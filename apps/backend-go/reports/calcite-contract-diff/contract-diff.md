# Contract Diff Report

**Generated:** 2026-02-24T14:41:39Z

## Summary

| Metric | Value |
|--------|-------|
| Total APIs | 33 |
| Passed | 33 |
| Failed | 0 |
| Parity | 100.0% |

## Configuration

- **Java Backend:** `http://127.0.0.1:18081`
- **Go Backend:** `http://127.0.0.1:18082`
- **Whitelist:** `testdata/contract-diff/critical-whitelist.yaml`
- **Timeout:** 30s
- **Retries:** 3

## Passed APIs

| Path | Method | Priority | Blocking |
|------|--------|----------|----------|
| `/templateManage/templateList` | POST | P0 | critical |
| `/templateManage/save` | POST | P0 | critical |
| `/templateMarket/searchTemplate` | GET | P0 | critical |
| `/datasource/list` | POST | P0 | critical |
| `/datasource/tree` | POST | P0 | critical |
| `/datasource/validate` | POST | P0 | critical |
| `/datasource/getTables` | POST | P0 | critical |
| `/datasource/previewData` | POST | P0 | critical |
| `/datasource/save` | POST | P0 | critical |
| `/datasource/delete/:id` | GET | P0 | critical |
| `/datasetTree/tree` | POST | P0 | critical |
| `/datasetData/tableField` | POST | P0 | critical |
| `/datasetData/previewData` | POST | P0 | critical |
| `/chartData/getData` | POST | P0 | critical |
| `/chart/getData` | POST | P0 | critical |
| `/exportCenter/exportTasks` | GET | P0 | critical |
| `/exportCenter/download/:id` | GET | P0 | critical |
| `/user/list` | POST | P0 | critical |
| `/user/create` | POST | P0 | critical |
| `/login/localLogin` | POST | P0 | critical |
| `/logout` | GET | P0 | critical |
| `/templateManage/delete` | POST | P1 | high |
| `/templateMarket/categories` | GET | P1 | high |
| `/datasource/types` | POST | P1 | high |
| `/datasource/syncApiTable` | POST | P1 | high |
| `/datasource/syncApiDs` | POST | P1 | high |
| `/datasource/listSyncRecord/:dsId/:page/:limit` | POST | P1 | high |
| `/datasetTree/get/:id` | POST | P1 | high |
| `/datasetData/previewSql` | POST | P1 | high |
| `/share/create` | POST | P0 | critical |
| `/share/validate` | POST | P0 | critical |
| `/org/list` | GET | P0 | critical |
| `/org/tree` | GET | P0 | critical |

