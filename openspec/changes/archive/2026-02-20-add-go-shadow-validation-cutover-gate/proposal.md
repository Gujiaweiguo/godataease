# Change: Go staging shadow validation and cutover gate

## Why
The archived hardening change `add-go-security-compatibility-readiness` left one operational item incomplete (`SEC-COMP-008`) due to missing staging prerequisites. To avoid mutating archived history while still completing migration risk closure, this follow-up change defines a dedicated execution path for shadow validation and cutover governance.

## What Changes
- Define staging readiness prerequisites (tooling, environment, gateway access) with explicit ownership and verification evidence.
- Execute and govern a bounded (up to 4-hour) shadow validation window for critical compatibility routes.
- Produce a structured mismatch and security incident report with Go/No-Go decision criteria.
- Add rollback trigger policy and route switchback procedure as mandatory cutover controls.
- Add OpenSpec delta for pre-cutover shadow gate requirements.

## Impact
- Affected specs:
  - `specs/api-compatibility-bridge/spec.md`
- Affected artifacts:
  - `apps/backend-go/scripts/shadow-gate/preflight_check.sh`
  - `apps/backend-go/scripts/shadow-gate/manage_shadow_rule.sh`
  - `apps/backend-go/scripts/shadow-gate/evaluate_shadow_gate.sh`
  - `apps/backend-go/scripts/shadow-gate/run_shadow_gate.sh`
  - `apps/backend-go/scripts/shadow-gate/run_shadow_window.sh`
  - `apps/backend-go/scripts/shadow-gate/rollback_rehearsal.sh`
  - `apps/backend-go/scripts/shadow-gate/generate_observability_baseline.py`
  - `apps/backend-go/scripts/shadow-gate/test_alert_policy.sh`
  - `apps/backend-go/scripts/shadow-gate/generate_shadow_classification_report.py`
  - `apps/backend-go/scripts/shadow-gate/verify_incident_channel.sh`
  - `apps/backend-go/testdata/contract-diff/observability/shadow-dashboard.json`
  - `apps/backend-go/testdata/contract-diff/observability/shadow-alert-policy.yaml`
  - `apps/backend-go/testdata/contract-diff/observability/README.md`
  - `apps/backend-go/.env.shadow-gate.go-only.example`
  - `openspec/changes/add-go-shadow-validation-cutover-gate/tasks.md`
  - `openspec/changes/add-go-shadow-validation-cutover-gate/design.md`
  - `openspec/changes/add-go-shadow-validation-cutover-gate/shadow-validation-runbook.md`
  - `openspec/changes/add-go-shadow-validation-cutover-gate/shadow-validation-report.md`
  - `openspec/changes/add-go-shadow-validation-cutover-gate/shadow-validation-evidence-template.md`
  - `openspec/changes/add-go-shadow-validation-cutover-gate/staging-prerequisite-checklist.md`
  - `openspec/changes/add-go-shadow-validation-cutover-gate/shadow-mismatch-security-classification.md`
  - `openspec/changes/add-go-shadow-validation-cutover-gate/go-no-go-decision.md`
  - `openspec/changes/add-go-shadow-validation-cutover-gate/rollback-signoff-package.md`
- Affected runtime systems (expected during implementation):
  - Staging traffic gateway and route switching controls
  - Contract diff/reporting jobs for critical interface set
  - Observability dashboards and alert channels for compatibility/security events
