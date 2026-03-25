# Test Gate Configuration

测试门禁配置，用于在 CI 和归档前强制执行测试。

## 快速开始

### 1. 复制 GitHub Actions 工作流

将以下文件复制到你的项目：

```
.github/workflows/test-gate.yml
```

### 2. 配置 openspec/config.yaml

在你的 `openspec/config.yaml` 中添加 `test_gate` 配置：

```yaml
test_gate:
  ci:
    - Unit tests
    - Integration tests
    - Lint and type check
  
  archive:
    - All CI tests must pass
    - E2E tests must pass
  
  allow_e2e_skip: true
  e2e_skip_requires_acknowledgment: true
```

### 3. 更新 archive skill（可选）

如果使用 OpenSpec，将 `openspec/templates/test-gate-skill.md` 中的步骤添加到你的 `.opencode/skills/openspec-archive-change/SKILL.md`。

## 测试策略

| 阶段 | 测试内容 | 触发时机 |
|------|---------|---------|
| CI | 单元 + 集成 | 每个 PR |
| Archive | CI + E2E | 归档前 |

## 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `allow_e2e_skip` | 是否允许跳过 E2E | `true` |
| `e2e_skip_requires_acknowledgment` | 跳过 E2E 是否需要确认 | `true` |

## 自定义测试命令

根据你的项目修改 `test-gate.yml` 中的测试命令：

```yaml
# Go 项目
- name: Backend unit tests
  run: make test

# Node.js 项目
- name: Frontend unit tests
  run: npm test
```
