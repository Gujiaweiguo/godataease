## Why

Frontend sync workflows already depend on dedicated `/sync/datasource/*`, `/sync/task/*`, `/sync/task/log/*`, and `/sync/summary/*` APIs, but the Go mainline only exposes partial datasource-side sync entrypoints. This leaves the sync center unusable on Go mainline and keeps SeaTunnel-based sync orchestration inaccessible from the primary frontend flows.

## What Changes

- Add a dedicated Go sync module HTTP surface for the frontend-consumed sync datasource, sync task, sync task log, and sync summary routes.
- Reuse and extend existing Go-side SeaTunnel orchestration and sync record persistence so sync task operations have deterministic lifecycle semantics.
- Define parity-safe request and response behavior for paging, status transitions, execution, cancellation, validation, and task log retrieval.
- Add verification coverage for sync routes at handler, service, and compatibility levels.

## Capabilities

### New Capabilities
- `sync-module-management`: Manage sync datasources, sync tasks, sync task logs, and sync summary APIs required by the frontend sync center.

### Modified Capabilities

## Impact

- Affected code likely includes:
  - `apps/backend-go/internal/integration/seatunnel/*`
  - `apps/backend-go/internal/service/*`
  - `apps/backend-go/internal/repository/*`
  - `apps/backend-go/internal/transport/http/handler/*`
- Affected APIs:
  - `/sync/datasource/*`
  - `/sync/task/*`
  - `/sync/task/log/*`
  - `/sync/summary/*`
- Related runtime concerns:
  - SeaTunnel task submission/status/cancel flows
  - persisted sync task and log records
  - frontend parity for sync-center pages
