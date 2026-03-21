# P0 Cut Line

This document freezes the first execution batch for `recover-broken-core-features`.

It exists to prevent the change from expanding back into a broad “fix all broken features” effort before the core BI paths are stabilized.

## Scope rule

P0 is limited to the governed BI critical-path families only:

- Datasource
- Dataset
- Dashboard
- Big-screen

P1 remains explicitly outside this batch:

- Export-center
- Audit

Deferred items remain outside both P0 implementation and P1 commitment until they are separately classified and costed.

## In P0

The following flow families are in P0 only when they are tied to governed reachability, initialization, contract, or state-consumption recovery.

### Datasource
Included:
- datasource management entry reachability
- datasource page initialization
- governed critical browse/list workflows
- explicit non-success behavior for governed datasource failures

Not included:
- new datasource capabilities
- broad datasource UX redesign
- low-value cosmetic issues without critical-path impact

### Dataset
Included:
- dataset management entry reachability
- dataset page initialization
- governed browse workflows
- governed field metadata workflows
- governed preview workflows
- deterministic failure semantics for authorization, dependency, route/resource, and business failure

Not included:
- new dataset authoring features
- broad modeling enhancements
- low-priority presentation defects

### Dashboard
Included:
- dashboard entry-chain reachability
- governed list/tree/detail/discovery flows required to reach usable page state
- payload-consumability recovery for the triggering frontend path
- explicit non-success behavior where governed dashboard flows fail

Not included:
- new dashboard editing features
- wide visualization refactors unrelated to governed recovery
- low-value visual polish tasks

### Big-screen
Included:
- big-screen entry-chain reachability
- governed tree/detail/discovery flows required to reach usable page state
- payload-consumability recovery for the triggering frontend path
- explicit non-success behavior where governed big-screen flows fail

Not included:
- new big-screen capabilities
- design or layout improvements outside recovery scope
- non-critical visual defects

## In P1

The following flow families are intentionally excluded from P0 and must not be pulled forward unless the cut line is revised explicitly.

### Export-center
Held in P1:
- export task query
- retry flows
- download flows
- export-specific route, contract, and page-init recovery

### Audit
Held in P1:
- audit page reachability
- filter query flows
- detail-read flows
- audit-specific route, contract, and page-init recovery

## Deferred

The following items are deferred even if they are discovered during P0 investigation:

- issues without a finished classification
- issues without bounded verification cost
- standalone real implementation gaps that do not block governed P0 usability
- low-value UI glitches
- broad compatibility rewrites across historical paths
- new product behavior
- changes that reopen system-management, RBAC, or menu information-architecture scope

## Entry criteria for P0 work

A flow may enter P0 execution only when all of the following are true:

1. the flow is represented in `feature-recovery-matrix.md`
2. the row is classified as:
   - route/access regression, or
   - API mismatch, or
   - page-init failure, or
   - state-sync failure, or
   - real implementation gap that directly blocks a governed BI critical path and remains bounded enough to verify inside P0
3. the row is explicitly labeled `P0`
4. the frontend caller / entry path is frozen
5. the backend owner is frozen
6. the current verification surface is frozen
7. a failing or missing verification target has been named
8. the expected usable state is explicit
9. the expected failure semantics are explicit

## Exit criteria for P0 completion

P0 is complete only when all of the following are true:

1. datasource governed flows reach usable initialized state
2. dataset governed flows reach usable initialized state
3. dashboard governed flows reach usable initialized state
4. big-screen governed flows reach usable initialized state
5. each recovered P0 flow has targeted regression or smoke evidence
6. failure semantics remain distinguishable for:
   - unauthenticated
   - forbidden
   - missing route/resource
   - explicit non-success unsupported, business, dependency, or real-gap failure
7. no remaining issue is silently left in P0 without classification
8. export-center and audit remain explicitly tracked as P1, not silently absorbed into the batch

## Scope change rule

P0 must not be widened during implementation unless all of the following happen explicitly:

1. the new item is added to `feature-recovery-matrix.md`
2. its classification is frozen
3. its verification cost is judged bounded
4. its addition does not block closure of the existing P0 BI flows
5. the cut-line document is updated intentionally rather than by implication

If any of these conditions are not met, the item stays in P1 or deferred.

## Current interpretation

This cut line reflects the intended first recovery batch described by the active change:

- stabilize datasource, dataset, dashboard, and big-screen first
- keep export-center and audit outside the first implementation wave
- require evidence and verification targets before repair
- prevent broad stabilization work from turning into an unbounded cleanup effort
