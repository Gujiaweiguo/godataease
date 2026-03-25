## 1. Permission Center Contract Freeze

- [x] 1.1 冻结统一权限中心的页面 IA、菜单/资源/行列三类页签与关键接口契约
- [x] 1.2 明确按用户/按资源双视角共享同一有效授权状态的判定口径
- [x] 1.3 明确角色仅作为授权载体使用，不在本 change 中回改角色生命周期规则

## 2. Backend Authorization Alignment

- [x] 2.1 对齐菜单权限查询/保存与角色菜单映射的统一治理语义
- [x] 2.2 对齐资源权限双视角配置与新增资源继承生效规则
- [x] 2.3 对齐行权限和列权限在统一权限中心中的查询/保存语义
- [x] 2.4 修正未授权访问的返回语义，确保与真实 404 保持可区分

## 3. Frontend Unified Permission Center

- [x] 3.1 在统一权限配置页中落地菜单权限、资源权限、行列权限三类托管工作流
- [x] 3.2 对齐按用户/按资源双视角切换与结果一致性展示
- [x] 3.3 验证统一权限中心与角色载体、菜单可见性、直接路由访问行为一致

## 4. Verification Gate

- [x] 4.1 执行菜单授权、资源授权、行列权限与拒绝语义的后端测试与回归矩阵
- [x] 4.2 执行统一权限配置页的前端联调、页签切换和双视角一致性验证
- [x] 4.3 形成“权限中心已消费稳定 foundation + role workflows 输出”的最终跨 change 验收记录

## Progress record

- unified permission center now keeps `menu/resource/data` inside one governed entry path and supports stable `?tab=` routing in `apps/frontend/src/views/system/permission/index.vue`
- all permission-center role selectors continue to consume `/role/byCurOrg`, preserving the role-carrier boundary from completed role workflows
- `/auth/busiTargetPermission` now accepts the official `{id,type,flag}` shape and returns a read-only resource-perspective payload instead of `501000`
- `apps/frontend/src/views/system/permission/ResourcePermission.vue` now exposes a first dual-entry slice: editable role-carrier view plus read-only by-resource preview backed by the new target query; save semantics and inheritance rollout remain pending
- resource-perspective backend preview now derives real user entries from the same persisted role/user grants used by carrier-side saves, keyed by governed resource type permission prefixes
- carrier-side resource permission saves now refresh the current resource-perspective preview in `ResourcePermission.vue`, and switching back to resource view re-queries the effective authorization state instead of showing stale data
- menu permission compat query/save now has handler-level integration coverage against MySQL, proving `/auth/menuPermission` and `/auth/saveMenuPer` round-trip through the same role-menu persistence model used by system routes
- `/system/permission` now has a hidden static route entry, and direct access redirects unauthenticated users to login before returning to the unified permission center path after authentication
- data permission backend now exposes a role-targeted row-permission pager so the unified center can inspect row rules by dataset + authorization carrier without inventing a fake column-permission carrier model
- `apps/frontend/src/views/system/permission/DataPermission.vue` now exposes a first dual-mode row-permission workflow (`按数据集` / `按角色`), while column permissions stay explicitly dataset-scoped to match the current backend persistence model
- row-permission create/edit flows now lock to the selected role when the unified center is in `按角色` mode, and unsupported `按系统变量` create-paths are no longer offered by the Go-backed UI slice
- `/auth/saveBusiTargetPer` now exits the `501000` placeholder path with a guarded contract: resource-perspective saves only mutate the explicitly listed role carriers and only within the selected resource-type permission prefix, leaving other resource types untouched
- `apps/frontend/src/views/system/permission/ResourcePermission.vue` now exposes a minimal editable by-resource slice for role-derived entries, while direct user grants remain read-only and resource-group inheritance stays explicitly pending instead of being faked as complete
- MySQL-backed resource-perspective queries no longer fail with `Error 3065` on `DISTINCT + ORDER BY` mismatches in `apps/backend-go/internal/repository/resource_perm_repo.go`, and handler integration coverage now exercises `/auth/busiTargetPermission` plus `/auth/saveBusiTargetPer` round-trips against MySQL
- manual QA on the local hybrid environment verified that unauthenticated access to `/#/system/permission?tab=resource` redirects to login and returns to the governed permission center after authentication, and `menu/resource/data` tab switches keep the `?tab=` route in sync
- manual QA also verified the current role-derived dual-view slice end-to-end: carrier-side dashboard permission changes appear in resource view, resource-view role edits round-trip back to the same carrier, while direct-user editing and inheritance remain explicitly out of scope for the current change slice
- `/auth/userPerspective` now exposes the existing backend user-perspective effective permission query, and `GetUserResources(...)` now respects `resourceType` filtering instead of returning cross-type grants
- `apps/frontend/src/views/system/permission/ResourcePermission.vue` now supports `按角色 / 按用户 / 按资源` three-way switching, with a read-only by-user effective-state view and optional user filtering inside resource view for same-scope comparison
- manual QA on the local hybrid environment verified the new 3.2 slice end-to-end: for user `test` under `仪表板` scope, the by-user view and the by-resource filtered view resolved to the same effective grants (`仪表板查看` / `仪表板编辑`) for resource `新建`
- backend regression evidence now covers the implemented permission-center matrix: menu permission query/save MySQL round-trips, resource query/save/userPerspective MySQL round-trips, non-role resource-target rejection, row-permission handler/service success + invalid targetId + unsupported targetType branches, and column-permission service save/update/delete/page mappings
- governed datasource / dataset / dashboard / screen resources now register into `sys_resource`/`sys_resource_perm` only when created beneath already governed parent folders, inherit the parent-folder effective permission gate on creation, and keep that gate consumable by resource-perspective queries plus runtime permission checks
- `/auth/saveBusiTargetPer` now persists the current resource-level permission gate from the submitted role-target union while preserving already-effective direct-user grants for the same governed instance, and `/auth/userPerspective` can align to the currently selected resource when a resource id is present so the unified center keeps by-user/by-resource comparison on the same governed instance
- `apps/frontend/src/views/system/permission/ResourcePermission.vue` now clears stale resource selections when switching into `按用户` mode so the by-user view remains type-scoped unless the current request explicitly chooses a governed resource instance for comparison
- targeted verification now covers inheritance hooks in `DatasourceService.Save`, `DatasetService.Create`, and `VisualizationService.Save` plus MySQL-backed compat-handler round trips after the resource-gate rollout

