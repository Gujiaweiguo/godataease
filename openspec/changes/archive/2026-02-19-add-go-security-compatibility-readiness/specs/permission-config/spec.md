## ADDED Requirements

### Requirement: Java-Parity Row Permission Execution Gate
The Go backend SHALL enforce row-level permissions with Java-equivalent filtering semantics on migration-critical query paths.

#### Scenario: Compare row-level query result with Java baseline
- **WHEN** the same role/user query is executed against parity fixtures in Java and Go
- **THEN** the visible row set returned by Go matches Java baseline exactly
- **AND** unauthorized rows are never exposed

### Requirement: Java-Parity Column Masking and Hiding Gate
The Go backend SHALL enforce column-level hide/mask rules equivalent to Java baseline behavior.

#### Scenario: Query protected dataset columns
- **WHEN** a user without full field permission queries protected columns
- **THEN** hidden columns are omitted from response
- **AND** masked columns use Java-compatible desensitization output

### Requirement: Security Regression Gate for Permission Negative Paths
The migration process SHALL run mandatory negative-path permission checks before release.

#### Scenario: Request unauthorized resource action
- **WHEN** a user attempts unauthorized read/export/modify actions
- **THEN** Go returns mapped deny semantics compatible with Java handling
- **AND** request is blocked without partial data leakage
