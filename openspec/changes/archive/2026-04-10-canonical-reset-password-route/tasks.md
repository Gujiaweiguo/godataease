## Tasks

- [ ] Add canonical user-management routes for default password and reset password
  - File: `apps/backend-go/internal/transport/http/handler/user_handler.go`
  - Register `/system/user/defaultPwd` and `/system/user/resetPwd/:id`

- [ ] Migrate frontend API helpers to canonical user-management paths
  - File: `apps/frontend/src/api/user.ts`
  - Update `defaultPwdApi` and `resetPwdApi`

- [ ] Add handler tests for canonical routes
  - File: `apps/backend-go/internal/transport/http/handler/user_handler_test.go`
  - Cover canonical defaultPwd success and canonical resetPwd invalid-id rejection

- [ ] Verification
  - `cd apps/backend-go && make test`
  - `cd apps/frontend && npm run lint && npm run ts:check`
