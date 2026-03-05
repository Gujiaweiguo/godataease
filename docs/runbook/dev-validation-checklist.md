# 开发验证清单 Runbook

本文档定义 DataEase 仓库在本地开发、提交前自检、PR 前验收的最小验证标准与推荐命令。

## 适用范围

- 主线后端：`apps/backend-go`
- 前端：`apps/frontend`
- 兼容性门禁：`make drift-check` 与 contract diff 流程

说明：`legacy/backend-java` 与 `legacy/sdk` 为只读备份，不纳入常规开发验证主流程。

## 前置环境

- Node.js `18+`
- Go `1.24+`
- MySQL `8.0+`
- Redis `7.0+`

数据库相关测试必须使用 MySQL 8，禁止 SQLite。

推荐测试库连接参数：

```bash
TEST_DB_HOST=172.19.0.2
TEST_DB_PORT=3306
TEST_DB_USER=root
TEST_DB_PASSWORD=Admin168
TEST_DB_NAME=dataease_test
```

## 分层验证策略

- L0 静态质量层（必须）
  - Frontend: `npm run lint` + `npm run ts:check`
  - Backend: `make lint` + `make test`
- L1 单元测试层（核心逻辑）
  - Frontend: `npm run test -- --run` 或 `npm run test:core`
  - Backend: `make test`
- L2 集成/契约层（接口与数据）
  - Backend: `make test-integration`
  - API 兼容性：`make drift-check`
- L3 端到端冒烟层（关键流程）
  - Frontend: `npm run e2e:system-smoke`（有环境时）

## 按改动类型执行矩阵

### 仅前端改动（最低要求）

```bash
cd apps/frontend
npm run lint
npm run ts:check
```

### 前端业务逻辑改动（store/composable/utils/api）

```bash
cd apps/frontend
npm run lint
npm run ts:check
npm run test -- --run
```

### 前端涉及关键路径（推荐）

```bash
cd apps/frontend
npm run lint
npm run ts:check
npm run test:core
npm run build:base
```

### 仅后端改动（最低要求）

```bash
cd apps/backend-go
make test
make lint
```

### 后端涉及 Repository/SQL/迁移

```bash
cd apps/backend-go
TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test make test-integration
```

### 后端涉及 API 兼容性逻辑

```bash
cd apps/backend-go
make drift-check
```

### 后端提交前全量推荐

```bash
cd apps/backend-go
make check
TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test make test-integration
make drift-check
```

## 前端按路径最小测试集

- 变更 `events/embedding`、`hooks/event`、embedded store/token utils：
  - `npm run test:affected:embedding`
- 变更 `views/visualized/data/datasource`、`api/datasource`、interactive store：
  - `npm run test:affected:datasource`
- 变更共享基础模块（`config/axios`、`store`、`utils`）：
  - `npm run test:core`
- 其他普通前端改动：
  - `npm run test:ci`

## 与 CI 对齐关系

- Frontend CI（`/.github/workflows/frontend.yml`）核心门禁：
  - `ts:check`、`lint`、`build:base`、`test:core`
  - 按路径补充 `test:affected:*`
- Backend CI（`/.github/workflows/go-backend.yml`）核心门禁：
  - `make test`、集成测试、`make drift-check`、linter
- Contract Diff（`/.github/workflows/go-contract-diff-gate.yml`）：
  - 兼容性对比报告与门禁

## 一键复制命令（提交前）

### Frontend

```bash
cd apps/frontend && npm run lint && npm run ts:check && npm run test:core && npm run build:base
```

### Backend

```bash
cd apps/backend-go && make check && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test make test-integration && make drift-check
```

## 常见失败与处理

- `ts:check` 失败：先修类型错误，再跑 `lint`，最后重跑全链路。
- `test-integration` 失败：优先检查 MySQL 8 可达性与 `TEST_DB_*` 参数。
- `drift-check` 失败：先确认接口变更是否预期，再更新兼容策略或修复状态漂移。
- E2E 冒烟失败：先校验环境变量与测试账号，再区分环境问题与业务回归。

## 变更建议

- 若改动涉及跨模块能力或兼容性策略，建议先在 `openspec/` 发起变更提案。
- 保持 PR 小步提交，每次提交保证当前分支可通过最小验证。
