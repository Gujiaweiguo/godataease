## 1. Preview routing contract

- [x] 1.1 Trace current `previewSql` request/response flow in backend and frontend, and document where `datasourceId` is currently ignored.
- [x] 1.2 Define the supported routing matrix for `previewSql` (local sync, direct datasource preview, explicit unsupported) and align compatibility error semantics with existing `code/data/msg` behavior.
- [x] 1.3 Add or update backend tests that lock the chosen routing contract, including explicit non-success behavior when datasource-aware direct preview is unsupported.

## 2. Direct preview execution foundation

- [x] 2.1 Introduce the minimal backend execution abstraction for datasource-aware direct preview without expanding `DatasetRepository` into a multi-driver execution layer.
- [ ] 2.2 Implement datasource eligibility, authorization, timeout, and result-size enforcement for supported direct preview paths.
- [ ] 2.3 Add backend unit/integration coverage for supported, unsupported, forbidden, and execution-failure direct preview scenarios.

## 3. Remaining field compatibility endpoints

- [ ] 3.1 Implement `/datasetField/multFieldValuesForPermissions` by reusing existing field enumeration and permission-filtering logic rather than a side-channel field model.
- [ ] 3.2 Implement `/datasetField/copilotFields` using governed dataset field metadata and explicit missing/unauthorized failure semantics.
- [ ] 3.3 Add backend regression tests for both field compatibility endpoints, including empty-success and explicit-failure scenarios.

## 4. Frontend alignment and verification

- [x] 4.1 Update affected frontend dataset API callers or views so preview and field endpoints surface the new explicit routing and failure semantics without relying on silent fallback behavior.
- [ ] 4.2 Add or update frontend tests for any changed preview or field-consumption assumptions.
- [ ] 4.3 Run backend and frontend verification (`make test`, affected integration tests if preview execution changes persistence or datasource access behavior, `npm run lint`, `npm run ts:check`, and affected frontend tests) and capture any unsupported datasource limitations in the final change notes.
