# threshold-management Specification

## Purpose
Define threshold alerting requirements for managing threshold definitions, previewing threshold matching results, and querying threshold instance history in the Go backend.

## Requirements

### Requirement: Threshold Definition CRUD
The system SHALL provide create, update, and form-info loading for threshold alert definitions.

#### Scenario: Create threshold definition
- **WHEN** an authenticated user submits a `POST /threshold/save` request with a valid `ThresholdCreator` payload containing name, chart reference, resource reference, threshold rules, and message configuration
- **THEN** the system persists a new `xpack_threshold_info` record with the provided fields, default `enable=true`, derived `creator`/`creator_name`/`create_time`/`oid` from the authenticated session
- **AND** returns success with Java-compatible response envelope

#### Scenario: Update threshold definition
- **WHEN** an authenticated user submits a `POST /threshold/edit` request with a valid `ThresholdCreator` payload containing an existing threshold `id`
- **THEN** the system updates the matching `xpack_threshold_info` record with the provided fields
- **AND** returns success with Java-compatible response envelope

#### Scenario: Load form info for editing
- **WHEN** an authenticated user submits a `GET /threshold/formInfo/{id}/{resourceTable}` for an existing threshold with `resourceTable="core"`
- **THEN** the system returns a `ThresholdCreator`-shaped response containing all editable fields including recipient lists and threshold rules
- **AND** the response SHALL include `showFieldValue` and `resourceTable` metadata

#### Scenario: Create threshold with missing required fields
- **WHEN** a create request omits required fields (name, chartId, resourceId, thresholdRules)
- **THEN** the system returns an explicit non-success response
- **AND** the response distinguishes the failure from authorization denial or missing resource

#### Scenario: Load form info for non-existent threshold
- **WHEN** a formInfo request targets a threshold ID that does not exist
- **THEN** the system returns an explicit non-success response indicating resource not found
- **AND** the response remains distinguishable from permission denial

### Requirement: Threshold Listing
The system SHALL provide paginated listing of threshold definitions with filtering.

#### Scenario: List thresholds with pagination
- **WHEN** an authenticated user submits a `POST /threshold/pager/{goPage}/{pageSize}` with a `ThresholdGridRequest` body
- **THEN** the system returns paginated `ThresholdGridVO` items including id, name, resourceId, resourceType, resourceName, chartId, chartType, chartName, status, enable, creator, createName, createTime
- **AND** the response includes total count, current page, and page size matching the Java IPage contract

#### Scenario: Filter thresholds by keyword
- **WHEN** the grid request includes a non-empty `keyword`
- **THEN** the system filters results where name matches the keyword (case-insensitive contains)

#### Scenario: Filter thresholds by status and enable state
- **WHEN** the grid request includes `statusList` or `enableList`
- **THEN** the system filters results to only include thresholds matching the specified states

#### Scenario: Filter thresholds by resource type
- **WHEN** the grid request includes `resourceTypeList`
- **THEN** the system filters results to thresholds whose `resource_type` is in the provided list

#### Scenario: Filter thresholds by chart ID
- **WHEN** the grid request includes a `chartId`
- **THEN** the system filters results to thresholds associated with the specified chart

### Requirement: Threshold Enable/Disable Switch
The system SHALL support toggling threshold enable state.

#### Scenario: Enable a threshold
- **WHEN** an authenticated user submits a `POST /threshold/switch` with `enable=true` for an existing threshold
- **THEN** the system sets the threshold's `enable` field to true
- **AND** returns success

#### Scenario: Disable a threshold
- **WHEN** an authenticated user submits a `POST /threshold/switch` with `enable=false` for an existing threshold
- **THEN** the system sets the threshold's `enable` field to false
- **AND** returns success

#### Scenario: Switch on non-existent threshold
- **WHEN** a switch request targets a threshold ID that does not exist
- **THEN** the system returns an explicit non-success response

### Requirement: Threshold Deletion
The system SHALL support deleting threshold definitions.

#### Scenario: Delete thresholds by IDs
- **WHEN** an authenticated user submits a `POST /threshold/delete/{resourceTable}` with a list of threshold IDs
- **THEN** the system removes the matching `xpack_threshold_info` records
- **AND** returns success

#### Scenario: Delete thresholds for a chart
- **WHEN** an authenticated user submits a `GET /threshold/deleteWithChart/{chartId}/{resourceTable}`
- **THEN** the system removes all `xpack_threshold_info` records associated with the specified chart ID
- **AND** returns success

#### Scenario: Delete with resourceTable snapshot returns success without side effects
- **WHEN** a delete request uses `resourceTable="snapshot"`
- **THEN** the system returns success without modifying records (snapshot behavior is out of scope)

### Requirement: Batch Recipient Update
The system SHALL support updating recipients across multiple thresholds at once.

#### Scenario: Batch update recipients
- **WHEN** an authenticated user submits a `POST /threshold/batchReci` with `idList` and recipient fields (uidList, ridList, emailList, larkGroupList, larksuiteGroupList, webhookList)
- **THEN** the system updates the recipient columns on all specified threshold records
- **AND** leaves non-recipient fields unchanged

#### Scenario: Batch update with empty ID list
- **WHEN** a batchReci request has an empty `idList`
- **THEN** the system returns success without modifying any records

### Requirement: Threshold Preview and Matching Engine
The system SHALL evaluate threshold rules against chart data and return a preview of matching results.

