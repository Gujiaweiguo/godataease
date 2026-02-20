# Plan v1: Contract Diff Runtime Engine（唯一执行计划）

> 执行基线：本文件是 `implement-go-contract-diff-runtime-engine` 的唯一执行计划。任何执行系统仅可依据本文件推进，不允许额外独立计划。

## Dependency Graph

`ENGINE-001 -> ENGINE-002 -> ENGINE-003 -> ENGINE-004 -> ENGINE-005 -> {ENGINE-006, ENGINE-007} -> ENGINE-008`

## Parallel Execution Waves

- **Wave 1（契约冻结）**: ENGINE-001
- **Wave 2（白名单执行入口）**: ENGINE-002
- **Wave 3（请求执行器）**: ENGINE-003
- **Wave 4（结构化差异）**: ENGINE-004
- **Wave 5（分级判定）**: ENGINE-005
- **Wave 6（并行报告产出）**: ENGINE-006 + ENGINE-007
- **Wave 7（CI 联调验收）**: ENGINE-008

## Parallel Batch Mapping

| Task | Wave | Parallel With | Blocks |
|------|------|---------------|--------|
| ENGINE-001 | 1 | None | ENGINE-002 |
| ENGINE-002 | 2 | None | ENGINE-003 |
| ENGINE-003 | 3 | None | ENGINE-004 |
| ENGINE-004 | 4 | None | ENGINE-005 |
| ENGINE-005 | 5 | None | ENGINE-006, ENGINE-007 |
| ENGINE-006 | 6 | ENGINE-007 | ENGINE-008 |
| ENGINE-007 | 6 | ENGINE-006 | ENGINE-008 |
| ENGINE-008 | 7 | None | Final completion |

## Task List

- [x] **ENGINE-001** 输入输出契约冻结
  - **Risk**: Medium
  - **Depends On**: None
  - **Input**:
    - `backend-go/testdata/contract-diff/critical-whitelist.yaml`
    - `backend-go/testdata/contract-diff/output-schema.md`
    - `backend-go/testdata/contract-diff/failure-taxonomy.md`
  - **Output**:
    - Runtime engine 输入参数和输出文件契约清单
    - 与现有 CI workflow 的参数映射表
  - **Acceptance Criteria**:
    - 明确 whitelist 字段最小必需集（`path/method/owner/priority/blockingLevel`）
    - 明确输出 JSON/Markdown 的必需字段与路径
  - **Rollback Plan**:
    - 回退至上一版契约定义并暂停后续实现

- [x] **ENGINE-002** Whitelist 读取与校验实现
  - **Risk**: Medium
  - **Depends On**: ENGINE-001
  - **Input**:
    - 契约清单
    - whitelist 文件
  - **Output**:
    - 可执行的 whitelist 解析逻辑（含字段缺失与非法值校验）
    - 非法配置错误输出（可被 CI 识别）
  - **Acceptance Criteria**:
    - 能稳定读取 critical/high API 列表
    - 字段缺失时返回确定性错误码并终止
  - **Rollback Plan**:
    - 回退到占位读取逻辑并恢复旧参数校验行为

- [x] **ENGINE-003** 并发请求执行器实现
  - **Risk**: High
  - **Depends On**: ENGINE-002
  - **Input**:
    - 解析后的 API 列表
    - Java/Go base URL、timeout、retries
  - **Output**:
    - 并发请求执行器（可配置并发度、超时与重试）
    - 每个接口的 Java/Go 原始响应采集结果
  - **Acceptance Criteria**:
    - 在相同输入下可复现执行结果（顺序可不同，结果一致）
    - 超时与连接错误可区分并正确记入结果集
  - **Rollback Plan**:
    - 切换为串行执行模式，保留请求结果采集能力

- [x] **ENGINE-004** 结构化 diff 核心实现
  - **Risk**: High
  - **Depends On**: ENGINE-003
  - **Input**:
    - Java/Go 响应采集结果
    - output schema 定义
  - **Output**:
    - `status/code/msg/payload schema/payload value` 多维 diff 结果
    - 统一字段路径表示（便于问题定位）
  - **Acceptance Criteria**:
    - 每条 API 结果包含完整 diff 子结构
    - payload schema diff 与 payload value diff 可被区分
  - **Rollback Plan**:
    - 回退为仅 status/code/msg 三维对比

