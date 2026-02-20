# Plan v1: Negative Security Contract Suite（唯一执行计划）

> 执行基线：本文件是 `add-go-contract-diff-negative-security-suite` 的唯一执行计划。任何执行系统仅可依据本文件推进，不允许额外独立计划。

## Dependency Graph

`NEGSEC-001 -> NEGSEC-002 -> NEGSEC-003 -> NEGSEC-004 -> {NEGSEC-005, NEGSEC-006} -> NEGSEC-007`

## Parallel Execution Waves

- **Wave 1（风险场景冻结）**: NEGSEC-001
- **Wave 2（用例与基线样本）**: NEGSEC-002
- **Wave 3（执行器扩展）**: NEGSEC-003
- **Wave 4（权限语义断言）**: NEGSEC-004
- **Wave 5（并行强化）**: NEGSEC-005 + NEGSEC-006
- **Wave 6（CI 联调验收）**: NEGSEC-007

## Task List

- [x] **NEGSEC-001** 负向安全契约范围冻结
  - **Risk**: High
  - **Depends On**: None
  - **Input**:
    - 现有 whitelist 与 failure taxonomy
    - 迁移高风险权限链路清单
  - **Output**:
    - 负向安全场景矩阵（`401/403`、行级、列脱敏、导出鉴权）
    - 场景优先级与阻断级别定义
  - **Acceptance Criteria**:
    - 每类权限风险至少 1 个阻断级场景
    - 场景期望结果可机读（状态码/错误码/数据可见性）
  - **Rollback Plan**:
    - 回退到最小安全场景集，仅保留阻断级用例

- [x] **NEGSEC-002** 安全基线样本与夹具落地
  - **Risk**: Medium
  - **Depends On**: NEGSEC-001
  - **Input**:
    - 负向安全场景矩阵
    - baseline fixture 规范
  - **Output**:
    - 按接口与身份上下文组织的负向样本夹具
    - 场景与夹具映射表（便于执行器消费）
  - **Acceptance Criteria**:
    - 所有阻断级场景可绑定到明确 fixture
    - fixture 可追溯到接口、身份、预期拒绝语义
  - **Rollback Plan**:
    - 回退至仅状态码断言夹具

- [x] **NEGSEC-003** 运行时执行器支持多身份与越权模拟
  - **Risk**: High
  - **Depends On**: NEGSEC-002
  - **Input**:
    - security fixture
    - runtime engine 执行入口
  - **Output**:
    - 执行器支持匿名/低权限/高权限身份上下文
    - 支持跨租户、跨组织、越权下载等请求构造
  - **Acceptance Criteria**:
    - 同一场景可在 Java/Go 两侧复现一致身份上下文
    - 身份注入失败时返回确定性错误并终止该场景
  - **Rollback Plan**:
    - 降级为单身份对比并保留场景占位

- [x] **NEGSEC-004** 401/403 与拒绝语义断言实现
  - **Risk**: High
  - **Depends On**: NEGSEC-003
  - **Input**:
    - 多身份执行结果
    - 错误码/消息规范
  - **Output**:
    - `401` 未认证与 `403` 无权限断言规则
    - 拒绝语义差异分类（状态码、错误码、错误消息）
  - **Acceptance Criteria**:
    - `401/403` 语义错位可稳定识别并标记为阻断
    - Java/Go 拒绝行为差异可追溯到具体接口与身份
  - **Rollback Plan**:
    - 回退为仅状态码断言，保留错误码/消息记录

- [x] **NEGSEC-005** 行级权限与列脱敏一致性断言
  - **Risk**: High
  - **Depends On**: NEGSEC-004
  - **Input**:
    - 数据查询类接口执行结果
    - 行级权限与脱敏规则
  - **Output**:
    - 行级过滤断言（应隐藏的数据不可见）
    - 列脱敏断言（敏感列脱敏策略一致）
  - **Acceptance Criteria**:
    - 任一越权可见记录或未脱敏字段触发阻断
    - 差异报告可定位到具体记录/字段
  - **Rollback Plan**:
    - 降级为字段存在性断言并保留明细日志

- [x] **NEGSEC-006** 导出下载鉴权负向套件
  - **Risk**: High
  - **Depends On**: NEGSEC-004
  - **Input**:
    - 导出任务与下载接口
    - token/session 场景样本
  - **Output**:
    - 失效 token、越权下载、跨用户文件访问断言
    - 导出链路负向报告（含阻断级结论）
  - **Acceptance Criteria**:
    - 未授权下载成功必须判为阻断级失败
    - 失效凭证行为在 Java/Go 两侧语义一致
  - **Rollback Plan**:
    - 暂退为只验证 token 失效场景

- [x] **NEGSEC-007** CI Gate 接入与验收
  - **Risk**: Medium
  - **Depends On**: NEGSEC-005, NEGSEC-006
  - **Input**:
    - 负向安全执行结果
    - gate 与报告归档策略
  - **Output**:
    - 负向安全 gate 执行记录（通过样本 + 阻断样本）
    - 归档报告（JSON/Markdown）与阻断结论一致性验证
  - **Acceptance Criteria**:
    - 阻断级权限语义回归可稳定拦截
    - 归档证据可追溯到场景、身份、接口
    - `openspec validate add-go-contract-diff-negative-security-suite --strict --no-interactive` 通过
  - **Rollback Plan**:
    - 暂停负向阻断，仅保留观察模式报告

## Execution Notes

- NEGSEC-003/004/005/006 是核心风险路径，任一失败需先修复后再推进。
- 负向安全断言与现有功能性 contract diff 结果必须分层展示，避免互相覆盖。
- 若安全语义基线更新，必须同步更新场景矩阵与夹具映射。
