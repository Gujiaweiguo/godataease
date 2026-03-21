## ADDED Requirements

### Requirement: Export Center Route-Level Hardening
The system SHALL preserve explicit export-center route and download semantics after the operational recovery batch is complete.

#### Scenario: Export-center download path remains explicit after auth
- **WHEN** a caller exercises export-center download behavior through a governed route
- **THEN** authorization, not-found, and business-state failures MUST remain explicit
- **AND** the path MUST not degrade into HTML fallback or silent success