- [x] **ENGINE-005** 稳定重试与失败分级输出
  - **Risk**: High
  - **Depends On**: ENGINE-004
  - **Input**:
    - diff 结果
    - failure taxonomy 与 gate thresholds
  - **Output**:
    - 失败分类（`STATUS_DIFF/CODE_DIFF/MSG_DIFF/...`）
    - 分级判定（critical/high/normal）与 gate pass/fail 结论
  - **Acceptance Criteria**:
    - `critical` 差异出现时 gate 必失败
    - 重试耗尽后错误分类稳定且可追踪
  - **Rollback Plan**:
    - 回退到分类但不阻断模式（仅输出告警）

- [x] **ENGINE-006** JSON 报告落盘与一致性校验
  - **Risk**: Medium
  - **Depends On**: ENGINE-005
  - **Input**:
    - 结构化 diff 结果
    - output schema
  - **Output**:
    - `contract-diff-report.json`
    - schema 一致性自检结果
  - **Acceptance Criteria**:
    - JSON 报告字段与 schema 文档一致
    - summary 统计（total/passed/failed/parity）准确
  - **Rollback Plan**:
    - 降级输出最小 JSON 并保留核心统计

- [x] **ENGINE-007** Markdown 报告渲染与故障证据
  - **Risk**: Medium
  - **Depends On**: ENGINE-005
  - **Input**:
    - 分类后的 diff 结果
  - **Output**:
    - `contract-diff-report.md`
    - Top failures 摘要与接口级证据索引
  - **Acceptance Criteria**:
    - 报告可快速定位失败接口与差异字段
    - 与 JSON 报告在统计与结论上保持一致
  - **Rollback Plan**:
    - 降级为摘要型 Markdown（仍保留失败列表）

- [x] **ENGINE-008** CI Gate 联调与验收
  - **Risk**: Medium
  - **Depends On**: ENGINE-006, ENGINE-007
  - **Input**:
    - 真实 runtime engine
    - CI workflow（`go-contract-diff-gate.yml`）
    - pass/fail 样本
  - **Output**:
    - CI 运行记录（至少一组通过样本 + 一组失败样本）
    - 退出码与 artifact 归档一致性结论
  - **Acceptance Criteria**:
    - pass 样本返回 0，fail 样本返回非 0
    - CI artifact 成功归档 JSON + Markdown
    - `openspec validate implement-go-contract-diff-runtime-engine --strict --no-interactive` 通过
  - **Rollback Plan**:
    - 回退 workflow 到 dry-run 非阻断模式并保留报告产物

## Suggested Acceptance Commands

> 说明：以下为执行批次的建议验收命令模板。命令名可按仓库实际脚本微调，但输入输出语义不可变更。

### Wave 1-2（ENGINE-001/002）

```bash
bash backend-go/scripts/contract-diff/run_contract_diff.sh --help
bash backend-go/scripts/contract-diff/run_contract_diff.sh --whitelist backend-go/testdata/contract-diff/critical-whitelist.yaml --java-base http://localhost:8100 --go-base http://localhost:8080 --out-dir backend-go/tmp/contract-diff --timeout 10 --retries 1
```

### Wave 3（ENGINE-003）

```bash
bash backend-go/scripts/contract-diff/run_contract_diff.sh --whitelist backend-go/testdata/contract-diff/critical-whitelist.yaml --java-base http://localhost:8100 --go-base http://localhost:8080 --out-dir backend-go/tmp/contract-diff --timeout 5 --retries 2
```

### Wave 4-5（ENGINE-004/005）

```bash
test -f backend-go/tmp/contract-diff/contract-diff-report.json
jq '.results[0].diff | keys' backend-go/tmp/contract-diff/contract-diff-report.json
jq '.summary' backend-go/tmp/contract-diff/contract-diff-report.json
```

### Wave 6（ENGINE-006/007）

