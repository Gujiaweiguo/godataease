# Calcite SQL Integration Compatibility Matrix

## Overview

- Change: `add-calcite-sql-integration`
- Last Updated: `2026-02-24`
- Scope: Calcite parse/validate integration, dataset preview SQL workflow, compatibility bridge SQL endpoint.

## Endpoint Status

| Endpoint | Method | Java Status | Go Status | Evidence |
|---|---|---|---|---|
| `/datasetData/previewSql` | `POST` | `exists` | `partial` | `apps/backend-go/internal/service/dataset_service.go`, `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go` |

## Task-to-Evidence Mapping

| Task | Status | Evidence |
|---|---|---|
| 1.1 ParseSQL/ValidateSQL | done | `apps/backend-go/internal/integration/calcite/calcite.go` |
| 1.2 timeout/retry + structured error | done | `apps/backend-go/internal/integration/calcite/calcite.go`, `apps/backend-go/internal/integration/calcite/calcite_test.go` |
| 1.3 dataset preview wiring | done | `apps/backend-go/internal/service/dataset_service.go`, `apps/backend-go/internal/service/dataset_service_test.go` |
| 1.4 compatibility bridge wiring | done | `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`, `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler_test.go` |
| 1.5 success/invalid/timeout/upstream tests | done | `apps/backend-go/internal/integration/calcite/calcite_test.go` |
| 2.1 integration/service verification | done | `go test ./internal/integration/calcite -v`, `go test ./internal/service -run "TestPreviewSQL_ValidateWithCalciteFirstWhenEnabled|TestPreviewSQL_BlockExecutionWhenCalciteUnavailable" -v` |
| 2.2 contract-diff verification | done | `apps/backend-go/reports/calcite-contract-diff/contract-diff.json`, `apps/backend-go/reports/calcite-contract-diff/contract-diff.md` |

## Contract Diff Result Snapshot

- Engine: `scripts/contract-diff/run_contract_diff.sh`
- Whitelist: `apps/backend-go/testdata/contract-diff/critical-whitelist.yaml`
- Result: `33/33 passed, 0 failed`
- Generated reports:
  - `apps/backend-go/reports/calcite-contract-diff/contract-diff.json`
  - `apps/backend-go/reports/calcite-contract-diff/contract-diff.md`

## Notes

- `goStatus` remains `partial` for `/datasetData/previewSql` because Calcite is integrated with fallback behavior and still tracked under ongoing migration parity scope.
- This matrix is the source-of-truth evidence link for task `2.3` of this change.
