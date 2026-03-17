## 0. 前置检查与盘点

- [x] 0.1 确认后端构建命令：`make build` → `apps/backend-go/dataease-backend`
- [x] 0.2 确认后端本地运行：`make run-local` 支持 localhost MySQL/Redis
- [x] 0.3 确认前端构建命令：`npm run build:base` → `apps/frontend/dist/`
- [x] 0.4 确认前端开发服务器：`npm run dev` → Vite HMR on :8080
- [x] 0.5 确认生产 compose 可用：`docker compose -f infra/compose/docker-compose.yml up -d`
- [x] 0.6 确认生产镜像健康检查：`/health` 返回 OK

## 1. Development workflow entry definition

- [x] 1.1 梳理并固定本地开发模式下依赖服务、前端 dev server、后端 run-local 的标准入口
  - 后端: `cd apps/backend-go && make run-local`
  - 前端: `cd apps/frontend && npm run dev`
  - 依赖: `docker compose -f infra/compose/docker-compose.yml up -d godataease-redis`
- [x] 1.2 为开发模式补充清晰的启动说明或脚本入口，避免默认使用 `docker compose up --build` 进行日常迭代
  - [x] 1.2.1 创建 `scripts/dev.sh` 或 Makefile `make dev` 聚合入口
  - [x] 1.2.2 在 AGENTS.md 中明确推荐开发模式入口
- [x] 1.3 明确开发模式与生产模式的端口、代理、健康检查与职责边界文档
  - [x] 1.3.1 端口约定已明确（均为 :8080）
  - [x] 1.3.2 补充开发模式代理配置说明（Vite proxy）
  - [x] 1.3.3 补充健康检查语义说明

## 2. Local fast-feedback implementation

- [x] 2.1 调整或补充开发辅助脚本/Compose override，使本地依赖服务可以独立启动
  - [x] 2.1.1 创建 `infra/compose/docker-compose.dev.yml` 或 override
  - [x] 2.1.2 挂载 `apps/frontend/dist/` 到容器前端静态资源路径
  - [x] 2.1.3 挂载 `apps/backend-go/dataease-backend` 到容器后端二进制路径
  - [x] 2.1.4 验证挂载后容器能正常启动并访问
  - [x] 2.1.5 添加 `make build-static` target 生成静态二进制（Alpine 兼容）
- [x] 2.2 统一前端开发模式的 API 代理与接口前缀约定，确保与本地后端运行模式兼容
  - [x] 2.2.1 检查 Vite proxy 配置是否指向 localhost:8088（与 config.yaml 一致）
  - [x] 2.2.2 确认 API 前缀与后端路由一致
- [x] 2.3 统一后端本地运行模式的配置加载方式，确保可直接复用本地 MySQL/Redis 依赖
  - 已通过 `make run-local` 环境变量覆盖实现
- [x] 2.4 补充最小环境检查或说明，降低 Node/Go 本地环境不一致带来的误用风险
  - [x] 2.4.1 在 AGENTS.md 或 README 中明确 Node 18+、Go 1.24+ 要求
  - [x] 2.4.2 可选：提供 `scripts/check-env.sh` 检查脚本

## 3. Production path preservation

- [x] 3.1 验证当前生产镜像构建路径在新工作流约定下仍然可作为 canonical delivery flow
  - `docker compose -f infra/compose/docker-compose.yml up -d --build` 可用
- [x] 3.2 明确文档中何时应使用镜像重建路径，何时应使用宿主机快速开发路径
  - [x] 3.2.1 在 README 中区分"开发模式"与"生产/集成模式"
  - [x] 3.2.2 说明 `--build` 仅用于生产/集成验证

## 4. Verification and adoption

- [x] 4.1 验证前端 `npm run dev` 与后端 `make run-local` 能在依赖服务已启动时协同工作
  - [x] 4.1.1 启动 Redis 容器
  - [x] 4.1.2 启动后端 `make run-local`
  - [x] 4.1.3 启动前端 `npm run dev`
  - [x] 4.1.4 验证登录页面可访问、API 调用正常
- [x] 4.2 验证本地代码修改不再需要重建 `godataease-app` 镜像即可完成普通开发调试
  - [x] 4.2.1 修改前端代码 → `npm run build:base` → 刷新页面验证
  - [x] 4.2.2 修改后端代码 → `make build` → 重启后端验证
- [x] 4.3 验证生产镜像路径仍可正常构建并通过健康检查
  - 已验证 `/health` 返回 `{"service":"dataease-backend","status":"ok"}`
- [x] 4.4 更新相关开发文档，并给出推荐工作流与常见误区说明
  - [x] 4.4.1 更新 README 开发模式章节
  - [x] 4.4.2 更新 AGENTS.md 开发命令说明
  - [x] 4.4.3 补充常见误区（如误用 `--build`）

## 进度统计

- 总任务: 37
- 已完成: 37
- 进行中: 0
- 待办: 0
