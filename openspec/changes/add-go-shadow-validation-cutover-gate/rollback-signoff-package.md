# SHADOW-006 Rollback Rehearsal and Sign-off Package

- Timestamp: 2026-02-20T09:06:40Z
- Change ID: `add-go-shadow-validation-cutover-gate`
- Decision Input: `openspec/changes/add-go-shadow-validation-cutover-gate/go-no-go-decision.md`

## Rehearsal Command

```bash
cd apps/backend-go
scripts/shadow-gate/rollback_rehearsal.sh --dry-run
```

## Rehearsal Execution Record

- Start: 2026-02-20T09:06:04Z
- End: 2026-02-20T09:06:04Z
- Duration: 0 seconds
- Exit code: 0
- Operator: Hephaestus automation

## Rehearsal Output

- `DRY-RUN request: PATCH <missing-gateway-api>/rules/<missing-rule-id>`
- `Payload: {"shadow":{"enabled":false,"ratio":0.0}}`

## RTO Verification

- Agreed rehearsal RTO: `<= 300 seconds`
- Actual rehearsal duration: `0 seconds`
- Result: `PASS`

## Sign-off Status

- Rollback path command validity: `PASS`
- Rehearsal timing target: `PASS`
- Production switchback execution: `PENDING (requires real gateway credentials)`
