## Phase 1: Backend Foundations
- [x] 1.1 Verify and extend schema for org/role/permission tables
- [x] 1.2 Add seed data (default admin, roles, menus, permissions)
- [x] 1.3 Implement org hierarchy and org-user associations
- [x] 1.4 Enforce datasource view-only permissions

## Phase 2: Backend Services and APIs
- [x] 2.1 Implement OrgService + OrgController (CRUD, tree, members)
- [x] 2.2 Implement RoleService + RoleController (CRUD, assignments, inheritance)
- [x] 2.3 Implement PermService + PermController (menu/resource/export permissions)
- [x] 2.4 Enforce datasource view-only permissions
- [x] 2.5 Integrate: permission checks in controllers, org context in auth/session

## Phase 3: Data Permissions
- [x] 3.1 Implement column permission model (disable/mask rules)
- [x] 3.2 Apply column permissions in query engine and masking output
- [x] 3.3 Implement permission inheritance resolution (resources + roles)

## Phase 4: Frontend Implementation
- [x] 4.1 User management pages (list/create/edit/profile, batch operations)
- [x] 4.2 Organization management pages (tree, members, statistics, switching)
- [x] 4.3 Permission configuration UI (menu/resource/row/column, by user/resource)
- [x] 4.4 Role management UI (assign users/permissions, inheritance)

## Phase 5: Integration
- [x] 5.1 Permission checks on relevant controllers
- [x] 5.2 Sync organization context in auth/session

## Phase 6: Testing
- [x] 6.1 Unit tests for services
- [x] 6.2 Integration tests for permission endpoints
- [x] 6.3 Row/column permission filtering tests
- [x] 6.4 Cross-organization data isolation tests

## Phase 7: Documentation
- [x] 7.1 Update development guide
- [x] 7.2 User guide for permission configuration
- [x] 7.3 API documentation (Knife4j)
