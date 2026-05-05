## ADDED Requirements

### Requirement: Paginated table data grid
The system SHALL render a paginated data grid for the selected leaf form, displaying submitted data rows with sortable columns. The grid SHALL use `searchTableData` API for data retrieval.

#### Scenario: Load table data for a form
- **WHEN** user selects a leaf form in the management tree
- **THEN** the system calls `searchTableData` with the form id, default page (1), and configured page size
- **AND** renders the returned data rows in a grid with column headers from the form field definitions
- **AND** displays pagination controls showing total count and current page

#### Scenario: Paginate through data
- **WHEN** user clicks a page number or next/previous
- **THEN** the system calls `searchTableData` with the updated currentPage and pageSize
- **AND** refreshes the grid with the new page of data

#### Scenario: Empty data state
- **WHEN** `searchTableData` returns zero rows for a form
- **THEN** the grid displays an empty-state message indicating no data has been submitted

### Requirement: Column-based search and filter
The system SHALL allow users to filter table data by column values using the searchParams mechanism of the `searchTableData` API.

#### Scenario: Filter by column value
- **WHEN** user enters a search term in a column filter input
- **THEN** the system constructs a SearchParam with the column field name and search value
- **AND** calls `searchTableData` with the search parameters appended
- **AND** the grid shows only matching rows

#### Scenario: Clear all filters
- **WHEN** user clicks "Clear Filters"
- **THEN** the system removes all search parameters and reloads the unfiltered first page

### Requirement: Row-level data editing
The system SHALL allow users to add, edit, and delete individual data rows in the table.

#### Scenario: Add new row
- **WHEN** user clicks "Add Row" and fills in field values in a row form
- **THEN** the system calls `saveRowData` with the form id and the field value map
- **AND** the new row appears in the grid after refresh

#### Scenario: Edit existing row
- **WHEN** user selects a row and modifies field values
- **THEN** the system calls `saveRowData` with the form id and updated field values including the row id
- **AND** the grid refreshes to show the updated data

#### Scenario: Delete single row
- **WHEN** user clicks delete on a row and confirms
- **THEN** the system calls `deleteRowData` with the form id and the row id
- **AND** the row is removed from the grid

#### Scenario: Batch delete rows
- **WHEN** user selects multiple rows via checkboxes and clicks "Batch Delete"
- **THEN** the system calls `batchDeleteRowData` with the form id and array of selected row ids
- **AND** all selected rows are removed from the grid

### Requirement: Commit log viewer
The system SHALL display a commit log panel showing the history of data changes for a form, using the `getCommitLogPage` API.

#### Scenario: View commit logs
- **WHEN** user opens the commit log panel for a form
- **THEN** the system calls `getCommitLogPage` and displays log entries with operator, operation type, and timestamp
- **AND** supports pagination through log entries

#### Scenario: Clear commit logs
- **WHEN** user triggers "Clear Log" with a clear type selection
- **THEN** the system calls `clearCommitLog` with the form id and clear type
- **AND** refreshes the log panel

### Requirement: Excel import and export workflow
The system SHALL support downloading an Excel template, uploading a filled Excel file, confirming the upload, and exporting form data via the existing Excel APIs.

#### Scenario: Download Excel template
- **WHEN** user clicks "Download Template"
- **THEN** the system calls `downloadExcelTemplate` with the form id
- **AND** downloads the template file to the user's browser

#### Scenario: Upload Excel file
- **WHEN** user selects an Excel file for upload
- **THEN** the system calls `uploadExcelFile` with the form id and file
- **AND** displays a preview of the parsed data for user confirmation

#### Scenario: Confirm uploaded Excel data
- **WHEN** user confirms the uploaded data preview
- **THEN** the system calls `confirmUpload` with the form id and upload id
- **AND** the data is persisted and the grid refreshes

#### Scenario: Export form data
- **WHEN** user clicks "Export Data"
- **THEN** the system calls `exportFormData` with the form id
- **AND** downloads the exported file to the user's browser

#### Scenario: Truncate all table data
- **WHEN** user triggers "Truncate" and confirms the destructive operation
- **THEN** the system calls `truncateTableData` with the form id
- **AND** the grid shows the empty state
