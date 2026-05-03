## MODIFIED Requirements

### Requirement: Threshold Preview and Matching Engine
The system SHALL evaluate threshold rules against chart data and return a preview of matching results. The Preview endpoint SHALL fetch chart data through an injected accessor interface and delegate to the evaluator engine for HTML generation.

#### Scenario: Preview threshold with matching data
- **WHEN** an authenticated user submits a `POST /threshold/preview` with a valid `ThresholdPreviewRequest` containing chartId, thresholdRules, and msgContent
- **THEN** the system fetches chart data for the specified chart through the injected `ThresholdChartDataAccessor` interface
- **AND** evaluates the threshold rules (filter tree) against the chart data rows using the evaluator engine
- **AND** returns generated HTML preview content with field values from matching rows substituted into the message template
- **AND** applies template style normalization to strip blue highlight backgrounds from the output
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

#### Scenario: Preview template styles are normalized
- **WHEN** the generated preview HTML contains span elements with blue highlight background styles
- **THEN** the system strips background-color styles from template span elements before returning the HTML
- **AND** the normalized HTML preserves all other styling and content

## ADDED Requirements

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
