## Context

Archived recovery work already established two facts:

1. Datasource list aliases have direct unauthenticated proof and hardening-level forbidden proof at the middleware boundary.
2. Live datasource list routes cannot simply adopt `CheckDatasourceView()` because there is no stable `resource_id` in the current list request contract.

This means the remaining work is a design problem, not a narrow bugfix.

## Goals / Non-Goals

**Goals**
- Decide what runtime semantics datasource list routes should have for authenticated-but-forbidden callers.
- Ensure the selected design is compatible with existing datasource callers and recovery guarantees.
- Produce implementation tasks that can be executed later without reopening the architectural question.

**Non-Goals**
- Implementing the selected runtime change in this change.
- Redesigning unrelated datasource detail or validation semantics.
- Expanding into dataset or broader permission-model refactors.

## Design Options

### Option A — Filtered list semantics
Treat datasource list as a collection query and filter the result set to only the resources the caller may view.

**Pros**
- Aligns naturally with collection endpoints
- Avoids introducing new required request fields
- Keeps compatibility aliases stable

**Cons**
- Forbidden becomes implicit (empty or reduced list) rather than explicit
- Harder to distinguish “no accessible datasources” from “query has no results”

### Option B — Explicit scoped list semantics
Require a stable scope identifier in the list request (for example datasource group, workspace, or another governing scope) and apply explicit forbidden semantics at that scope boundary.

**Pros**
- Preserves explicit forbidden outcomes
- Keeps runtime semantics closer to existing detail-path permission checks

**Cons**
- Changes request semantics
- Requires identifying a real stable scope already meaningful to callers

### Option C — Keep list broad, move explicit forbidden to detail-only paths
Accept that list endpoints remain auth-only and enforce explicit forbidden only for detail/read operations.

**Pros**
- Minimal implementation impact
- Fits the current route shape

**Cons**
- Does not satisfy the blocked hardening expectation for list-path forbidden runtime semantics
- Leaves a semantic asymmetry between list and detail paths

## Option Comparison Against Discovery

### Option A — Filtered list semantics

**Fit with current discovery:** weak

Why it does not fit well today:
- current backend list semantics do not perform permission-based filtering
- current frontend callers expect broad collection loading rather than “only what I can see” semantics
- silently changing list semantics to filtered behavior would alter caller expectations without an explicit request-contract change

### Option B — Explicit scoped list semantics

**Fit with current discovery:** not viable today

Why it is not viable today:
- no stable governing scope exists in the current runtime request shape
- frontend wrappers centralize datasource list usage and only inject `busiFlag`, optional `weight`, and occasional UI helper params like `leaf`/`id`
- none of those fields currently behave like a durable permission boundary comparable to detail-path `resourceId` semantics

### Option C — Keep list broad, move explicit forbidden to detail-only paths

**Fit with current discovery:** strongest

Why it fits best today:
- current runtime behavior is already auth-only list behavior
- current frontend callers already assume broad collection loading
- compatibility aliases remain stable without introducing new request semantics
- explicit forbidden semantics stay attached to paths that already carry stable resource identity

## Selected Direction

Select **Option C**.

Datasource list runtime semantics should be documented and preserved as **auth-only collection-load behavior**, while explicit forbidden semantics remain a requirement for datasource detail/read paths that already carry stable resource identity.

This keeps the design aligned with both current runtime behavior and current caller expectations, and avoids introducing hidden semantic drift or a premature request-contract redesign.

## Rejected Options

### Why Option A is rejected for now
- it would introduce permission-based list filtering that does not exist today
- it would silently redefine what callers mean by a datasource list query
- it would require additional user/resource filtering behavior in service or repository code without a caller-driven contract change

### Why Option B is rejected for now
- it depends on a stable governing scope that discovery did not find
- adding such a scope would itself be a contract redesign rather than a narrow runtime clarification
- compatibility callers would need coordinated changes before the route semantics could be considered safe

## Recommended Direction

Given the current discovery evidence, **Option C** is the recommended and selected direction.

The design implication is straightforward:
- list aliases remain auth-only collection-load paths
- forbidden semantics remain a live runtime guarantee only for detail/read paths with stable resource identity
- middleware-only `403` proof for list aliases remains useful as a component-level boundary test, but not as a statement about live list runtime behavior

## Key Questions To Resolve

1. Do current datasource list callers already carry a stable scope that can be treated as a permission boundary?
2. If not, is changing datasource list requests acceptable for existing compatibility callers?
3. If explicit forbidden remains impossible at list runtime, should the spec formally declare filtered/auth-only list semantics instead?

## Validation Plan

The future implementation change should not begin until this change answers the three questions above and selects a single direction.

## Minimal Implementation Planning

### Backend changes

- keep current datasource list aliases (`/api/ds/list`, `/api/datasource/list`, `/de2api/datasource/list`) as JWT-auth-only routes
- do **not** attach `CheckDatasourceView()` to live list routes unless a future change first introduces a stable governing scope
- ensure datasource detail/read paths remain the place where explicit forbidden semantics are guaranteed at runtime

### Frontend changes

- keep current wrapper usage unchanged (`listDatasources`, `getDsTree`, `getDatasourceList`)
- do not introduce new scope parameters into datasource list callers as part of this design choice
- align caller and test expectations so datasource list remains broad/auth-only rather than resource-addressed

### Regression coverage needed

- preserve auth-only runtime proof for datasource list aliases (unauthenticated fails, authenticated list still behaves as collection load)
- preserve explicit forbidden runtime proof for datasource detail/read paths
- keep middleware-level forbidden proof for datasource list aliases only as a component/boundary proof, not as evidence of live list runtime semantics
