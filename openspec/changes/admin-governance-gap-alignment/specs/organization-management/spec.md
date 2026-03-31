## ADDED Requirements

### Requirement: Organization Administration Must Follow Frozen Official Semantics
Organization administration MUST align with the frozen official baseline for tree management, organization isolation, and delete-policy preconditions.

#### Scenario: Administrator manages organization tree
- **WHEN** an administrator creates, queries, or edits organizations
- **THEN** the organization workflow MUST preserve multi-level tree semantics and organization isolation
- **AND** the same baseline MUST govern downstream user, role, and permission workflows

### Requirement: Organization Delete Policy Must Be Deterministic and Verifiable
The system MUST define a deterministic delete policy for organizations that is compatible with child-organization constraints and documented resource disposition behavior.

#### Scenario: Administrator deletes organization under governed workflow
- **WHEN** an administrator deletes an organization from the governed administration flow
- **THEN** the system MUST reject deletion when child organizations still exist
- **AND** the system MUST apply the configured resource disposition policy consistently when deletion is allowed

### Requirement: Organization Tree API Must Remain a Canonical Administrative Source
The governed organization UI and contract checks MUST treat the organization tree response as a canonical source for organization administration.

#### Scenario: Organization administration needs hierarchical data
- **WHEN** the admin workflow needs tree-structured organization data
- **THEN** the system MUST provide a tree contract that can be consumed consistently by governed organization management flows
- **AND** the existence of a flat list path MUST NOT be treated as sufficient proof of full organization-tree parity
