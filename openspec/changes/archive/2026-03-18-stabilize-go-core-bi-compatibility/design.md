## Context

The repository has already migrated the primary runtime backend from Java to Go, while preserving Java-era route contracts through compatibility handlers and frontend compatibility branches. The unstable area is no longer broad platform bring-up; it is the narrow but critical BI execution chain that starts at datasource access, continues through dataset discovery and preview, and ends at dashboard and big-screen rendering plus resource operations.

This change treats those flows as a single release-readiness surface. The main constraints are: existing frontend clients still depend on compatibility endpoints and Java-shaped envelopes, permission behavior has to remain semantically correct during migration, and current governance gates exist but do not yet define a clear required scope for the four core BI flow families.

## Goals / Non-Goals

**Goals:**
- Define a governed stabilization boundary for datasource, dataset, dashboard, and big-screen flows.
- Make compatibility behavior explicit for canonical routes, alias routes, frontend compatibility endpoints, and resource-tree payloads.
- Prevent migration-time ambiguity between unauthenticated, unauthorized, missing, and not-yet-implemented behavior.
- Establish a test-first execution order so future implementation work starts from failing parity and regression cases instead of ad hoc fixes.
- Make release gating and regression evidence part of the change scope for these BI-critical flows.

**Non-Goals:**
- Introducing new BI capabilities or redesigning charting behavior.
- Rewriting the frontend compatibility layer wholesale.
- Changing legacy Java code or defining full parity for every historical endpoint.
- Expanding this change into unrelated platform modules such as login, audit, or organization management.

## Decisions

### Decision: Treat the core BI path as one stabilization unit
The change will group datasource, dataset, dashboard, and big-screen workflows into a single critical path rather than stabilizing each page in isolation.

- **Why:** most user-visible failures are cross-step failures, such as a datasource being visible but unusable in dataset flows, or a dashboard tree loading while downstream operations fail on malformed nodes.
- **Alternative considered:** split the work into independent module hardening changes. Rejected because it would hide compatibility and permission failures that only appear across module boundaries.

### Decision: Use contract-first hardening before behavioral expansion
Implementation work driven by this change must first freeze expected request/response envelopes, route reachability, and error semantics for the critical BI APIs.

- **Why:** current instability is primarily migration drift, not missing product direction. Contract-first hardening reduces false positives caused by placeholder success or mismatched envelopes.
- **Alternative considered:** continue direct bugfixing from pages upward. Rejected because page-level fixes can accidentally normalize incorrect backend behavior and make parity drift harder to detect.

### Decision: Keep compatibility aliases and canonical routes jointly governed
Core BI endpoints will be considered healthy only when both canonical Go paths and migration compatibility aliases are routable and semantically aligned for in-scope flows.

- **Why:** frontend code still consumes compatibility routes in places, so canonical correctness alone is insufficient.
- **Alternative considered:** declare canonical routes authoritative and allow compatibility gaps temporarily. Rejected because it would not match the current deployment reality.

### Decision: Treat permission semantics as a first-class stability concern
The change will explicitly govern `401`, `403`, `404`, and deterministic non-success behavior for partially migrated or enterprise-only endpoints that touch core BI flows.

- **Why:** migration instability often appears as misclassified permission failures, misleading missing-route errors, or placeholder success that hides incomplete logic.
- **Alternative considered:** keep permission semantics inside module-specific fixes only. Rejected because the same ambiguity spans datasource, dataset, dashboard, and compatibility handlers.

### Decision: Use additive spec deltas for stabilization requirements
This change will add new stabilization requirements to existing capabilities instead of rewriting the existing baseline requirements.

- **Why:** the current specs already cover the basic capability surface. Additive deltas let the change define release-readiness expectations with lower risk of losing prior baseline detail.
- **Alternative considered:** modify existing requirements in place. Rejected because it would create larger deltas and make archive-time reasoning harder.

## Risks / Trade-offs

- **[Risk] Stabilization scope grows into full migration parity program** → **Mitigation:** constrain the change to four flow families and the exact capability set named in the proposal.
- **[Risk] Compatibility checks still miss frontend-consumed edge cases** → **Mitigation:** require explicit resource-tree and frontend compatibility endpoint coverage in specs and tasks.
- **[Risk] Permission fixes mask route gaps or vice versa** → **Mitigation:** separate route reachability, authorization semantics, and placeholder-success prohibition in both specs and tasks.
- **[Risk] Release gate effort slows feature work** → **Mitigation:** limit required gates to BI-critical routes and reuse existing compat/shadow infrastructure where possible.
- **[Risk] Enterprise or xpack endpoints remain partially implemented** → **Mitigation:** require explicit deterministic non-success semantics instead of fake success for in-scope but unsupported paths.

## Migration Plan

1. Freeze the critical BI endpoint matrix and identify which routes are canonical, alias, compatibility-only, or intentionally unavailable.
2. Add failing regression coverage for datasource, dataset, dashboard, big-screen, and related permission flows.
3. Repair backend routing, handler, service, and permission behavior until the governed matrix passes.
4. Align frontend compatibility handling with the stabilized backend contracts and remove any placeholder-success assumptions for in-scope flows.
5. Run minimum verification gates for backend, frontend, compatibility drift, and targeted regression evidence before considering the change implementation complete.

Rollback for future implementation remains straightforward because this change does not prescribe data-model migration; the primary rollback path is to revert behavior changes and restore the prior route/handler set while preserving evidence artifacts.

## Open Questions

- Whether any chart-specific read paths should be added as dependent regression cases once dashboard and big-screen resource loading is stabilized.
- Whether the current required-gate whitelist already contains all BI-critical routes or needs a dedicated expansion in the implementation phase.
- Whether some frontend compatibility branches can be retired immediately after stabilization, or should be left for a follow-up convergence change.
