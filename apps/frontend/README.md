# DataEase Frontend

DataEase 前端主工程，基于 Vue 3 + TypeScript + Vite，负责仪表板、数据集、数据源、系统管理等核心页面。

## 环境要求

- Node.js 18+

## 本地开发

```bash
cd apps/frontend
npm install
npm run dev
```

默认开发地址：`http://localhost:8080`

## 构建

```bash
# 基础构建（默认）
npm run build:base

# 分布式构建
npm run build:distributed

# 库模式构建
npm run build:lib
```

## 质量检查

```bash
# ESLint
npm run lint

# Stylelint
npm run lint:stylelint

# TypeScript 类型检查
npm run ts:check
```

说明：当前仓库没有统一的 `npm test` 脚本，通常以 `lint + ts:check` 作为前端改动的基础质量门禁。

## 关键目录

- `src/views/`: 页面视图
- `src/components/`: 通用组件
- `src/api/`: API 封装
- `src/store/`: Pinia 状态管理
- `src/router/`: 路由配置
- `src/utils/`: 工具函数

## 开发约定

- 推荐使用 `<script setup lang="ts">`
- 遵循仓库内 ESLint / Prettier / Stylelint 规范
- 优先使用 `@/*` 别名导入 `src/*`
