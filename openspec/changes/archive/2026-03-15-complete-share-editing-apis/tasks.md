## 1. Share edit contract setup

- [x] 1.1 Add Go route registration for `/share/editUuid`, `/share/editExp`, and `/share/editPwd`
- [x] 1.2 Define request DTOs and validation rules for share UUID, expiration, and password edits
- [x] 1.3 Ensure edited share fields are surfaced consistently through existing `/share/detail/:resourceId` responses

## 2. Share editing implementation

- [x] 2.1 Implement share UUID edit flow with format and uniqueness validation semantics compatible with the current frontend dialog
- [x] 2.2 Implement share expiration edit flow, including valid update and clear/invalid handling behavior
- [x] 2.3 Implement share password edit flow for auto-generated and custom password updates

## 3. Verification

- [x] 3.1 Add handler and service tests covering successful and rejected UUID, expiration, and password edits
- [x] 3.2 Add compatibility checks confirming the frontend edit-and-reload flow works with `/share/detail/:resourceId`
- [x] 3.3 Document that `/share/query` and `/share/proxyInfo` remain out of scope for this change
