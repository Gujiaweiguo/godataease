# Design: Establish Comprehensive Test Infrastructure

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Test Infrastructure                   │
├─────────────────────────────────────────────────────────┤
│  L0: Static Quality                                     │
│  ├── Frontend: lint + ts:check                          │
│  └── Backend: golangci-lint                             │
├─────────────────────────────────────────────────────────┤
│  L1: Unit Tests                                         │
│  ├── Frontend: vitest                                   │
│  └── Backend: go test                                   │
├─────────────────────────────────────────────────────────┤
│  L2: Integration Tests                                  │
│  ├── Backend: go test -tags=integration                 │
│  └── Database: MySQL 8                                  │
├─────────────────────────────────────────────────────────┤
│  L3: E2E Tests                                          │
│  └── Playwright: critical user flows                    │
└─────────────────────────────────────────────────────────┘
```

## Current State

| Component | Unit | Integration | E2E | Coverage |
|-----------|------|-------------|-----|----------|
| Backend | ✅ | ✅ | ❌ | `make test`=28.6%，service integration gate=84.5% |
| Frontend | ✅ | ❌ | ⚠️ | Unknown |

## Decisions

1. **CI 只运行单元+集成**：快速反馈（5分钟内）
2. **归档前运行全部**：包括 E2E，确保质量
3. **E2E 可跳过**：但需要显式确认
4. **覆盖率基线**：service integration coverage gate 为 84.5%，全量单测覆盖率现状需单独治理

## Implementation

Phase 1: 验证现有测试基础设施
- 确认 test-gate.yml 工作流正常
- 确认 e2e.yml 支持 workflow_call
- 验证 openspec-archive-change skill 的门禁逻辑

Phase 2: 补全测试覆盖
- 识别覆盖率缺口
- 补充关键路径测试
- 提高前端测试覆盖率

Phase 3: 门禁验证
- 在当前分支运行全量测试
- 验证归档流程能自动触发测试
- 验证测试失败时归档被阻止
