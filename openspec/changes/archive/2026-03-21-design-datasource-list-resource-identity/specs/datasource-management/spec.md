## ADDED Requirements

### Requirement: Datasource List Permission Model Must Be Explicitly Defined
The system SHALL explicitly define the runtime permission semantics for datasource list endpoints and their compatibility aliases.

#### Scenario: Datasource list runtime semantics are chosen deliberately
- **WHEN** datasource list endpoints are hardened after recovery
- **THEN** the system MUST choose and document whether list behavior is filtered, scope-bound with explicit forbidden outcomes, or intentionally auth-only
- **AND** the chosen behavior MUST be justified against existing callers and compatibility paths

#### Scenario: Datasource list permission behavior remains consistent with caller expectations
- **WHEN** a client calls a datasource list route through a governed canonical or compatibility alias
- **THEN** the runtime permission behavior MUST remain consistent across those aliases
- **AND** regression coverage MUST exist for the selected permission model
