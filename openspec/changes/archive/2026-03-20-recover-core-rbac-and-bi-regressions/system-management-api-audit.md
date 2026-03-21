## System-management API audit

This note captures the frontend API usage classification for the recovered system-management scope.

### Canonical dedicated clients

- `apps/frontend/src/api/org.ts`
  - `orgListApi`
  - `orgCreateApi`
  - `orgUpdateApi`
  - `orgDeleteApi`
  - `orgTreeApi`
  - `queryUserOptionsApi`
- `apps/frontend/src/api/menu.ts`
  - `menuQueryApi`
  - `menuDetailApi`
  - `menuCreateApi`
  - `menuUpdateApi`
  - `menuDeleteApi`
  - `menuUpdateSortApi`
  - `menuUpdateHiddenApi`

### Compatibility-aliased but currently valid clients

- `apps/frontend/src/api/auth.ts`
  - `queryUserApi` → `/user/byCurOrg`
  - `queryRoleApi` → `/role/byCurOrg`
  - `resourceTreeApi` → `/auth/busiResource/:flag`
  - `resourcePerApi` → `/auth/busiPermission`
  - `resourcePerSaveApi` → `/system/role/permission/save`
  - `menuTreeApi` → `/auth/menuResource`
  - `roleMenuAuthApi` / `roleMenuAuthSaveApi` → `/roleMenu/auth`

These paths are compatibility aliases rather than clean domain-specific REST surfaces, but they are now part of the recovered, verified admin flows and should be treated as canonical for the current recovered scope until a later API cleanup change is scheduled.

### Duplicate client paths removed during recovery

- Removed duplicate `queryUserOptionsApi` export from `apps/frontend/src/api/auth.ts`.
- `apps/frontend/src/views/system/user/index.vue` already imports `queryUserOptionsApi` from `apps/frontend/src/api/org.ts`, so `org.ts` is now the single source of truth for organization-option loading in the recovered user-management flow.

### Classification summary

- **Aligned**: `org.ts` organization CRUD/tree APIs, `menu.ts` menu administration APIs.
- **Compatibility-aliased**: user/role list bootstrap and role permission/menu authorization APIs currently used by recovered admin pages.
- **Mismatched but fixed during this change**: `/user/org/option` previously bridged to user options; now correctly bridged to organization list semantics expected by the user-management page.
- **Remaining implementation gap**: data-permission APIs are recovered as a compatibility subset only, documented separately in `permission-admin-recovery-notes.md`.
