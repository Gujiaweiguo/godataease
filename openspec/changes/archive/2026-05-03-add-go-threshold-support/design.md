# Design: Go Threshold Support (First Slice)

## Overview

Add threshold definition CRUD, preview/matching engine, and instance history to the Go backend. This is the first viable slice: tables and GORM models already exist (`xpack_threshold_info`, `xpack_threshold_instance`), but Go has no repository, service, handler, or routes. Scheduled dispatch, external notification channels, report, and data-filling are explicitly out of scope.

---

## 1. Layering & File Layout

Follow the established Go backend pattern (chart module as reference):

```
internal/
├── domain/
│   ├── auto/
│   │   ├── xpack_threshold_info.gen.go        # existing generated model
│   │   └── xpack_threshold_instance.gen.go    # existing generated model
│   └── threshold/
│       ├── threshold.go                        # domain DTOs: CreateRequest, GridRequest, GridVO, etc.
│       └── filter_tree.go                      # FilterTreeObj / FilterTreeItem deserialization types
├── repository/
│   └── threshold_repo.go                       # ThresholdRepository (concrete struct)
├── service/
│   ├── threshold_service.go                    # ThresholdService + ThresholdRepo interface
│   └── threshold_evaluator.go                  # filter-tree matching engine (pure functions)
└── transport/http/
    └── handler/
        └── threshold_handler.go                # ThresholdHandler + RegisterThresholdRoutes()
```

### Wiring (in `internal/transport/http/router.go`)

```
NewRouter():
  thresholdRepo    = repository.NewThresholdRepository(db)
  thresholdService = service.NewThresholdService(thresholdRepo)
  thresholdService.SetChartService(chartService)           // for preview data access
  thresholdService.SetChartRepo(chartRepo)                 // for chart metadata
  thresholdHandler = handler.NewThresholdHandler(thresholdService)
  // stored in Router struct

registerAPIRoutes():
  handler.RegisterThresholdRoutes(thresholdGroup, r.thresholdHandler, r.permMiddleware)
```

Cross-cutting dependencies injected via setter methods (same as `ChartService.SetRowPermissionService()`):
- `SetChartService(*ChartService)` — needed for preview to resolve chart data
- `SetChartRepo(ChartRepository)` — needed for chart metadata in rule evaluation

### Repository Interface (defined in service layer)

```go
type ThresholdRepository interface {
    Create(ctx context.Context, info *auto.XpackThresholdInfo) error
    Update(ctx context.Context, info *auto.XpackThresholdInfo) error
    GetByID(ctx context.Context, id int64) (*auto.XpackThresholdInfo, error)
    DeleteByIDs(ctx context.Context, ids []int64) error
    DeleteByChartID(ctx context.Context, chartID int64) error
    UpdateEnable(ctx context.Context, id int64, enable bool) error
    UpdateRecipients(ctx context.Context, ids []int64, users, roles, emails, larkGroups, larksuiteGroups, webhooks string) error
    Pager(ctx context.Context, req *threshold.GridRequest, goPage, pageSize int) ([]*threshold.GridVO, int64, error)
    ExistsByChartID(ctx context.Context, chartID int64) (bool, error)
    InstancePager(ctx context.Context, req *threshold.InstanceRequest, goPage, pageSize int) ([]*threshold.InstanceVO, int64, error)
}
```

---

## 2. Threshold Rule Engine

### Scope

Port the core condition-matching engine from Java `ChartViewThresholdManage` as a **reusable evaluator** used by:
1. **Preview endpoint** (`POST /threshold/preview`) — evaluate a proposed rule against live chart data and return HTML preview
2. **Future scheduler** — the same evaluator will be called by scheduled jobs (not in this slice)

### Architecture

```
threshold_evaluator.go (pure functions, no DB dependency)
├── EvaluateThreshold(rules FilterTreeObj, rows []map[string]interface{}, fields FieldMap) ([]map[string]interface{}, error)
├── MatchesConditionTree(row, tree, fieldMap) bool
├── MatchesConditionItem(row, item, fieldMap) bool
├── RowMatch(row, item, fieldDTO) bool
├── FormatDynamicValue(rows, item) string    // min/max/avg
└── ConvertRulesToText(chart, rules) string   // human-readable summary
```

**Key design decisions:**

1. **FilterTreeObj as domain type** — Define Go structs matching the Java `FilterTreeObj` / `FilterTreeItem` JSON structure. Stored as `longtext` in DB, deserialized on evaluation.

2. **Operator support** — Port the full operator set from Java for data types:
   - String (deType=0): eq, not_eq, in, not_in, like, not_like, null, not_null, empty, not_empty
   - Numeric (deType=2/3): eq, not_eq, gt, ge, lt, le
   - Time (deType=1): eq, not_eq, gt, ge, lt, le (numeric comparison after stripping non-digits)

3. **Dynamic values** — Resolve `min`/`max`/`average` from actual data rows before matching (same as Java). Dynamic time values (relative dates) supported but simplified.

