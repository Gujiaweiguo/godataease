# DataEase Agent Guide

This guide is for agentic coding tools working in this repository.
Follow existing project conventions, keep changes minimal, and prefer verifiable commands.

## 沟通偏好
- 默认使用中文回复
- 使用私有仓库地址：`https://github.com/Gujiaweiguo/godataease.git`

## Repository Layout
- `core/core-backend/`: Spring Boot backend (Java 21)
- `core/core-frontend/`: Vue 3 + TypeScript frontend (Vite)
- `sdk/`: SDK modules (Maven multi-module)
- `openspec/`: spec-driven change management

## Environment Requirements
- Java 21+
- Maven 3.8+
- Node.js 18+
- MySQL 8.0+
- Redis 7.0+

## Build, Lint, Test, Run

### Repo Root
Run in `/opt/code/dataease`:
- Build root modules: `mvn clean install`
- Build without tests: `mvn clean install -DskipTests`
- Docker dev stack: `docker compose up -d`
- Docker app URL: `http://localhost:8100`
- Docker API docs: `http://localhost:8100/doc.html`

### Frontend (`core/core-frontend`)
Run in `/opt/code/dataease/core/core-frontend`:
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
- Build scripts set `NODE_OPTIONS` memory in `core/core-frontend/package.json`.
- NPM registry is configured in `core/core-frontend/.npmrc`.
- There is no standard `npm test` script currently; use lint + ts check as quality gates.

### Backend (`core/core-backend`)
Run in `/opt/code/dataease/core/core-backend`:
- Run app: `mvn spring-boot:run`
- Run app (standalone profile): `mvn spring-boot:run -Dspring-boot.run.profiles=standalone`
- Package standalone: `mvn clean package -Pstandalone`
- DB migration: `mvn flyway:migrate`

Testing (JUnit 4):
- Run tests (explicitly enable): `mvn test -DskipTests=false`
- Single test class: `mvn test -Dtest=PermissionManageTest -DskipTests=false`
- Single test method: `mvn test -Dtest=PermissionManageTest#testMethodName -DskipTests=false`
- FQCN test class: `mvn test -Dtest=io.dataease.dataset.manage.PermissionManageTest -DskipTests=false`
- FQCN method example: `mvn test -Dtest=EmbeddedTokenUtilTest#testTokenGeneration -DskipTests=false`

Important:
- `core/core-backend/pom.xml` sets Surefire `<skip>true>` by default.
- If tests appear skipped, always add `-DskipTests=false`.

### SDK (`sdk`)
Run in `/opt/code/dataease/sdk`:
- Build all SDK modules: `mvn clean install`
- Modules include: `common`, `api`, `distributed`, `extensions`

## Code Style and Conventions

### Source of Truth
- Frontend formatting: `core/core-frontend/.editorconfig`, `core/core-frontend/prettier.config.js`
- Frontend lint: `core/core-frontend/.eslintrc.js`, `core/core-frontend/stylelint.config.js`
- Frontend types: `core/core-frontend/tsconfig.json`
- Backend build/test behavior: `pom.xml`, `core/core-backend/pom.xml`
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

### Backend (Java + Spring Boot)
Architecture and package conventions:
- Package prefix: `io.dataease.<module>`.
- Layering: controller / service / mapper / entity / dto.
- Controllers typically use `@RestController` + `@RequestMapping`.

Code style:
- Follow existing class naming: `*Controller`, `*Service`, `*Mapper`, `*DTO`.
- Keep service logic focused; avoid mixing controller concerns into services.
- Prefer existing project response/error patterns over introducing a new global style in small changes.

Testing:
- JUnit 4 is used in backend test modules.
- Tests are under `core/core-backend/src/test/java`.
- For quick validation, run single test class/method first.

Error handling:
- Maintain module-consistent exception behavior.
- Do not swallow exceptions silently.
- If touching existing `try/catch` blocks, preserve response contracts expected by current callers.

## Single-Test Quick Reference
- Backend class: `mvn test -Dtest=PermissionManageTest -DskipTests=false`
- Backend method: `mvn test -Dtest=PermissionManageTest#testMethodName -DskipTests=false`
- Backend FQCN: `mvn test -Dtest=io.dataease.dataset.manage.PermissionManageTest -DskipTests=false`

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
