## 1. Menu permission canonical API migration

- [x] 1.1 Update `MenuPermission.vue` to use canonical role-menu APIs for role-selected query and save paths.
- [x] 1.2 Keep initial no-role tree load behavior stable without broad UI workflow changes.

## 2. Focused regression coverage

- [x] 2.1 Add frontend unit test covering canonical role-menu load/save invocation from menu permission tab.

## 3. Verification

- [x] 3.1 Run focused frontend test for menu permission tab migration.
- [x] 3.2 Run frontend validation (`npm run lint`, `npm run ts:check`).
