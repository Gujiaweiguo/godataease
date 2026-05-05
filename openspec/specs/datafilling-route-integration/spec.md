## ADDED Requirements

### Requirement: DataFilling routes in main frontend router
The system SHALL register native Vue Router routes for all DataFilling pages, replacing the Xpack jsname-based dynamic component loading.

#### Scenario: Admin management route
- **WHEN** the frontend application starts
- **THEN** the router includes a route for the DataFilling admin management page (e.g., `/data-filling/manage`)
- **AND** this route renders the native form management tree component

#### Scenario: Form editor route
- **WHEN** the frontend application starts
- **THEN** the router includes a route for the form editor (e.g., `/data-filling/manage/form/:id?`)
- **AND** this route renders the native form editor component
- **AND** the optional `:id` parameter determines new-form vs. edit-form mode

#### Scenario: Table data route
- **WHEN** the frontend application starts
- **THEN** the router includes a route for table data management (e.g., `/data-filling/manage/data/:formId`)
- **AND** this route renders the native table data grid component

#### Scenario: Task management route
- **WHEN** the frontend application starts
- **THEN** the router includes a route for task management (e.g., `/data-filling/manage/tasks/:formId`)
- **AND** this route renders the native task management component

#### Scenario: User fill route
- **WHEN** the frontend application starts
- **THEN** the router includes a route for the user fill page (e.g., `/data-filling/fill`)
- **AND** this route renders the native user task list and fill form components

### Requirement: DataFilling menu entry in navigation
The system SHALL register a DataFilling section in the main application navigation menu.

#### Scenario: DataFilling visible in navigation
- **WHEN** a user with DataFilling access loads the main navigation
- **THEN** a "Data Filling" menu item is visible under the data section
- **AND** clicking it navigates to the admin management route for admin users, or the user fill route for regular users

### Requirement: ChartView embedded entry replacement
The system SHALL replace Xpack dynamic component loading for DataFilling in the chart embedded view with native route navigation.

#### Scenario: ChartView opens DataFilling management
- **WHEN** the chart embedded view receives a component name of "DataFilling"
- **THEN** the system navigates to the native admin management route instead of loading `AsyncXpackComponent` with the jsname path

#### Scenario: ChartView opens DataFilling editor
- **WHEN** the chart embedded view receives a component name of "DataFillingEditor"
- **THEN** the system navigates to the native form editor route instead of loading `AsyncXpackComponent`

#### Scenario: ChartView opens DataFilling handler
- **WHEN** the chart embedded view receives a component name of "DataFillingHandler"
- **THEN** the system navigates to the native user fill route instead of loading `AsyncXpackComponent`

### Requirement: Panel App entry replacement
The system SHALL replace Xpack dynamic component loading for DataFilling in the panel App.vue with native route navigation.

#### Scenario: Panel switches to DataFilling management
- **WHEN** panel App.vue `changeCurrentComponent` receives "DataFilling"
- **THEN** the system navigates to the native admin management route instead of rendering `XpackComponent` with jsname

#### Scenario: Panel switches to DataFilling editor
- **WHEN** panel App.vue `changeCurrentComponent` receives "DataFillingEditor"
- **THEN** the system navigates to the native form editor route

#### Scenario: Panel switches to DataFilling handler
- **WHEN** panel App.vue `changeCurrentComponent` receives "DataFillingHandler"
- **THEN** the system navigates to the native user fill route

### Requirement: Workbranch shortcut entry replacement
The system SHALL replace the Xpack component in the workbranch ShortcutTable data-filling tab with the native user fill component.

#### Scenario: Workbranch data-filling tab renders native component
- **WHEN** user switches to the "data-filling" tab in the workbranch shortcut
- **THEN** the system renders the native user task list component instead of `XpackComponent` with the jsname path
- **AND** the loaded callback updates the todo count badge

### Requirement: Mobile home entry replacement
The system SHALL replace the Xpack component in the mobile home data-filling tab with the native user fill component.

#### Scenario: Mobile data-filling tab renders native component
- **WHEN** user switches to the "data-filling" tab in the mobile home view
- **THEN** the system renders a mobile-responsive version of the native user task list component instead of `XpackComponent`
- **AND** the loaded callback updates the todo count badge

### Requirement: Removal of DataFilling Xpack jsname paths
The system SHALL no longer reference the base64-encoded DataFilling jsname paths in any frontend code.

#### Scenario: No remaining jsname references for DataFilling
- **WHEN** the migration is complete
- **THEN** the base64 strings `L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2luZGV4`, `L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2Zvcm0vaW5kZXg=`, `L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvZmlsbC9UYWJQYW5lVGFibGU=`, and `L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvZmlsbC9UYWJQYW5l` are absent from the frontend source code
- **AND** no `AsyncXpackComponent` or `XpackComponent` rendering paths remain for DataFilling component names

### Requirement: Graceful fallback for unauthorized access
The system SHALL handle unauthorized access to DataFilling routes consistently with the rest of the application.

#### Scenario: User without DataFilling permission accesses route
- **WHEN** a user without DataFilling access navigates to a DataFilling route
- **THEN** the system redirects to an appropriate fallback page or shows an access-denied message consistent with the application's existing permission handling pattern
