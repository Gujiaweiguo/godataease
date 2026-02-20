# Plan v1: Java/Go Contract Diff CI Gate（唯一执行计划）

> 执行基线：本文件是 `add-go-contract-diff-ci-gate` 的唯一执行计划。任何执行系统仅可依据本文件推进，不允许额外独立计划。

## Dependency Graph

`CI-GATE-001 -> CI-GATE-002 -> CI-GATE-003 -> {CI-GATE-004, CI-GATE-005} -> CI-GATE-006`

## Parallel Execution Waves

- **Wave 1（基础定义）**: CI-GATE-001
- **Wave 2（执行器落地）**: CI-GATE-002
- **Wave 3（CI 接入）**: CI-GATE-003
- **Wave 4（并行门禁能力）**: CI-GATE-004 + CI-GATE-005
- **Wave 5（联调验收）**: CI-GATE-006

## Task List

- [x] **CI-GATE-001** 关键接口白名单冻结
  - **Risk**: Medium
  - **Depends On**: None
  - **Parallel Group**: Wave 1
  - **Input**:
    - Java/Go 兼容矩阵与关键链路清单
    - 现有兼容性规格：`openspec/specs/api-compatibility-bridge/spec.md`
  - **Output**:
    - 白名单文件（建议位置）：`backend-go/testdata/contract-diff/critical-whitelist.yaml`
    - 白名单治理规则（建议位置）：`backend-go/testdata/contract-diff/whitelist-governance.md`
  - **Acceptance Criteria**:
    - 白名单覆盖所有 P0/P1 关键接口
    - 每个接口包含：`path`、`method`、`owner`、`priority`、`blockingLevel`
    - 白名单变更规则明确评审责任人和生效流程
  - **Rollback Plan**:
    - 回退到上一个已批准白名单版本并重新触发 gate

- [x] **CI-GATE-002** Contract diff 执行器标准化
  - **Risk**: Medium
  - **Depends On**: CI-GATE-001
  - **Parallel Group**: Wave 2
  - **Input**:
    - 白名单文件
    - Java/Go 对比执行入口与测试环境参数
  - **Output**:
    - 执行脚本（建议位置）：`backend-go/scripts/contract-diff/run_contract_diff.sh`
    - 标准输出格式说明（建议位置）：`backend-go/testdata/contract-diff/output-schema.md`
  - **Acceptance Criteria**:
    - 支持白名单过滤、重试、超时控制
    - 同一输入下重复执行结果一致
    - 输出包含：`endpoint`、`status`、`codeDiff`、`msgDiff`、`payloadDiff`
  - **Rollback Plan**:
    - 保留旧执行入口并允许一键切换回旧脚本

- [x] **CI-GATE-003** CI Job 接入与触发策略落地
  - **Risk**: High
  - **Depends On**: CI-GATE-002
  - **Parallel Group**: Wave 3
  - **Input**:
    - 现有 Go CI workflow
    - 标准化 contract diff 执行脚本
  - **Output**:
    - CI workflow（建议新增）：`.github/workflows/go-contract-diff-gate.yml`
    - 触发策略：PR + 受保护分支；可配置并发与重试
  - **Acceptance Criteria**:
    - PR 流程中 contract diff job 可见且默认启用
    - 受保护分支合并后自动复跑并保留结果
    - gate 失败会阻断合并（branch protection）
  - **Rollback Plan**:
    - 通过 workflow 开关降级为仅告警模式

- [x] **CI-GATE-004** 失败阈值与阻断规则实施
  - **Risk**: High
  - **Depends On**: CI-GATE-003
  - **Parallel Group**: Wave 4（与 CI-GATE-005 并行）
  - **Input**:
    - 标准 diff 输出
    - 白名单阻断等级配置
  - **Output**:
    - 阈值配置（建议位置）：`backend-go/testdata/contract-diff/gate-thresholds.yaml`
    - 失败分类规范（建议位置）：`backend-go/testdata/contract-diff/failure-taxonomy.md`
  - **Acceptance Criteria**:
    - 默认阈值落地：`overallParity >= 99%`、`requiredApiPassRate = 100%`
    - `blockingLevel=critical` 的差异出现即失败
    - 阈值边界条件（刚好达标/刚好不达标）判定一致
  - **Rollback Plan**:
    - 回退到前一版阈值配置并保留失败样本用于复盘

- [x] **CI-GATE-005** 报告归档与审计留痕
  - **Risk**: Medium
  - **Depends On**: CI-GATE-003
  - **Parallel Group**: Wave 4（与 CI-GATE-004 并行）
  - **Input**:
    - CI job 产出的原始 diff 数据
  - **Output**:
    - 机器可读报告：`contract-diff-report.json`
    - 人类可读报告：`contract-diff-report.md`
    - 归档规范（建议位置）：`backend-go/testdata/contract-diff/report-archive-policy.md`
  - **Acceptance Criteria**:
    - 每次 gate 执行都归档 JSON + Markdown 两类报告
    - 报告命名可追溯到 `commit SHA` 与 `PR number`
    - 失败报告可定位到具体接口与差异字段
  - **Rollback Plan**:
    - 归档失败时降级为 job 日志输出并标记非阻断告警

