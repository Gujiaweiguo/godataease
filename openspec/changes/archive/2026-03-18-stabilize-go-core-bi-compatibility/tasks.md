## 1. Critical-flow baseline

- [x] 1.1 Inventory the canonical and compatibility endpoints for datasource, dataset, dashboard, and big-screen flows.
- [x] 1.2 Map each endpoint to its owning backend handler/service and frontend caller.
- [x] 1.3 Mark each in-scope endpoint as full, partial, stub, or missing based on current runtime behavior.
- [x] 1.4 Freeze the expected `code/data/msg` envelope and error semantics for every governed BI route family.
- [x] 1.5 Identify which BI routes must become required-gate coverage before release.

## 2. Datasource route and contract hardening

- [x] 2.1 Add failing backend regression coverage for canonical datasource list queries.
- [x] 2.2 Add failing backend regression coverage for governed datasource compatibility aliases.
- [x] 2.3 Add failing tests for datasource validation timeout and invalid-parameter behavior.
- [x] 2.4 Fix datasource routing gaps so governed list and validation routes no longer return `404`.
- [x] 2.5 Align datasource success and failure envelopes with Java-compatible semantics.
- [x] 2.6 Add or update permission-negative tests for unauthorized datasource access.

## 3. Dataset route and contract hardening

- [x] 3.1 Add failing backend regression coverage for dataset tree payload shape.
- [x] 3.2 Add failing backend regression coverage for dataset field metadata parity.
- [x] 3.3 Add failing backend regression coverage for dataset preview timeout and execution failure semantics.
- [x] 3.4 Fix dataset handlers or services so tree, fields, and preview routes return deterministic non-success behavior on failure.
- [x] 3.5 Verify dataset responses do not silently drop required metadata while still reporting success.
- [x] 3.6 Add or update tests for dataset operations blocked by unauthorized datasource dependencies.

## 4. Dashboard and big-screen stabilization

- [x] 4.1 Inventory the dashboard and big-screen detail, tree, and resource-operation preparation routes used by the frontend.
- [x] 4.2 Add failing regression coverage for visualization resource-tree contract shape.
- [x] 4.3 Add failing regression coverage for dashboard and big-screen detail payload completeness.
- [x] 4.4 Fix tree handlers so in-scope resource-tree routes are reachable through canonical and governed compatibility paths.
- [x] 4.5 Fix malformed or incomplete tree node payloads required for copy, move, delete, and selection preparation.
- [x] 4.6 Verify visualization endpoints do not return placeholder success when business data is missing.

## 5. Permission semantic stabilization

- [x] 5.1 Build a backend regression matrix for `401`, `403`, `404`, and deterministic unavailable behavior across the four BI flow families.
- [x] 5.2 Add failing tests for authenticated-but-unauthorized access to datasource, dataset, dashboard, and big-screen routes.
- [x] 5.3 Fix middleware and handler mappings so permission denial is no longer misclassified as generic `404`.
- [x] 5.4 Add failing tests for permission-filtered resource trees to ensure required node fields remain intact.
- [x] 5.5 Fix permission-filtered tree generation so filtered payloads remain structurally valid for frontend consumers.
- [x] 5.6 Verify role/menu/resource compatibility APIs used by BI administration flows return explicit non-success instead of placeholder success.

## 6. Frontend compatibility convergence

- [x] 6.1 Inventory frontend BI API wrappers and pages that still call compatibility endpoints for datasource, dataset, dashboard, and big-screen flows.
- [x] 6.2 Identify frontend branches that assume Java-only payload details or placeholder success behavior.
- [x] 6.3 Update frontend API handling so stabilized backend non-success semantics are surfaced correctly.
- [x] 6.4 Update or add unit tests for frontend compatibility adapters that consume BI resource trees and detail payloads.
- [x] 6.5 Add targeted smoke coverage for the four core BI flows using the stabilized frontend/backend contracts.

## 7. Compatibility gate and evidence hardening

- [x] 7.1 Update the governed compatibility matrix or whitelist metadata for every in-scope BI endpoint family.
- [x] 7.2 Add strict-compat or contract-diff coverage for newly governed BI routes that were previously outside the required gate scope.
- [x] 7.3 Ensure runtime route status and governed metadata stay aligned for full, partial, stub, and missing states.
- [x] 7.4 Capture regression evidence for datasource, dataset, dashboard, and big-screen route families.
- [x] 7.5 Record any remaining waivers with owner, reason, and expiry instead of leaving implicit migration exceptions.

## 8. Final verification and release readiness

- [x] 8.1 Run backend minimum verification for the stabilized BI scope (`make test`, targeted integration coverage, and `make drift-check` where applicable).
- [x] 8.2 Run frontend minimum verification for the stabilized BI scope (`npm run lint`, `npm run ts:check`, and targeted Vitest coverage).
- [x] 8.3 Re-run strict compatibility checks or equivalent release gates for the governed BI endpoint set.
- [x] 8.4 Execute manual smoke verification for datasource, dataset, dashboard, and big-screen user flows.
- [x] 8.5 Document any remaining gaps, blocked items, or follow-up changes before marking the change implementation-ready.
