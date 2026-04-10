## Why

Watermark pages and runtime watermark rendering still call `/user/personInfo` and `/user/ipInfo`, but the Go backend does not currently register these endpoints. This causes request failures and breaks watermark identity resolution in the frontend.

## What Changes

- Add compatibility endpoints `/user/personInfo` and `/user/ipInfo` in the Go compatibility bridge
- Implement `PersonInfo` and `IPInfo` handlers in `user_handler.go` using authenticated context + user lookup fallback
- Add handler tests for both endpoints
- Keep existing frontend API paths unchanged to minimize risk

## Capabilities

### New Capabilities
- _(none)_

### Modified Capabilities
- `watermark-management`: restore watermark identity endpoint availability for watermark UI/runtime callers
- `api-compatibility-bridge`: extend `/user/*` compatibility mapping with person/ip info endpoints

## Impact

- Backend handlers: `apps/backend-go/internal/transport/http/handler/user_handler.go`
- Compatibility route mapping: `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
- Backend tests: `apps/backend-go/internal/transport/http/handler/user_handler_test.go`
- No frontend behavior change required; existing `personInfoApi` / `ipInfoApi` calls remain valid
