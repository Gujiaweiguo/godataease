# watermark-management Specification

## Purpose
TBD - created by archiving change add-watermark-service. Update Purpose after archive.
## Requirements
### Requirement: User can query watermark settings
The system SHALL provide an API to retrieve the current watermark configuration.

#### Scenario: Query watermark returns default when no settings exist
- **WHEN** user calls GET /watermark/find
- **THEN** system returns default watermark settings with enable=false

#### Scenario: Query watermark returns saved settings
- **WHEN** user calls GET /watermark/find after saving settings
- **THEN** system returns the previously saved watermark configuration

### Requirement: User can save watermark settings
The system SHALL provide an API to update the watermark configuration.

#### Scenario: Save watermark with valid settings
- **WHEN** user calls POST /watermark/save with valid settingContent JSON
- **THEN** system saves the settings and returns the updated watermark object

#### Scenario: Save watermark with empty content uses defaults
- **WHEN** user calls POST /watermark/save with empty settingContent
- **THEN** system saves default watermark settings

### Requirement: Watermark settings include required fields
The system SHALL support the following watermark configuration fields:
- enable: boolean to toggle watermark display
- enablePanelCustom: boolean to allow panel-level customization
- type: watermark type (userName, custom, etc.)
- content: watermark content template with variables
- watermark_color: hex color code
- watermark_x_space: horizontal spacing
- watermark_y_space: vertical spacing
- watermark_fontsize: font size

#### Scenario: Default settings are valid JSON
- **WHEN** no watermark settings exist
- **THEN** system returns valid JSON with all required fields

### Requirement: Watermark API requires authentication
The system SHALL require authentication for watermark API access.

#### Scenario: Unauthenticated request is rejected
- **WHEN** unauthenticated user calls watermark API
- **THEN** system returns 401 Unauthorized

### Requirement: Watermark uses upsert pattern
The system SHALL use an upsert pattern for watermark persistence, creating or updating a single default watermark record.

#### Scenario: Multiple saves update the same record
- **WHEN** user saves watermark settings multiple times
- **THEN** only one watermark record exists with the latest settings

### Requirement: Watermark identity compatibility endpoints remain available
The system SHALL provide compatibility user identity endpoints used by watermark preview and runtime rendering.

#### Scenario: Watermark settings page loads user person info
- **WHEN** watermark settings page requests user identity fields for watermark preview
- **THEN** backend MUST provide `/user/personInfo` with `id`, `account`, `name`, and `ip` fields

#### Scenario: Runtime watermark resolver loads current IP info
- **WHEN** runtime watermark logic requests current user/ip identity
- **THEN** backend MUST provide `/user/ipInfo` with `account`, `name`, and `ip` fields