#### Scenario: Preview threshold with matching data
- **WHEN** an authenticated user submits a `POST /threshold/preview` with a valid `ThresholdPreviewRequest` containing chartId, thresholdRules, and msgContent
- **THEN** the system fetches chart data for the specified chart
- **AND** evaluates the threshold rules (filter tree) against the chart data rows
- **AND** returns generated HTML preview content with field values from matching rows substituted into the message template
- **AND** returns the result as a string in the response data field

#### Scenario: Preview with no matching rows
- **WHEN** the threshold rules evaluated against chart data produce zero matching rows
- **THEN** the system returns an empty string (no alert content)
- **AND** does not treat this as an error condition

#### Scenario: Preview for chart with no data
- **WHEN** the referenced chart returns empty data or the chart cannot be found
- **THEN** the system returns an explicit non-success response with an error message indicating chart data is unavailable
- **AND** the response distinguishes chart-data-unavailability from malformed rules

#### Scenario: Preview with unsupported chart type
- **WHEN** the chart type does not support threshold evaluation
- **THEN** the system returns an explicit non-success response naming the unsupported chart type
- **AND** does not silently return empty content

#### Scenario: Preview with malformed threshold rules
- **WHEN** the thresholdRules JSON cannot be parsed into a valid filter tree
- **THEN** the system returns an explicit non-success response indicating malformed rules
- **AND** does not silently skip or ignore the malformed rules

#### Scenario: Rule evaluation with AND logic
- **WHEN** the filter tree specifies AND logic with multiple conditions
- **THEN** the evaluator requires all conditions to match for a row to be included

#### Scenario: Rule evaluation with OR logic
- **WHEN** the filter tree specifies OR logic with multiple conditions
- **THEN** the evaluator requires at least one condition to match for a row to be included

#### Scenario: Dynamic value resolution for numeric conditions
- **WHEN** a threshold condition references a dynamic value (min, max, average)
- **THEN** the evaluator computes the dynamic value from the actual chart data rows before comparison

### Requirement: Chart-Linked Threshold Lookup
The system SHALL support checking whether a chart has associated thresholds.

#### Scenario: Check if chart has thresholds
- **WHEN** an authenticated user submits a `GET /threshold/anyThreshold/{chartId}/{resourceTable}`
- **THEN** the system returns `true` if at least one active threshold record exists for the chart
- **AND** returns `false` otherwise

### Requirement: Threshold Instance History Listing
The system SHALL provide paginated listing of threshold execution instances.

#### Scenario: List threshold instances with pagination
- **WHEN** an authenticated user submits a `POST /threshold/instancePager/{goPage}/{pageSize}` with a `ThresholdInstanceRequest` body
- **THEN** the system returns paginated `ThresholdInstanceVO` items including id, taskId, name, execTime, status, content, msg
- **AND** the response includes total count, current page, and page size

#### Scenario: Filter instances by threshold ID
- **WHEN** the instance request includes a `thresholdId`
- **THEN** the system filters instances to only those linked to the specified threshold

#### Scenario: Filter instances by keyword
- **WHEN** the instance request includes a `keyword`
- **THEN** the system filters instances by content or name matching the keyword

### Requirement: Explicit Non-Success on Unsupported Operations
The system SHALL return explicit non-success responses for operations that fall outside the first slice scope.

#### Scenario: resourceTable snapshot is requested
- **WHEN** any endpoint receives `resourceTable="snapshot"` as a path or body parameter
- **THEN** for write operations the system returns success without side effects
- **AND** for read operations the system returns empty results
- **AND** the system does not treat snapshot requests as errors

#### Scenario: Unsupported chart/resource target for threshold
- **WHEN** a threshold references a chart type or resource type that the preview engine cannot evaluate
- **THEN** the system returns an explicit non-success response naming the unsupported target type
- **AND** the response remains distinguishable from missing-resource or authorization failures

### Requirement: Chart Data Accessor Interface
The system SHALL provide a `ThresholdChartDataAccessor` interface that ThresholdService depends on to retrieve chart data for preview evaluation.

#### Scenario: Accessor returns rows and fields for a valid chart
- **WHEN** the accessor's `GetChartDataForThreshold` method is called with a valid chart ID and resource table
- **THEN** the accessor returns a slice of row maps, a slice of FieldDTO objects, and no error
- **AND** the rows contain the actual chart data values keyed by field DataeaseName
- **AND** the FieldDTO slice contains field metadata (ID, Name, DataeaseName, DeType) for all chart fields

#### Scenario: Accessor returns error for missing chart
- **WHEN** the accessor's `GetChartDataForThreshold` method is called with a chart ID that does not exist
- **THEN** the accessor returns an error indicating the chart data is unavailable

#### Scenario: Accessor returns empty rows for chart with no data
- **WHEN** the accessor's `GetChartDataForThreshold` method is called for a chart that has no data rows
- **THEN** the accessor returns an empty row slice, the available fields, and no error

### Requirement: Preview Template Style Normalization
The system SHALL normalize preview HTML template styles by stripping blue highlight backgrounds from span elements.

#### Scenario: Preview template styles are normalized
- **WHEN** the generated preview HTML contains span elements with blue highlight background styles
- **THEN** the system strips background-color styles from template span elements before returning the HTML
- **AND** the normalized HTML preserves all other styling and content
