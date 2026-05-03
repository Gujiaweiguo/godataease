# Tasks: Go Threshold Support (First Slice)

Tasks are ordered by dependency. Each task is sized to be completable in approximately 2 hours or less. Verification criteria are explicit.

---

## Phase 1: Domain & Foundation

### Task 1.1 — Define threshold domain types
**Files**: `internal/domain/threshold/threshold.go`
**What**: Create Go request/response DTO structs mirroring the Java DTOs:
- `CreateRequest` (maps to Java `ThresholdCreator`)
- `BaseReciDTO` (recipient configuration)
- `GridRequest` / `GridVO` (pager filter + response)
- `InstanceRequest` / `InstanceVO` (instance history)
- `SwitchRequest` (enable/disable)
- `BatchReciRequest` (batch recipient update)
- `PreviewRequest` (preview evaluation)
- `PreviewResponse` (preview result string)

**Verify**: `go build` succeeds, types compile.

### Task 1.2 — Define filter tree domain types
**Files**: `internal/domain/threshold/filter_tree.go`
**What**: Create Go structs for the threshold rules JSON format:
- `FilterTreeObj` (recursive: items + logic "and"/"or" + sub-trees)
- `FilterTreeItem` (fieldId, term, value, filterType, enumVal, deType, etc.)
- JSON tags matching the Java serialization format

**Verify**: Unit test: deserialize known Java filter-tree JSON samples, verify round-trip.

### Task 1.3 — Create threshold repository interface and implementation
**Files**: `internal/repository/threshold_repo.go`
**What**: Implement `ThresholdRepository` concrete struct with GORM:
- `Create`, `Update`, `GetByID`, `DeleteByIDs`, `DeleteByChartID`
- `UpdateEnable`, `UpdateRecipients`
- `Pager` (with keyword, status, enable, resourceType, chartId filters)
- `InstancePager` (with thresholdId, keyword filters)
- `ExistsByChartID`

Define `ThresholdRepository` interface in `internal/service/threshold_service.go`.

**Verify**: `go build` succeeds. Integration tests in Task 3.1.

---

## Phase 2: Core Service Logic

### Task 2.1 — Implement threshold evaluator (matching engine)
**Files**: `internal/service/threshold_evaluator.go`
**What**: Port core matching logic from Java `ChartViewThresholdManage`:
- `EvaluateThreshold()` — main entry point
- `MatchesConditionTree()` — recursive AND/OR tree walker
- `RowMatch()` — per-row per-condition matching
- `FormatDynamicValue()` — min/max/avg from data rows
- Support operators: eq, not_eq, gt, ge, lt, le (numeric/time); eq, not_eq, in, not_in, like, not_like, null, not_null, empty, not_empty (string)

**Verify**: Comprehensive unit tests with table-driven cases covering:
- Each operator for each data type
- AND/OR nesting
- Dynamic value resolution
- Malformed tree handling

### Task 2.2 — Implement threshold evaluator (text conversion)
**Files**: `internal/service/threshold_evaluator.go` (continued)
**What**: Port `ConvertRulesToText()` logic:
- Convert filter tree to human-readable string
- Field name resolution from chart metadata
- Replace `[触发告警]` and `[检测时间]` placeholders

**Verify**: Unit tests with known Java input/output pairs.

### Task 2.3 — Implement threshold evaluator (HTML preview generation)
**Files**: `internal/service/threshold_evaluator.go` (continued)
**What**: Port preview HTML generation:
- Template substitution (`[检测时间]`, `[触发告警]`, `[告警数据]`)
- Field value replacement in `<span id="changeText-{fieldId}">`
- HTML table generation for alert data (limited by `thresholdLimit`)
- `showFieldValue` support

**Verify**: Unit tests comparing output with expected HTML fragments.

### Task 2.4 — Implement threshold service (CRUD methods)
**Files**: `internal/service/threshold_service.go`
**What**: Implement service methods:
- `Create(ctx, req)` — validate, derive creator/org from context, persist
- `Edit(ctx, req)` — validate, update existing
- `FormInfo(ctx, id, resourceTable)` — load and return creator-shaped response
- `SwitchEnable(ctx, req)` — toggle enable field
- `Delete(ctx, ids, resourceTable)` — remove by IDs
- `DeleteWithChart(ctx, chartId, resourceTable)` — remove by chart ID
- `BatchReci(ctx, req)` — update recipient fields on multiple records
- `Pager(ctx, req, goPage, pageSize)` — paginated list with filters
- `AnyThreshold(ctx, chartId, resourceTable)` — boolean existence check
- `InstancePager(ctx, req, goPage, pageSize)` — instance history pagination

