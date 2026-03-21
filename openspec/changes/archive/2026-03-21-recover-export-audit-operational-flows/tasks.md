## 1. Operational evidence baseline and scope freeze

- [x] 1.1 Build an operational recovery matrix for export-center and audit governed flows.
- [x] 1.2 For each in-scope issue, classify the current symptom as route/access regression, API mismatch, page-init failure, state-sync failure, or real implementation gap.
- [x] 1.3 Freeze the frontend caller or entry path for each governed export-center and audit flow.
- [x] 1.4 Freeze the backend handler, service, or repository owner for each governed export-center and audit flow.
- [x] 1.5 Freeze the current verification surface for each governed export-center and audit flow.

## 2. Failing verification targets and semantic freeze

- [x] 2.1 Define failing or missing verification for export-center query, retry, and download workflows.
- [x] 2.2 Define failing or missing verification for audit page reachability, filter-query, and detail-read workflows.
- [x] 2.3 Freeze governed failure semantics for unauthenticated, forbidden, missing route/resource, and explicit non-success outcomes across the operational follow-up batch.

## 3. Export-center recovery lane

- [x] 3.1 Add failing regression coverage for governed export-center query and retry flows.
- [x] 3.2 Repair export-center route, contract, or page-init gaps required by the governed flows.
- [x] 3.3 Verify export-center failures surface as explicit non-success outcomes instead of silent empty success.
- [x] 3.4 Add targeted regression or smoke coverage proving recovered export-center flows remain stable.

## 4. Audit recovery lane

- [x] 4.1 Add failing regression coverage for governed audit page entry and query initialization flows.
- [x] 4.2 Add failing regression coverage for governed audit detail-read flows.
- [x] 4.3 Repair audit route, contract, page-init, or detail-read gaps required by the governed flows.
- [x] 4.4 Verify audit failures remain distinguishable across authorization, missing route/resource, and business-query execution failure.
- [x] 4.5 Add targeted regression or smoke coverage proving recovered audit flows remain stable.

## 5. Operational closeout

- [x] 5.1 Capture regression evidence for export-center and audit recovery.
- [x] 5.2 Run frontend lint, typecheck, and affected tests for the recovered operational scope.
- [x] 5.3 Run backend tests required for the recovered operational scope.
- [x] 5.4 Capture release-readiness evidence before widening scope again.
