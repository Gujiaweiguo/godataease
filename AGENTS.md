# DataEase Agent Guide

This file is for agentic coding tools working in this repo.
Follow the repo conventions, use the documented commands, and keep changes focused.

## Repository Layout
- `core/core-backend/`: Spring Boot backend (Java 21)
- `core/core-frontend/`: Vue 3 + TypeScript frontend (Vite)
- `sdk/`: SDK modules (Maven)
- `openspec/`: spec-driven change management

## Environment Requirements
- Java 21+
- Maven 3.8+
- Node.js 18+
- MySQL 8.0+
- Redis 7.0+

## Build, Run, and Test Commands

### Frontend (core/core-frontend)
From `core/core-frontend/`:
- Install deps: `npm install`
- Dev server: `npm run dev` (Vite; http://localhost:5173)
- Build (base): `npm run build:base`
- Build (distributed): `npm run build:distributed`
- Build (lib): `npm run build:lib`
- Lint (ESLint): `npm run lint`
- Lint (Stylelint): `npm run lint:stylelint`
- Type check: `npm run ts:check` (vue-tsc --noEmit)

Notes:
- Node memory is bumped in build scripts (see `core/core-frontend/package.json`).
- NPM registry uses `https://registry.npmmirror.com/` (`core/core-frontend/.npmrc`).

### Backend (core/core-backend)
From repo root:
- Build all: `mvn clean install`
- Build without tests: `mvn clean install -DskipTests`

From `core/core-backend/`:
- Run app: `mvn spring-boot:run`
- Run app with profile: `mvn spring-boot:run -Dspring-boot.run.profiles=standalone`
- Package (standalone): `mvn clean package -Pstandalone`
- Package (desktop): `mvn clean package -Pdesktop -U -Dmaven.test.skip=true`
- Package (distributed): `mvn clean package -Pdistributed`
- DB migration: `mvn flyway:migrate`

Tests:
- `mvn test` (JUnit 4)
- Single test class: `mvn test -Dtest=PermissionManageTest`
- Single test method: `mvn test -Dtest=PermissionManageTest#testMethodName`

Important: `core/core-backend/pom.xml` sets Surefire `<skip>true>`.
To run tests, you may need to override with `-DskipTests=false` or adjust the POM locally.

### Docker (optional dev stack)
From repo root:
- `docker-compose up -d`
- Frontend: `http://localhost:8100`
- Backend API: `http://localhost:8100/api`
- API docs: `http://localhost:8100/doc.html`

## Code Style and Conventions

### General
- Use UTF-8, LF line endings, 2-space indentation (`core/core-frontend/.editorconfig`).
- Avoid trailing whitespace; Markdown files may keep it (`.editorconfig`).
- Keep PRs small and focused (`CONTRIBUTING.md`).

### Frontend (Vue 3 + TypeScript)
Config sources:
- ESLint: `core/core-frontend/.eslintrc.js`
- Prettier: `core/core-frontend/prettier.config.js`
- Stylelint: `core/core-frontend/stylelint.config.js`
- TS config: `core/core-frontend/tsconfig.json`

Key ESLint rules:
- `vue/multi-word-component-names`: off
- `@typescript-eslint/no-explicit-any`: off
- `vue/no-setup-props-destructure`: off
- Uses `plugin:vue/vue3-essential`, `@typescript-eslint/recommended`, `prettier`

Prettier defaults:
- 2 spaces, 100 chars/line
- Single quotes
- No semicolons
- No trailing commas
- Arrow parens: avoid

Stylelint highlights:
- Uses `stylelint-config-standard` and `stylelint-order`
- Vue-specific pseudos allowed: `deep`, `global`, `v-deep`, `v-global`, `v-slotted`
- Property ordering enforced (see `stylelint.config.js`)

TypeScript settings:
- `target`/`module`: `esnext`
- `noUnusedLocals`: true
- `noUnusedParameters`: true
- `noImplicitAny`: false
- Path alias: `@/*` -> `./src/*`

Vue conventions:
- `<script setup lang="ts">`
- Use composition API, Pinia stores
- Keep components in `src/components` and views in `src/views`

### Backend (Java + Spring Boot)
Style guide is implied by structure in `development_guide.md`:
- Packages: `io.dataease.<module>`
- Layers: controller / service / mapper / entity / dto
- Controllers use `@RestController` + `@RequestMapping`
- Prefer small, focused services

Testing:
- JUnit 4 dependency in `core/core-backend/pom.xml`
- Test classes live under `core/core-backend/src/test/java/`

## Single Test Examples
- `mvn test -Dtest=io.dataease.dataset.manage.PermissionManageTest`
- `mvn test -Dtest=EmbeddedTokenUtilTest#testTokenGeneration`

## OpenSpec Workflow (Required for major changes)
- Use OpenSpec for new capabilities, breaking changes, architecture shifts
- Read `openspec/AGENTS.md` before starting a proposal or large change

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
