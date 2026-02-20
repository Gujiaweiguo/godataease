# Shadow Observability Baseline

Generated artifacts for SHADOW-002:

- `shadow-dashboard.json`
- `shadow-alert-policy.yaml`

Generate or refresh:

```bash
cd apps/backend-go
scripts/shadow-gate/generate_observability_baseline.py --whitelist testdata/contract-diff/critical-whitelist.yaml --out-dir testdata/contract-diff/observability
```

Alert rehearsal:

```bash
cd apps/backend-go
scripts/shadow-gate/test_alert_policy.sh --out tmp/shadow-gate/alert-test-pass.md --mismatch-rate 0.30 --security-incidents 0 --sev1 0 --sev2 0
scripts/shadow-gate/test_alert_policy.sh --out tmp/shadow-gate/alert-test-trigger.md --mismatch-rate 1.20 --security-incidents 1 --sev1 0 --sev2 0
```
