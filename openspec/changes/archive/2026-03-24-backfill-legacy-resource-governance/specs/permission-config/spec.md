## MODIFIED Requirements

### Requirement: Resource permission supports governed assignment views

#### Scenario: Resource permission tab loads backfilled historical resources
- **WHEN** a historical datasource, dataset, dashboard, or big-screen resource has been backfilled into the governed resource model
- **THEN** the resource permission tab MUST expose that resource through the same governed resource-view workflow used by newly governed resources
- **AND** the UI MUST NOT require a legacy-only permission path for that resource

#### Scenario: Backfilled historical resource stays consistent across assignment perspectives
- **WHEN** an administrator inspects a backfilled historical resource from the unified permission center
- **THEN** the by-user and by-resource perspectives MUST resolve against the same effective authorization state
- **AND** switching perspectives MUST NOT produce conflicting effective grants for the same governed resource

## ADDED Requirements

### Requirement: Backfilled Historical Resources Must Converge to the Governed Authorization Model
The system SHALL bring historical resources that are backfilled into the unified resource model under the same effective authorization semantics used by newly governed resources.

#### Scenario: Historical resource receives inherited effective permission after backfill
- **WHEN** a historical datasource, dataset, dashboard, or big-screen resource is backfilled under a governed parent group
- **THEN** the system MUST calculate and expose inherited effective grants for that resource
- **AND** unified permission queries MUST treat that resource as governed instead of falling back to legacy-only type-scoped semantics

#### Scenario: Historical resource save/query round-trip uses governed state
- **WHEN** an administrator queries or saves permission changes for a backfilled historical resource
- **THEN** the system MUST persist and return authorization state through the same governed permission model used by newly governed resources
- **AND** runtime authorization checks MUST observe the same effective result after refresh

#### Scenario: Historical resource cannot be safely governed automatically
- **WHEN** a historical resource lacks a valid parent scope, governing identity, or organization boundary required by the governed model
- **THEN** the system MUST skip automatic governance for that resource
- **AND** the skip result MUST be auditable and classifiable for follow-up remediation
- **AND** the system MUST NOT pretend the resource is fully governed when it is not
