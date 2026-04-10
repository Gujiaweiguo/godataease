## Design Overview

This is a compatibility-to-canonical migration slice for user password reset related endpoints.

### 1) Canonical route registration

Update `RegisterUserRoutes` in `user_handler.go` to register:

- `GET /system/user/defaultPwd` → `h.GetDefaultPassword`
- `POST /system/user/resetPwd/:id` → `h.ResetPasswordCompat` with existing audit middleware configuration:
  - `ActionType: audit.ActionTypeUserAction`
  - `ActionName: "RESET_USER_PASSWORD"`
  - `ResourceType: audit.ResourceTypeUser`

The handler implementation is intentionally reused to keep behavior parity with compat routes.

### 2) Frontend API migration

Update two frontend API helpers in `apps/frontend/src/api/user.ts`:

- `defaultPwdApi` from `/user/defaultPwd` to `/system/user/defaultPwd`
- `resetPwdApi` from `/user/resetPwd/:uid` to `/system/user/resetPwd/:uid`

No component-level changes are required because call signatures stay the same.

### 3) Test coverage

Add handler-level tests in `user_handler_test.go` via `RegisterUserRoutes`:

- canonical default password endpoint success (`GET /api/system/user/defaultPwd`)
- canonical reset password endpoint invalid-id rejection (`POST /api/system/user/resetPwd/invalid`)

### 4) Compatibility strategy

This change does **not** remove compat bridge aliases. It only introduces canonical routes and migrates frontend callers so downstream cleanup can safely remove compat dependencies later.
