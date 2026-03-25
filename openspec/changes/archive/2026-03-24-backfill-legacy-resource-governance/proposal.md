## Why

当前 `align-permission-center-with-xpack` 已经完成统一权限中心的主流程闭环，但它明确只覆盖“已纳管”或“新创建且位于已纳管父级下”的资源。对于历史上已经存在、但从未注册进 `sys_resource` / `sys_resource_perm` 的 datasource、dataset、dashboard、screen，系统仍然保留旧的类型级或散落语义。

这会带来一个持续性的治理裂缝：同样出现在统一权限中心里的资源类型，新增资源和历史资源的授权来源、继承行为、查询结果与运行时鉴权并不完全一致。继续在这个状态下扩展 direct-user grants 或更复杂的资源视角编辑，会把“新资源可治理、老资源半治理”的边界固化成长期负担。

这个 change 需要把问题单独收敛：为历史资源建立可审计、可回滚、可验证的补纳管与权限继承补算机制，让统一权限中心、资源视角查询、用户视角查询与运行时鉴权真正落到同一套治理模型上。

## What Changes

- 为历史 datasource / dataset / dashboard / screen 建立补注册流程，将未纳管资源补齐到统一的资源身份模型中。
- 对齐历史资源的父级定位、资源类型映射和治理边界，明确哪些资源可被自动纳管，哪些异常数据需要显式跳过或记录。
- 为已补纳管的历史资源补算资源组/父级继承权限，使其在统一权限中心中的表现与新资源一致。
- 对齐资源视角、用户视角和运行时鉴权对历史资源的有效授权判定，避免“中心可见但运行时报错”或“运行时可见但中心缺失”的语义漂移。
- 建立补纳管过程的审计、幂等和回滚边界，确保迁移型治理不会破坏现有已生效授权。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `permission-config`: 扩展统一权限中心对历史资源的治理覆盖范围，使历史资源与新纳管资源遵循同一有效授权模型。
- `datasource-management`: 明确历史数据源资源的注册、继承与授权对齐行为。
- `dataset-management`: 明确历史数据集资源的注册、继承与授权对齐行为。
- `visualization-management`: 明确历史仪表板、大屏等可视化资源的注册、继承与授权对齐行为。

## Impact

- Affected backend:
  - `apps/backend-go/internal/service/resource_perm_service.go`
  - `apps/backend-go/internal/service/datasource_service.go`
  - `apps/backend-go/internal/service/dataset_service.go`
  - `apps/backend-go/internal/service/visualization_service.go`
  - `apps/backend-go/internal/repository/resource_perm_repo.go`
  - 与 `sys_resource` / `sys_resource_perm` 相关的资源注册、继承补算与查询路径
- Affected frontend:
  - `apps/frontend/src/views/system/permission/ResourcePermission.vue`
  - 可能涉及统一权限中心中资源树/资源详情的历史资源展示一致性
- Affected verification:
  - 历史资源补纳管后的 by-user / by-resource 一致性验证
  - 历史资源运行时访问与统一权限中心展示一致性验证
  - 补纳管任务的幂等、审计、异常跳过与回滚验证
- Affected specs:
  - `openspec/specs/permission-config/spec.md`
  - `openspec/specs/datasource-management/spec.md`
  - `openspec/specs/dataset-management/spec.md`
  - `openspec/specs/visualization-management/spec.md`
