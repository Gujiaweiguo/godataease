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
- Go: 1.21+
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
- Test: `make test`
- Lint: `golangci-lint run`

### Frontend (`apps/frontend`)
Run in `/opt/code/godataease/apps/frontend`:
- Install dependencies: `npm install`
- Dev server: `npm run dev` (Vite, default `http://localhost:5173`)
- Build (base): `npm run build:base`
- Build (distributed): `npm run build:distributed`
- Build (library): `npm run build:lib`
- Lint (ESLint): `npm run lint`
- Lint (Stylelint): `npm run lint:stylelint`
- Type check: `npm run ts:check`
- Preview build: `npm run preview`

Notes:
- Build scripts set `NODE_OPTIONS` memory in `apps/frontend/package.json`.
- NPM registry is configured in `apps/frontend/.npmrc`.
- There is no standard `npm test` script currently; use lint + ts check as quality gates.

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

<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->
