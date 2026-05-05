## Why

DataFilling is the only major feature in this repo whose frontend UX is entirely provided by Xpack dynamic components, loaded at runtime via base64-encoded `jsname` paths. There are zero native Vue pages for DataFilling in the `apps/frontend` source tree, even though the full API layer (`src/api/datafilling.ts`) already exists and the Go backend is complete. This means DataFilling breaks whenever the Xpack plugin is missing or changes its internal routes, and the feature can't be developed, tested, or debugged without the plugin. Migrating to native in-repo Vue pages removes this dependency and puts DataFilling on equal footing with every other module.

## What Changes

- **Replace all Xpack dynamic component rendering** for DataFilling in `ChartView.vue`, `panel/App.vue`, `ShortcutTable.vue`, and `mobile/home/index.vue` with native Vue route navigation or in-repo components.
- **Build native admin form management page**: folder/form tree, form CRUD operations, folder create/rename/move/delete.
- **Build native form editor page**: visual form builder for defining fields (text, number, date, select, etc.), field settings, index configuration, datasource binding.
- **Build native table data management page**: paginated data grid with column search, inline/batch row editing, row delete, commit log viewer, and Excel template download/upload/confirm flow.
- **Build native task and sub-task management pages**: task creation with scheduling (one-time, rate-based), recipient selection (users/roles), task start/stop, sub-task list with user progress, sub-task user detail view.
- **Build native user-task fill page**: the form that end-users fill out when assigned a task, including data rows, append via Excel, submit/finish flow.
- **Add DataFilling routes and menu entries** to the main frontend router and navigation menu, replacing the current Xpack path-based routing (`/menu/data/data-filling/manage/index`, `manage/form/index`, `fill/TabPaneTable`, `fill/TabPane`).
- **Wire all embedded/panel/chart/mobile/workbranch entry points** to the new native routes instead of loading `AsyncXpackComponent` / `XpackComponent` with DataFilling jsname values.
- **BREAKING**: Any external integration that depends on the Xpack jsname-based DataFilling component paths will stop working. The base64-encoded route strings (`L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcv...`) will no longer resolve to Xpack components. Internal callers (chart embedded view, panel component switcher, workbranch shortcut tab, mobile home tab) will be rewired to native routes.

## Capabilities

### New Capabilities

- `datafilling-form-management-ui`: Admin form tree view, folder and form CRUD, datasource selection, folder operations (create/rename/move/delete).
- `datafilling-form-editor-ui`: Visual form builder/editor page for defining form fields, configuring field types and validation, setting table indexes, and binding to a datasource.
- `datafilling-table-data-ui`: Paginated data grid for viewing and managing submitted form data, column-based search/filter, inline row editing, batch operations, commit log panel, and Excel import/export workflow.
- `datafilling-task-management-ui`: Task creation and editing with scheduling options, recipient configuration, task lifecycle (start/stop/delete), sub-task listing with progress tracking, sub-task user detail view.
- `datafilling-user-fill-ui`: End-user facing fill form page for completing assigned tasks, data row entry, Excel append, submit and finish flow.
- `datafilling-route-integration`: Vue Router configuration, menu registration, and replacement of Xpack component loading in ChartView, Panel App, Workbranch ShortcutTable, and Mobile home with native route navigation.

### Modified Capabilities

(None. The existing backend specs `data-filling`, `data-filling-dml`, `data-filling-excel`, `data-filling-tasks`, `data-filling-user-tasks` define API behavior that does not change. This change is purely frontend UI.)

## Impact

**Affected frontend files (Xpack entry points to rewire):**
- `src/views/chart/ChartView.vue` - DataFilling/DataFillingEditor/DataFillingHandler component switching
- `src/pages/panel/App.vue` - panel-embedded DataFilling component switching
- `src/views/workbranch/ShortcutTable.vue` - data-filling tab and fill shortcut
- `src/views/mobile/home/index.vue` - mobile data-filling tab

**New frontend code:**
- New view files under `src/views/data-filling/` (or similar path matching existing project conventions)
- New route definitions in the router
- Menu entries for DataFilling section
- New Pinia stores if state management is needed for form editor or task pages

**Existing frontend code (unchanged API layer):**
- `src/api/datafilling.ts` - all API functions remain as-is; new pages consume them directly

**Dependencies:**
- Vue 3, Element Plus, Pinia (already in project)
- No new npm packages expected unless the form editor requires a drag-drop library

**Risk surface:**
- The Xpack jsname-based component loading is removed for DataFilling, so any Xpack plugin that still registers these components will silently stop being called. This is intentional but needs the embedded/panel entry points tested across all four surfaces.
- Mobile view may need responsive layout adjustments since the current Xpack component had its own mobile styling.
- Route path naming must not collide with existing routes.
