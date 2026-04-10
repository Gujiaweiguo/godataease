## Why

Password reset and default-password query in user management are still consumed through compatibility bridge paths (`/user/resetPwd/:uid`, `/user/defaultPwd`) while the rest of user lifecycle APIs already use canonical `/system/user/*` routes. This leaves a avoidable compat dependency in T6 and prevents the canonical route group from owning the audited reset-password workflow.

## What Changes

- Add canonical route `GET /system/user/defaultPwd` in `RegisterUserRoutes`, reusing `GetDefaultPassword`
- Add canonical route `POST /system/user/resetPwd/:id` in `RegisterUserRoutes`, reusing `ResetPasswordCompat` with existing audit middleware config
- Migrate frontend `defaultPwdApi` to `/system/user/defaultPwd`
- Migrate frontend `resetPwdApi` to `/system/user/resetPwd/:uid`
- Add handler tests covering canonical default-password and canonical reset-password invalid-id behavior

## Capabilities

### New Capabilities
- _(none)_

### Modified Capabilities
- `user-management`: canonicalize default-password and reset-password endpoints under `/system/user/*` while preserving compat bridge aliases

## Impact

- Backend: `apps/backend-go/internal/transport/http/handler/user_handler.go`
- Backend tests: `apps/backend-go/internal/transport/http/handler/user_handler_test.go`
- Frontend: `apps/frontend/src/api/user.ts`
- No breaking change: compat bridge routes are retained
