# Release Note: Go Shadow Validation Gate

Date: 2026-02-20
Scope: main branch direct update

## Summary

- Added an executable go-only shadow validation pipeline for migration gating.
- Added compatibility aliases for critical Java-era routes in Go handlers.
- Added observability baseline artifacts (dashboard + alert policy) and automated classification reporting.
- Completed OpenSpec change `add-go-shadow-validation-cutover-gate` with full execution evidence.

## Delivered Components

- Shadow gate scripts under `apps/backend-go/scripts/shadow-gate/`:
  - preflight, bounded window runner (max 4h), gate decision, alert rehearsal, incident visibility check, rollback rehearsal
- Compatibility route handlers:
  - `template_handler.go`
  - `export_handler.go`
  - `compatibility_bridge_handler.go`
- Observability artifacts:
  - `apps/backend-go/testdata/contract-diff/observability/shadow-dashboard.json`
  - `apps/backend-go/testdata/contract-diff/observability/shadow-alert-policy.yaml`

## Evidence and Governance

- 4-hour real-cadence go-only window run completed with PASS.
- SHADOW-004 classification report generated.
- SHADOW-005 decision record and SHADOW-006 rollback signoff package generated.
- OpenSpec strict validation passed.

## Validation Highlights

- `go test ./internal/transport/http/handler`
- `openspec validate add-go-shadow-validation-cutover-gate --strict --no-interactive`
- pass/fail rehearsals for gate decision and alert triggering
- incident-channel delivery verification captured as evidence

## Operational Note

- Current cutover governance is data-driven by threshold and route-level classification outputs.
- Non-blocking watchlist items remain under post-cutover observation.
