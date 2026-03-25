# DataEase Go Backend

Go implementation of DataEase backend services.

## Structure

- `cmd/` - Application entry points
- `internal/` - Private application code
- `configs/` - Configuration files
- `deployments/` - Docker and Kubernetes manifests

## Build

```bash
make build
```

## Run

```bash
make run
```

For local backend development with an external MySQL and Docker Redis, use:

```bash
DATABASE_HOST=<external-mysql-host> \
DATABASE_PORT=3306 \
DATABASE_NAME=dataease_dev \
DATABASE_USER=root \
DATABASE_PASSWORD=<password> \
REDIS_HOST=127.0.0.1 \
REDIS_PORT=16379 \
make run-local
```

You can also override hosts/ports explicitly:

```bash
DATABASE_HOST=127.0.0.1 DATABASE_PORT=3306 REDIS_HOST=127.0.0.1 REDIS_PORT=16379 make run-local
```

## Integration Config (Optional)

The backend supports optional gRPC integrations for Calcite and SeaTunnel.

```env
CALCITE_GRPC_ADDR=
CALCITE_GRPC_TIMEOUT_SEC=10
CALCITE_GRPC_MAX_RETRIES=1
SEATUNNEL_GRPC_ADDR=
SEATUNNEL_GRPC_TIMEOUT_SEC=15
SEATUNNEL_GRPC_MAX_RETRIES=1
```

- Keep `CALCITE_GRPC_ADDR` empty to disable Calcite SQL validation integration.
- Keep `SEATUNNEL_GRPC_ADDR` empty to disable SeaTunnel sync orchestration integration.
- These values are loaded through `integration.calcite` and `integration.seatunnel` in `configs/config.yaml`.
