# Project Context

## Purpose

DataEase 是开源的 BI（商业智能）工具，帮助用户通过拖拽方式快速制作数据可视化图表，支持多种数据源连接。

**核心目标**：
- 简单易用：零门槛上手，拖拽式操作
- 开源开放：GPL v3 许可，按月迭代
- 全场景支持：多平台安装和嵌入
- 安全分享：多种数据分享方式

## Tech Stack

### 前端
- **框架**: Vue.js 3 + TypeScript
- **构建工具**: Vite 4
- **UI 组件库**: Element Plus
- **图表库**: AntV (G2Plot, L7, S2)、ECharts
- **状态管理**: Pinia
- **路由**: Vue Router 4

### 后端（Go 主线）
- **语言**: Go 1.24+
- **HTTP 框架**: Gin
- **ORM**: GORM
- **缓存**: Redis (go-redis/v9)
- **日志**: Zap
- **配置**: Viper
- **认证**: JWT (golang-jwt/v5)

### 后端（Java 备份 - 只读）
- **框架**: Spring Boot 3.3.0 (Java 21)
- **ORM**: MyBatis Plus 3.5.6
- **SQL 处理**: Apache Calcite 1.35.24

### 基础设施
- **数据库**: MySQL 8.0+
- **缓存**: Redis 7.0+
- **容器**: Docker
- **监控**: Prometheus + OpenTelemetry

## Project Conventions

### Code Style

**通用规则**：
- UTF-8 编码，LF 换行
- 2 空格缩进
- Markdown 以外的文件去除尾部空格
- 保持修改聚焦，bugfix 避免大范围重构

**前端 (Vue 3 + TypeScript)**：
- 使用 `<script setup lang="ts">`
- 使用 Composition API 和 Pinia stores
- ESLint: `plugin:vue/vue3-essential` + `@typescript-eslint/recommended` + `prettier`
- Prettier: 单引号、无分号、无尾逗号、print width 100
- 使用 `@/*` 别名导入 `src/*` 路径
- 类型导入使用 `import type`

**后端 (Go)**：
- 遵循 Go 标准代码规范
- 使用 `golangci-lint` 进行静态检查
- 分层架构：domain → service → repository → transport

### Architecture Patterns

**目录结构**：
```
godataease/
├── apps/                    # 运行时应用
│   ├── backend-go/         # Go 后端（主线）
│   └── frontend/           # Vue 3 前端
├── legacy/                  # 历史备份（只读）
│   ├── backend-java/       # Java 后端备份
│   └── sdk/                # Java SDK 模块
├── infra/                   # 部署与运维
│   ├── compose/            # Docker Compose 配置
│   ├── assets/             # 运维资产
│   └── scripts/            # 部署脚本
├── docs/                    # 文档
└── openspec/               # OpenSpec 规范
```

**Go 后端分层**：
- `internal/domain/` - 领域模型
- `internal/service/` - 业务逻辑
- `internal/repository/` - 数据访问
- `internal/transport/` - HTTP/WebSocket 层

**前端分层**：
- `src/api/` - API 请求函数
- `src/components/` - 通用组件
- `src/views/` - 页面视图
- `src/store/` - Pinia 状态管理
- `src/router/` - 路由配置

### Testing Strategy

**前端**：
- 单元测试：Vitest
- 代码检查：ESLint + Stylelint
- 类型检查：`npm run ts:check`

**后端**：
- 单元测试：`make test` (go test)
- 静态检查：`golangci-lint run`
- API 契约检查：Go Contract Diff Gate (CI)

### Git Workflow

**分支策略**：
- `main` - 主分支，始终保持可部署状态
- 功能开发在 feature 分支进行

**提交规范**：
- 使用语义化提交前缀：
  - `feat:` 新功能
  - `fix:` Bug 修复
  - `ci:` CI 配置变更
  - `docs:` 文档更新
  - `refactor:` 重构
  - `test:` 测试相关
  - `chore:` 其他杂项

**示例**：
```
feat(auth): add two-factor authentication
fix(chart): resolve data rendering issue
ci: add frontend quality gates
```

## Domain Context

DataEase 是 BI 数据可视化平台，核心领域包括：

- **数据源管理**：连接多种数据库和 API
- **数据集管理**：数据建模和转换
- **图表管理**：可视化图表创建和配置
- **仪表板**：图表组合和布局
- **权限管理**：用户、角色、组织权限控制
- **嵌入与分享**：仪表板嵌入和分享功能

## Important Constraints

### 技术约束
- Go 版本：1.24+
- Node.js 版本：18+
- MySQL 版本：8.0+
- Redis 版本：7.0+

### 业务约束
- `legacy/` 目录为只读备份，仅允许安全补丁和应急修复
- 所有大型功能变更必须通过 OpenSpec 流程

### 安全约束
- 敏感配置通过环境变量注入
- 不在代码中硬编码密钥或凭证

## External Dependencies

### 核心服务
- **MySQL**：主数据存储
- **Redis**：缓存和会话存储

### 可选服务
- **Prometheus**：监控指标收集
- **OpenTelemetry**：分布式追踪

### 外部 API
- **SQLBot**：AI 智能问数集成（可选）

## Development Commands

### 前端
```bash
cd apps/frontend
npm install          # 安装依赖
npm run dev          # 开发服务器 (http://localhost:8080)
npm run build:base   # 构建
npm run lint         # ESLint 检查
npm run ts:check     # TypeScript 类型检查
```

### Go 后端
```bash
cd apps/backend-go
make build           # 构建
make run             # 运行 (需要配置数据库)
make run-local       # 本机 MySQL/Redis 运行
make test            # 测试
golangci-lint run    # 静态检查
```

### Docker
```bash
docker network create my-net
docker compose -f infra/compose/docker-compose.yml up -d --build
# 访问: http://localhost:8080
# API 文档: http://localhost:8080/doc.html
```
