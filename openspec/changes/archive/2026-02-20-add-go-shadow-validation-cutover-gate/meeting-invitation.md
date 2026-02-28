# Go/No-Go Approval Meeting Invitation

## 邮件/日历邀请模板

---

### 主题

```
[审批会议] Java → Go 后端迁移生产切换 - Go/No-Go 决策
```

---

### 正文

```
各位好，

Java → Go 后端迁移项目已完成所有技术验证，现召开 Go/No-Go 审批会议，请相关人员准时参加。

会议信息
--------
日期：[待定]
时间：[待定]（预计 30 分钟）
地点：[会议室 / 线上会议链接]

参会人员（必须）
--------
- Engineering Manager（工程经理） - 决策审批
- Release Manager（发布经理） - 决策审批  
- Observability Engineer（可观测性工程师） - 决策审批

参会人员（可选）
--------
- Platform SRE Lead - 技术支持
- API Compatibility Owner - 技术支持
- Gateway Operations Lead - 技术支持

会议议程
--------
1. 项目背景与迁移概述（5分钟）
2. 质量门禁验证结果（10分钟）
3. Shadow 验证结果回顾（10分钟）
4. 风险评估与回滚方案（5分钟）
5. Q&A 与审批签署（5分钟）

决策要点
--------
✅ 单元测试通过
✅ 构建成功（32MB 二进制）
✅ Lint 检查通过（57 项低优先级警告）
✅ Shadow 验证 4h 通过（mismatch 0.30% < 1%）
✅ 安全事件：0 critical
✅ Sev-1/Sev-2 回归：0
✅ 回滚演练（dry-run）通过

待决事项
--------
⚠️ 三方审批签署待完成
⚠️ 真实回滚演练待执行（需生产凭证）

预期产出
--------
- Go/No-Go 正式决策
- 审批签署完成
- 切换窗口确认

会前准备
--------
请提前查阅以下文档：
1. 审批签署模板：openspec/changes/archive/2026-02-20-add-go-shadow-validation-cutover-gate/cutover-approval-signoff.md
2. Go/No-Go 决策记录：openspec/changes/archive/2026-02-20-add-go-shadow-validation-cutover-gate/go-no-go-decision.md
3. Shadow 验证报告：openspec/changes/archive/2026-02-20-add-go-shadow-validation-cutover-gate/shadow-validation-report.md

如有问题，请提前反馈。

谢谢！

[会议组织者姓名]
[联系方式]
```

---

### 日历事件描述

```
Java → Go 后端迁移生产切换审批会议

决策事项：是否批准执行生产环境流量切换

关键指标：
- Shadow 验证：✅ PASS (0.30% mismatch)
- 安全事件：✅ 0 critical
- 回归：✅ 0 Sev-1/Sev-2

必参会人员：Engineering Manager, Release Manager, Observability Engineer

文档链接：[链接]
```

---

## 快速邀请文本（即时通讯）

```
📅 审批会议邀请

主题：Java→Go迁移 Go/No-Go 决策
时间：[待定]
时长：30分钟

关键结论：
✅ Shadow 4h验证通过
✅ Mismatch 0.30% < 1%
✅ 0安全事件，0回归

待决：三方审批签署

请查阅：cutover-approval-signoff.md
```
