## 1. Implementation

### 1.1 Backend Foundations
- [ ] 1.1.1 Verify and extend schema for org/role/permission tables
- [ ] 1.1.2 Add seed data (default admin, roles, menus, permissions)
- [ ] 1.1.3 Implement org hierarchy and org-user associations

### 1.2 Backend Services and APIs
- [ ] 1.2.1 Implement OrgService + OrgController (CRUD, tree, members)
- [ ] 1.2.2 Implement RoleService + RoleController (CRUD, assignments, inheritance)
- [ ] 1.2.3 Implement PermService + PermController (menu/resource/export permissions)
- [ ] 1.2.4 Enforce datasource view-only permissions

### 1.3 Data Permissions
- [ ] 1.3.1 Implement column permission model (disable/mask rules)
- [ ] 1.3.2 Apply column permissions in query engine and masking output
- [ ] 1.3.3 Implement permission inheritance resolution (resources + roles)

### 1.4 Frontend Implementation
- [ ] 1.4.1 User management pages (list/create/edit/profile, bulk operations)
- [ ] 1.4.2 Organization management pages (tree, members, statistics, switching)
- [ ] 1.4.3 Permission configuration UI (menu/resource/row/column, by user/resource)
- [ ] 1.4.4 Role management UI (assign users/permissions, inheritance)

### 1.5 Integration
- [ ] 1.5.1 Permission checks on relevant controllers
- [ ] 1.5.2 Sync organization context in auth/session

### 1.6 Testing
- [ ] 1.6.1 Unit tests for services
- [ ] 1.6.2 Integration tests for permission endpoints
- [ ] 1.6.3 Row/column permission filtering tests
- [ ] 1.6.4 Cross-organization data isolation tests

### 1.7 Documentation
- [ ] 1.7.1 Update development guide
- [ ] 1.7.2 User guide for permission configuration
- [ ] 1.7.3 API documentation (Knife4j)
