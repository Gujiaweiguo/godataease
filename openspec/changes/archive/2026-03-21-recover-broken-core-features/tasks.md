## 1. P0 evidence baseline and scope freeze

- [x] 1.1 Build a feature recovery matrix for datasource, dataset, dashboard, and big-screen governed flows.
- [x] 1.2 For each P0 candidate issue, classify the current symptom as route/access regression, API mismatch, page-init failure, state-sync failure, or real implementation gap.
- [x] 1.3 Freeze the frontend caller or entry path for each governed P0 flow.
- [x] 1.4 Freeze the backend handler, service, or repository owner for each governed P0 flow.
- [x] 1.5 Freeze the current verification surface for each governed P0 flow, including existing unit, integration, and smoke coverage or explicit lack of coverage.
- [x] 1.6 Publish the P0 cut line and explicitly hold export-center and audit flows in P1.
- [x] 1.7 Mark every in-scope issue as P0, P1, or deferred based on user-path impact and bounded verification cost.

## 2. P0 failing verification targets and semantic freeze

- [x] 2.1 Define failing or missing verification for datasource entry reachability, initialization, and critical browse workflows.
- [x] 2.2 Define failing or missing verification for dataset entry reachability, initialization, browse, field, and preview workflows.
- [x] 2.3 Define failing or missing verification for dashboard and big-screen entry-chain, list/tree/detail, and discovery workflows.
- [x] 2.4 Freeze governed failure semantics for unauthenticated, forbidden, missing route/resource, and explicit non-success outcomes for unsupported, dependency, business, and bounded real-gap failures across the P0 BI flow families.
- [x] 2.5 Record the expected usable initialized state for each governed P0 flow before repair begins.

## 3. Datasource P0 recovery lane

- [x] 3.1 Add failing regression coverage for governed datasource entry and initialization flows.
- [x] 3.2 Add failing regression coverage for governed datasource browse or list workflows.
- [x] 3.3 Repair datasource route, contract, or page-init gaps required by the governed P0 flows.
- [x] 3.4 Verify datasource failures surface as explicit non-success outcomes instead of silent empty success.
- [x] 3.5 Add targeted regression or smoke coverage proving recovered datasource P0 flows remain stable.

## 4. Dataset P0 recovery lane

- [x] 4.1 Add failing regression coverage for governed dataset entry and initialization flows.
- [x] 4.2 Add failing regression coverage for governed dataset browse, field, and preview workflows.
- [x] 4.3 Repair dataset route, contract, page-init, or preview gaps required by the governed P0 flows.
- [x] 4.4 Verify dataset failure semantics remain distinguishable across authorization, dependency, missing route/resource, and business execution failure.
- [x] 4.5 Add targeted regression or smoke coverage proving recovered dataset P0 flows remain stable.

## 5. Visualization P0 recovery lane

- [x] 5.1 Add failing regression coverage for governed dashboard entry-chain and discovery workflows.
- [x] 5.2 Add failing regression coverage for governed big-screen entry-chain and discovery workflows.
- [x] 5.3 Repair visualization route, detail, discovery, payload-shape, or page-init gaps required by the governed P0 flows.
- [x] 5.4 Verify dashboard and big-screen payloads remain consumable by the frontend path that triggered the flow.
- [x] 5.5 Add targeted regression or smoke coverage proving recovered visualization P0 flows remain stable.

## 6. P1 backlog remains explicit

- [x] 6.1 Keep export-center query, retry, and download recovery work tracked as P1 and outside the first execution batch.
- [x] 6.2 Keep audit page reachability, filter-query, and detail-read recovery work tracked as P1 and outside the first execution batch.
- [x] 6.3 Record any newly discovered export-center or audit issues without pulling them into P0 unless the cut line is revised explicitly.

## 7. P0 regression evidence and batch closeout

- [x] 7.1 Capture regression evidence for datasource, dataset, dashboard, and big-screen recovery.
- [x] 7.2 Run frontend lint, typecheck, and affected tests for the recovered P0 scope.
- [x] 7.3 Run backend tests required for the recovered P0 scope.
- [x] 7.4 Confirm recovered P0 flows preserve distinguishable failure semantics instead of degrading into placeholder success or misleading blank-success states.
- [x] 7.5 Record any remaining real implementation gaps separately from recovered access-path, compatibility, and initialization regressions.
- [x] 7.6 Capture release-readiness evidence for the completed P0 batch before widening scope.
- [x] 7.7 Confirm export-center and audit still remain explicitly tracked as P1 before starting the next batch.