4. **Preview HTML generation** — Port the template substitution logic:
   - Replace `[检测时间]` with current timestamp
   - Replace `[触发告警]` with human-readable rule text
   - Replace `<span id="changeText-{fieldId}">` with actual field values from matching rows
   - Support `[告警数据]` placeholder with HTML table of matching rows (capped by `thresholdLimit`, default 5)

5. **No scheduler/notifications** — The evaluator is purely a data-matching engine. Notification dispatch, Quartz scheduling, and external channel integration are explicitly deferred.

### Error handling

- Chart data empty or fetch failure → return `valid=false` with error message (not panic)
- Unsupported chart type → return `valid=false` with chart type info
- Malformed filter tree JSON → return error, do not silently skip
- No matching rows → return `valid=false` with no error message (same as Java behavior)

---

## 3. Compatibility Route Surface

All routes mirror the legacy Java `ThresholdApi` surface, registered under both `/api/threshold/` and `/de2api/threshold/` base paths.

### Route Table

| Method | Path | Handler | Auth | Permission |
|--------|------|---------|------|------------|
| POST | `/threshold/save` | Save | JWT | menu-auth |
| POST | `/threshold/edit` | Edit | JWT | menu-auth |
| POST | `/threshold/pager/:goPage/:pageSize` | Pager | JWT | menu-auth |
| GET | `/threshold/formInfo/:id/:resourceTable` | FormInfo | JWT | menu-auth |
| POST | `/threshold/switch` | SwitchEnable | JWT | menu-auth |
| POST | `/threshold/delete/:resourceTable` | Delete | JWT | menu-auth |
| POST | `/threshold/batchReci` | BatchReci | JWT | menu-auth |
| POST | `/threshold/instancePager/:goPage/:pageSize` | InstancePager | JWT | menu-auth |
| POST | `/threshold/preview` | Preview | JWT | menu-auth |
| GET | `/threshold/anyThreshold/:chartId/:resourceTable` | AnyThreshold | JWT | none |
| GET | `/threshold/deleteWithChart/:chartId/:resourceTable` | DeleteWithChart | JWT | menu-auth |

### Compatibility notes

- `resourceTable` path parameter accepts `"core"` or `"snapshot"` — this slice implements `"core"` only; `"snapshot"` returns empty results without error.
- Request/response payloads match the Java DTO shapes exactly (field names, nesting, defaults).
- `IPage<ThresholdGridVO>` Java response maps to Go `{list: [], total: N, current: P, size: S}` envelope.
- `formInfo` returns a `ThresholdCreator`-shaped response (all editable fields + recipient lists).

---

## 4. Recipient Fields & Repeat-Send Metadata

### Decision: Store but do not dispatch

The `xpack_threshold_info` table already has recipient columns (`reci_users`, `reci_roles`, `reci_emails`, `reci_lark_groups`, `reci_larksuite_groups`, `reci_webhooks`) and `repeat_send`.

**This slice:**
- **Reads and writes** all recipient fields in CRUD operations (create/edit/formInfo/batchReci)
- **Stores** `reciFlagList` as a serialized list in a JSON column or as a structured field alongside recipients
- **Does NOT** implement notification dispatch or channel resolution
- **Stores** `repeat_send` as-is; the field is returned in formInfo but has no behavioral effect yet

### Recipient field handling

- `BaseReciDTO` fields (uidList, ridList, emailList, larkGroupList, larksuiteGroupList, webhookList) are serialized to JSON strings for storage in their respective DB columns
- `reciFlagList` is serialized alongside as a JSON-encoded `[]int`
- `batchReci` updates only the recipient columns, leaving other threshold fields unchanged

---

## 5. Test Strategy

### Unit Tests

**File**: `internal/service/threshold_service_test.go`

- Test the evaluator with hand-crafted filter trees and data rows
- Cover each operator for each data type (string, numeric, time)
- Test dynamic value resolution (min, max, average)
- Test malformed rule handling
- Use manual fake repository (following `fakeChartRepo` pattern)

**File**: `internal/service/threshold_evaluator_test.go`

- Pure function tests for `MatchesConditionTree`, `RowMatch`, `FormatDynamicValue`
- Table-driven tests for operator coverage

### Integration Tests

**File**: `internal/repository/threshold_repo_integration_test.go` (`//go:build integration`)

- CRUD against real MySQL (existing integration test suite pattern)
- Pager queries with filters
- Instance pager queries
- Chart-linked lookup and deletion

**File**: `internal/service/threshold_service_integration_test.go` (`//go:build integration`)

- End-to-end preview: create chart, create threshold, call preview, verify matching
- FormInfo round-trip: create threshold, load formInfo, verify all fields

### Test data

- Reuse existing chart integration test data setup
- Add threshold-specific test fixtures (filter tree JSON samples, expected match results)

---

## 6. Out of Scope (This Slice)

- Scheduled threshold execution (Quartz → Go scheduler)
- Notification dispatch (email, lark, webhook, in-app)
- Report module threshold integration
- Data-filling module threshold integration
- `resourceTable = "snapshot"` behavior
- Chart senior visual threshold (conditional formatting) — separate feature already handled in chart rendering
- Threshold job CRUD (create/delete Quartz jobs on save/delete)
