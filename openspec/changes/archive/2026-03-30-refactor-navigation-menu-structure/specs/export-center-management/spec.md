## ADDED Requirements

### Requirement: Export center remains reachable through governed toolbox navigation
The system SHALL keep export-center functionality reachable through the governed Toolbox menu after shell navigation restructuring.

#### Scenario: Authorized user opens export center from toolbox child menu
- **WHEN** an authorized user expands Toolbox and selects Data Export Center
- **THEN** the system MUST open the export-center workflow using the same business semantics as before the restructure
- **AND** the workflow MUST remain reachable without depending on a header More menu entry

#### Scenario: Export-center access respects menu authorization after restructure
- **WHEN** a role loses authorization to the export-center child menu or its toolbox parent
- **THEN** the export-center entry MUST disappear from governed navigation on next authorization refresh
- **AND** retained users with authorization MUST continue to reach the workflow through Toolbox
