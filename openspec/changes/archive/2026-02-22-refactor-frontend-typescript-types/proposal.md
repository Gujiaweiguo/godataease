# Change: Refactor Frontend TypeScript Types

## Why

前端代码存在 1132 个 TypeScript 类型错误，影响代码质量和可维护性。主要问题包括：

1. `unknown` 类型泛滥导致属性访问失败（约 60% 错误）
2. string/number 类型不匹配（约 20% 错误）
3. 未使用变量声明（约 5% 错误）
4. 缺失的类型定义和导入（约 15% 错误）

这些错误虽然不阻塞构建（`npm run build:base` 通过），但：
- 无法将 `npm run ts:check` 作为 CI 质量门禁
- IDE 类型提示失效，降低开发效率
- 潜在的运行时类型错误风险

## What Changes

### 类型定义增强
- 为 Vue 组件添加正确的 props/emits 类型定义
- 为 reactive 对象添加 interface 约束
- 为 API 响应添加类型定义

### 类型断言优化
- 使用类型守卫替代 `as any` 类型断言
- 为 `unknown` 类型添加运行时类型检查
- 规范 Element DOM 操作的类型处理

### 代码清理
- 移除未使用的变量和导入
- 统一类型导入方式（`import type`）

### 错误分布（Top 10）

| 文件 | 错误数 | 类别 |
|------|--------|------|
| `common_antv.ts` | 88 | chart |
| `chart/editor/index.vue` | 88 | chart |
| `CanvasCore.vue` | 36 | visualization |
| `LinkageSet.vue` | 25 | visualization |
| `table-pivot.ts` | 23 | chart |
| `TableHeaderGroupConfig.vue` | 23 | chart |
| `common_table.ts` | 22 | chart |
| `chart-mix.ts` | 21 | chart |
| `datasource/form/index.vue` | 20 | datasource |
| `Shape.vue` | 20 | visualization |

## Impact

- **Affected specs**: frontend-type-safety (新建)
- **Affected code**: `apps/frontend/src/`
- **主要模块**:
  - `src/components/data-visualization/` (~150 errors)
  - `src/components/visualization/` (~80 errors)
  - `src/views/chart/` (~250 errors)
  - `src/custom-component/` (~200 errors)
  - 其他模块 (~450 errors)

## Approach

分四个阶段进行，每个阶段独立可交付：

1. **Phase 1: 高频错误文件** (Top 10, ~300 errors)
2. **Phase 2: data-visualization 模块** (~150 errors)
3. **Phase 3: chart 模块** (~250 errors)
4. **Phase 4: 其他模块** (~430 errors)

每个 Phase 完成后验证：
- `npm run lint` 通过
- `npm run build:base` 通过
- 错误数减少确认

## Success Criteria

- [ ] `npm run ts:check` 错误数从 1132 降至 0
- [ ] `npm run lint` 无新增警告
- [ ] `npm run build:base` 成功
- [ ] 所有修复不改变运行时行为

## Risks

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 类型断言引入运行时错误 | 高 | 每次修改后运行构建验证 |
| 过度使用 `any` 类型 | 中 | 代码审查，优先使用具体类型 |
| 修复时间超出预期 | 低 | 按模块分批，可随时暂停 |
