## MODIFIED Requirements

### Requirement: Canonical Build and Infra Entry
系统 SHALL 统一构建与部署入口到迁移后的目录拓扑，禁止旧入口继续作为事实来源，并明确区分本地快速开发模式与生产镜像模式的标准入口。

#### Scenario: Canonical entry enforcement for production and integration
- **WHEN** 开发或运维执行生产镜像构建、集成验证或部署
- **THEN** 系统 SHALL 通过 `infra/` 与新目录入口执行
- **AND** `Dockerfile`、Compose、脚本不再依赖旧目录结构
- **AND** 生产与集成场景 MUST continue to treat the image-based path as the canonical delivery flow

#### Scenario: Sanctioned local host-run development entry
- **WHEN** 开发者执行本地快速迭代开发
- **THEN** 系统 SHALL 允许通过 `apps/backend-go/` 的本地运行入口和 `apps/frontend/` 的开发服务器入口完成应用调试
- **AND** 该本地开发入口 MUST be documented as a sanctioned development workflow rather than an ad hoc workaround
- **AND** 普通代码修改 MUST NOT require rebuilding the production app image to be testable in local development mode

#### Scenario: Development and production contract alignment
- **WHEN** 开发者在本地开发模式与生产镜像模式之间切换
- **THEN** 系统 SHALL 保持约定一致的端口职责、代理语义、依赖服务边界和健康检查说明
- **AND** 进程载体可以不同，但运行约定 MUST remain explicit and stable across both modes
