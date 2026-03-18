# Regression Evidence

## Backend interactiveTree parity

- `go test ./internal/transport/http/handler -run 'TestFrontendCompatHandler_InteractiveTreeUsesAuthorizedMenus|TestFrontendCompatHandler_InteractiveTreeReturnsRealDataVNodes|TestFrontendCompatHandler_InteractiveTreeFiltersUnauthorizedVisualizationScopes|TestBuildVisualizationTreeContractShape' -count=1`
- `TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -tags=integration ./internal/service -run 'TestVisualizationServiceIntegration_(InteractiveTree|Detail|Detail_Completeness)$' -count=1`

Covered outcomes:
- dashboard interactive tree returns real visualization nodes instead of synthetic authorization placeholders
- dataV interactive tree returns real visualization nodes instead of synthetic authorization placeholders
- unauthorized visualization scopes remain filtered to empty trees
- node contract preserves `id`, `pid`, `leaf`, `weight`, `extraFlag`, `extraFlag1`, and `children`

## Frontend compatibility evidence

- `npm run test -- --run tests/unit/store/interactive.test.ts`

Covered outcomes:
- interactive store preserves real dashboard/dataV resource nodes from `queryBusiTreeApi`
- derived state such as `leafNodeCount` and `anyManage` remains correct for real resource trees
- no frontend caller requires synthetic-root-only semantics

## Governance evidence

- whitelist metadata for `/dataVisualization/interactiveTree` updated from `partial` to `full`
- `go test ./internal/transport/http/handler -run 'TestStatusConsistencyWithWhitelist|TestNoPlaceholderSuccessInCompatibilityBridge' -count=1`
- `./scripts/check-status-drift.sh`
- `npx playwright test e2e/interactive/interactive.spec.ts --grep @system-smoke`
- `npx playwright test e2e/official-example/official-example.spec.ts --grep @system-smoke`

Covered outcomes:
- governance metadata is aligned with runtime status for `/dataVisualization/interactiveTree`
- no compatibility status drift is detected after promoting the endpoint to `full`
- dashboard and big-screen interactive entry smoke paths remain healthy in targeted frontend smoke coverage
