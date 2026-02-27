# Production Cutover Execution Checklist

## Document Information

| Field | Value |
|-------|-------|
| Change ID | `add-go-shadow-validation-cutover-gate` |
| Execution Date | TBD |
| Execution Window | TBD |
| Commander | TBD |

---

## Pre-Cutover (T-24h)

### Infrastructure

- [ ] Verify Go backend binary built and tested
- [ ] Verify Docker image pushed to registry
- [ ] Verify database backups completed
- [ ] Verify Redis backups completed
- [ ] Verify rollback scripts ready
- [ ] Verify monitoring dashboards accessible

### Communication

- [ ] Notify stakeholders of maintenance window
- [ ] Confirm incident response team availability
- [ ] Confirm on-call engineer assigned
- [ ] Post maintenance notice to users

### Access

- [ ] Verify production gateway credentials available
- [ ] Verify database access credentials
- [ ] Verify Redis access credentials
- [ ] Verify admin panel access

---

## Cutover Day (T-0)

### Pre-Switch (T-30min)

- [ ] Final build verification: `make build && make test`
- [ ] Verify all approvers have signed off
- [ ] Verify monitoring team ready
- [ ] Verify incident channel active
- [ ] Record baseline metrics (error rate, latency, throughput)

### Switch Execution (T-0)

| Step | Action | Owner | Status | Time |
|------|--------|-------|--------|------|
| 1 | Announce cutover start in incident channel | Commander | [ ] | |
| 2 | Stop traffic to Java backend | SRE | [ ] | |
| 3 | Verify Java backend stopped | SRE | [ ] | |
| 4 | Start Go backend | SRE | [ ] | |
| 5 | Verify Go backend healthy | SRE | [ ] | |
| 6 | Route traffic to Go backend | SRE | [ ] | |
| 7 | Verify traffic flowing | SRE | [ ] | |
| 8 | Announce cutover complete | Commander | [ ] | |

### Immediate Verification (T+5min)

- [ ] Error rate < 1%
- [ ] P99 latency within baseline
- [ ] No critical alerts
- [ ] Sample API calls successful
- [ ] User login working
- [ ] Dashboard loading

---

## Post-Cutover Monitoring

### T+15min

- [ ] Error rate stable
- [ ] No new incidents
- [ ] Performance metrics normal

### T+1h

- [ ] All critical paths tested
- [ ] No Sev-1/Sev-2 issues
- [ ] User feedback normal

### T+4h

- [ ] Full system health check
- [ ] Database integrity verified
- [ ] Cache hit rates normal

### T+24h

- [ ] Overnight metrics review
- [ ] No anomalies detected
- [ ] Close incident channel (if no issues)

### T+48h

- [ ] Final health check
- [ ] Update documentation
- [ ] Archive cutover artifacts
- [ ] Close change record

---

## Rollback Procedure

### Rollback Triggers

| Condition | Threshold | Action |
|-----------|-----------|--------|
| Error rate | > 5% for 5 min | Immediate rollback |
| Mismatch rate | >= 1% | Immediate rollback |
| Security incident | Any critical | Immediate rollback |
| Sev-1 regression | Any | Immediate rollback |

### Rollback Steps

| Step | Action | Owner | Status | Time |
|------|--------|-------|--------|------|
| 1 | Announce rollback in incident channel | Commander | [ ] | |
| 2 | Stop traffic to Go backend | SRE | [ ] | |
| 3 | Start Java backend | SRE | [ ] | |
| 4 | Verify Java backend healthy | SRE | [ ] | |
| 5 | Route traffic to Java backend | SRE | [ ] | |
| 6 | Verify traffic flowing | SRE | [ ] | |
| 7 | Announce rollback complete | Commander | [ ] | |
| 8 | Begin incident review | All | [ ] | |

---

## Communication Templates

### Cutover Start

```
🚀 PRODUCTION CUTOVER STARTED

Time: [TIMESTAMP]
Change: Java → Go Backend Migration
Commander: [NAME]
Expected duration: 10-15 minutes

Monitoring: [GRAFANA_URL]
Incident channel: [SLACK_CHANNEL]
```

### Cutover Complete

```
✅ PRODUCTION CUTOVER COMPLETE

Time: [TIMESTAMP]
Status: SUCCESS
Go backend is now serving production traffic

Next check: T+15min
```

### Rollback Initiated

```
🚨 ROLLBACK INITIATED

Time: [TIMESTAMP]
Reason: [REASON]
Commander: [NAME]

Reverting to Java backend...
```

### Rollback Complete

```
⏪ ROLLBACK COMPLETE

Time: [TIMESTAMP]
Status: Java backend restored
Traffic: 100% to Java

Incident review: [TIME]
```

---

## Contacts

| Role | Name | Contact |
|------|------|---------|
| Commander | | |
| SRE Lead | | |
| Engineering Manager | | |
| Release Manager | | |
| Observability Engineer | | |

---

## References

- [Approval Sign-Off](./cutover-approval-signoff.md)
- [Go/No-Go Decision](./go-no-go-decision.md)
- [Rollback Sign-Off Package](./rollback-signoff-package.md)
