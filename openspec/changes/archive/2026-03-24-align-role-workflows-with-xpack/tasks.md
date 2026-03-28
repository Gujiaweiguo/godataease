## 1. Role Workflow Contract Freeze

- [x] 1.1 冻结角色 CRUD、成员挂载、成员移除、继承约束和唯一角色策略的接口契约
- [x] 1.2 明确组织用户与外部用户两条成员挂载路径的输入校验与审计要求
- [x] 1.3 形成“仅消费 foundation 主数据，不重复定义组织/身份规则”的边界清单

## 2. Backend Role Workflow Alignment

- [x] 2.1 对齐角色创建、编辑、启停用、详情与查询行为
- [x] 2.2 实现组织用户与外部用户的角色成员挂载流程及幂等关系维护
- [x] 2.3 实现移除成员时的最后角色安全策略并固定确定性系统行为
- [x] 2.4 实现自定义角色继承边界校验，拒绝超出父级授权上限的配置

## 3. Frontend Hosted Role Administration

- [x] 3.1 在用户管理页角色页签中对齐角色列表、成员管理和继承约束工作流
- [x] 3.2 保证用户页签与角色页签切换时组织上下文持续一致
- [x] 3.3 移除或降级独立角色入口，确保托管式 IA 与 capability 定义一致

## 4. Verification Gate

- [x] 4.1 执行角色成员添加/移除、唯一角色策略和继承约束的后端测试
- [x] 4.2 执行用户管理页角色页签的前端交互与组织上下文回归验证
- [x] 4.3 形成“permission center 依赖稳定角色载体已满足”的跨 change 验收记录

## Cross-change readiness record

- hosted role workflows now converge in `apps/frontend/src/views/system/user/RoleTab.vue`, which consumes `/role/byCurOrg` for current-org role queries and the `/system/role/*` compatibility aliases for create/edit/delete updates
- the standalone `apps/frontend/src/views/system/role/index.vue` entry is downgraded to a bridge that redirects into `/system/user?tab=role`, so role lifecycle and membership semantics no longer exist as a duplicated independent surface
- backend verification for member mount/unmount, last-role protection, and inheritance boundaries is already green, and frontend hosted-role verification is green via `npm run lint` and `npm run ts:check`
- permission center can now consume stable role carriers and current-org role listings without owning role lifecycle or member-management semantics
