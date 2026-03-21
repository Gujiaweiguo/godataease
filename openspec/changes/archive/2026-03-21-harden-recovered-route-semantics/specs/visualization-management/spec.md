## ADDED Requirements

### Requirement: Visualization Detail Hardening After Recovery
The system SHALL preserve explicit detail-path semantics for dashboard and big-screen flows after the primary recovery batch is complete.

#### Scenario: Dashboard detail missing-resource behavior stays explicit at the boundary
- **WHEN** a dashboard detail request targets a missing resource
- **THEN** the frontend-facing boundary MUST preserve an explicit missing-resource response
- **AND** the response MUST remain distinguishable from permission denial

#### Scenario: Big-screen deeper detail paths remain consumable after hardening
- **WHEN** a big-screen detail or edit path is exercised beyond preview-only coverage
- **THEN** the route and detail payload MUST remain consumable by the intended frontend path
- **AND** failures MUST remain explicit instead of degrading into generic feature absence
