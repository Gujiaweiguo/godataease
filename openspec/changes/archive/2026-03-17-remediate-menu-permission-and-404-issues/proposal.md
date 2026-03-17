# Change: 全面治理菜单权限缺失与 404 问题

## Why

当前项目存在多处菜单点击后出现“无权限”或 `Request failed with status code 404` 的问题，说明菜单可见性、前端路由、接口路径、后端能力与权限模型之间存在不一致。

这类问题不是单点缺陷，而是导航配置、角色授权、路由注册、页面初始化请求和后端资源暴露之间的系统性偏差。如果继续以零散 bugfix 方式处理，容易出现以下问题：

- 修复一个菜单后其他菜单仍然失效
- 同类问题反复出现但缺少统一判定标准
- 无法证明“核心菜单已全面修复”
- 后续改动再次引入同类问题而无法及时发现

因此需要一次面向“菜单访问链路”的专项测试与修复，将问题归类、批量修复，并形成后续可重复执行的验证基线。

## What Changes

### Investigation and Repair Scope
- 对核心菜单进行系统化巡检，覆盖菜单展示、路由跳转、页面加载、接口请求和权限校验链路
- 识别并修复以下类型问题：
  - 菜单存在但前端路由不存在
  - 路由存在但组件路径或重定向错误
  - 页面初始化请求命中不存在的接口
  - 后端接口已迁移或改名但前端未同步
  - 菜单展示逻辑与角色授权结果不一致
  - 无权限与资源不存在在运行时表现混淆
- 建立菜单访问异常的分类与处理规则
- 补充最小回归验证机制，降低后续回归风险

### Frontend
- 梳理菜单树、路由表、权限守卫与页面初始化 API 的映射关系
- 修复错误路由、错误组件路径、错误跳转和错误 API 地址
- 统一“无权限”和“404”场景下的前端反馈与兜底行为

### Backend
- 校验菜单相关接口、权限校验链路和资源注册一致性
- 修复前端实际依赖但后端不存在或路径不一致的接口
- 确保未授权与资源不存在返回可区分的语义结果

### Verification
- 建立核心角色 × 核心菜单的专项验证矩阵
- 执行核心菜单点击冒烟与关键页面初始化验证
- 记录剩余已知问题、范围边界与风险项

## Capabilities

### New Capabilities
- `menu-access-governance`

### Modified Capabilities
- `menu-management`
- `permission-config`

## Impact

### Affected frontend areas
- `apps/frontend/src/router/**`
- `apps/frontend/src/permission.ts`
- `apps/frontend/src/layout/**`
- `apps/frontend/src/views/**`
- `apps/frontend/src/api/**`

### Affected backend areas
- `apps/backend-go/internal/transport/http/**`
- `apps/backend-go/internal/service/**`
- `apps/backend-go/internal/repository/**`

### Affected specs
- `openspec/specs/menu-management/spec.md`
- `openspec/specs/permission-config/spec.md`
- `openspec/specs/menu-access-governance/spec.md` (new)
