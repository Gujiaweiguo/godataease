## Context

当前 DataEase Go 后端的菜单数据通过硬编码方式返回，前端路由解析依赖固定的菜单结构。角色管理页面无法配置菜单授权，导致所有用户看到相同的菜单集。此设计文档描述如何实现数据驱动的菜单管理和角色菜单授权机制。

**当前状态**:
- `core_menu` 表存储菜单定义，但未用于动态输出
- `/api/roleRouter/query`、`/api/auth/menuResource` 返回硬编码菜单
- 无 `sys_role_menu` 关联表

**约束**:
- 保持前端路由解析结构兼容
- 迁移期间兼容端点不可中断
- 管理员角色需保持全量菜单访问

## Goals / Non-Goals

**Goals:**
- 实现基于 `sys_role_menu` 的角色菜单授权模型
- 菜单 API 支持动态 CRUD、排序、显隐控制
- 兼容端点从数据库动态输出菜单，消除硬编码
- 前端可配置角色菜单授权并即时生效

**Non-Goals:**
- 企业版 xpack 功能实现
- 非菜单资源权限（`sys_perm`）语义重构
- 前端路由架构重设计

## Decisions

### D1: 角色菜单关系表设计
**选择**: 新增 `sys_role_menu` 表，存储 `(role_id, menu_id)` 映射
**理由**: 简单 M:N 关系，支持增量授权配置
**替代方案**: 在 `sys_role` 中存储 JSON 数组 — 查询和约束复杂

```sql
CREATE TABLE sys_role_menu (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  role_id BIGINT NOT NULL,
  menu_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_role_menu (role_id, menu_id),
  INDEX idx_menu (menu_id)
);
```

### D2: 管理员角色识别策略
**选择**: 优先 `role_code='admin'`，找不到时回退 `role_id=1`
**理由**: 兼容现有数据和显式编码两种场景
**替代方案**: 仅用 `role_id=1` — 新环境可能不匹配

### D3: 菜单删除策略
**选择**: 有子节点时拒绝删除
**理由**: 防止悬空菜单，简化级联逻辑
**替代方案**: 级联删除或软删除 — 增加复杂度和恢复风险

### D4: 兼容端点改造策略
**选择**: 保持 `/de2api/*` 路径，内部调用统一菜单组装服务
**理由**: 前端无需改动路由配置，渐进迁移
**替代方案**: 新增 `/api/v2/*` 端点 — 前端需同步改动

### D5: 授权数据缓存策略
**选择**: 每次请求实时查询，不引入 Redis 缓存
**理由**: 简化首次实现，菜单变更即时生效
**替代方案**: Redis 缓存 + TTL — 增加缓存一致性问题

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| 迁移期间菜单不可见 | Bootstrap 脚本初始化管理员全量授权；开关化回退路径 |
| 授权配置错误导致锁定 | 保留硬编码应急分支，通过配置开关启用 |
| 前端路由解析字段不兼容 | 输出结构与现有硬编码 payload 完全一致 |
| 性能下降（实时查询） | 后续迭代可引入缓存，当前优先正确性 |

## Migration Plan

### 阶段 1: 数据准备
1. 执行 `sys_role_menu` DDL 迁移
2. 运行 bootstrap 脚本，初始化管理员角色全量菜单映射
3. 验证 rollback 脚本可恢复

### 阶段 2: 服务层实现
1. 实现 `RoleMenuRepository` 和 `RoleMenuService`
2. 实现统一菜单组装服务 `MenuAssemblyService`
3. 单元测试覆盖授权过滤逻辑

### 阶段 3: API 上线
1. 新增菜单管理 API 和角色授权 API
2. 改造兼容端点为动态输出
3. 通过配置开关控制新旧路径

### 阶段 4: 前端集成
1. 菜单管理页面连接新 API
2. 角色管理页面添加授权配置
3. 动态路由消费后端授权树

### 阶段 5: 验证与发布
1. 集成测试和 parity 报告
2. 数据回滚演练
3. 安全审计和发布签收

### 回滚策略

#### 数据库回滚脚本

文件: `apps/backend-go/migrations/mysql/20260222_rollback_sys_role_menu.sql`

```sql
-- 1. 可选：自动备份现有数据（通过 system_parameters 控制开关）
SET @backup_enabled = IFNULL((SELECT VALUE FROM system_parameters WHERE param_key = 'migration.backup_enabled'), 'true');

-- 2. 检查表是否存在
SET @table_exists = (
    SELECT COUNT(*)
    FROM information_schema.tables
    WHERE table_schema = DATABASE()
    AND table_name = 'sys_role_menu'
);

-- 3. 备份到临时表（如果启用）
IF @table_exists > 0 AND @backup_enabled = 'true' THEN
    CREATE TABLE IF NOT EXISTS sys_role_menu_backup_20260222 AS
    SELECT * FROM sys_role_menu;
END IF;

-- 4. 删除表
DROP TABLE IF EXISTS sys_role_menu;
```

#### 回滚执行步骤

1. **备份当前数据**（可选但推荐）:
   ```bash
   docker exec mysql8 mysqldump -u root -pAdmin168 dataease_dev sys_role_menu > /tmp/sys_role_menu_backup_$(date +%Y%m%d).sql
   ```

2. **执行回滚脚本**:
   ```bash
   docker exec -i mysql8 mysql -u root -pAdmin168 dataease_dev < apps/backend-go/migrations/mysql/20260222_rollback_sys_role_menu.sql
   ```

3. **恢复数据**（如需恢复到回滚前状态）:
   ```bash
   docker exec mysql8 mysql -u root -pAdmin168 dataease_dev -e "
     CREATE TABLE sys_role_menu AS SELECT * FROM sys_role_menu_backup_20260222;
   "
   ```

4. **服务层回退**: 通过配置 `menu.hardcoded_fallback: true` 启用硬编码菜单模式

- 服务: 通过配置开关切回硬编码分支
- 前端: Feature toggle 回退旧菜单渲染

## Open Questions

1. 是否需要支持菜单授权的批量导入/导出？ — 建议 v2 迭代
2. 菜单排序是否需要支持拖拽式交互？ — 取决于前端实现