**Verify**: `go build` succeeds. Integration tests in Task 3.2.

### Task 2.5 — Implement threshold service (preview method)
**Files**: `internal/service/threshold_service.go` (continued)
**What**: Implement `Preview(ctx, req)`:
- Fetch chart metadata via `ChartService`
- Fetch chart data rows via `ChartRepository`
- Parse threshold rules JSON into filter tree
- Call evaluator to get matching rows
- Generate HTML preview
- Return string content

**Verify**: Integration test with real chart data in Task 3.2.

---

## Phase 3: Handler & Routes

### Task 3.1 — Implement threshold handler
**Files**: `internal/transport/http/handler/threshold_handler.go`
**What**: Create `ThresholdHandler` struct and `RegisterThresholdRoutes()` function:
- Handler methods: Save, Edit, Pager, FormInfo, SwitchEnable, Delete, BatchReci, InstancePager, Preview, AnyThreshold, DeleteWithChart
- Route registration with JWT auth middleware
- Request binding and response envelope usage (following existing handler patterns)
- `defer recoverServicePanic(c)` in each method

**Verify**: `go build` succeeds.

### Task 3.2 — Wire threshold module in router
**Files**: `internal/transport/http/router.go`
**What**:
- Add `thresholdHandler` field to `Router` struct
- Construct `thresholdRepo → thresholdService → thresholdHandler` in `NewRouter()`
- Inject `chartService` and `chartRepo` into `thresholdService` via setters
- Call `handler.RegisterThresholdRoutes()` in `registerAPIRoutes()` under `/threshold` group
- Register under both `/api/` and `/de2api/` base paths

**Verify**: `go build` succeeds. Manual smoke test: `curl /api/threshold/pager/1/10` returns valid response.

---

## Phase 4: Testing

### Task 4.1 — Repository integration tests
**Files**: `internal/repository/threshold_repo_integration_test.go`
**What**: Integration tests using existing MySQL test suite:
- CRUD round-trip (create → get → update → get → delete)
- Pager with various filter combinations
- InstancePager queries
- DeleteByChartID cascading
- ExistsByChartID checks

**Verify**: `TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test make test-integration` passes.

### Task 4.2 — Service unit tests
**Files**: `internal/service/threshold_service_test.go`, `internal/service/threshold_evaluator_test.go`
**What**:
- Evaluator pure-function tests (table-driven, all operators, all types)
- Service CRUD tests with fake repository
- Preview logic tests with mock chart data
- Error handling tests (malformed rules, missing chart, unsupported type)

**Verify**: `go test ./internal/service/... -run Threshold -v` passes.

### Task 4.3 — Service integration tests
**Files**: `internal/service/threshold_service_integration_test.go`
**What**: End-to-end service tests with real DB:
- Create threshold → FormInfo round-trip
- Preview with real chart data
- Pager filtering correctness
- Instance listing

**Verify**: `TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -tags=integration -v ./internal/service/... -run Threshold` passes.

### Task 4.4 — Handler smoke tests
**Files**: `internal/transport/http/handler/threshold_handler_test.go`
**What**: HTTP handler tests (using `httptest` or similar):
- Verify route registration produces reachable endpoints
- Verify request binding and response envelope format
- Verify error responses for malformed requests

**Verify**: `go test ./internal/transport/http/handler/... -run Threshold -v` passes.

---

## Task Dependency Graph

```
1.1 ─┐
1.2 ─┤
     ├── 2.1 ── 2.2 ── 2.3 ──┐
1.3 ─┤                        ├── 2.4 ── 2.5 ── 3.1 ── 3.2 ── 4.1
     │                        │                              4.2
     │                        │                              4.3
     └────────────────────────┘                              4.4
```

Critical path: 1.1/1.2/1.3 → 2.1 → 2.2 → 2.3 → 2.4 → 2.5 → 3.1 → 3.2

Testing tasks (4.x) can start as soon as their corresponding implementation tasks are complete.

---

## Effort Estimate

| Phase | Tasks | Est. Hours |
|-------|-------|------------|
| Phase 1: Domain & Foundation | 1.1, 1.2, 1.3 | 4h |
| Phase 2: Core Service Logic | 2.1, 2.2, 2.3, 2.4, 2.5 | 10h |
| Phase 3: Handler & Routes | 3.1, 3.2 | 3h |
| Phase 4: Testing | 4.1, 4.2, 4.3, 4.4 | 6h |
| **Total** | **14 tasks** | **~23h** |
