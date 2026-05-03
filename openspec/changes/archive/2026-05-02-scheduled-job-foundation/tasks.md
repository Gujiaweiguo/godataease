## 1. Registry and startup wiring

- [x] 1.1 Inventory the current scheduler startup path and document where scheduled jobs can be registered without scattering direct `AddFunc` calls across modules.
- [x] 1.2 Introduce a centralized job registry structure that captures stable job metadata (`job key`, `cron expression`, `description`, `enabled state`) and add unit tests for registry validation.
- [x] 1.3 Wire the application startup path to load only enabled jobs from the centralized registry and verify with tests that disabled jobs are not registered.

## 2. Execution wrapper and distributed lock semantics

- [x] 2.1 Add a shared scheduled-job execution wrapper that records `success`, `skipped`, and `failed` outcomes, and cover each outcome with unit tests.
- [x] 2.2 Refine Redis-backed distributed execution so lock contention is classified as `skipped` instead of `failed`, with tests covering lock-not-acquired behavior.
- [x] 2.3 Ensure lock release/expiry behavior remains safe after job execution and add tests for normal completion and error paths.

## 3. Foundation sample job

- [x] 3.1 Select and implement one low-risk sample job that exercises the full registry → scheduler → distributed wrapper path without mutating core business data, and verify it in tests.
- [x] 3.2 Register the sample job through the centralized registry and verify that it can be enabled, triggered, and observed through the new execution outcome model.

## 4. Observability, rollback, and verification

- [x] 4.1 Add diagnostic logging or equivalent observable execution output for every scheduled-job attempt and verify that operators can distinguish `success`, `skipped`, and `failed` runs.
- [x] 4.2 Add configuration or registration controls that allow the runtime to return to a no-job-active state without deleting code, and verify rollback behavior with tests.
- [x] 4.3 Update runbook or developer/operator documentation for enabling, disabling, validating, and rolling back scheduled jobs, and verify the documented commands against the implementation.
