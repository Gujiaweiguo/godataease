## Context

DataFilling is a feature-complete backend module with a full API layer in `src/api/datafilling.ts` (402 lines covering forms, table data, tasks, sub-tasks, user tasks, Excel, and options). The Go backend exposes all endpoints under `/data-filling/*`. However, the frontend has zero native Vue pages. All DataFilling UI is loaded at runtime through Xpack dynamic components using base64-encoded jsname paths in four entry points:

- **ChartView.vue** — `AsyncXpackComponent` with jsname for DataFilling/DataFillingEditor/DataFillingHandler
- **panel/App.vue** — `XpackComponent` with jsname for the same three component names
- **ShortcutTable.vue** — `XpackComponent` for the data-filling tab (fill view + count loader)
- **mobile/home/index.vue** — `XpackComponent` for the data-filling tab (fill view + count loader)

The existing frontend patterns (dataset, datasource) use standalone route pages at paths like `/module-dataset`, `/dataset-embedded`, with tree + detail layouts built in-repo. Routes are hash-based (`createWebHashHistory`), defined in `src/router/index.ts` as flat hidden routes. Menu entries are driven by backend-returned route config resolved through `src/router/establish.ts`, which supports both in-repo component resolution and Xpack plugin loading via the `plugin` flag.

This design replaces the Xpack dependency entirely with native Vue 3 pages, following established project conventions.

## Goals / Non-Goals

**Goals:**

- Full native Vue 3 implementation of all DataFilling pages: admin management, form editor, table data, task management, user fill
- Route integration matching the existing `module-*` and `*-embedded` patterns
- Rewire all four Xpack entry points (ChartView, Panel, Workbranch, Mobile) to native routes/components
- Mobile-responsive fill pages that work within the existing mobile layout
- Consume the existing `src/api/datafilling.ts` API layer directly with no API changes

**Non-Goals:**

