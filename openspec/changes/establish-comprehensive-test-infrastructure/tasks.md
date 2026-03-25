# Tasks: Establish Comprehensive Test Infrastructure

## Phase 1: 基础设施验证

- [ ] 验证 test-gate.yml 工作流语法正确
- [ ] 验证 e2e.yml 支持 workflow_call
- [ ] 验证 openspec-archive-change skill 中的测试门禁步骤

## Phase 2: 测试覆盖分析

- [ ] 运行 `make test` 生成覆盖率报告
- [ ] 运行 `make test-integration` 验证集成测试
- [ ] 识别覆盖率缺口（<85% 的模块）
- [ ] 列出需要补充测试的关键路径

## Phase 3: 单元测试补全

- [ ] 为核心服务补充单元测试
- [ ] 为核心仓库补充单元测试
- [ ] 验证覆盖率 >= 85%

## Phase 4: E2E 测试补全

- [ ] 识别需要 E2E 覆盖的核心用户流程
- [ ] 补充 Playwright 测试用例
- [ ] 验证 E2E 测试在 CI 中通过

## Phase 5: 门禁验证

- [ ] 在当前分支运行 `gh workflow run test-gate.yml`
- [ ] 验证归档流程能自动触发测试
- [ ] 验证测试失败时归档被阻止
