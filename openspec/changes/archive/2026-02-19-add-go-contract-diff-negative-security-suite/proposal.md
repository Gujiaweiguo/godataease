# Change: 增加 Go Contract Diff 负向安全契约测试套件

## Plan Version
Plan v1（唯一执行计划，执行以本变更 `tasks.md` 为准）

## Why
当前 contract diff 主要覆盖正向功能契约，对权限语义的负向场景覆盖不足。迁移过程中最容易出现高风险回归的是安全边界：`401/403` 语义混淆、行级权限失效、列脱敏缺失、导出下载鉴权绕过。缺少强约束会导致“功能看似可用但权限已退化”的隐性风险。

## What Changes
- 增加负向安全契约用例集，覆盖 `401/403`、行级权限、列脱敏、导出下载鉴权。
- 建立安全契约断言规则（状态码、错误码、错误消息、数据可见性）并接入 contract diff。
- 扩展运行时执行能力，支持多身份上下文与越权请求模拟。
- 将负向安全报告纳入 gate 证据归档，确保权限语义回归可阻断、可追溯。

## Deliverables
- Negative Security Contract Matrix
  - 高风险权限场景清单与期望行为基线
- Runtime Suite Extension
  - contract diff 执行器支持负向安全场景驱动
- Security Assertions
  - 401/403、行级过滤、列脱敏、导出鉴权断言规则
- Gate & Evidence
  - 负向安全报告归档与 CI 阻断策略对齐

## Impact
- Affected specs:
  - `specs/api-compatibility-bridge/spec.md`
- Affected code/config (expected):
  - `backend-go/scripts/contract-diff/*`
  - `backend-go/testdata/contract-diff/*`
  - `.github/workflows/go-contract-diff-gate.yml`（如需增加负向安全 job / 参数）

## Execution Policy
本变更只允许一个执行计划：`openspec/changes/add-go-contract-diff-negative-security-suite/tasks.md`。
不生成、不维护任何独立计划文档。