- Backend API changes (the Go backend and API contract are stable)
- Modifying the Xpack plugin system itself (only removing DataFilling's dependency on it)
- Building a drag-and-drop library from scratch (evaluate existing project dependencies first; fallback to reorder controls if no suitable library exists)
- Offline/PWA support for fill pages
- Permission model changes (use existing route guards and menu visibility)

## Decisions

### D1: View directory structure — `src/views/data-filling/`

**Decision**: Place all DataFilling views under `src/views/data-filling/` as a new top-level view directory, parallel to `visualized/`, `chart/`, `workbranch/`, etc.

**Rationale**: DataFilling is a distinct domain feature, not a sub-concern of dataset or datasource. A top-level directory gives it the same organizational weight as other modules. The path `data-filling` (hyphenated) matches the API base path `/data-filling` and the convention used by `visualized/data/` patterns.

**Structure**:

```
src/views/data-filling/
├── manage/                        # Admin section
│   ├── index.vue                  # Admin shell: tree sidebar + content area
│   ├── FormTree.vue               # Folder/form tree component (extracted)
│   ├── form/
│   │   └── index.vue              # Form editor page
│   │   └── FormFieldEditor.vue    # Individual field config block
│   │   └── FormFieldList.vue      # Field list with reorder
│   ├── data/
│   │   └── index.vue              # Table data grid page
│   │   └── CommitLog.vue          # Commit log drawer/panel
│   │   └── ExcelUploader.vue      # Upload/confirm flow component
│   └── task/
│       └── index.vue              # Task list + editor page
│       └── TaskEditor.vue         # Task creation/edit form
│       └── SubTaskList.vue        # Sub-task table with progress
│       └── SubTaskUsers.vue       # User detail drawer
├── fill/                          # User-facing fill section
│   ├── index.vue                  # User task list page
│   └── TaskFill.vue               # Fill form for a specific task
└── components/                    # Shared across manage and fill
    ├── FormSchemaRenderer.vue     # Renders form fields from JSON schema
    ├── DataGrid.vue               # Reusable paginated data table
    └── SearchFilter.vue           # Column search filter bar
```

**Alternative considered**: Place under `src/views/visualized/data/data-filling/`. Rejected because DataFilling has its own admin tree, its own fill UX, and its own mobile entry — it's more than a data sub-module.

### D2: Route design — flat hidden routes, two groups

**Decision**: Add routes as flat hidden entries in `src/router/index.ts`, following the `module-*` (full page) and `*-embedded` (embedded/panel context) patterns already used by dataset and datasource.

**Routes**:

| Path | Name | Component | Purpose |
|------|------|-----------|---------|
| `/data-filling-manage` | `data-filling-manage` | `manage/index.vue` | Admin management (tree + content) |
| `/data-filling-editor` | `data-filling-editor` | `manage/form/index.vue` | Form editor (receives params via query) |
| `/data-filling-data` | `data-filling-data` | `manage/data/index.vue` | Table data grid |
| `/data-filling-tasks` | `data-filling-tasks` | `manage/task/index.vue` | Task management |
| `/data-filling-fill` | `data-filling-fill` | `fill/index.vue` | User task list |
| `/data-filling-fill-task` | `data-filling-fill-task` | `fill/TaskFill.vue` | Fill form for a task |

All routes are `hidden: true` (not shown in sidebar directly; menu entry is backend-driven). Route parameters (formId, taskId, pid) are passed as query parameters (`?formId=123`) rather than path segments (`:formId`), matching the pattern used by dataset-embedded-form which reads `route.query` for context. This avoids the need for nested route children and keeps the routing flat.

**Menu integration**: The backend-returned menu config already defines the DataFilling menu entry (since the Xpack plugin currently registers it). After migration, the menu entry's `component` field should resolve to the in-repo view path `data-filling/manage/index` instead of being flagged as `plugin: true`. If the menu is still served with `plugin: true`, `establish.ts` will wrap it in XpackComponent — so the backend menu config must be updated to point to the native component path and remove the `plugin` flag for DataFilling entries.

### D3: Xpack entry point rewiring strategy

**Decision**: Replace conditional Xpack component rendering with native route navigation or direct component import, depending on context.

**ChartView.vue** (`src/views/chart/ChartView.vue`):
- Currently: `name.includes('DataFilling')` → set jsname → render `AsyncXpackComponent`
- After: `name === 'DataFilling'` → `router.push('/data-filling-manage')`, `'DataFillingEditor'` → `router.push({ path: '/data-filling-editor', query: { formId } })`, `'DataFillingHandler'` → `router.push('/data-filling-fill')`
- Since ChartView is itself a route-level component, navigation replaces the component rather than rendering within it

**panel/App.vue** (`src/pages/panel/App.vue`):
- Currently: `val.includes('DataFilling')` → render `XpackComponent` with jsname
- After: Same mapping as ChartView — navigate to native routes
- The `isDataFilling` ref and conditional rendering block are removed; DataFilling names go through the same `nextTick` component switch as other names, but with native route navigation instead

**ShortcutTable.vue** (`src/views/workbranch/ShortcutTable.vue`):
- Currently: `<XpackComponent jsname="..." v-if="activeName === 'data-filling'" />` for the tab, plus a separate XpackComponent for count loading
- After: Replace with `<UserTaskList v-if="activeName === 'data-filling'" />` using direct component import. The todo count is fetched internally by the UserTaskList component via `getUserTaskTodoCount`
- This is an inline component swap, not a route navigation, because the workbranch has its own tab layout

**mobile/home/index.vue** (`src/views/mobile/home/index.vue`):
- Currently: `<XpackComponent>` for the data-filling tab
- After: Same inline component swap — import `UserTaskList` directly. The component renders a mobile-responsive layout internally (see D7)

**Key principle**: Route navigation for ChartView and Panel (which are navigation hubs). Direct component import for Workbranch and Mobile (which embed the component in their own layout).

### D4: State management — minimal Pinia stores

**Decision**: Use one Pinia store for admin tree state. Everything else uses local component state via `ref`/`reactive`.

**Store**: `src/store/modules/data-filling.ts`

The store holds:
- `formTree`: the loaded tree data (avoid re-fetching on admin sub-page navigation)
- `selectedNodeId`: currently selected tree node id
- `expandedNodeIds`: tree expansion state (persisted across navigations)

All other state (form editor fields, table data rows, task lists, fill form data) is local to the page component that manages it. This matches the dataset pattern where the tree state is in `interactive` store but the form editor holds its own local state.

**Alternative considered**: Stores for table data cache and task state. Rejected — these are page-scoped and stale data is worse than a fresh API call on page entry.

### D5: Form schema rendering strategy

**Decision**: Build a `FormSchemaRenderer.vue` component that takes a JSON `forms` string (the form field definition) and renders the appropriate input controls for each field type. Shared between the form editor preview and the user fill page.

The `forms` field in the API is a JSON string describing form fields with properties like:
- Field type (text, number, decimal, date, select, etc.)
- Label, placeholder, required flag
- Type-specific settings (e.g., date format, select options, decimal precision)

`FormSchemaRenderer` maps each field type to an Element Plus input component:
- `text` / `nvarchar` → `el-input`
- `number` → `el-input-number`
- `decimal` → `el-input-number` with precision config
- `date` / `datetime` → `el-date-picker`
- `select` → `el-select` with options from `listColumnData` or static config

For the form editor, `FormFieldList.vue` wraps a list of `FormFieldEditor.vue` blocks, each rendering field configuration UI. Reordering uses a simple move-up/move-down button pair initially, with the option to integrate a drag-drop library later. This avoids adding a new npm dependency in the initial migration.

### D6: Admin page layout — tree sidebar + content area

**Decision**: The admin management page (`manage/index.vue`) uses a two-panel layout with an `el-aside` tree sidebar and a `el-main` content area, following the dataset `index.vue` pattern.

- The tree sidebar (`FormTree.vue`) shows the folder/form hierarchy
- Selecting a leaf form loads the appropriate content view in the right panel: table data grid, task list, or form info summary
- Right-panel content is switched via a `ref` component binding (not route children), keeping navigation fast without full page reloads
- "Edit Form" and "New Form" actions navigate to the form editor route (full page transition)

This mirrors how `dataset/index.vue` has a tree sidebar and opens the dataset form editor as a separate route.

### D7: Responsive/mobile handling

**Decision**: The fill components (`fill/index.vue`, `fill/TaskFill.vue`) use CSS responsive design with Element Plus's responsive grid, not separate mobile-specific view files.

- On desktop/workbranch: full-width task list with table columns
- On mobile: single-column card layout for tasks, simplified fill form with stacked fields
- Responsive breakpoint: use `@vueuse/core`'s `useWindowSize` (already used in dataset form) to detect mobile viewport and conditionally render compact layouts
- The mobile home entry imports the same `fill/index.vue` component — it adapts via CSS media queries

**Alternative considered**: Separate `src/views/mobile/data-filling/` views. Rejected — the fill UX is simple enough that responsive CSS handles both contexts, and duplicating views creates maintenance burden. If the mobile layout diverges significantly in the future, a mobile-specific wrapper component can be introduced.

### D8: API consumption boundary

**Decision**: All API calls go through the existing `src/api/datafilling.ts`. No wrappers, no store-based API caching. Components import and call API functions directly in composable functions or `<script setup>` blocks.

Pattern in each view component:
```typescript
import { getFormTree, createForm, deleteForm, ... } from '@/api/datafilling'

const loading = ref(false)
const treeData = ref([])

const loadTree = async () => {
  loading.value = true
  try {
    treeData.value = await getFormTree()
  } finally {
    loading.value = false
  }
}
```

Error handling follows the existing pattern: the axios interceptor in `src/config/axios` handles auth errors and network failures globally. Component-level `try/catch` wraps calls for user-facing error messages via `ElMessage.error()`.

### D9: Backend menu config update

**Decision**: The backend-returned menu/route config for DataFilling entries must be updated to remove the `plugin: true` flag and point to the native component path.

Currently, the menu entries for DataFilling are registered by the Xpack plugin with `plugin: true`, causing `establish.ts` to resolve them as `XpackComponent` with jsname. After migration:

1. The backend should return DataFilling menu entries with `component: 'data-filling/manage/index'` and `plugin: false` (or absent)
2. `establish.ts`'s `resolveViewComponent` will resolve this to `../views/data-filling/manage/index.vue` via the glob import

If backend menu config cannot be changed immediately, a fallback in `establish.ts` can intercept known DataFilling plugin paths and remap them to native components. However, this is a transitional hack — the clean path is updating the backend menu config.

## Risks / Trade-offs

**[Risk] Xpack plugin still registers DataFilling components** → The Xpack plugin may still attempt to register DataFilling routes/components. After migration, these registrations become dead code. No runtime conflict because native routes take precedence in the router, and the entry points no longer call `XpackComponent` for DataFilling names. Mitigation: verify that removing the jsname paths doesn't cause console errors from the plugin trying to mount to a nonexistent target.

**[Risk] Mobile layout divergence** → The current Xpack mobile DataFilling component may have mobile-specific interactions that aren't captured by responsive CSS alone. Mitigation: test on mobile viewport early and introduce a mobile wrapper component only if needed. The mobile view is limited to the fill experience (task list + fill form), which is straightforward to make responsive.

**[Risk] Form schema format mismatch** → The `forms` JSON string format returned by `getFormById` must match what the form editor serializes on save. The current API doesn't document this schema explicitly. Mitigation: parse the existing forms structure from real API responses during implementation. The `FormSchemaRenderer` must handle whatever format the backend stores. Add type definitions in a shared `types.ts` file under the data-filling views directory.

**[Risk] Drag-drop for form editor field reorder** → Using move-up/move-down buttons instead of drag-drop is less polished but avoids adding a dependency. Trade-off: acceptable for initial migration; can add `vuedraggable` or similar in a follow-up if UX feedback demands it.

**[Trade-off] Flat routes with query params vs. nested routes** → Query params are simpler but make URLs less semantic (`/data-filling-editor?formId=123` vs `/data-filling/manage/form/123`). This matches the existing dataset pattern and avoids the complexity of nested route resolution.

**[Trade-off] Single store for tree only** → Keeping most state local means data is re-fetched on page entry. This is acceptable because DataFilling data changes frequently and stale cache would be confusing. Only the tree structure benefits from cross-page persistence.

## Migration Plan

1. **Create view scaffolding**: Add `src/views/data-filling/` directory structure with all Vue files as empty shells
2. **Implement shared components**: `FormSchemaRenderer`, `DataGrid`, `SearchFilter` — these are used by multiple pages
3. **Implement admin pages**: FormTree → form editor → table data → task management (build from leaf dependencies up)
4. **Implement fill pages**: User task list → fill form
5. **Add routes**: Register all routes in `src/router/index.ts`
6. **Rewire entry points**: ChartView → Panel → Workbranch → Mobile (one at a time, test each)
7. **Update menu config**: Coordinate backend change to remove `plugin` flag from DataFilling menu entries
8. **Remove Xpack references**: Delete all DataFilling-related jsname strings and conditional Xpack rendering blocks
9. **Test all surfaces**: Verify each of the four entry points independently, plus direct URL navigation

**Rollback strategy**: Each entry point rewiring is independent. If ChartView breaks, only that file needs reverting. The Xpack jsname strings and conditional blocks can be restored by reverting the specific file. New view files and routes can coexist with Xpack (the entry points determine which is used). Full rollback is `git revert` on the merge commit.

## Open Questions

1. **Form schema JSON format**: What is the exact JSON structure of the `forms` field? Must be reverse-engineered from `getFormById` responses or backend code before implementing `FormSchemaRenderer` and the form editor.
2. **Backend menu config ownership**: Who controls the menu/route entries returned by the backend? If the Go backend serves these, we need a coordinated change to remove `plugin: true` for DataFilling routes.
3. **Drag-drop library**: Should we introduce `vuedraggable` (or equivalent) for form field reorder now, or defer to a follow-up? The initial migration can ship with move-up/move-down buttons.
4. **Commit log drawer vs. page**: Should the commit log be a slide-over drawer within the table data page, or a separate route? Current design assumes drawer for UX continuity.
