## ADDED Requirements

### Requirement: Row-Level Permission Filtering
The system SHALL support row-level permission filtering for dataset queries based on
configured row permission rules.

#### Scenario: Apply row permission filters to a query
- **WHEN** a user executes a dataset query with row permission rules configured
- **THEN** the system applies those rules to the query
- **THEN** rows outside the allowed scope are excluded from results
