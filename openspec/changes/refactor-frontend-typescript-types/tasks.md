# Tasks: Refactor Frontend TypeScript Types

## Phase 1: 高频错误文件 (Top 10) ✅ COMPLETED

**目标**: 修复错误最多的 10 个文件，减少约 300 个错误

**结果**: 错误从 1132 减少到 895（减少 237 个，21% 改善）

- [x] 1.1 `src/views/chart/components/js/panel/common/common_antv.ts` (88 errors → 47 partial)
- [x] 1.2 `src/views/chart/components/editor/index.vue` (88 errors → 68 partial)
- [x] 1.3 `src/components/data-visualization/canvas/CanvasCore.vue` (36 errors → 1 partial)
- [x] 1.4 `src/components/visualization/LinkageSet.vue` (25 errors → 0 complete)
- [x] 1.5 `src/views/chart/components/js/panel/charts/table/table-pivot.ts` (23 errors → 0 complete)
- [x] 1.6 `src/views/chart/components/editor/editor-style/components/table/TableHeaderGroupConfig.vue` (23 errors → 0 complete)
- [x] 1.7 `src/views/chart/components/js/panel/common/common_table.ts` (22 errors → 0 complete)
- [x] 1.8 `src/views/chart/components/js/panel/charts/others/chart-mix.ts` (21 errors → partial fix)
- [x] 1.9 `src/views/visualized/data/datasource/form/index.vue` (20 errors → 0 complete)
- [x] 1.10 `src/components/data-visualization/canvas/Shape.vue` (20 errors → 0 complete)
- [x] 1.11 Phase 1 验证: `npm run lint` ✅ | `npm run build:base` ✅ | `npm run ts:check` 895 errors

**Commits**:
- `78ed609` fix(frontend): resolve type errors in Phase 1 files
- `e0e9da2` fix(frontend): resolve type errors in Phase 1 remaining files

## Phase 2: data-visualization 模块 🚧 IN PROGRESS

**目标**: 修复 `src/components/data-visualization/` 下所有类型错误

- [ ] 2.1 `src/components/data-visualization/RealTimeGroup.vue`
- [ ] 2.2 `src/components/data-visualization/RealTimeGroupInner.vue`
- [ ] 2.3 `src/components/data-visualization/RealTimeListTree.vue`
- [ ] 2.4 `src/components/data-visualization/RealTimeTab.vue`
- [ ] 2.5 `src/components/data-visualization/canvas/ComponentWrapper.vue`
- [ ] 2.6 `src/components/data-visualization/canvas/ContextMenuDetails.vue`
- [ ] 2.7 `src/components/data-visualization/canvas/DePreview.vue`
- [ ] 2.8 `src/components/data-visualization/canvas/PGrid.vue`
- [ ] 2.9 其他 data-visualization 文件
- [ ] 2.10 Phase 2 验证: 运行 `npm run lint && npm run build:base && npm run ts:check`

## Phase 3: chart 模块

**目标**: 修复 `src/views/chart/` 和 `src/components/chart/` 下所有类型错误

- [ ] 3.1 `src/views/chart/components/js/panel/charts/map/tooltip-carousel.ts`
- [ ] 3.2 `src/views/chart/components/js/panel/charts/bar/range-bar.ts`
- [ ] 3.3 `src/views/chart/components/views/components/ChartComponentS2.vue`
- [ ] 3.4 其他 chart 文件 (约 20 个)
- [ ] 3.5 Phase 3 验证: 运行 `npm run lint && npm run build:base && npm run ts:check`

## Phase 4: 其他模块

**目标**: 修复剩余所有类型错误

- [ ] 4.1 `src/custom-component/` 模块 (~200 errors)
- [ ] 4.2 `src/components/visualization/` 模块 (~80 errors)
- [ ] 4.3 `src/components/dashboard/` 模块
- [ ] 4.4 `src/views/visualized/` 模块
- [ ] 4.5 `src/hooks/` 和 `src/events/` 模块
- [ ] 4.6 `src/utils/` 和 `src/config/` 模块
- [ ] 4.7 其他文件
- [ ] 4.8 Phase 4 验证: 运行 `npm run lint && npm run build:base && npm run ts:check`

## 最终验证

- [ ] 5.1 `npm run ts:check` 错误数降至 0
- [ ] 5.2 `npm run lint` 无新增警告
- [ ] 5.3 `npm run build:base` 成功
- [ ] 5.4 代码审查确认无行为变更

## 提交规范

每个 Phase 完成后创建独立提交：

```
fix(frontend): resolve type errors in [module] - Phase N

- Fix TS2339: Property does not exist on 'unknown'
- Fix TS2322: Type assignment errors
- Remove unused variables (TS6133)
- Add type guards and interfaces

Reduces ts:check errors from X to Y
```
