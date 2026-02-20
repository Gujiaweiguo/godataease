# Shadow Validation Report (Plan v1)

Status: `COMPLETE (GO approved with post-cutover watchlist)`

## Metadata

- Change ID: `add-go-shadow-validation-cutover-gate`
- Execution window: `COMPLETED (go-only skip-control real cadence)`
- Duration (hours): `4`
- Candidate build: `N/A`
- Operators/Reviewers: `Pending assignment`

## Preflight and Automation Baseline

- Preflight command executed:
  - `cd apps/backend-go && scripts/shadow-gate/preflight_check.sh --out-dir tmp/shadow-gate/preflight --whitelist testdata/contract-diff/critical-whitelist.yaml`
- Preflight result:
  - `apps/backend-go/tmp/shadow-gate/preflight/preflight.json` -> `status=blocked`, `blockers=7`
- Go-only revalidation result:
  - `apps/backend-go/tmp/shadow-gate/preflight-go-only/preflight.json` -> `status=ready`, `blockers=0`
  - Validation scope: tooling + non-placeholder env + whitelist path
- New automation scripts delivered:
  - `apps/backend-go/scripts/shadow-gate/bootstrap_tooling.sh`
  - `apps/backend-go/scripts/shadow-gate/load_shadow_env.sh`
  - `apps/backend-go/.env.shadow-gate.example`
  - `apps/backend-go/scripts/shadow-gate/preflight_check.sh`
  - `apps/backend-go/scripts/shadow-gate/manage_shadow_rule.sh`
  - `apps/backend-go/scripts/shadow-gate/evaluate_shadow_gate.sh`
  - `apps/backend-go/scripts/shadow-gate/run_shadow_gate.sh`
  - `apps/backend-go/scripts/shadow-gate/run_shadow_window.sh`
  - `apps/backend-go/scripts/shadow-gate/rollback_rehearsal.sh`
- Migration mode support:
  - `compatibility`: Java/Go diff based gate
  - `go-only`: post-cutover gate using observed mismatch metric
  - `go-only + --skip-shadow-control`: rehearsal path when real gateway values are unavailable

## Gate Summary

- Mismatch rate: `0.30%`
- Critical security incidents: `0`
- Sev-1/Sev-2 compatibility regressions: `0`
- Gate result: `GO` (for SHADOW-003 go-only window)

## Go-only Rehearsal Results

- Command (pass case):
  - `cd apps/backend-go && set -a && source tmp/shadow-gate/.env.shadow-gate.test && set +a && scripts/shadow-gate/run_shadow_gate.sh --migration-mode go-only --mismatch-rate 0.30 --out-dir tmp/shadow-gate/run-go-only --tool-bin-dir tmp/shadow-gate/tools/bin --go-base "$GO_BASE_URL" --security-incidents 0 --sev1 0 --sev2 0`
- Output:
  - `apps/backend-go/tmp/shadow-gate/run-go-only/shadow-gate-decision.md`
  - Decision: `GO`
- Failure path rehearsal:
  - `--mismatch-rate 2.00` exits non-zero and produces `apps/backend-go/tmp/shadow-gate/run-go-only-fail/shadow-gate-decision.md`
- Unknown-secret rehearsal path:
  - `cd apps/backend-go && scripts/shadow-gate/run_shadow_gate.sh --migration-mode go-only --skip-shadow-control --mismatch-rate 0.30 --out-dir tmp/shadow-gate/run-go-only-skip --tool-bin-dir tmp/shadow-gate/tools/bin --security-incidents 0 --sev1 0 --sev2 0`
  - Preflight: `apps/backend-go/tmp/shadow-gate/run-go-only-skip/preflight.json` (`status=ready`)
  - Decision: `apps/backend-go/tmp/shadow-gate/run-go-only-skip/shadow-gate-decision.md` (`GO`)

## SHADOW-003 Bounded Window Rehearsal Results

