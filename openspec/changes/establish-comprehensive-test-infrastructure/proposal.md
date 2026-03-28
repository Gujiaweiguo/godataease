# Proposal: Establish Comprehensive Test Infrastructure

## Problem Statement

当前项目缺乏统一的测试基础设施管理，测试覆盖不完整，无法保证归档前的质量门禁。

- 单元测试：部分模块有测试，但覆盖率不统一
- 集成测试：依赖外部数据库，本地运行困难
- E2E 测试：只有 workflow_dispatch，没有自动化触发

## Proposed Solution

建立完整的测试基础设施，包括：

1. **单元测试补充**
   - 确保所有核心模块有测试覆盖
   - 建立覆盖率基线（当前：backend service integration gate 为 84.5%，`make test` 全量单测覆盖率为 28.6%）

2. **集成测试规范化**
   - 使用 test-gate.yml 工作流
   - 确保 CI 和归档前自动运行
   - 明确 84.5% 基线对应 backend service integration coverage gate，而非整体单测覆盖率

3. **E2E 测试自动化**
   - 更新 e2e.yml 支持 workflow_call
   - Playwright 覆盖核心用户流程

## Affected Components

- `.github/workflows/` (CI/CD 配置)
- `apps/backend-go/internal/` (单元/集成测试)
- `apps/frontend/` (E2E 测试)

## Success Criteria

- [ ] CI 中所有测试通过
- [ ] 归档前自动运行 E2E 测试
- [ ] 测试覆盖率 >= 85%
- [ ] 核心用户流程有 E2E 覆盖

## Non-Goals

- 重构现有测试框架
- 建立性能测试（未来 work）

## Implementation Plan

Phase 1: 基础设施
- 验证 test-gate.yml 工作流
- 更新 e2e.yml 支持被调用

Phase 2: 测试补全
- 识别测试覆盖缺口
- 补充关键路径测试

Phase 3: 门禁验证
- 验证 CI 门禁
- 验证归档门禁