- [x] **CI-GATE-006** 联调验收与基线固化
  - **Risk**: Medium
  - **Depends On**: CI-GATE-004, CI-GATE-005
  - **Parallel Group**: Wave 5
  - **Input**:
    - 完整 CI gate 流程
    - 至少一组通过样本和一组失败样本
  - **Output**:
    - Gate 验收记录（建议位置）：`openspec/changes/add-go-contract-diff-ci-gate/gate-validation-report.md`
    - 版本化基线与维护说明（建议位置）：`backend-go/testdata/contract-diff/baseline-policy.md`
  - **Acceptance Criteria**:
    - 成功样本不误报，失败样本可稳定阻断
    - Gate 结果与报告归档一致
    - OpenSpec 严格校验通过
  - **Rollback Plan**:
    - 回退到仅构建测试的 CI 基线，保留 gate 配置以便二次启用

## Execution Notes

- 高风险任务（CI-GATE-003/004）失败会阻断后续任务。
- 任务状态仅可在对应验收标准满足后更新。
- 白名单、阈值、归档三类产物必须同步评审，不允许单独漂移。
- 所有“建议位置”可在执行时按仓库约定调整，但产物语义与验收标准不可弱化。

## Command Templates（执行参考）

> 以下命令为执行模板，实际脚本名可按仓库约定微调；但输入输出契约与验收目标不可变更。

### CI-GATE-001 白名单冻结

```bash
# 1) 生成/更新关键接口白名单
python backend-go/scripts/contract-diff/build_whitelist.py \
  --matrix openspec/changes/add-go-security-compatibility-readiness/compatibility-matrix.md \
  --out backend-go/testdata/contract-diff/critical-whitelist.yaml

# 2) 基础结构校验
python backend-go/scripts/contract-diff/validate_whitelist.py \
  --in backend-go/testdata/contract-diff/critical-whitelist.yaml
```

### CI-GATE-002 执行器标准化

```bash
# 1) 本地执行 contract diff
bash backend-go/scripts/contract-diff/run_contract_diff.sh \
  --whitelist backend-go/testdata/contract-diff/critical-whitelist.yaml \
  --java-base "$JAVA_BASE_URL" \
  --go-base "$GO_BASE_URL" \
  --out-dir backend-go/tmp/contract-diff

# 2) 输出结构校验
python backend-go/scripts/contract-diff/validate_output_schema.py \
  --in backend-go/tmp/contract-diff/contract-diff-report.json
```

### CI-GATE-003 CI Job 接入

```bash
# 1) 本地检查 workflow 语法（可选）
yamllint .github/workflows/go-contract-diff-gate.yml

# 2) 通过 PR 触发后在 Actions 中确认 job 运行
# - 目标：job 名称可见，且失败可阻断合并
```

### CI-GATE-004 阈值与阻断

```bash
# 1) 阈值校验（成功样本）
python backend-go/scripts/contract-diff/evaluate_gate.py \
  --threshold backend-go/testdata/contract-diff/gate-thresholds.yaml \
  --report backend-go/testdata/contract-diff/samples/pass-report.json

# 2) 阈值校验（失败样本）
python backend-go/scripts/contract-diff/evaluate_gate.py \
  --threshold backend-go/testdata/contract-diff/gate-thresholds.yaml \
  --report backend-go/testdata/contract-diff/samples/fail-report.json
```

### CI-GATE-005 报告归档

```bash
# 1) 生成 Markdown 报告
python backend-go/scripts/contract-diff/render_report_md.py \
  --in backend-go/tmp/contract-diff/contract-diff-report.json \
  --out backend-go/tmp/contract-diff/contract-diff-report.md

# 2) 打包归档
tar -czf backend-go/tmp/contract-diff/contract-diff-artifacts.tgz \
  backend-go/tmp/contract-diff/contract-diff-report.json \
  backend-go/tmp/contract-diff/contract-diff-report.md
```

### CI-GATE-006 联调验收

```bash
# 1) OpenSpec 严格校验
openspec validate add-go-contract-diff-ci-gate --strict --no-interactive

# 2) 验收记录检查（示例）
test -f openspec/changes/add-go-contract-diff-ci-gate/gate-validation-report.md
```

## Definition of Done（DoD）Checklist

> 规则：任务从 `- [ ]` 改为 `- [x]` 前，必须满足对应 DoD 清单。

### CI-GATE-001 DoD
- [x] 白名单文件已提交并可被版本追踪
- [x] 每条白名单记录都包含 `path/method/owner/priority/blockingLevel`
- [x] 白名单治理规则文档已提交并评审

### CI-GATE-002 DoD
- [x] 执行脚本支持白名单过滤、超时、重试
- [x] 同一输入重复执行结果一致（至少两次）
- [x] 输出 schema 校验通过

### CI-GATE-003 DoD
- [x] PR 触发 contract diff gate job
- [x] 受保护分支触发复跑
- [x] branch protection 已将 gate 设置为必需检查（需在 GitHub 设置）

### CI-GATE-004 DoD
- [x] 阈值配置文件落盘并纳入版本控制
- [x] `critical` 级差异被正确阻断
- [x] 边界样本（刚好达标/刚好不达标）判定符合预期

### CI-GATE-005 DoD
- [x] 每次执行均产出 JSON 与 Markdown 报告
- [x] artifact 命名包含 commit 或 PR 维度
- [x] 报告可定位具体接口与字段差异

### CI-GATE-006 DoD
- [x] 至少一组通过样本与一组失败样本已留档
- [x] `gate-validation-report.md` 已提交并包含误报评估
- [x] `openspec validate add-go-contract-diff-ci-gate --strict --no-interactive` 通过

## Sign-off Matrix

| Area | Owner Role | Required Sign-off |
|------|------------|-------------------|
| Whitelist governance | API/module owner | Yes |
| Threshold policy | Tech lead / Architect | Yes |
| CI branch protection | Repo maintainer | Yes |
| Report archival policy | QA/Release owner | Yes |
