## Why

当前仓库的本地开发体验主要依赖 `godataease-app` 容器承载前后端运行结果，但该容器使用的是发布型镜像工作流：前端需要先构建 `dist`，后端需要先编译二进制，再重建镜像后才能看到代码变更。这使得日常开发反馈周期过长，也让“开发模式”和“生产模式”之间出现了不必要的操作割裂。为了兼顾开发效率与环境统一性，需要一套不必频繁重建镜像、但仍保持应用运行在容器中的标准开发模式。

## What Changes

- 为仓库定义一套明确的本地开发工作流，使前端构建产物与后端二进制可以通过文件挂载方式进入 `godataease-app` 容器，不再默认依赖重建镜像才能生效。
- 明确开发模式与生产模式的职责边界：两者都保持应用运行在容器中，但开发模式允许通过宿主机构建产物 + 容器挂载实现快速迭代，生产模式继续保持当前镜像化交付路径。
- 统一两种模式的运行约定，尽量复用相同的端口、入口、健康检查语义和 app 容器职责，减少“开发和生产是两套完全不同运行环境”的认知负担。
- 补充必要的开发入口、脚本和文档支持，降低团队成员选择正确本地工作流的成本。

## Capabilities

### New Capabilities
- `local-dev-workflow`: 定义 DataEase 前端 `dist` 与后端二进制通过文件挂载进入 app 容器的本地开发模式约定。

### Modified Capabilities
- `backend-go-architecture`: 增补 Go 主线在开发模式与生产镜像模式下共享 app 容器运行约束的一致性要求。

## Impact

- Affected code: `infra/compose/`, `apps/backend-go/Makefile`, `apps/frontend/package.json`, 相关开发脚本、Docker 运行时配置与文档。
- Affected systems: 本地 Docker app 容器、宿主机构建产物输出路径、生产镜像构建路径。
- Dependencies: Docker Compose、Node/Vite 构建链路、Go 本地构建链路、现有 MySQL/Redis 开发依赖。
