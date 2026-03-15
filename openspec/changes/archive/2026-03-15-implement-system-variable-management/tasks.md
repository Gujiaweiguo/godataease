## 1. Variable definition contract setup

- [x] 1.1 Add Go route registration for `/sysVariable/create`, `/sysVariable/edit`, `/sysVariable/detail/:id`, `/sysVariable/delete/:id`, and `/sysVariable/query`
- [x] 1.2 Define request and response DTOs for system variable definition CRUD and query flows
- [x] 1.3 Keep system variable behavior isolated from existing `sysParameter` and dataset SQL variable logic

## 2. Variable definition and value implementation

- [x] 2.1 Implement create, edit, detail, delete, and query flows for system variable definitions
- [x] 2.2 Implement `/sysVariable/value/create`, `/sysVariable/value/edit`, `/sysVariable/value/delete/:id`, and `/sysVariable/value/batchDel` flows
- [x] 2.3 Implement `/sysVariable/value/selected/:page/:limit` and `/sysVariable/value/selected/:id` selection endpoints with deterministic paging and lookup behavior

## 3. Verification

- [x] 3.1 Add handler, service, and repository tests for system variable definition and value management
- [x] 3.2 Add compatibility verification for the frontend contract defined in `apps/frontend/src/api/variable.ts`
- [x] 3.3 Document that dataset SQL variable parsing remains out of scope for this change
