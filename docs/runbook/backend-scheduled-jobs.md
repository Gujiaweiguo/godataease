# Backend Scheduled Jobs Runbook

## Scope

This runbook documents the scheduled-job foundation introduced for the Go backend. The current foundation uses a centralized registry, `robfig/cron`, and Redis-backed single-node execution locking.

## Current startup path

The backend startup path is:

1. `cmd/api/main.go` initializes application config and logger.
2. Database and Redis clients are initialized.
3. `internal/job/jobs.NewRegistry` builds the centralized scheduled-job registry.
4. `internal/job.NewScheduler` loads enabled jobs from the registry and starts the scheduler.
5. HTTP serving starts after scheduler registration completes.

Scheduled jobs MUST be registered through the centralized registry rather than direct ad hoc `AddFunc` calls in business module startup logic.

## Current sample job

- Job key: `scheduler-foundation-sample-heartbeat`
- Purpose: prove the registry → scheduler → distributed wrapper → diagnostics path is live
- Risk profile: low-risk logging heartbeat only, no core business data mutation

## Enable or disable the sample job

Use configuration or environment variable:

- Config key: `scheduler.sample_job_enabled`
- Env var: `SCHEDULER_SAMPLE_JOB_ENABLED`

Example:

```bash
SCHEDULER_SAMPLE_JOB_ENABLED=true make run-local
```

Rollback to a no-job-active state by disabling the sample job:

```bash
SCHEDULER_SAMPLE_JOB_ENABLED=false make run-local
```

## Verification commands

Run focused scheduler tests:

```bash
cd apps/backend-go
go test ./internal/job/...
```

Run backend baseline validation:

```bash
cd apps/backend-go
make test
```

## Expected diagnostics

Each scheduled-job attempt should emit an outcome that distinguishes:

- `success`
- `skipped`
- `failed`

Lock contention in multi-node execution MUST appear as `skipped`, not `failed`.

## Rollback guidance

If scheduled-job activation causes operational issues:

1. Disable affected jobs through registry-backed configuration.
2. Restart the backend so no scheduled jobs are registered.
3. Re-run `go test ./internal/job/...` to confirm the foundation still passes with jobs disabled.
4. Only remove code in a later follow-up if the registry/config rollback is insufficient.
