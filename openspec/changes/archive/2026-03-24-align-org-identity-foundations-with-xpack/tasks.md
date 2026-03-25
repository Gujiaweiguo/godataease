## 1. Contract Freeze and Foundation Baseline

- [x] 1.1 冻结组织、用户、登录、基础角色相关接口的当前响应基线与字段映射清单
- [x] 1.2 明确 foundation 输出的共享字段：current user、current org、available orgs、built-in role classification
- [x] 1.3 为组织上下文、用户归属和基础角色查询补齐契约测试骨架

## 2. Backend Foundation Alignment

- [x] 2.1 对齐组织树、组织切换和叶子组织删除相关服务/接口语义
- [x] 2.2 对齐用户的组织作用域成员基线与管理员基础操作契约
- [x] 2.3 对齐登录 bootstrap 与当前用户资料接口的组织感知身份上下文
- [x] 2.4 对齐系统级/组织级内置角色的基础分类与共享查询行为

## 3. Frontend Identity Bootstrap Alignment

- [x] 3.1 调整用户 store/bootstrap 流程以消费稳定的组织感知身份上下文
- [x] 3.2 对齐组织管理页与用户管理页对 current org / available orgs 的使用口径
- [x] 3.3 验证 foundation 输出可被后续 role workflows 与 permission center 直接复用

## 4. Verification Gate

- [x] 4.1 执行 foundation 相关后端测试与契约校验
- [x] 4.2 执行前端基础状态初始化、组织切换与系统管理入口回归验证
- [x] 4.3 形成“role workflows 依赖 foundation 已满足”的跨 change 验收记录

## Cross-change readiness record

- foundation runtime contract now returns stable `id/name/oid/language/currentOrg/availableOrgs` from both `/login/localLogin` and `/user/info`
- org switching is closed-loop through JWT `org_id` claim, middleware context extraction, and `/user/switch/:id` token reissue
- organization-scoped user membership baseline is persisted via `sys_user_role(user_id, role_id, org_id)` with a default built-in organization role
- role workflows can now consume stable org context, selected org, and org-scoped role queries without redefining identity bootstrap semantics
