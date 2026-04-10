## Design Overview

This is a compatibility fill slice for watermark identity endpoints.

### 1) Endpoint placement

The frontend currently calls `/user/personInfo` and `/user/ipInfo`. These are Java-era compatibility endpoints, so we register them under the existing compatibility bridge user group (`RegisterCompatibilityBridgeRoutes`).

### 2) Handler behavior

Add two handlers to `UserHandler`:

- `PersonInfo`
  - requires authenticated user context (`user_id`)
  - resolves account/name using `loadUserByID` when available, otherwise falls back to context username
  - returns `{ id, account, name, ip, model }`
- `IPInfo`
  - same identity resolution
  - returns `{ account, name, ip }`

Identity resolution logic is centralized in `resolveWatermarkIdentity`.

### 3) Route registration

In `RegisterCompatibilityBridgeRoutes`, add:

- `GET /user/personInfo` -> `user.PersonInfo`
- `GET /user/ipInfo` -> `user.IPInfo`

### 4) Test strategy

Add handler tests that verify:

- success payload shape of `PersonInfo`
- success payload shape of `IPInfo`

### 5) Compatibility strategy

This change intentionally preserves existing frontend API paths and only backfills missing backend compatibility endpoints.
