# Design: Frontend TypeScript Types Refactoring

## Context

前端代码库（`apps/frontend/src/`）包含约 8934 个模块，存在 1132 个 TypeScript 类型错误。这些错误主要集中在：

1. **Chart 模块** (`src/views/chart/`) - 图表渲染和编辑
2. **Visualization 模块** (`src/components/data-visualization/`) - 数据可视化画布
3. **Custom Component 模块** (`src/custom-component/`) - 自定义组件

当前状态：
- `npm run build:base` 成功（Vite 不强制类型检查）
- `npm run ts:check` 失败（vue-tsc 报告类型错误）
- `npm run lint` 通过（ESLint 配置较宽松）

## Goals / Non-Goals

### Goals
- 修复所有 1132 个 TypeScript 类型错误
- 不改变任何运行时行为
- 保持代码可读性和可维护性
- 建立类型安全最佳实践

### Non-Goals
- 不重构代码逻辑
- 不升级依赖版本
- 不添加新功能
- 不修改 API 接口

## Decisions

### D1: 类型断言策略

**决策**: 优先使用类型守卫，避免 `as any`

```typescript
// ❌ 避免
const value = (obj as any).property

// ✅ 推荐
interface MyType {
  property: string
}
const value = (obj as MyType).property

// ✅ 更好：类型守卫
function isMyType(obj: unknown): obj is MyType {
  return typeof obj === 'object' && obj !== null && 'property' in obj
}
if (isMyType(obj)) {
  const value = obj.property
}
```

**原因**: `as any` 会完全绕过类型检查，引入潜在运行时错误。

### D2: Vue 组件类型定义

**决策**: 使用 Composition API + `<script setup lang="ts">` + interface

```typescript
// ✅ 推荐
interface Props {
  modelValue: string
  disabled?: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()
```

**原因**: 提供完整的类型推断和 IDE 支持。

### D3: Reactive 状态类型

**决策**: 使用泛型参数定义类型

```typescript
// ❌ 避免
const state = reactive({
  items: [],
  loading: false
})

// ✅ 推荐
interface State {
  items: Item[]
  loading: boolean
}

const state = reactive<State>({
  items: [],
  loading: false
})
```

**原因**: 避免 `items: never[]` 类型推断问题。

### D4: API 响应类型

**决策**: 为 API 响应创建接口，使用 `IResponse<T>` 泛型

```typescript
interface UserInfo {
  id: string
  name: string
}

// API 返回类型
type UserInfoResponse = IResponse<UserInfo>
```

**原因**: 统一 API 响应格式，便于错误处理。

### D5: DOM 元素类型

**决策**: 使用 `HTMLElement` 及其子类型，配合类型断言

```typescript
// ❌ 避免
const el = document.querySelector('.my-element')
el.style.width = '100px' // Error: Property 'style' does not exist on 'Element'

// ✅ 推荐
const el = document.querySelector<HTMLElement>('.my-element')
if (el) {
  el.style.width = '100px'
}
```

**原因**: `querySelector` 返回 `Element | null`，需要正确处理。

## Risks / Trade-offs

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| 类型断言隐藏真实错误 | 中 | 高 | 每次修改后运行 `build:base` 验证 |
| 过度使用 `any` 绕过检查 | 中 | 中 | 代码审查，限制 `any` 使用 |
| 修复时间超出预期 | 低 | 低 | 按模块分批，可随时暂停 |
| 类型定义与实际数据不匹配 | 中 | 高 | 运行时验证关键数据结构 |

## Migration Plan

### Phase 1: 高频错误文件 (1-2 周)

1. 按错误数排序文件
2. 从错误最多的文件开始
3. 每修复 5-10 个文件验证一次

### Phase 2-4: 模块化修复 (2-4 周)

1. 按模块划分工作
2. 每个模块独立提交
3. 模块完成后验证

### 验证检查点

每个阶段完成后：
```bash
cd apps/frontend
npm run lint           # 无新增警告
npm run build:base     # 构建成功
npm run ts:check       # 错误数减少
```

## Open Questions

1. **是否需要为所有 API 创建类型定义？**
   - 当前部分 API 使用 `IResponse<any>`
   - 建议：优先修复高频使用 API

2. **是否引入 Zod/io-ts 进行运行时验证？**
   - 当前仅静态类型检查
   - 建议：本次不引入，后续单独考虑

3. **是否需要更新 tsconfig.json 严格性？**
   - 当前 `noImplicitAny: false`
   - 建议：本次不修改，避免扩大范围
