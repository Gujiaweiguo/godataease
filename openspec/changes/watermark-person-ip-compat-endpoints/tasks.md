## Tasks

- [ ] Add `PersonInfo` and `IPInfo` handlers to user handler
  - File: `apps/backend-go/internal/transport/http/handler/user_handler.go`

- [ ] Register compatibility routes for watermark identity endpoints
  - File: `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
  - Add `GET /user/personInfo` and `GET /user/ipInfo`

- [ ] Add handler tests
  - File: `apps/backend-go/internal/transport/http/handler/user_handler_test.go`
  - Cover success payload for both endpoints

- [ ] Verification
  - `cd apps/backend-go && make test`