```bash
test -f backend-go/tmp/contract-diff/contract-diff-report.json
test -f backend-go/tmp/contract-diff/contract-diff-report.md
jq '.metadata,.summary' backend-go/tmp/contract-diff/contract-diff-report.json
```

### Wave 7（ENGINE-008）

```bash
openspec validate implement-go-contract-diff-runtime-engine --strict --no-interactive
```

## Definition of Done（DoD）Checklist

> 规则：任务从 `- [ ]` 改为 `- [x]` 前，必须满足对应 DoD 清单。

### ENGINE-001 DoD
- [x] 输入参数契约文档已冻结并可追溯版本
- [x] 输出文件契约（JSON/Markdown 路径与字段）已冻结
- [x] 与 CI workflow 参数映射完成且无冲突

### ENGINE-002 DoD
- [x] whitelist 解析支持 `path/method/owner/priority/blockingLevel`
- [x] 字段缺失/非法值返回确定性错误码
- [x] critical/high API 清单提取结果稳定可复现

### ENGINE-003 DoD
- [x] 并发度、timeout、retries 参数生效
- [x] Java/Go 请求结果完整采集并可追踪到接口
- [x] timeout 与 connection error 分类可区分

### ENGINE-004 DoD
- [x] 每个接口结果都包含 `status/code/msg/payload schema/payload value` diff
- [x] payload schema diff 与 payload value diff 已拆分
- [x] 差异路径表示统一且可定位到字段

### ENGINE-005 DoD
- [x] 失败类别映射到 taxonomy（如 `STATUS_DIFF/CODE_DIFF/...`）
- [x] `critical` 差异可稳定触发 gate fail
- [x] 重试耗尽后的最终状态与失败原因确定且可追踪

### ENGINE-006 DoD
- [x] `contract-diff-report.json` 成功落盘
- [x] JSON 字段与 `output-schema.md` 一致
- [x] summary 统计（total/passed/failed/parity）准确

### ENGINE-007 DoD
- [x] `contract-diff-report.md` 成功落盘
- [x] 报告含 Top failures 与接口级差异证据索引
- [x] Markdown 与 JSON 的统计和结论一致

### ENGINE-008 DoD
- [x] pass/fail 样本均完成 CI 联调验证
- [x] CI artifact 成功归档 JSON + Markdown
- [x] `openspec validate implement-go-contract-diff-runtime-engine --strict --no-interactive` 通过

## Sign-off Matrix

| Area | Owner Role | Required Sign-off |
|------|------------|-------------------|
| Runtime Engine 参数与执行语义 | Tech Lead / Backend Owner | Yes |
| Whitelist 读取与字段治理 | API/Module Owner | Yes |
| 失败分级与阈值映射 | Architect / Tech Lead | Yes |
| CI Gate 与 Artifact 归档 | Repo Maintainer | Yes |
| 报告可读性与验收口径 | QA/Release Owner | Yes |

## Go/No-Go Gates（评审阻断条件）

> 任一 `No-Go` 命中即不得进入执行或合并；仅当全部 `Go` 条件满足时可推进。

### Go（可推进）
- [x] ENGINE-001 ~ ENGINE-008 对应 DoD 全部满足
- [x] Sign-off Matrix 所有 `Required Sign-off = Yes` 已完成签收
- [x] CI pass/fail 样本均验证通过，且退出码符合预期（pass=0 / fail!=0）
- [x] JSON/Markdown 报告已归档并可按 commit/PR 追溯
- [x] `openspec validate implement-go-contract-diff-runtime-engine --strict --no-interactive` 通过

### No-Go（阻断推进）
- [x] 任何 `critical` 失败未触发 gate fail
- [x] whitelist 必需字段缺失仍可执行（校验失效）
- [x] 输出字段与 `output-schema.md` 不一致或统计不准确
- [x] CI workflow 与脚本参数不一致导致误判
- [x] 必要签收角色缺失或验收证据不完整

## Execution Notes

- ENGINE-003/004/005 为高风险关键路径，任一失败阻断后续任务。
- 运行时引擎实现必须与现有 whitelist/threshold/report 文档保持兼容。
- 如需修改输出字段，必须同步更新 schema 与 CI 消费侧。
