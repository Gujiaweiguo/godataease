# Release Note: gRPC Integration Config Unification

Date: 2026-02-24
Scope: Go backend config and deployment docs

## Summary

- Unified Calcite and SeaTunnel integration settings into the backend config path.
- Removed direct service-level environment reads for integration addresses.
- Added deployment-facing environment variable templates and compose passthrough.

## What Changed

- Config model and env binding:
  - `apps/backend-go/internal/app/config.go`
  - New config keys:
    - `integration.calcite.address`
    - `integration.calcite.timeout_sec`
    - `integration.calcite.max_retries`
    - `integration.seatunnel.address`
    - `integration.seatunnel.timeout_sec`
    - `integration.seatunnel.max_retries`
- Runtime wiring (config -> service):
  - `apps/backend-go/internal/transport/http/router.go`
  - Uses `SetCalciteConfig` and `SetSeatunnelConfig` during service initialization.
- Service cleanup:
  - `apps/backend-go/internal/service/dataset_service.go`
  - `apps/backend-go/internal/service/datasource_service.go`
  - Removed direct `os.Getenv("CALCITE_GRPC_ADDR")` / `os.Getenv("SEATUNNEL_GRPC_ADDR")` reads.
- Deployment and docs updates:
  - `apps/backend-go/configs/config.yaml`
  - `apps/backend-go/configs/config.example.yaml`
  - `infra/compose/.env.example`
  - `infra/compose/.env.prod.example`
  - `infra/compose/docker-compose.yml`
  - `apps/backend-go/README.md`
  - `docs/quick-start.md`

## Environment Variables

- `CALCITE_GRPC_ADDR` (empty means disabled)
- `CALCITE_GRPC_TIMEOUT_SEC` (default `10`)
- `CALCITE_GRPC_MAX_RETRIES` (default `1`)
- `SEATUNNEL_GRPC_ADDR` (empty means disabled)
- `SEATUNNEL_GRPC_TIMEOUT_SEC` (default `15`)
- `SEATUNNEL_GRPC_MAX_RETRIES` (default `1`)

## Operational Notes

- Configuration now has a single source of truth: app config.
- Compose templates now pass the integration env vars into `dataease-app` container.
- If `*_ADDR` is empty, corresponding integration remains disabled without fallback implicit reads.

## Verification

- `go build ./...`
- `go test ./internal/service/... -count=1`
- `go test ./internal/integration/calcite ./internal/integration/seatunnel -count=1`
- `go test ./internal/transport/http/handler/... -count=1`
- `docker compose -f infra/compose/docker-compose.yml config`
