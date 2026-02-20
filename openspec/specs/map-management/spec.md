# map-management Specification

## Purpose
TBD - created by archiving change implement-go-map-module. Update Purpose after archive.
## Requirements
### Requirement: Map Area Tree Query
The system SHALL provide map area tree query functionality for geographic visualization.

#### Scenario: Query world tree
- **WHEN** user requests world tree
- **THEN** the system returns hierarchical area structure
- **AND** root node is "世界村" with id "000"
- **AND** parent-child relationships are preserved

#### Scenario: Area node structure
- **WHEN** area node is returned
- **THEN** each node has id, level, name, pid, custom flag, and children
- **AND** custom flag indicates if it's a custom area

### Requirement: Map API Endpoints
The system SHALL provide the following API endpoint for map management.

#### Scenario: World tree endpoint
- **WHEN** GET /api/map/worldTree is called
- **THEN** returns root area node with nested children

