## Why

The frontend share dialog already depends on editable share state, but the Go mainline only implements share creation, validation, detail lookup, switcher, and ticket APIs. Without edit APIs for UUID, expiration, and password, the existing share UI cannot complete its main editing flow on Go mainline.

## What Changes

- Add Go share editing APIs for UUID updates, expiration updates, and password updates used by the existing frontend share dialog.
- Define deterministic validation and persistence behavior for share UUID edits so the frontend can use the same endpoint during inline editing.
- Align share detail responses with the post-edit reload flow used by current frontend components.
- Add backend verification coverage for share editing request validation, persistence rules, and compatibility behavior.

## Capabilities

### New Capabilities

### Modified Capabilities
- `share-management`: expand share management requirements to cover editable share link state for UUID, expiration, and password-protected sharing flows.

## Impact

- Affected code likely includes:
  - `apps/backend-go/internal/domain/share/*`
  - `apps/backend-go/internal/service/share_service.go`
  - `apps/backend-go/internal/repository/*share*`
  - `apps/backend-go/internal/transport/http/handler/share_handler.go`
- Affected APIs:
  - `/share/editUuid`
  - `/share/editExp`
  - `/share/editPwd`
  - existing `/share/detail/:resourceId` reload path
- Affected frontend call sites:
  - `apps/frontend/src/views/share/share/ShareHandler.vue`
  - `apps/frontend/src/views/share/share/ShareVisualHead.vue`