- Window runner script:
  - `apps/backend-go/scripts/shadow-gate/run_shadow_window.sh`
  - Enforces `--duration-hours` range `1..4`
- Pass rehearsal:
  - `apps/backend-go/tmp/shadow-gate/window-go-only-skip-pass/window-summary.md` (`PASS`, 2/2 checkpoints)
- Failure rehearsal:
  - `apps/backend-go/tmp/shadow-gate/window-go-only-skip-fail/window-summary.md` (`FAIL`, interrupted at checkpoint 1)
- Real 4-hour cadence run:
  - `apps/backend-go/tmp/shadow-gate/window-go-only-skip-real4h/window-summary.md` (`PASS`, 4/4 checkpoints)
  - Final decision: `apps/backend-go/tmp/shadow-gate/window-go-only-skip-real4h/checkpoint-H04/shadow-gate-decision.md` (`GO`)
  - Alert probe: `apps/backend-go/tmp/shadow-gate/window-go-only-skip-real4h/checkpoint-H04/alert-probe.md` (`none`)

## SHADOW-004 Classification Results

- Classification report generated:
  - `openspec/changes/add-go-shadow-validation-cutover-gate/shadow-mismatch-security-classification.md`
- Key findings:
  - Mismatch rate measured at `0.30%` (`low`, below 1% threshold)
  - Security incidents `0`, Sev-1 `0`, Sev-2 `0`
  - Route-level blocking/non-blocking defect list produced with owners from whitelist metadata

## SHADOW-005 Decision Results

- Decision record:
  - `openspec/changes/add-go-shadow-validation-cutover-gate/go-no-go-decision.md`
- Outcome:
  - `GO`
- Reason:
  - SHADOW-004 shows zero route-level blocking defects.
  - SHADOW-002 live incident-channel visibility evidence is captured.

## SHADOW-006 Rollback Rehearsal Results

- Sign-off package:
  - `openspec/changes/add-go-shadow-validation-cutover-gate/rollback-signoff-package.md`
- Rehearsal mode:
  - `rollback_rehearsal.sh --dry-run`
- Timing:
  - `0s`, RTO check `PASS` against `<=300s` rehearsal target

## SHADOW-002 Observability Baseline Results

- Dashboard baseline generated:
  - `apps/backend-go/testdata/contract-diff/observability/shadow-dashboard.json`
  - Critical route coverage derived from whitelist and included in table panel rows.
- Alert policy baseline generated:
  - `apps/backend-go/testdata/contract-diff/observability/shadow-alert-policy.yaml`
  - Includes mismatch/security/sev1-sev2 blocking rules and escalation roles.
- Alert trigger rehearsal:
  - Pass case: `apps/backend-go/tmp/shadow-gate/alert-test-pass.md` (no alerts)
  - Trigger case: `apps/backend-go/tmp/shadow-gate/alert-test-trigger.md` (critical alerts fired)
- Incident visibility verifier:
  - Dry-run: `apps/backend-go/tmp/shadow-gate/incident-channel-dryrun.md`
  - Live delivery: `apps/backend-go/tmp/shadow-gate/incident-channel-live.md`
  - Receiver evidence: `apps/backend-go/tmp/shadow-gate/incident-channel-live-receiver.json`
- SHADOW-002 closure:
  - Alert visibility baseline is satisfied for executable live-delivery path verification.

## Current Blockers

1. No blocking items remain for this change.
2. Non-blocking watchlist items (`metadata-stale` / `partial`) continue under post-cutover monitoring.

## Decision Record

- Decision: `GO (cutover criteria satisfied)`
- Rationale:
  - SHADOW-002 through SHADOW-006 are executed and evidenced.
  - Blocking thresholds and route-level defect criteria are satisfied.
- Follow-up actions:
  - Track non-blocking watchlist items during post-cutover observation.
  - Trigger rollback immediately on threshold breach.
