## 1. Scaffolding and Routes

- [x] 1.1 Create directory structure `src/views/data-filling/` with subdirectories: `manage/`, `manage/form/`, `manage/data/`, `manage/task/`, `fill/`, `components/`. Add empty placeholder `index.vue` in each leaf directory.
- [x] 1.2 Create `src/views/data-filling/types.ts` with TypeScript interfaces for form field definitions, table data, tasks, sub-tasks, and user tasks — derived from the types already exported in `src/api/datafilling.ts`. Extend with UI-specific types (e.g., `FormFieldConfig` for the parsed `forms` JSON schema).
- [x] 1.3 Create `src/store/modules/data-filling.ts` Pinia store with `formTree`, `selectedNodeId`, `expandedNodeIds` state; `fetchTree` action calling `getFormTree`; and tree mutation helpers. Verify store is importable with no errors.
- [x] 1.4 Register six flat hidden routes in `src/router/index.ts`: `/data-filling-manage`, `/data-filling-editor`, `/data-filling-data`, `/data-filling-tasks`, `/data-filling-fill`, `/data-filling-fill-task`. Each lazy-imports its corresponding view component. All routes `hidden: true`, no children.
- [x] 1.5 Verify routes resolve: run `npm run dev` and confirm no console errors. Navigate to `/#/data-filling-manage` and confirm the placeholder component renders without crashing.

## 2. Shared Components

- [x] 2.1 Implement `FormSchemaRenderer.vue` — accepts a `forms` JSON string prop and a `modelValue` data object, renders each field as the appropriate Element Plus input control (el-input, el-input-number, el-date-picker, el-select) based on field type. Emit `update:modelValue` on changes. Handle unknown field types with a fallback text input.
- [x] 2.2 Implement `DataGrid.vue` — generic paginated table component. Props: `columns` (field definitions), `data` (rows), `total`, `currentPage`, `pageSize`. Emits: `page-change`, `selection-change`. Uses `el-table` with `el-pagination`. Supports checkbox selection for batch operations.
- [x] 2.3 Implement `SearchFilter.vue` — renders column filter inputs based on column definitions. Props: `columns`, `modelValue` (array of SearchParam). Emits: `update:modelValue`, `clear`. Builds SearchParam objects matching the `searchTableData` API contract.
- [x] 2.4 Verify shared components: run `npm run lint` and `npm run ts:check` on the new files. Fix any type or lint errors.

## 3. Admin Management Shell and Form Tree

- [x] 3.1 Implement `manage/FormTree.vue` — folder/form tree using `el-tree`. Calls `getFormTree` via the data-filling store on mount. Renders folder and leaf nodes with icons. Supports expand/collapse, context menu (new folder, new form, rename, move, delete). Emits `select` event with the selected node.
- [x] 3.2 Implement folder operations in FormTree: create folder (calls `createForm` with `nodeType: 'folder'`), rename (calls `renameForm`), move (calls `moveForm` via drag-and-drop or move dialog), delete (calls `deleteForm` with cascade confirmation for non-empty folders). All operations update the tree in-place without full reload.
- [x] 3.3 Implement `manage/index.vue` admin shell — two-panel layout with `el-aside` (FormTree) and `el-main` (content area). Right panel shows a summary/empty state when no form is selected, and switches to table data or task views when a leaf form is selected (controlled by `selectedNodeId` and a content-type ref). Toolbar with "New Form" button that navigates to the editor route.
- [ ] 3.4 Verify admin shell: run dev server, navigate to `/#/data-filling-manage`. Confirm tree loads, folder CRUD works, and right panel reacts to tree selection. Run `npm run lint` and `npm run ts:check`.

## 4. Form Editor Page

