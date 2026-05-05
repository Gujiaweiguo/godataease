## ADDED Requirements

### Requirement: Form tree navigation and display
The system SHALL render a native folder/form tree on the DataFilling admin management page, displaying folders and leaf forms in a hierarchical structure. The tree SHALL be fetched via the existing `getFormTree` API.

#### Scenario: Initial page load displays form tree
- **WHEN** an authorized admin user navigates to the DataFilling management page
- **THEN** the system calls `getFormTree` and renders the returned tree structure with folder nodes and leaf form nodes
- **AND** each folder node is expandable/collapsible
- **AND** each leaf form node shows the form name

#### Scenario: Empty state when no forms exist
- **WHEN** `getFormTree` returns an empty array
- **THEN** the system displays an empty-state placeholder with a prompt to create the first folder or form

### Requirement: Folder operations
The system SHALL allow admin users to create, rename, move, and delete folders within the DataFilling form tree.

#### Scenario: Create folder
- **WHEN** user clicks "New Folder" and provides a folder name
- **THEN** the system calls `createForm` with nodeType set to folder and the given name and pid
- **AND** the new folder appears in the tree at the correct position

#### Scenario: Rename folder
- **WHEN** user selects a folder and triggers rename with a new name
- **THEN** the system calls `renameForm` with the folder id and new name
- **AND** the folder label updates in the tree without a full reload

#### Scenario: Move folder or form to another parent
- **WHEN** user drags a tree node to a different parent folder
- **THEN** the system calls `moveForm` with the node id and the target parent pid
- **AND** the tree repositions the node under the new parent

#### Scenario: Delete folder
- **WHEN** user deletes a folder that contains no child forms
- **THEN** the system calls `deleteForm` with the folder id
- **AND** the folder is removed from the tree

#### Scenario: Delete folder with children prevented or confirmed
- **WHEN** user attempts to delete a folder that contains child forms or sub-folders
- **THEN** the system prompts a confirmation warning about cascading deletion of all children
- **AND** only proceeds if the user confirms

### Requirement: Form CRUD from tree context
The system SHALL allow admin users to create new leaf forms, delete forms, and navigate to the form editor from the tree.

#### Scenario: Create new form under a folder
- **WHEN** user selects a folder and clicks "New Form"
- **THEN** the system opens the native form editor page with the parent folder pid pre-set
- **AND** the form editor uses the existing `createForm` API on save

#### Scenario: Delete leaf form
- **WHEN** user selects a leaf form and triggers delete
- **THEN** the system prompts for confirmation
- **AND** on confirm calls `deleteForm` and removes the node from the tree

#### Scenario: Navigate to form editor from tree
- **WHEN** user double-clicks or selects "Edit" on a leaf form node
- **THEN** the system navigates to the native form editor route with the form id as a route parameter

### Requirement: Datasource selection for forms
The system SHALL present a datasource selector when creating or editing a form, populated by the existing `listDatasourceList` and `listDatasourceListAll` APIs.

#### Scenario: Datasource dropdown populated on form creation
- **WHEN** user opens the form editor for a new form
- **THEN** the system calls `listDatasourceList` and populates a dropdown with available datasources
- **AND** the selected datasource id is included in the `createForm` request
