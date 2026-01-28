# Change: Add comprehensive audit logging system

## Why
Current permission and user management system lacks audit trail functionality. Organizations need to track:
- Who performed what action
- When the action occurred
- What resources or data were affected
- Before and after state for sensitive operations

This is critical for:
- Security compliance (SOC 2, GDPR, etc.)
- Forensic analysis in case of data breaches
- User activity monitoring and accountability
- Operational auditing and troubleshooting

## What Changes

### New Specifications

#### Specification: `audit-logs`
**Purpose**: Provide comprehensive audit logging across the system to track all user actions, permission changes, data access, and system operations for security compliance and operational monitoring.

**Requirements**:
1. **AUD-LOG-001**: Audit log entry must have timestamp, user ID, action type, and target resource
2. **AUD-LOG-002**: Audit logs must be immutable once created (read-only for most users)
3. **AUDIT-LOG-003**: System must retain audit logs for minimum 90 days
4. **AUDIT-LOG-004**: Audit logs must support filtering by user, action type, date range, and resource type
5. **AUDIT-LOG-005**: Critical actions (user deletion, permission changes) must require explicit confirmation
6. **AUDIT-LOG-006**: Audit logs must be exportable in CSV/JSON format
7. **AUDIT-LOG-007**: System must log failed login attempts with IP address and timestamp
8. **AUDIT-LOG-008**: Audit logs must support real-time monitoring and alerts for suspicious activities
9. **AUDIT-LOG-009**: Permission changes must record who changed what for whom (actor, target, change type)
10. **AUDIT-LOG-010**: Data export operations must be logged with record count and format
11. **AUD-LOG-011**: Audit logs must integrate with existing user/org/role/permission system
12. **AUDIT-LOG-012**: System must provide audit log query API with pagination
13. **AUD-LOG-013**: Audit events must be categorized (USER_ACTION, PERMISSION_CHANGE, DATA_ACCESS, SYSTEM_CONFIG)

### Affected specs

- `embedded-bi`: ADD requirement for embedding access audit logging
- `permission-config`: ADD requirements for permission change audit logging
- `organization-management`: ADD requirements for org member management audit
- `user-management`: ADD requirements for user CRUD operation audit logging

## Impact

- **Affected code**:
  - New: `io.dataease.audit.*` (audit service packages)
  - New: Audit log database tables and Flyway migrations
  - New: Audit APIs in controllers
  - New: Audit interceptor for automatic logging

- **Backend services**:
  - Extend `UserService`, `OrgService`, `RoleService`, `PermissionService` with audit logging
  - Create new `AuditService` for log management

- **Frontend components**:
  - Audit log viewer page with filtering and export
  - Real-time audit dashboard for administrators
  - Audit settings and configuration page

- **Database**:
  - New tables: `de_audit_log`, `de_audit_log_detail`
  - Flyway migration: `V{version}__audit_logs.sql`

- **Breaking changes**: None (pure additive feature)

## Implementation Plan

### Phase 1: Database & Infrastructure
- [ ] Create audit log tables schema
- [ ] Create Flyway migration script
- [ ] Implement AuditService with CRUD operations
- [ ] Add audit log indexing for performance

### Phase 2: Backend Integration
- [ ] Add audit logging to UserService (create, update, delete)
- [ ] Add audit logging to OrgService (create, update, delete, member changes)
- [ ] Add audit logging to RoleService (create, update, delete, permission assignments)
- [ ] Add audit logging to PermissionService (create, update, delete)
- [ ] Implement failed login attempt logging
- [ ] Add data export operation logging

### Phase 3: API Layer
- [ ] Create AuditController with query endpoints
- [ ] Implement filtering by user, action, date range, resource
- [ ] Implement pagination (page, pageSize)
- [ ] Add export endpoint (CSV, JSON)
- [ ] Add real-time monitoring endpoint

### Phase 4: Frontend
- [ ] Create audit log viewer page
- [ ] Create audit dashboard with statistics
- [ ] Add audit log settings page
- [ ] Implement real-time alerts for suspicious activities
- [ ] Integrate with existing user/org/role/permission pages

### Phase 5: Testing & Documentation
- [ ] Unit tests for AuditService
- [ ] Integration tests for audit API
- [ ] Write API documentation
- [ ] Performance testing with large audit log volumes
- [ ] Create audit log retention policy configuration