- [x] 4.1 Implement `manage/form/FormFieldEditor.vue` — single field configuration block. Props: `field` (FormFieldConfig), `index`. Renders field type selector (dropdown of available types), label input, required toggle, and type-specific settings (e.g., date format, decimal precision, select options with `getDatasourceOptions` / `listColumnData` lookup). Emits `update:field` and `remove`.
- [x] 4.2 Implement `manage/form/FormFieldList.vue` — list of FormFieldEditor blocks with move-up/move-down reorder buttons. "Add Field" button opens a type picker, then appends a new field with defaults. Maintains field order in a reactive array. Emits `update:fields` with the ordered array.
- [x] 4.3 Implement `manage/form/index.vue` form editor page. Reads `route.query.pid` for new-form parent and `route.query.formId` for edit mode. Loads existing form via `getFormById` if editing. Contains: FormFieldList, datasource selector (`listDatasourceList`), table name input, "use existing table" toggle (calls `getBuiltInTables` when enabled), index configuration section (hidden when useExistingTable). Save button serializes fields to JSON `forms` string and calls `createForm` or `updateForm`.
- [ ] 4.4 Verify form editor: create a new form under a folder (saves and appears in tree), edit an existing form (loads fields, modifies, saves). Confirm "use existing table" toggle shows/hides index section. Run `npm run lint` and `npm run ts:check`.

## 5. Table Data Page

- [x] 5.1 Implement `manage/data/index.vue` table data page. Reads `formId` from route query or parent component prop. Fetches form definition via `getFormById` to build column config, then fetches data via `searchTableData` with pagination. Uses `DataGrid` component. Toolbar: "Add Row", "Batch Delete", "Export Data", "Download Template", "Upload Excel", "Truncate", "Commit Log" actions.
- [x] 5.2 Implement row editing: "Add Row" opens a dialog using `FormSchemaRenderer` for input. On confirm calls `saveRowData`. Inline edit on row double-click opens the same dialog pre-populated, calls `saveRowData` with the row id. Single row delete calls `deleteRowData` with confirmation. Batch delete calls `batchDeleteRowData` for selected rows.
- [x] 5.3 Implement column search: integrate `SearchFilter` component above the grid. On filter change, construct SearchParam array and re-fetch via `searchTableData`. "Clear Filters" resets searchParams to empty and reloads page 1.
- [x] 5.4 Implement `manage/data/CommitLog.vue` — drawer component. Calls `getCommitLogPage` on open, renders log entries in a table (committer, operation, timestamp, count). Paginated. "Clear Log" button with type selector calls `clearCommitLog`.
- [x] 5.5 Implement `manage/data/ExcelUploader.vue` — upload dialog. "Download Template" calls `downloadExcelTemplate` and triggers browser download. File upload calls `uploadExcelFile`, displays parsed data preview in a table. "Confirm" calls `confirmUpload`. "Export Data" calls `exportFormData`.
- [x] 5.6 Implement "Truncate" action with double-confirmation dialog calling `truncateTableData`. After truncation, grid reloads showing empty state.
- [ ] 5.7 Verify table data page: select a form with data, paginate, search by column, add/edit/delete rows, open commit log, download/export Excel. Run `npm run lint` and `npm run ts:check`.

## 6. Task Management Pages

