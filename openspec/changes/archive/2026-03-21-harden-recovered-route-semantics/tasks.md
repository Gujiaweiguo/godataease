## 1. Residual hardening baseline

- [x] 1.1 Build a hardening matrix from the archived recovery evidence for datasource, dashboard, big-screen, export-center, and audit residual gaps.
- [x] 1.2 Freeze the frontend caller and backend owner for each residual hardening path.
- [x] 1.3 Freeze the verification surface and currently-missing hardening target for each path.

## 2. Datasource hardening lane

- [x] 2.1 Add direct forbidden regression coverage for datasource list aliases.
- [x] 2.2 Hand off datasource runtime alias-semantics redesign to `design-datasource-list-resource-identity` because the live list routes lack a stable resource identifier and require a broader design decision.

## 2b. Hardening semantic freeze

- [x] 2.3 Freeze residual hardening semantics for unauthenticated, forbidden, missing route/resource, and explicit non-success outcomes across the current change scope.

## 3. Visualization hardening lane

- [x] 3.1 Add handler- or route-level missing-resource coverage for dashboard detail paths.
- [x] 3.2 Add deeper big-screen detail/edit semantic coverage beyond current preview/discovery smoke.
- [x] 3.3 Repair visualization boundary behavior only where the new hardening coverage reveals a concrete mismatch.

## 4. Operational route-smoke hardening lane

- [x] 4.1 Add route-level smoke or boundary verification for export-center download-path semantics.
- [x] 4.2 Add route-level smoke or boundary verification for audit page-entry/detail flows beyond current unit/handler coverage.
- [x] 4.3 Repair operational route-level boundary behavior only where the new hardening coverage reveals a concrete mismatch.

## 5. Hardening closeout

- [x] 5.1 Record hardening evidence for all completed slices.
- [x] 5.2 Run focused frontend/backend verification for the touched hardening slices.
- [x] 5.3 Confirm no new broad regressions were introduced while tightening the archived recovery work.
