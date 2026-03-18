# Regression Evidence

## Backend aggregate parity

- `go test ./internal/transport/http/handler -run 'TestFrontendCompatHandler_InteractiveTree(ReturnsDatasetAndDatasourceNodes|HandlesDatasetAndDatasourceLoaderErrorsDeterministically|UsesAuthorizedMenus|ReturnsRealDataVNodes)' -count=1`

Covered outcomes:
- `interactiveTree` now returns real dataset nodes for authorized dataset scope
- `interactiveTree` now returns real datasource tree payloads for authorized datasource scope
- dataset/datasource loader failures degrade to deterministic empty lists instead of breaking the aggregate response

## Frontend aggregate parity

- `npm run test -- --run tests/unit/store/interactive.test.ts`

Covered outcomes:
- batched interactive loading preserves dataset and datasource nodes
- bootstrap initialization uses the batched interactive path rather than dataset/datasource direct tree calls
- aggregate store state remains valid for all four BI domains

## Governance evidence

- `apps/backend-go/testdata/contract-diff/critical-whitelist.yaml`
  - interactiveTree governance notes updated to reflect dataset/datasource aggregate discovery coverage