- [x] 6.1 Implement `manage/task/index.vue` task list page. Reads `formId` from route query or parent prop. Fetches task list via `getTaskPageList`. Renders task table with columns: name, status, schedule type, last exec status, next exec time. Toolbar: "Create Task", "Delete" (batch). Row actions: "Edit", "Start", "Stop", "Execute Now", "View Sub-tasks".
- [x] 6.2 Implement task lifecycle actions: "Start" calls `startTask`, "Stop" calls `stopTask`, "Execute Now" calls `executeNowTask`, batch "Delete" calls `deleteTasks` with confirmation. Each action refreshes the task list on success.
- [x] 6.3 Implement `manage/task/TaskEditor.vue` — dialog/drawer for creating and editing tasks. Fields: name, fillType, fitType, fitColumn, rateType, rateVal, oneTimeType, startTime, endTime, publishRangeTime, publishRangeTimeType. Recipient configuration: reciFlagList checkboxes, uidList user picker, ridList role picker. Form filter/ext settings: formExtSetting and formFilterSetting JSON inputs. Save calls `saveTask`. Edit mode loads via `getTaskInfo`.
- [x] 6.4 Implement `manage/task/SubTaskList.vue` — expandable section or drawer for a task's sub-tasks. Calls `getSubTaskPageList` with taskId. Renders sub-task table with start/end time, exec status, total/unfinished counts. Delete action calls `deleteSubTasks`. "View Users" action opens SubTaskUsers.
- [x] 6.5 Implement `manage/task/SubTaskUsers.vue` — drawer showing users assigned to a sub-task. Calls `getSubTaskUsersList` with subTaskId and type filter. Tab toggle between "Finished" and "Unfinished" re-fetches with the type parameter. Renders user name, status, finish time, data id.
- [ ] 6.6 Verify task management: create a task, edit it, start/stop it, view sub-tasks, view sub-task users, delete tasks. Run `npm run lint` and `npm run ts:check`.

## 7. User Fill Pages

- [x] 7.1 Implement `fill/index.vue` user task list page. Calls `getUserTaskPageList` on mount. Renders task cards/table with task name, form name, start/end time, status badge, progress (finishCount/totalCount), expired indicator. Search input filters by `taskName`. Clicking a task navigates to fill form. Internal `getUserTaskTodoCount` call emits loaded count for parent badge integration.
- [x] 7.2 Implement `fill/TaskFill.vue` fill form page. Reads `subTaskId` from route query. Calls `getUserTaskData` to load form definition and existing data rows. Renders form fields via `FormSchemaRenderer` for each data row. Actions: "Add Row", "Delete Row" (calls `deleteUserTaskData`), "Submit" (calls `saveUserTaskData`), Excel append via `uploadExcelFile` + `userTaskConfirmUpload`.
- [x] 7.3 Implement expiration handling: expired tasks are visually marked with a dimmed style and disabled fill action in the task list. Fill form enforces fillType — single-fill tasks prevent re-submission if already finished.
- [x] 7.4 Implement responsive styles for fill pages: use `useWindowSize` from `@vueuse/core` to detect viewport. Below 768px, task list switches from table to card layout; fill form stacks fields vertically with full-width inputs.
- [ ] 7.5 Verify fill pages: view task list, search by name, click into fill form, add/edit/delete rows, submit, test Excel append, verify expired tasks are disabled. Test at mobile viewport (320-768px). Run `npm run lint` and `npm run ts:check`.

## 8. Admin Page Integration (Wire Sub-Pages into Shell)

- [x] 8.1 Wire table data page into admin shell right panel: when a leaf form is selected, render `manage/data/index.vue` inline with the formId. Add tab or toolbar switch between "Data", "Tasks" views in the right panel.
- [x] 8.2 Wire task management page into admin shell right panel: "Tasks" tab renders `manage/task/index.vue` inline with the formId.
- [x] 8.3 Wire "New Form" action: tree context menu "New Form" on a folder navigates to `/#/data-filling-editor?pid=<folderId>`. "Edit" on a leaf form navigates to `/#/data-filling-editor?formId=<id>`. Editor "Back" button navigates to `/#/data-filling-manage`.
- [ ] 8.4 Verify full admin flow: tree CRUD → create form → edit form → view data → manage tasks. All transitions work without page reload for inline content, full navigation for editor. Run `npm run lint` and `npm run ts:check`.

## 9. Workbranch and Mobile Entry Rewiring

