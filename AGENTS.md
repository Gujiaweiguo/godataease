# DataEase Agent Guide

This guide is for agentic coding tools working in this repository.
Follow existing project conventions, keep changes minimal, and prefer verifiable commands.

## 沟通偏好
- 默认使用中文回复
- 使用私有仓库地址：`https://github.com/Gujiaweiguo/godataease.git`

## Repository Layout

```
godataease/
├── apps/                    # 运行时应用
│   ├── backend-go/         # Go 后端（主线）
│   └── frontend/           # Vue 3 前端
├── legacy/                  # 历史备份（只读）
│   ├── backend-java/       # Java 后端备份
│   └── sdk/                # Java SDK 模块
├── infra/                   # 部署与运维
│   ├── assets/             # 运维资产（地图数据等）
│   ├── compose/            # Docker Compose 配置
│   └── scripts/            # 部署脚本
├── docs/                    # 文档
└── openspec/               # OpenSpec 规范
```

## Environment Requirements
- Go: 1.24+
- Node.js: 18+
- MySQL: 8.0+
- Redis: 7.0+

## Build, Lint, Test, Run

### Repo Root
Run in `/opt/code/godataease`:
- Validate repo aggregator: `mvn -N validate`
- Docker dev stack: `docker compose -f infra/compose/docker-compose.yml up -d`
- Docker app URL: `http://localhost:8080`
- Docker API docs: `http://localhost:8080/doc.html`
- Legacy Java emergency operations: see `legacy/README-READONLY.md`

### Go Backend (`apps/backend-go`)
Run in `/opt/code/godataease/apps/backend-go`:
- Build: `make build`
- Run: `make run`
- Run (local DB/Redis): `make run-local`
- Test: `make test`
- Lint: `golangci-lint run`

### Frontend (`apps/frontend`)
Run in `/opt/code/godataease/apps/frontend`:
- Install dependencies: `npm install`
- Dev server: `npm run dev` (Vite, default `http://localhost:8080`)
- Build (base): `npm run build:base`
- Build (distributed): `npm run build:distributed`
- Build (library): `npm run build:lib`
- Lint (ESLint): `npm run lint`
- Lint (Stylelint): `npm run lint:stylelint`
- Type check: `npm run ts:check`
- Unit test (Vitest): `npm run test -- --run`
- Test (CI smoke): `npm run test:ci`
- Test (core suite): `npm run test:core`
- Coverage: `npm run test:coverage`
- Coverage (core suite): `npm run test:coverage:core`
- Preview build: `npm run preview`

Notes:
- Build scripts set `NODE_OPTIONS` memory in `apps/frontend/package.json`.
- NPM registry is configured in `apps/frontend/.npmrc`.
- Frontend has Vitest configured; current `test:ci` is a lightweight smoke entry and does not represent full feature coverage.

## 测试策略（AI 执行约定）

### 分层测试策略（推荐）
- L0 静态质量层（必须）：Frontend `lint + ts:check`；Backend `golangci-lint + go test`。
- L1 单元测试层（核心逻辑）：优先覆盖纯函数、store/composable、service/domain 逻辑。
- L2 集成/契约层（接口与数据）：Backend 仓储/HTTP 集成测试、契约差异检查（contract diff / drift-check）。
- L3 端到端冒烟层（关键流程）：登录、权限、数据源创建/编辑/校验等关键用户路径。

### AI 最低验证要求（提交前）
- 修改 Frontend 代码至少执行：`npm run lint`、`npm run ts:check`。
- 若修改 Frontend 业务逻辑（store/composable/utils/api 处理），应补充或执行对应 Vitest 用例（`npm run test -- --run`）。
- 修改 Backend 代码至少执行：`make test`。
- 若修改 Repository/SQL/迁移/持久化逻辑，增加执行 `make test-integration`（环境不满足时需在结论中明确说明未执行原因与风险）。
- 若修改 API 兼容相关逻辑，增加执行 `make drift-check`（必要时结合 contract diff workflow）。

### 按改动路径触发测试矩阵（Frontend）
- 变更 `events/embedding`、`hooks/event`、`embedded store/token utils`：执行 `npm run test:affected:embedding`。
- 变更 `views/visualized/data/datasource`、`api/datasource`、`interactive store`：执行 `npm run test:affected:datasource`（数据源专属最小测试集）。
- 变更 `config/axios`、`store`、`utils` 等共享基础模块：执行 `npm run test:core`。
- 其他普通前端变更：至少执行 `npm run test:ci`（smoke）。

### CI 与门禁建议（当前仓库适配）
- Frontend `ts:check` 建议尽快改为阻断（当前工作流为 non-blocking）。
- Frontend `test:ci` 建议从示例测试扩展为“变更影响范围测试”或“核心模块测试集合”。
- Backend 保持 `make test` 阻断；`integration` 标签测试建议加入夜间定时任务或手动门禁流程。

### Legacy (Read Only)
- `legacy/backend-java/` 与 `legacy/sdk/` 为只读备份，不承接常规功能开发。
- 仅允许安全补丁、应急修复和迁移对照改动。
- 详细命令与审批规则见 `legacy/README-READONLY.md`。

## Code Style and Conventions

### Source of Truth
- Frontend formatting: `apps/frontend/.editorconfig`, `apps/frontend/prettier.config.js`
- Frontend lint: `apps/frontend/.eslintrc.js`, `apps/frontend/stylelint.config.js`
- Frontend types: `apps/frontend/tsconfig.json`
- Backend build/test behavior: `legacy/pom.xml`, `legacy/backend-java/pom.xml`
- Project development conventions: `development_guide.md`, `CONTRIBUTING.md`

### General Formatting
- UTF-8, LF, 2-space indentation.
- Trim trailing spaces except Markdown.
- Keep edits focused; avoid broad refactors in bugfixes.

### Frontend (Vue 3 + TypeScript)
Framework and structure:
- Prefer `<script setup lang="ts">`.
- Use Composition API and Pinia stores.
- Views under `src/views`, reusable components under `src/components`.

Lint and format rules:
- ESLint extends `plugin:vue/vue3-essential` + `@typescript-eslint/recommended` + `prettier`.
- Prettier: single quotes, no semicolons, no trailing commas, print width 100.
- Stylelint enforces property order and supports Vue-specific pseudo selectors.

Types and imports:
- Respect `noUnusedLocals` and `noUnusedParameters`.
- `noImplicitAny` is `false`; avoid adding unnecessary `any` in new code.
- Use alias imports via `@/*` for `src/*` paths.
- Keep import blocks readable: third-party imports first, then alias imports, then relative imports.
- Use `import type` for type-only imports where applicable.

Naming:
- Components/files generally use PascalCase or project-existing naming in same folder.
- Composables use `useXxx` naming.
- Store accessors follow existing patterns like `useXxxStoreWithOut`.

Error handling:
- API calls in UI logic should use `try/catch` and user-facing message feedback.
- Keep response code checks consistent with existing patterns (`code === '000000'` where used).

## Legacy Java Note
- Java 代码规范与应急命令统一维护在 `legacy/README-READONLY.md`，避免主线文档混入双栈细节。

## Cursor / Copilot Rules
- No `.cursorrules` found in this repository.
- No `.cursor/rules/` directory found in this repository.
- No `.github/copilot-instructions.md` found in this repository.

## OpenSpec Workflow (Required for major changes)
- Use OpenSpec for new capabilities, breaking changes, and architecture shifts.
- Read `openspec/AGENTS.md` before starting proposals or large changes.
