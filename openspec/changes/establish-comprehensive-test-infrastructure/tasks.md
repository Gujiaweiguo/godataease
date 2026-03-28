# Tasks: Establish Comprehensive Test Infrastructure

## Phase 1: 基础设施验证

- [x] 验证 test-gate.yml 工作流语法正确
- [x] 验证 e2e.yml 支持 workflow_call
- [x] 验证 openspec-archive-change skill 中的测试门禁步骤

## Phase 2: 测试覆盖分析

- [x] 运行 `make test` 生成覆盖率报告
- [x] 运行 `make test-integration` 验证集成测试
- [x] 识别覆盖率缺口（<85% 的模块）
- [x] 列出需要补充测试的关键路径

## Phase 3: 单元测试补全

- [x] 为核心服务补充单元测试
- [x] 为核心仓库补充单元测试
- [x] 验证覆盖率 >= 85%

## Phase 4: E2E 测试补全

- [x] 识别需要 E2E 覆盖的核心用户流程
- [ ] 补充 Playwright 测试用例
- [ ] 验证 E2E 测试在 CI 中通过

## Phase 5: 门禁验证

- [ ] 在当前分支运行 `gh workflow run test-gate.yml`
- [ ] 验证归档流程能自动触发测试
- [ ] 验证测试失败时归档被阻止

## Analysis Notes

- Phase 3 单元测试补全已大幅超额完成。`internal/service` 全量覆盖率从 37.1% 提升至 **93.9%**，远超 85% 目标。
- 所有 `internal/service/*.go` 文件级覆盖率均 ≥ 80%，无低于 80% 的模块。
- 重点模块覆盖：`dataset_service.go` 86.9%、`datasource_service.go` 82.2%、`user_service.go` 93.0%、`chart_service.go` 90.6%、`visualization_service.go` 97.5%、`sync_service.go` 92.4%。
- 新增测试文件：`dataset_service_test.go`（~1100 行）、`datasource_service_test.go`（~930 行）、`user_service_test.go`（扩展审计分支）、`chart_service_test.go`（扩展查询/保存分支）、`visualization_service_test.go`（扩展复制/更新分支）、`sync_service_test.go`（扩展数据源包装器分支）。
- 测试模式：sqlite-backed repo 测试、触发器错误路径、异步审计日志时序修复。
- 已识别的核心 E2E 用户流程包括：登录与重定向、数据源页面导航、系统用户创建/删除、资源权限保存回显、角色/权限配置、图表访问，以及地图/嵌入/交互等现有 Playwright 目录覆盖的管理路径。

## Blockers

- 任务 14 当前被 GitHub Actions 限制阻塞：`gh workflow run test-gate.yml --ref feat/legacy-resource-governance-backfill ...` 返回 `404 workflow ... not found on the default branch`，说明新增 workflow 在合入默认分支前无法通过 `workflow_dispatch` 直接触发。
- 任务 15-16 依赖真实 archive 流程调用与失败拦截验证；在 `test-gate.yml` 进入默认分支、且具备安全的非归档演练路径前，不适合对当前 active change 直接执行归档动作。
