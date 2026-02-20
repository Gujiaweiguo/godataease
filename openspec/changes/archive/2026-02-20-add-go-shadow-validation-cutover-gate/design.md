## Context

The repository currently has no active OpenSpec changes. A prior archived change completed technical parity work but left operational shadow validation blocked by staging prerequisites. This design isolates the remaining operational gate into a dedicated, auditable change without rewriting archived records.

## Goals / Non-Goals

- Goals:
  - Establish a deterministic pre-cutover shadow gate for critical Java/Go compatibility routes.
  - Define objective Go/No-Go thresholds and rollback triggers.
  - Ensure evidence traceability for audit and release approval.
- Non-Goals:
  - Re-implement route compatibility logic already delivered in archived changes.
  - Expand required-gate scope beyond the existing critical interface whitelist.

## Decisions

- Decision: Use `api-compatibility-bridge` as the capability target for this operational gate.
  - Why: The gate evaluates contract parity and security semantics for compatibility routes.
- Decision: Keep tasks as operation-first with explicit metadata (input/output/acceptance/rollback).
  - Why: Cutover risk is primarily operational and requires command-level verification.
- Decision: Keep archived change immutable and create follow-up change instead.
  - Why: Preserves historical integrity while enabling completion tracking.

## Risks / Trade-offs

- Risk: Staging prerequisites can remain blocked and delay cutover.
  - Mitigation: SHADOW-001 introduces owner-based gap closure and explicit ETA tracking.
- Risk: Shadow evidence quality may be insufficient for decision.
  - Mitigation: SHADOW-002 requires dashboard/alert baseline before shadow run.
- Risk: Pressure to bypass gate due timeline.
  - Mitigation: SHADOW-005 defines mandatory threshold policy and approval record.

## Migration Plan

1. Approve this change and assign task owners.
2. Execute SHADOW-001..SHADOW-006 in dependency order.
3. Archive this change only after all tasks are marked `[x]` with report evidence.

## Open Questions

- Which team owns gateway switchback execution during incident-triggered rollback?
- What is the exact on-call escalation chain for Sev-1 compatibility regressions in staging?
