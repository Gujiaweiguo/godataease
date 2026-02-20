# SEC-COMP-008 Evidence Template (Plan v1)

Use this template during the 48h shadow run. Atlas/Hephaestus should fill all sections before marking SEC-COMP-008 complete.

## 0) Run Metadata

- Start Time (UTC):
- End Time (UTC):
- Mirror Ratio:
- Route Allowlist Version:
- Java Build/Tag:
- Go Build/Tag:
- Operator:

## 1) Hourly Contract Mismatch Summary

| Hour Window | Requests Compared | Mismatch Count | Mismatch Rate | Top 3 Routes | Notes |
|------------|-------------------|----------------|---------------|--------------|-------|
| H+01 |  |  |  |  |  |
| H+02 |  |  |  |  |  |
| ... |  |  |  |  |  |
| H+48 |  |  |  |  |  |

## 2) Security Probe Results (Every 4h)

| Window | Unauthorized Export Download | Row Bypass Probe | Column Leakage Probe | Verdict | Evidence Link |
|--------|-------------------------------|------------------|----------------------|---------|---------------|
| H+04 | PASS/FAIL | PASS/FAIL | PASS/FAIL |  |  |
| H+08 | PASS/FAIL | PASS/FAIL | PASS/FAIL |  |  |
| ... | ... | ... | ... | ... | ... |
| H+48 | PASS/FAIL | PASS/FAIL | PASS/FAIL |  |  |

## 3) Incident Summary

- Sev-1 count:
- Sev-2 count:
- Critical security incidents count:
- Incident IDs:

## 4) Final Gate Decision

- Critical-route mismatch rate (<1%): PASS/FAIL
- Zero critical security incidents: PASS/FAIL
- Zero Sev-1/Sev-2 regressions: PASS/FAIL
- Go/No-Go: GO / NO-GO
- Decision owner:
- Decision timestamp:

## 5) Rollback Record (if triggered)

- Rollback triggered: YES/NO
- Trigger timestamp:
- Trigger reason:
- Switchback completion timestamp:
- Post-rollback health verdict:
