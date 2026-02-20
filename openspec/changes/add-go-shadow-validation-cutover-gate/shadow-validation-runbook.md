# Shadow Validation Runbook (Plan v1)

## Defaults

- Mirror ratio: `5%`
- Observation window: `4h` continuous (hard cap 4h)
- Contract threshold: mismatch rate `< 1%`
- Security threshold: `0` critical incidents
- Stability threshold: `0` Sev-1/Sev-2 regressions

## Scope

- Critical compatibility interfaces defined by `apps/backend-go/testdata/contract-diff/critical-whitelist.yaml`
- Staging Java baseline and Go candidate services
- Gateway-level shadow route rule and switchback controls

## Preconditions

- [ ] SHADOW-001 preflight is `ready` (or approved exception documented)
- [ ] SHADOW-002 dashboards and alerts validated
- [ ] Candidate build is deployed and smoke checks passed

## Required Variables

- `JAVA_BASE_URL`
- `GO_BASE_URL`
- `STAGING_GATEWAY_API`
- `SHADOW_RULE_ID`
- `AUTH_TOKEN`

If Java has already been retired after cutover, use `go-only` mode and provide observed mismatch metric directly.

If real gateway control-plane values are unavailable, use `--skip-shadow-control` for rehearsal-only gating.

Environment bootstrap helpers:

```bash
cd apps/backend-go
cp .env.shadow-gate.example .env.shadow-gate
scripts/shadow-gate/load_shadow_env.sh --env-file .env.shadow-gate --migration-mode go-only
scripts/shadow-gate/bootstrap_tooling.sh --install-dir tmp/shadow-gate/tools/bin
```

Go-only rehearsal bootstrap (no real gateway values required):

```bash
cd apps/backend-go
cp .env.shadow-gate.go-only.example .env.shadow-gate
scripts/shadow-gate/load_shadow_env.sh --env-file .env.shadow-gate --migration-mode go-only --skip-shadow-control
scripts/shadow-gate/bootstrap_tooling.sh --install-dir tmp/shadow-gate/tools/bin
```

## Command-Level Workflow

0. Build observability baseline artifacts (SHADOW-002)

```bash
cd apps/backend-go
scripts/shadow-gate/generate_observability_baseline.py --whitelist testdata/contract-diff/critical-whitelist.yaml --out-dir testdata/contract-diff/observability
scripts/shadow-gate/test_alert_policy.sh --out tmp/shadow-gate/alert-test-pass.md --mismatch-rate 0.30 --security-incidents 0 --sev1 0 --sev2 0
scripts/shadow-gate/test_alert_policy.sh --out tmp/shadow-gate/alert-test-trigger.md --mismatch-rate 1.20 --security-incidents 1 --sev1 0 --sev2 0
scripts/shadow-gate/verify_incident_channel.sh --out tmp/shadow-gate/incident-channel-dryrun.md --dry-run
```

Artifacts:

- `testdata/contract-diff/observability/shadow-dashboard.json`
- `testdata/contract-diff/observability/shadow-alert-policy.yaml`
- `tmp/shadow-gate/alert-test-pass.md`
- `tmp/shadow-gate/alert-test-trigger.md`
- `tmp/shadow-gate/incident-channel-dryrun.md`

Live incident-channel closure command (required to close SHADOW-002):

```bash
cd apps/backend-go
scripts/shadow-gate/verify_incident_channel.sh --webhook-url "$SHADOW_INCIDENT_WEBHOOK_URL" --out tmp/shadow-gate/incident-channel-live.md
```

1. Preflight readiness check

```bash
cd apps/backend-go
scripts/shadow-gate/preflight_check.sh --out-dir tmp/shadow-gate/preflight --whitelist testdata/contract-diff/critical-whitelist.yaml
```

2. Enable shadow rule (5%)

```bash
cd apps/backend-go
scripts/shadow-gate/manage_shadow_rule.sh --action enable --ratio 0.05
```

3. Execute contract diff collection during shadow window

```bash
cd apps/backend-go
scripts/contract-diff/run_contract_diff.sh --java-base "$JAVA_BASE_URL" --go-base "$GO_BASE_URL" --out-dir tmp/shadow-gate/contract-diff --whitelist testdata/contract-diff/critical-whitelist.yaml
```

4. Evaluate Go/No-Go decision from collected evidence

```bash
cd apps/backend-go
scripts/shadow-gate/evaluate_shadow_gate.sh --contract-report tmp/shadow-gate/contract-diff/contract-diff.json --out tmp/shadow-gate/shadow-gate-decision.md --security-incidents 0 --sev1 0 --sev2 0
```

4b. Evaluate Go/No-Go in go-only mode (no Java baseline)

```bash
cd apps/backend-go
scripts/shadow-gate/evaluate_shadow_gate.sh --mismatch-rate 0.30 --out tmp/shadow-gate/shadow-gate-decision.md --security-incidents 0 --sev1 0 --sev2 0
```

4c. One-command go-only rehearsal (no real gateway values)

```bash
cd apps/backend-go
scripts/shadow-gate/run_shadow_gate.sh --migration-mode go-only --skip-shadow-control --mismatch-rate 0.30 --out-dir tmp/shadow-gate/run-go-only --tool-bin-dir tmp/shadow-gate/tools/bin --security-incidents 0 --sev1 0 --sev2 0
```

4d. Bounded shadow window runner (max 4h)

```bash
cd apps/backend-go
scripts/shadow-gate/run_shadow_window.sh --duration-hours 4 --migration-mode go-only --skip-shadow-control --mismatch-rate 0.30 --out-dir tmp/shadow-gate/window-go-only-skip --tool-bin-dir tmp/shadow-gate/tools/bin --security-incidents 0 --sev1 0 --sev2 0
```

5. Disable shadow rule (normal close or rollback)

```bash
cd apps/backend-go
scripts/shadow-gate/manage_shadow_rule.sh --action disable
```

6. Rollback rehearsal (can run as dry-run before real switchback)

```bash
cd apps/backend-go
scripts/shadow-gate/rollback_rehearsal.sh --dry-run
```

## Exit Criteria

- [ ] `>= 4h` continuous evidence collected and total window `<= 4h`
- [ ] Contract mismatch rate `< 1%`
- [ ] Critical security incidents = `0`
- [ ] Sev-1/Sev-2 compatibility regressions = `0`
- [ ] Decision report is produced in `tmp/shadow-gate/shadow-gate-decision.md`

## Rollback Trigger

Trigger switchback immediately if any blocking threshold is violated:

- mismatch rate `>= 1%`
- any critical security incident
- any Sev-1/Sev-2 compatibility regression

Rollback command:

```bash
cd apps/backend-go
scripts/shadow-gate/manage_shadow_rule.sh --action disable
```
