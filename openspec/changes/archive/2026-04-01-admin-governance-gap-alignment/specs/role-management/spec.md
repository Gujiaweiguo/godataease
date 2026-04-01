## ADDED Requirements

### Requirement: Role Semantics Must Reuse Foundation Organization Context
The system MUST execute governed role workflows using the same organization context established by IAM foundation semantics.

#### Scenario: Administrator manages roles within an organization
- **WHEN** an administrator queries, creates, edits, or assigns members to roles
- **THEN** the workflow MUST consume the current governed organization context
- **AND** role member discovery MUST remain consistent with that context across user-management and permission flows

### Requirement: Last-Role Policy Must Be Deterministic and Explicitly Recorded
The system MUST define and document a deterministic last-role policy before governed role-member removal is considered complete.

#### Scenario: Administrator removes a user's last governed role
- **WHEN** an administrator removes the user's last remaining role under the governed semantics
- **THEN** the system MUST execute the configured policy deterministically
- **AND** the chosen policy MUST be explicit, testable, and traceable in governance records

### Requirement: Custom Role Grants Must Respect Inherited Permission Ceilings
The system MUST validate that custom roles cannot exceed the permissions of their allowed inherited role baseline.

#### Scenario: Administrator attempts to exceed inherited grants
- **WHEN** an administrator configures grants beyond a custom role's inherited ceiling
- **THEN** the system MUST reject the request with a governed validation error
- **AND** role-management and permission-center workflows MUST agree on the same effective ceiling
