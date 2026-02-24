# Change: Add Calcite SQL Integration

## Why
The Go backend still returns `not implemented` for Calcite SQL parsing and validation, which leaves dataset SQL workflows without parity to Java behavior.

## What Changes
- Add production-ready Calcite gRPC integration for SQL parse and validate operations.
- Define deterministic timeout, retry, and error mapping behavior for Calcite RPC calls.
- Integrate Calcite validation into dataset SQL preview and related compatibility workflows.

## Impact
- Affected specs: `calcite-sql-integration`, `datasource-management`
- Affected code:
  - `apps/backend-go/internal/integration/calcite/calcite.go`
  - `apps/backend-go/internal/service/dataset_service.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
