## ADDED Requirements

### Requirement: Visualization Entry-Chain Recovery
The system SHALL keep dashboard and big-screen entry chains recoverable as a governed stabilization surface.

#### Scenario: Visualization entry path reaches usable page state
- **WHEN** a user enters dashboard or big-screen flows from a governed menu or route entry
- **THEN** the page MUST reach a usable initialized state for in-scope list, tree, or detail workflows
- **AND** broken route or discovery behavior MUST be classified explicitly instead of appearing as generic feature absence

#### Scenario: Visualization recovery preserves discovery-path integrity
- **WHEN** a recovery fix is applied to dashboard or big-screen discovery flows
- **THEN** tree/detail/resource-discovery payloads MUST remain consumable by the frontend path that triggered the flow
- **AND** the recovered flow MUST have targeted regression or smoke coverage
