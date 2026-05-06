## 1. Organization-scoped governance baseline

- [x] 1.1 Audit governed organization, role, and user write paths to identify every service entry that must consume the active organization context
- [x] 1.2 Add organization-scope enforcement to governed IAM write services and cover rejection paths with backend tests
- [x] 1.3 Verify governed organization-scope behavior through canonical and compatibility route checks

## 2. Role and user foundation semantics

- [x] 2.1 Encode the built-in role baseline and organization-scoped role classification in shared role queries and fixtures
- [x] 2.2 Preserve and document the last-role block policy across role-member removal flows, including explicit rejection coverage
- [x] 2.3 Align governed user membership baseline handling with organization scope so downstream role workflows reuse the same context contract

## 3. Organization delete and compatibility policy

- [x] 3.1 Keep organization delete behavior deterministic with child rejection, soft delete, and auditable deferred resource disposition coverage
- [x] 3.2 Classify legacy IAM route families into permanent shim, frontend migration, and dual-support transition buckets in code/config/tests
- [x] 3.3 Verify `/user/org/option` and other governed compatibility aliases preserve canonical semantics under the selected route-family policy

## 4. Regression and rollout safety

- [x] 4.1 Add backend regression coverage for organization isolation, last-role rejection, and organization delete policy
- [x] 4.2 Run compatibility and frontend contract checks for governed IAM routes (`make test`, `make drift-check`, relevant frontend checks)
- [x] 4.3 Capture rollout and fallback notes for feature-flag or compat-layer rollback before downstream C2/C3 implementation starts
