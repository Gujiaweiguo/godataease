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

For local MySQL/Redis on `localhost`, use:

```bash
make run-local
```

You can also override hosts/ports explicitly:

```bash
DATABASE_HOST=127.0.0.1 DATABASE_PORT=3306 REDIS_HOST=127.0.0.1 REDIS_PORT=6379 make run-local
```
