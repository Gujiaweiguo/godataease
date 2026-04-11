## ADDED Requirements

### Requirement: Datasource Detail Must Return Resolved Creator Name
The system SHALL return a resolved creator user name (not just the raw user ID) in datasource detail responses.

#### Scenario: Datasource detail includes creator display name
- **WHEN** a client calls `GET /datasource/get/{id}` or `GET /datasource/hidePw/{id}` for a valid datasource
- **THEN** the response MUST include a `creator` field containing the resolved user name corresponding to `create_by`
- **AND** if the user cannot be found, the system MUST fall back to the raw `create_by` value

#### Scenario: Datasource detail includes updater display name
- **WHEN** a datasource has been updated (non-empty `update_by` field)
- **THEN** the response MUST include an `updater` field containing the resolved user name corresponding to `update_by`
- **AND** if `update_by` is empty or the user cannot be found, the system MUST return an empty string
