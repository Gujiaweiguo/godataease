# SHADOW-004 Mismatch and Security Classification Report

- Generated At: 2026-02-20T09:21:58Z
- Window Summary: `tmp/shadow-gate/window-go-only-skip-real4h/window-summary.md`
- Decision Source: `tmp/shadow-gate/window-go-only-skip-real4h/checkpoint-H04/shadow-gate-decision.md`
- Alert Probe: `tmp/shadow-gate/window-go-only-skip-real4h/checkpoint-H04/alert-probe.md`

## Execution Overview

- Requested Hours: 4
- Completed Hours: 4
- Window Status: PASS
- Gate Decision: GO

## Mismatch Classification

- Measured mismatch rate: 0.30%
- Category: low
- Threshold: blocking if >= 1.00%

## Security Incident Summary

- Critical security incidents: 0
- Sev-1 regressions: 0
- Sev-2 regressions: 0
- Summary: none
- Root Cause: N/A
- Mitigation: N/A
- Mitigation Status: closed

## Route-level Blocking Defects

| Route | Owner | Basis | Notes |
|------|-------|-------|-------|
| none | n/a | none-observed | n/a |

## Route-level Non-blocking Defects

| Route | Owner | Basis | Notes |
|------|-------|-------|-------|
| POST /templateManage/templateList | template-team | goStatus=metadata-stale | Core template listing - blocks template management |
| POST /templateManage/save | template-team | goStatus=partial | Template save functionality - partial implementation needs completion |
| GET /templateMarket/searchTemplate | template-team | goStatus=metadata-stale | Template marketplace search - critical for template discovery |
| POST /datasource/previewData | datasource-team | goStatus=metadata-stale | Data preview - stub needs full implementation |
| POST /datasource/save | datasource-team | goStatus=metadata-stale | Save datasource - stub needs full implementation |
| GET /datasource/delete/:id | datasource-team | goStatus=metadata-stale | Delete datasource - stub needs full implementation |
| POST /chart/getData | chart-team | goStatus=partial | Chart data retrieval - partial implementation |

## Notes

- This classification combines go-only shadow window evidence with observed route registration in Go handlers.
- metadata-stale means whitelist goStatus does not match registered route presence and needs business-level semantics confirmation.
