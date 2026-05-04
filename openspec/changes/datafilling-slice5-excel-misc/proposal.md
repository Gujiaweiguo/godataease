## Why

The DataFilling module Go migration (slices 1–4) covers form CRUD, DML, task scheduling, and user-task workflows. The remaining endpoints — Excel import/export, extra details for cross-datasource lookups, datasource column options, and form template retrieval — are still only in the legacy Java backend. These are required for feature parity and unblock the frontend's Excel upload/download workflows.

## What Changes

- **Excel template download**: Generate `.xlsx` from form field definitions using `excelize/v2`, stream to client as file download.
- **Excel upload**: Accept multipart `.xlsx` file, parse rows into `DfExcelData` (fields + row data), store parsed result server-side for confirmation.
- **Confirm upload**: Insert confirmed Excel rows into the form's physical table via existing DML provider.
- **User task confirm upload**: Same confirm flow scoped to a SubInstance (user-task context).
- **Extra details**: Cross-datasource lookup for select-type fields — query an external table by column value and return extra column data.
- **Datasource column options**: List distinct values from an external datasource table/column (separate from the form's own `listColumnData`).
- **Template by user task item**: Return the form's JSON template configuration for a given SubInstance item.
- **Data export**: Export all form table data to `.xlsx` file and stream to client.

## Capabilities

### New Capabilities
- `data-filling-excel`: Excel template download, file upload/parsing, confirm upload (both admin and user-task contexts), and data export for DataFilling forms.

### Modified Capabilities
- `data-filling`: Add extra details endpoint for cross-datasource lookups on select-type fields.
- `data-filling`: Add datasource column options endpoint (distinct values from external datasource table).
- `data-filling`: Add template retrieval endpoint (form config for user-task item).

## Impact

- **Dependencies**: `github.com/xuri/excelize/v2` (already in `go.mod` at v2.10.1, not yet imported).
- **Domain types**: New types — `DfExcelData`, `RowDataDatum`, `ExtraDetailsRequest`, `ExtraDetails`, `ExtraColumnItem`, `DatasourceOptionsRequest`, `ColumnOption`.
- **Service layer**: 8 new service methods on `DataFillingService`.
- **Handler layer**: 8 new endpoint handlers (7 on `DataFillingHandler`, 1 on `UserTaskHandler`).
- **Routes**: 8 new route registrations under `/data-filling/...`.
- **Rollback**: All changes are additive; removing the new handlers and routes restores previous behavior.
