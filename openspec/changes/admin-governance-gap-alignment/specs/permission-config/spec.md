## ADDED Requirements

### Requirement: Permission Center Alignment Must Follow Semantic Sequencing
The governed permission-alignment workflow MUST sequence menu/resource alignment before row/column and deferred P2 expansion.

#### Scenario: Team sequences permission-center work
- **WHEN** maintainers execute permission-center alignment
- **THEN** menu and resource authorization consistency MUST be stabilized before row/column and deferred P2 work is treated as complete
- **AND** deferred items MUST NOT block already-approved P0/P1 semantic corrections

### Requirement: User-View and Resource-View Permission Workflows Must Converge
The system MUST keep by-user and by-resource permission workflows semantically consistent for governed resources.

#### Scenario: Administrator compares two governed permission perspectives
- **WHEN** an administrator inspects the same governed permission state from user-view and resource-view workflows
- **THEN** both workflows MUST resolve to the same effective authorization result
- **AND** target-query or target-save gaps MUST be classified as incomplete rather than treated as equivalent behavior

### Requirement: Row and Column Permissions Must Enforce Governed Runtime Behavior
The system MUST treat row filters, disabled columns, and masked columns as runtime-enforced governed behavior, and MUST explicitly trace or defer whitelist and system-variable dimensions until they are implemented.

#### Scenario: Governed data-access rule is evaluated at runtime
- **WHEN** a governed dataset request is evaluated against row or column permissions
- **THEN** the system MUST apply the configured enforcement semantics instead of relying on placeholder middleware or UI-only interpretation
- **AND** whitelist and system-variable dimensions MUST be either traceable in the governed permission model or explicitly marked as deferred instead of being silently accepted
