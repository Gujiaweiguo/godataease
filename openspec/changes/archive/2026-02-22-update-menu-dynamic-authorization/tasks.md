# Tasks: Update Menu Dynamic Authorization

> 单一权威任务清单，与 `.sisyphus/plans/update-menu-dynamic-authorization-v1.md` 同步。

## Execution Order

- Wave 1 (foundation): `MENU-AUTH-001`
- Wave 2 (parallel core): `MENU-AUTH-002`, `MENU-AUTH-003`
- Wave 3 (parallel API): `MENU-AUTH-004`, `MENU-AUTH-005`
- Wave 4 (cutover): `MENU-AUTH-006`
- Wave 5 (integration): `MENU-AUTH-007`, `MENU-AUTH-008`
- Final verification wave: `MENU-AUTH-F1`, `MENU-AUTH-F2`, `MENU-AUTH-F3`

## Defaults Applied (Plan v1)

- bootstrap 管理员角色识别优先使用 `role_code='admin'`，找不到时回退 `role_id=1`
- 菜单删除策略默认为"有子节点时拒绝删除"
- 兼容写死兜底路径以 feature flag 方式保留，默认在生产关闭

---

## Tasks

 [x] **MENU-AUTH-001 | 建立角色菜单关系基础（Wave 1）**
  - 依赖: None
  - 风险等级: HIGH
  - 输入:
    - `core_menu`, `sys_role` 现状
    - 目标变更 `update-menu-dynamic-authorization`
  - 输出:
    - `sys_role_menu` 迁移脚本（含唯一约束、索引、外键）
    - bootstrap 管理员菜单授权初始化脚本
    - rollback SQL 脚本
  - 验收标准:
    - `sys_role_menu` 存在且 `(role_id, menu_id)` 唯一
    - bootstrap 管理员拥有全量菜单映射
    - rollback 脚本可在演练库回退并通过校验
  - 回滚方案:
    - 执行 `DROP TABLE/RESTORE` 或 `TRUNCATE + backup restore` 原子事务
    - 回滚后验证 `role-menu` 记录数与快照一致

 [x] **MENU-AUTH-002 | 后端角色菜单仓储与服务层（Wave 2）**
  - 依赖: `MENU-AUTH-001`
  - 风险等级: MEDIUM
  - 输入:
    - `sys_role_menu` 新表结构
    - 现有 role/menu service 模式
  - 输出:
    - RoleMenu repository/service
    - 幂等保存逻辑（全量覆盖或差量更新）
  - 验收标准:
    - 相同授权请求重复提交不产生重复关系
    - 角色不存在/菜单不存在时返回结构化错误
    - 单元测试覆盖创建、覆盖、撤销三类路径
  - 回滚方案:
    - 通过 feature flag 关闭调用路径，恢复旧服务分支
    - 保留表结构不删，仅停止读写以保障快速回退

 [x] **MENU-AUTH-003 | 菜单组装服务按角色过滤（Wave 2）**
  - 依赖: `MENU-AUTH-001`
  - 风险等级: MEDIUM
  - 输入:
    - `core_menu`
    - 当前用户角色上下文
    - role-menu 映射
  - 输出:
    - 统一菜单组装服务（menu tree + role router）
    - 管理员兜底策略（仅受配置开关控制）
  - 验收标准:
    - 同一用户在 menu tree 与 role router 的可见菜单一致
    - 非授权菜单不出现在输出 payload
    - 无角色用户返回空菜单（非 500）
  - 回滚方案:
    - 切回兼容旧组装器（保留旧路径开关）
    - 记录差异日志用于二次切换前对账

 [x] **MENU-AUTH-004 | 菜单管理 API（CRUD/排序/显隐）（Wave 3）**
  - 依赖: `MENU-AUTH-001`
  - 风险等级: MEDIUM
  - 输入:
    - 菜单领域模型与 service
    - 管理端菜单维护需求
  - 输出:
    - 菜单查询/创建/更新/删除/排序/显隐 API
    - 参数校验与错误码规范
  - 验收标准:
    - CRUD 全链路可执行且排序稳定
    - 删除含子菜单节点时按策略拒绝或受控处理
    - API 响应保持统一 `{code,data,msg}` 合约
  - 回滚方案:
    - 回滚到 query-only 路由集合
    - 新增端点下线但不破坏旧端点可用性

 [x] **MENU-AUTH-005 | 角色-菜单授权 API（Wave 3）**
  - 依赖: `MENU-AUTH-001`, `MENU-AUTH-002`
  - 风险等级: MEDIUM
  - 输入:
    - role/menu 领域服务
    - 角色管理页授权交互需求
  - 输出:
    - 查询角色菜单授权 API
    - 保存角色菜单授权 API
  - 验收标准:
    - 查询返回可用于树形勾选的授权数据
    - 保存后重新查询结果一致
    - 授权撤销后相关菜单即刻失效（下次拉取生效）
  - 回滚方案:
    - 关闭角色菜单写入入口并回退到只读模式
    - 恢复到切换前快照授权数据

 [x] **MENU-AUTH-006 | 兼容端点去写死并统一组装（Wave 4）**
  - 依赖: `MENU-AUTH-002`, `MENU-AUTH-003`, `MENU-AUTH-005`
  - 风险等级: HIGH
  - 输入:
    - `/api/roleRouter/query`
    - `/api/auth/menuResource`
    - 统一菜单组装服务
  - 输出:
    - compatibility 端点改为动态数据输出
    - 写死菜单逻辑移除或仅保留开关化应急分支
  - 验收标准:
    - compatibility 与 canonical 菜单可见范围一致
    - 不再依赖固定硬编码菜单数组
    - 前端路由解析结构字段兼容
  - 回滚方案:
    - 通过配置开关回到应急兼容分支
    - 保留切换日志和差异对账结果

 [x] **MENU-AUTH-007 | 前端管理与动态路由集成（Wave 5）**
  - 依赖: `MENU-AUTH-004`, `MENU-AUTH-005`, `MENU-AUTH-006`
  - 风险等级: MEDIUM
  - 输入:
    - 新菜单管理 API
    - 新角色菜单授权 API
  - 输出:
    - 菜单管理页面能力（至少可维护菜单树）
    - 角色管理中菜单授权配置能力
    - 动态路由消费后端授权树
  - 验收标准:
    - 管理员可新增菜单并在顶部菜单可见
    - 角色授权变更后对应用户菜单即时变化（重新登录/刷新后）
    - 前端无 `/api/api/...` 路径拼接错误
  - 回滚方案:
    - 前端保留 feature toggle，失败时回退旧菜单渲染策略
    - 保持后端新接口在线但前端不启用

 [x] **MENU-AUTH-008 | 自动化验证与准入门（Wave 5）**
  - 依赖: `MENU-AUTH-006`, `MENU-AUTH-007`
  - 风险等级: LOW
  - 输入:
    - 关键接口清单
    - 角色样本账号（admin + restricted role）
  - 输出:
    - 集成测试与回归脚本
    - parity 报告（compat vs canonical）
  - 验收标准:
    - `go test` 与前端 type/lint 校验通过
    - 关键菜单接口 200/403 语义符合预期
    - parity 报告无阻断级差异
  - 回滚方案:
    - 准入门失败则禁止发布，回退到上一个稳定镜像
    - 保留测试证据用于缺陷修复后复测

 [x] **MENU-AUTH-F1 | 数据回滚演练审计（Final）**
  - 依赖: `MENU-AUTH-008`
  - 风险等级: HIGH
  - 输入:
    - 迁移前快照
    - rollback SQL
  - 输出:
    - 回滚演练记录与恢复时间指标
  - 验收标准:
    - 回滚后数据完整性与快照一致
    - 回滚耗时满足发布窗口
  - 回滚方案:
    - 如演练失败，冻结发布并修复脚本后重演

 [x] **MENU-AUTH-F2 | 安全与权限一致性审计（Final）**
  - 依赖: `MENU-AUTH-008`
  - 风险等级: MEDIUM
  - 输入:
    - 授权矩阵
    - API 访问日志
  - 输出:
    - 权限一致性审计报告
  - 验收标准:
    - 未授权菜单不可见且直访返回 403
    - admin 无误阻断
  - 回滚方案:
    - 发现阻断级越权或误拒绝时，立即切回旧授权路径

 [x] **MENU-AUTH-F3 | 发布门禁签收（Final）**
  - 依赖: `MENU-AUTH-F1`, `MENU-AUTH-F2`
  - 风险等级: LOW
  - 输入:
    - F1/F2 报告
    - 变更检查清单
  - 输出:
    - 发布签收记录（可追溯）
  - 验收标准:
    - 所有阻断项关闭
    - Atlas/Hephaestus 双方确认执行完毕并签收
  - 回滚方案:
    - 任一阻断项未关闭则停止发布并保留当前稳定版本

---

## Dependency Matrix

- `MENU-AUTH-001` -> blocks `002,003,004,005`
- `MENU-AUTH-002` -> blocks `005,006`
- `MENU-AUTH-003` -> blocks `006`
- `MENU-AUTH-004` -> blocks `007`
- `MENU-AUTH-005` -> blocks `006,007`
- `MENU-AUTH-006` -> blocks `007,008`
- `MENU-AUTH-007` -> blocks `008,F1,F2,F3`
- `MENU-AUTH-008` -> blocks `F1,F2,F3`

## Risk Policy

- `HIGH`: schema/cutover/auth risks; requires rollback-ready scripts before merge
- `MEDIUM`: API contract drift risk; requires integration tests
- `LOW`: tooling/test documentation risks; requires CI pass

## Success Criteria

- No hardcoded menu tree in compatibility runtime path
- Role-based menu visibility works for menu and route payloads
- Unauthorized direct menu route access returns 403 semantics
- Compatibility and canonical menu endpoints are parity-verified