## Cross-change acceptance record (implemented scope)

### Dependency confirmation: foundation outputs

- the unified permission center consumes the stable org/identity boundary delivered by `align-org-identity-foundations-with-xpack`, including current-org JWT context, `/user/info` bootstrap payloads, `/user/switch/:id`, and authenticated `/api` routing that now preserves org-aware permission requests
- permission-center behavior does not redefine org selection or identity bootstrap semantics; it reuses the already-validated foundation contracts as prerequisites for governed permission queries, saves, and direct-route access

### Dependency confirmation: role-workflow outputs

- the unified permission center consumes stable role carriers from `align-role-workflows-with-xpack`, especially `/role/byCurOrg` as the current-org role source for menu/resource/data permission selectors
- permission-center resource and menu authorization flows continue to treat roles as carriers only; they do not take ownership of role lifecycle, member mounting, last-role guardrails, or inheritance-boundary decisions already validated in the completed role-workflow change

### Implemented-scope acceptance evidence

- menu permission governance is verified through the unified center and MySQL-backed handler integration coverage for `/auth/menuPermission` and `/auth/saveMenuPer`
- resource permission governance is verified for the implemented scope: carrier-side save/query, resource-side effective-state preview, guarded `/auth/saveBusiTargetPer`, MySQL-backed `/auth/busiTargetPermission` round-trips, and read-only `/auth/userPerspective` alignment with the same underlying grants
- row/column permission governance is verified for the implemented scope: role-targeted row-permission pager, handler/service regression coverage for valid and invalid row-target requests, and service-level save/update/delete/page coverage for dataset-scoped column permissions
- frontend runtime verification confirms governed `menu/resource/data` tab routing, direct login redirect/return for `/#/system/permission`, and by-user/by-resource effective-state consistency for the implemented resource-permission slice

### Verification evidence summary

- backend regression matrix now covers implemented menu/resource/row/column semantics plus denial branches, including MySQL integration for menu/resource permission compat handlers
- local hybrid manual QA has already validated direct-route access, `?tab=` synchronization, role/resource dual-view coherence, and the newer by-user/by-resource alignment slice under the current implementation boundary

### Remaining blockers and non-goals

- no remaining blocker prevents this change from closing: task `2.2` resource inheritance is now implemented for newly governed datasource / dataset / dashboard / screen resources created beneath already governed folders
- existing legacy resources that have never been registered into `sys_resource` continue to fall back to the previously implemented type-scoped semantics until they are governed through the unified resource workflow; this record does not claim a retroactive backfill of all historical resources

### Current acceptance status

- cross-change consumption is validated end to end: the permission center consumes stable outputs from the completed foundation and role-workflow changes without redefining them, and newly governed resources now inherit folder-level effective authorization state through the same governed center
- final closeout is complete for this change: resource inheritance, unified permission-center queries/saves, and cross-change acceptance evidence have all been verified within the implemented scope
