# Staging Prerequisite Checklist (SHADOW-001)

## Snapshot

- Timestamp: 2026-02-20T01:47:17Z
- Source report: `apps/backend-go/tmp/shadow-gate/preflight/preflight.json`
- Overall status: `blocked`

## Revalidation Snapshot (Go-only rehearsal)

- Timestamp: 2026-02-20T03:15:46Z
- Source report: `apps/backend-go/tmp/shadow-gate/preflight-go-only/preflight.json`
- Mode: `go-only`
- Overall status: `ready`
- Notes: readiness verified with local rehearsal env file `apps/backend-go/tmp/shadow-gate/.env.shadow-gate.test`; production secrets/endpoints still require real values from staging owners.

## Revalidation Snapshot (Go-only skip-control)

- Timestamp: 2026-02-20T05:14:38Z
- Source report: `apps/backend-go/tmp/shadow-gate/preflight-go-only-skip/preflight.json`
- Mode: `go-only --skip-shadow-control`
- Overall status: `ready`
- Notes: control-plane and endpoint env checks are intentionally marked `n/a` for rehearsal when real staging secrets are unavailable.

## Readiness Matrix

| Item | Status | Owner | ETA | Mitigation |
|------|--------|-------|-----|------------|
| `jq` | Ready | Platform SRE Lead | 2026-02-24 | N/A |
| `curl` | Ready | Platform SRE Lead | 2026-02-24 | N/A |
| `kubectl` | Blocked | Platform SRE Lead | 2026-02-24 | Provision kubectl in staging runner image |
| `helm` | Blocked | Platform SRE Lead | 2026-02-24 | Install Helm v3 and pin chart repo access |
| `JAVA_BASE_URL` | Blocked | Platform SRE Lead | 2026-02-24 | Publish Java staging endpoint in CI secret store |
| `GO_BASE_URL` | Blocked | Platform SRE Lead | 2026-02-24 | Publish Go candidate endpoint in CI secret store |
| `STAGING_GATEWAY_API` | Blocked | Gateway Operations Lead | 2026-03-02 | Provide gateway control-plane endpoint and access policy |
| `SHADOW_RULE_ID` | Blocked | Gateway Operations Lead | 2026-03-02 | Create dedicated shadow rule and share immutable rule ID |
| `AUTH_TOKEN` | Blocked | Engineering Manager | 2026-03-01 | Issue short-lived service token with least privilege |
| `critical-whitelist.yaml` | Ready | API Compatibility Owner | 2026-02-28 | N/A |

## Gap List

1. Missing cluster tooling (`kubectl`, `helm`) in execution environment.
2. Missing staging service endpoint variables for Java/Go targets.
3. Missing gateway control-plane credentials (`STAGING_GATEWAY_API`, `SHADOW_RULE_ID`, `AUTH_TOKEN`).

## Bootstrap Helpers

- Tooling bootstrap: `apps/backend-go/scripts/shadow-gate/bootstrap_tooling.sh`
- Env loader: `apps/backend-go/scripts/shadow-gate/load_shadow_env.sh`
- Env template: `apps/backend-go/.env.shadow-gate.example`
- Go-only template: `apps/backend-go/.env.shadow-gate.go-only.example`
- Go-only preflight mode for post-cutover operations: `--migration-mode go-only`
- Rehearsal mode without control-plane secrets: `--migration-mode go-only --skip-shadow-control`

## Sign-off

- Prepared by: Atlas/Hephaestus automation
- Reviewer: Pending assignment
- Approval status: Pending