- [x] 9.1 In `src/views/workbranch/ShortcutTable.vue`: replace the `<XpackComponent jsname="L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvZmlsbC9UYWJQYW5lVGFibGU=" v-if="activeName === 'data-filling'" />` with `<FillTaskList v-if="activeName === 'data-filling'" />` importing `fill/index.vue` as `FillTaskList`. Remove the standalone XpackComponent that loads with `@loaded="loadedDataFilling"` — replace with a native `getUserTaskTodoCount` call that emits to the parent for badge display. Remove all DataFilling-related Xpack imports.
- [x] 9.2 In `src/views/mobile/home/index.vue`: replace the `<XpackComponent>` for data-filling tab with a direct import of `fill/index.vue`. The component renders responsively at mobile viewport (already handled by D7). Wire the `loadedDataFilling` callback to use `getUserTaskTodoCount` natively. Remove DataFilling Xpack imports.
- [ ] 9.3 Verify workbranch: open workbranch, switch to data-filling tab, confirm native task list renders, todo count badge displays. Verify mobile: open mobile home, switch to data-filling tab, confirm native fill list renders at mobile width.

## 10. ChartView and Panel Entry Rewiring

- [x] 10.1 In `src/views/chart/ChartView.vue`: replace the `name.includes('DataFilling')` conditional block (lines ~99-116). Map `'DataFilling'` → `router.push('/data-filling-manage')`, `'DataFillingEditor'` → `router.push({ path: '/data-filling-editor', query: { formId } })`, `'DataFillingHandler'` → `router.push('/data-filling-fill')`. Remove `dataFillingPath` ref and `AsyncXpackComponent` import if no longer needed.
- [x] 10.2 In `src/pages/panel/App.vue`: replace the `val.includes('DataFilling')` conditional block (lines ~62-73). Same mapping as ChartView — use `router.push` to native routes. Remove `isDataFilling` ref, `dataFillingPath` ref, and the conditional template block that renders `<component :is="currentComponent" :jsname="dataFillingPath">`.
- [ ] 10.3 Verify ChartView: trigger component switch to DataFilling names and confirm route navigation to native pages. Verify Panel: trigger panel component switch to DataFilling names and confirm route navigation.

## 11. Menu Config and Plugin Transition

- [x] 11.1 Investigate how DataFilling menu entries are currently served by the backend. Check if the Go backend returns menu config with `plugin: true` for DataFilling routes. Document the exact menu/route payload structure.
- [ ] 11.2 If the backend serves DataFilling menus with `plugin: true`, update the backend menu config to return `component: 'data-filling/manage/index'` with `plugin: false` (or absent). This ensures `establish.ts` resolves the view via the glob import to the native component.
- [ ] 11.3 If backend menu update is deferred, add a transitional remap in `establish.ts`: intercept known DataFilling plugin component paths and remap to native `data-filling/manage/index`. Mark this as a temporary measure with a TODO comment.
- [ ] 11.4 Verify menu entry: log in as admin, confirm "Data Filling" appears in navigation, click it, confirm it loads the native `manage/index.vue` instead of Xpack component.

## 12. Cleanup and Validation

- [x] 12.1 Search entire `apps/frontend/src` for all four base64 jsname strings: `L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2luZGV4`, `L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2Zvcm0vaW5kZXg=`, `L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvZmlsbC9UYWJQYW5lVGFibGU=`, `L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvZmlsbC9UYWJQYW5l`. Confirm all are removed. Remove any unused Xpack/XpackComponent imports in the modified files.
- [x] 12.2 Search for `'DataFilling'`, `'DataFillingEditor'`, `'DataFillingHandler'` string literals. Confirm none still reference Xpack component loading — all should route to native pages or have been removed.
- [x] 12.3 Run full frontend quality suite: `npm run lint`, `npm run ts:check`, `npm run lint:stylelint` in `apps/frontend`. Fix all errors and warnings in new and modified files.
- [x] 12.4 Run `npm run test:core` in `apps/frontend`. Confirm no regressions in existing tests.
- [ ] 12.5 Manual end-to-end verification: (1) Admin: tree CRUD → create/edit form → view data → manage tasks. (2) Fill: task list → fill form → submit. (3) Workbranch: data-filling tab. (4) Mobile: data-filling tab. (5) ChartView: DataFilling component switch. (6) Panel: DataFilling component switch. (7) Direct URL: navigate to each of the six routes. Document any issues found.
