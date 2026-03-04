# 前端系统冒烟清单与 E2E 映射

本文档给出“系统级关键路径”的最小冒烟清单，并映射到当前 Playwright 自动化用例，便于按优先级补齐。

## 1. 覆盖范围定义

系统冒烟的目标是验证以下关键链路可用：

1. 登录
2. 权限拦截（未登录访问受限页）
3. 数据源创建与校验
4. 可视化页面加载（图表/地图/嵌入）

## 2. 现状概览（基于当前仓库）

- E2E 脚本目录：`apps/frontend/e2e`
- 可执行命令：`npm run e2e`
- 当前活跃（非 `fixme`）测试约 8 条（主要是基础与未登录访问）
- 当前 `test.fixme` 约 41 条（主要集中在“登录后业务流”）

说明：`fixme` 数量来自 `apps/frontend/e2e/**/*.spec.ts` 的静态统计。

## 3. 系统冒烟清单（推荐最小集）

| ID | 关键路径 | 验收目标 | 当前自动化映射 | 状态 |
| --- | --- | --- | --- | --- |
| SYS-SMK-001 | 应用可访问 | 访问 `/` 页面可正常渲染 body | `apps/frontend/e2e/smoke.spec.ts` -> `application should load` | 已覆盖 |
| SYS-SMK-002 | 未登录权限拦截 | 未登录访问受限页会被拦截或回到登录上下文 | `apps/frontend/e2e/smoke.spec.ts` -> `should redirect to login when not authenticated` | 已覆盖 |
| SYS-SMK-003 | 登录页与错误凭据 | 登录页元素存在；错误账号有报错/停留登录页 | `apps/frontend/e2e/auth/login.spec.ts` -> `should display login page` + `should show error with invalid credentials` | 已覆盖 |
| SYS-SMK-004 | 登录成功 | 有效账号可登录成功并离开登录页 | `apps/frontend/e2e/auth/login.spec.ts` -> `SYS-SMK-004 @system-smoke should login successfully with valid credentials` | 已启用 |
| SYS-SMK-005 | 数据源列表加载 | 登录后进入数据源页并出现核心元素 | `apps/frontend/e2e/datasource/datasource.spec.ts` -> `SYS-SMK-005 @system-smoke should navigate to datasource list` | 已启用 |
| SYS-SMK-006 | 数据源创建入口 | 能看到“新建数据源”入口并拉起创建弹层 | `apps/frontend/e2e/datasource/datasource.spec.ts` -> `SYS-SMK-006 @system-smoke should display create datasource button` + `should open create datasource dialog` | 部分启用（入口可见性已启用；弹层仍 fixme） |
| SYS-SMK-007 | 数据源类型展示 | 创建弹层可见至少一种数据源类型 | `apps/frontend/e2e/datasource/datasource.spec.ts` -> `should show datasource types in creation dialog` | 待启用（fixme） |
| SYS-SMK-008 | 图表编辑页加载 | 登录后进入 `/chart`，编辑器主框架可见 | `apps/frontend/e2e/chart/chart.spec.ts` -> `should navigate to chart editor` | 待启用（fixme） |
| SYS-SMK-009 | 地图能力入口 | 图表页可见地图类图表入口或 map 相关配置 | `apps/frontend/e2e/map/map.spec.ts` -> `should display map chart type options in chart editor` | 待启用（fixme） |
| SYS-SMK-010 | 嵌入页基本可用 | 关键嵌入页（dataset/datasource/preview）可加载 | `apps/frontend/e2e/embedding/embedding.spec.ts` 对应 `should load ... page` 系列 | 待启用（fixme） |

## 4. 清单到脚本分组映射

| 模块 | 文件 | 主要用途 | 当前建议 |
| --- | --- | --- | --- |
| Smoke | `apps/frontend/e2e/smoke.spec.ts` | 应用可达性、基础未登录访问 | 保持必跑 |
| Auth | `apps/frontend/e2e/auth/login.spec.ts` | 登录基础流程 | 优先启用成功登录用例 |
| Datasource | `apps/frontend/e2e/datasource/datasource.spec.ts` | 数据源管理关键路径 | 第二优先级启用 |
| Chart | `apps/frontend/e2e/chart/chart.spec.ts` | 图表编辑器主流程 | 第三优先级启用 |
| Map | `apps/frontend/e2e/map/map.spec.ts` | 地图图表分支 | 第三优先级启用 |
| Embedding | `apps/frontend/e2e/embedding/embedding.spec.ts` | 嵌入相关路径 | 第四优先级启用 |
| Interactive | `apps/frontend/e2e/interactive/interactive.spec.ts` | 资源树与交互树能力 | 第四优先级启用 |

## 5. 缺口与风险

1. **登录成功链路默认 `fixme`**：导致“鉴权后系统路径”大多无法在 CI 自动验证。
2. **数据源系统路径未纳入活跃冒烟**：当前只覆盖了页面可见性，不覆盖创建-校验主链路。
3. **路由命名存在不一致风险**：部分脚本使用 `/datasource`，部分使用 `/data/datasource`，建议统一到真实生产路由。
4. **账号权限差异会影响冒烟稳定性**：`SYS-SMK-006` 依赖数据源管理权限，无权限账号会按设计跳过。

## 6. 分阶段落地建议（两周）

### 第 1 周（先通链路）

- 启用并稳定：`SYS-SMK-004/005/006`
- 目标：登录后可进入并操作数据源入口

### 第 2 周（扩覆盖）

- 启用并稳定：`SYS-SMK-008/009/010`
- 目标：图表与嵌入关键页面可用

## 7. 建议执行方式

- PR 阶段：保持现有 `quality` 门禁 + 可选 E2E 手动执行
- Nightly 阶段：已接入 `frontend.yml` 的 `system_smoke` job，默认每日自动执行
- 周复盘：记录“通过率、失败原因、修复时长、是否误报”

## 8. CI / Nightly 执行入口

- 手动触发：GitHub Actions -> `Frontend CI` -> `Run workflow`，勾选 `run_system_smoke=true`
- 定时触发：`Frontend CI` 每日按 cron 自动运行 `system_smoke`（非阻断）
- 本地命令：`npm run e2e:system-smoke`

### 必需 Secrets

- `E2E_BASE_URL`：可访问的测试环境地址（例如 `http://localhost:18080`）
- `E2E_PASSWORD`：测试账号密码

### 可选 Secrets（失败告警）

- `FRONTEND_SYSTEM_SMOKE_SLACK_WEBHOOK_URL`
- `FRONTEND_SYSTEM_SMOKE_WECOM_WEBHOOK_URL`

## 9. Workflow Summary 示例

`Frontend CI` 的 `system_smoke` job 会在 `GITHUB_STEP_SUMMARY` 输出结构化表格，示例如下：

| Trigger | Result | Total | Passed | Failed | Skipped | Flaky |
| --- | --- | --- | --- | --- | --- | --- |
| schedule | success | 3 | 2 | 0 | 1 | 0 |

若缺少必需 secrets（`E2E_BASE_URL`、`E2E_PASSWORD`），示例为：

| Trigger | Result | Total | Passed | Failed | Skipped | Flaky |
| --- | --- | --- | --- | --- | --- | --- |
| workflow_dispatch | skipped (missing secrets) | - | - | - | - | - |

---

已支持按标签执行最小系统冒烟集：`npm run e2e:system-smoke`。
