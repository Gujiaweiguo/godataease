# SEC-COMP-008 Shadow Validation Runbook (Plan v1)

This runbook executes SEC-COMP-008 using the approved defaults and serves as the operational checklist for Atlas/Hephaestus.

## Defaults (Approved)

- Mirror ratio: `5%` of whitelisted critical routes
- Observation window: `48h` continuous
- Gate threshold: critical-route mismatch rate `< 1%`
- Security gate: zero critical incidents (row/column leakage, unauthorized export download)
- Stability gate: zero Sev-1/Sev-2 on compatibility routes

## Prerequisites

- Staging Go and Java backends are both reachable
- Shadow routing capability exists at gateway/ingress layer
- Metrics and logs are queryable (gateway + application)
- Rollback switch can disable mirror and route back to Java primary

## Required Variables

Set these before running commands:

- `JAVA_BASE_URL`
- `GO_BASE_URL`
- `STAGING_GATEWAY_API`
- `SHADOW_RULE_ID`
- `AUTH_TOKEN` (staging token with read scope)

## Execution Steps

1. Enable 5% mirror on critical-route allowlist
2. Start 48h timer and snapshot baseline metrics
3. Collect hourly mismatch report (`status/code/msg/payload`) for allowlist
4. Run negative security probes every 4h:
   - unauthorized export download
   - row-level access bypass attempts
   - column leakage checks on protected fields
5. At T+48h, produce final mismatch/security summary

## Reference Commands

```bash
# 1) Enable shadow rule (example API shape; adapt to gateway)
curl -sS -X PATCH "$STAGING_GATEWAY_API/rules/$SHADOW_RULE_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"shadow":{"enabled":true,"ratio":0.05}}'

# 2) Sample contract check (single endpoint example)
curl -sS "$JAVA_BASE_URL/api/templateManage/templateList" -H "X-DE-TOKEN: $AUTH_TOKEN" > /tmp/java-template.json
curl -sS "$GO_BASE_URL/api/templateManage/templateList" -H "X-DE-TOKEN: $AUTH_TOKEN" > /tmp/go-template.json

# 3) Disable shadow rule (rollback)
curl -sS -X PATCH "$STAGING_GATEWAY_API/rules/$SHADOW_RULE_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"shadow":{"enabled":false,"ratio":0.0}}'
```

## Acceptance Checklist

- [ ] 48h mirror fully executed
- [ ] Critical-route mismatch rate < 1%
- [ ] Zero critical security incidents
- [ ] No Sev-1/Sev-2 regression on compatibility routes
- [ ] Go/No-Go recommendation documented

## Rollback Procedure

1. Disable shadow traffic immediately
2. Route all traffic to Java stable profile
3. Freeze cutover decision and open incident record
4. Attach mismatch/security evidence and root-cause candidates
