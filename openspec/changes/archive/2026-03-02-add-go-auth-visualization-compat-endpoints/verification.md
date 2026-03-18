# Verification Evidence

## Backend

- Command: `go test ./internal/transport/http/handler -run "TestResolveBusiTypes|TestBuildVisualizationTreeValidation|TestContractDiffTemplateRoutes"`
  - Result: PASS
- Command: `make test`
  - Result: PASS
- Command: `make build`
  - Result: PASS

## Compatibility Endpoint Regression

- Command: `docker compose -f infra/compose/docker-compose.yml up -d`
  - Result: `godataease-app` and `godataease-redis` started
- Command: `./scripts/compat-checks/run_auth_visualization_compat.sh`
  - Result: PASS (`18/18`)
  - Coverage: `/api/*` and `/de2api/*` for auth permission family, role aliases, and `/dataVisualization/tree`

## Frontend Regression

- Command: `npm run test:ci`
  - Result: PASS
- Command: `npm run ts:check`
  - Result: FAIL (pre-existing frontend typing issues not introduced by this change)
- Command: `npm run test:core`
  - Result: FAIL (pre-existing test failures in embedding/event and TokenManager related suites)

## Notes

- This change does not modify frontend source code.
- Frontend failures were recorded for transparency as requested by task 3.3.
