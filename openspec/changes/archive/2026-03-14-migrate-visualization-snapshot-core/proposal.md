## Why

The current Go visualization implementation only mirrors `data_visualization_info` into `snapshot_data_visualization_info`, while the legacy Java implementation treats visualization editing, publishing, recovery, deletion, and copy as coordinated multi-table operations across metadata, chart views, linkage, jump, and outer-parameter data. We need to align the Go backend with that behavior now because recent compatibility fixes exposed that the current simplified model cannot preserve legacy publish/recover semantics once visualization workflows move beyond metadata-only changes.

## What Changes

- Align visualization lifecycle semantics so `snapshot_*` represents draft/edit state and core/main tables represent published state.
- Expand Go visualization persistence from metadata-only mirroring to coordinated multi-table snapshot/core synchronization.
- Add Go-side support for snapshot/core chart-view persistence and lifecycle orchestration used by save, publish, recover, delete, and copy workflows.
- Extend visualization lifecycle handling to include linkage, jump, and outer-parameter child data that already exists as generated models in Go.
- Define phased migration rules so Go can adopt legacy-compatible behavior without rewriting unrelated modules.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `visualization-management`: Change visualization lifecycle requirements so draft editing, publishing, recovery, and deletion follow legacy-compatible multi-table snapshot/core behavior instead of metadata-only mirroring.

## Impact

- Affected Go code: `apps/backend-go/internal/domain/visualization`, `internal/domain/auto`, `internal/repository`, `internal/service/visualization_service.go`, and `internal/transport/http/handler/visualization_handler.go`.
- Affected persistence model: `data_visualization_info`, `snapshot_data_visualization_info`, `core_chart_view`, missing `snapshot_core_chart_view`, existing visualization linkage/jump/outer-params snapshot tables, and threshold-related integration points.
- Affected APIs: visualization save/publish/recover/delete/copy compatibility endpoints and any handlers that currently treat publish/recover as metadata-only operations.
- Affected migration scope: phased backend implementation, integration tests, and future compatibility verification against legacy Java visualization behavior.
