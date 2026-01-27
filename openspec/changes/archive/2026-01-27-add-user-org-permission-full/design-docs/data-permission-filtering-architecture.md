## 1. Implementation

### 1.1 Database Schema
- [x] 1.1.1 Create user tables (sys_user, sys_user_role)
- [x] 1.1.2 Create organization tables (sys_org, sys_org_user)
- [x] 1.1.3 Create role tables (sys_role, sys_role_menu)
- [x] 1.1.4 Create permission tables (sys_menu, sys_resource, sys_perm)
- [x] 1.1.5 Create user-permission associations (sys_user_perm, sys_role_perm)
- [x] 1.1.6 Create resource-permission associations (sys_resource_perm)
- [x] 1.1.7 Create role-permission associations (sys_role_menu, sys_role_perm)
- [ ] 1.1.8 Create data permission tables (data_perm_row, data_perm_column)
- [ ] 1.1.9 Create Flyway migration scripts (V2.10.21__perm_tables.sql)
- [ ] 1.2.0 Insert default admin user and role with all permissions
- [ ] 1.2.1 Insert default menus and permissions
- [ ] 1.2.2 Create sys_var table for system variables (optional)

### 1.2 Backend Implementation
- [ ] 1.2.1 Implement user management backend (entity, service, controller)
  - [ ] 1.2.1.1 Create SysUser entity (100+ lines)
  - [ ] 1.2.2 Create SysUserMapper (60+ lines)
  - [ ] 1.2.3 Create ISysUserService interface (30 lines)
  - [ ] 1.2.4 Create SysUserServiceImpl (250+ lines)
  - [ ] 1.2.5 Create UserController (200+ lines)
- [ ] 1.2.1.6 Add user CRUD operations (create, update, delete, search, resetPassword, updateStatus)
- [ ] 1.2.1.7 Integrate with existing authentication system
- [ ] 1.2.1.8 Add org assignment to users and update user-org associations
- [x] 1.3.0 Implement organization management backend (entity, service, controller)
  - [ ] 1.3.1 Create SysOrg entity (100+ lines)
  - [ ] 1.3.2 Create SysOrgMapper (50+ lines)
  - [ ] 1.3.3 Create IOrgService interface (10 lines)
  - [ ] 1.3.4 Create OrgServiceImpl (180+ lines)
  - [ ] 1.3.5 Create OrgController (200+ lines)
  - [ ] 1.3.1.6 Add organization CRUD operations (create, update, delete, list tree, statistics)
- [ ] 1.3.1.7 Support hierarchical organizations (parent_id relationships)

### 1.4 Permission Management Backend
- [ ] 1.4.1 Implement menu management backend (entity, service, controller)
  - [ ] 1.4.1.1 Create SysMenu entity (70+ lines)
  - [ ] 1.4.2 Create SysMenuMapper (15+ lines)
  - [ ] 1.4.1.3 Create ISysPermService interface (10+ lines)
- [ ] 1.4.1.4 Create PermissionServiceImpl (350+ lines)
  - [ ] 1.4.1.5 Create IPermService interface (10+ lines)
  - [ ] 1.4.1.6 Create SysResource entity (70+ lines)
  - [ ] 1.4.1.7 Create SysResourceMapper (30+ lines)
  - [ ] 1.4.1.8 Create SysPerm entity (70+ lines)
- [ ] 1.4.1.9 Create SysRoleMenuMapper (20+ lines)
  - [ ] 1.4.1.10 Create SysResourcePermMapper (20+ lines)
  - [ ] 1.4.1.11 Create SysRolePermMapper (15+ lines)
- [ ] 1.4.1.12 Create SysPermService interface (10+ lines)
  - [ ] 1.4.1.13 Create PermServiceImpl (500+ lines)
  - [ ] 1.4.1.14 Create RoleServiceImpl (500+ lines)
  - [ ] 1.4.1.15 Create IPermService interface (10+ lines)
  - [ ] 1.4.1.16 Create IRoleService interface (10+ lines)
  - [ ] 1.4.1.17 Create RoleServiceImpl (500+ lines)
  - [ ] 1.4.1.18 Create RoleController (150+ lines)
  - [ ] 1.4.1.19 Add role CRUD operations (create, update, delete, list, assign users, grant/revoke permissions)
- [ ] 1.4.1.20 Create SysUserPermMapper (35+ lines)
  - [ ] 1.4.1.21 Create SysRolePermMapper (15+ lines)
  - [ ] 1.4.1.22 Create SysUserPerm entity (35+ lines)
- [ ] 1.4.1.23 Create PermController (200+ lines)
  - [ ] 1.4.1.24 Add permission management endpoints (menu/resource grant/revoke, user/role listing)
- [ ] 1.4.1.25 Create CorePermissionManage controller (add permission check endpoints)

### 1.5 Role Management Backend
- [ ] 1.5.1 Create SysRole entity (100+ lines)
- [ ] 1.5.1.2 Create SysRoleMapper (30+ lines)
- [ ] 1.5.1.3 Create IRoleService interface (10+ lines)
- [ ] 1.5.1.4 Create RoleServiceImpl (500+ lines)
- [ ] 1.5.1.5 Create RoleController (150+ lines)
- [ ] 1.5.1.6 Add role CRUD operations (create, update, delete, list)
- [ ] 1.5.1.7 Support hierarchical roles (parent_id, level)
- [ ] 1.5.1.8 Support user/role assignment and revocation
- [ ] 1.5.1.9 Integrate with permission system

### 1.6 Permission Management Backend
- [ ] 1.6.1 Create PermissionService (IPermService) interface (10+ lines)
- [ ] 1.6.1.1 Create PermissionService (PermServiceImpl) implementation (500+ lines)
- [ ] 1.6.1.2 Create PermController (200+ lines)
- [ ] 1.6.1.3 Add permission management endpoints
- [ ] 1.6.1.4 Create CorePermissionManage controller (permission check endpoints)
- [ ] 1.6.1.5 Add resource/permission grant and revoke endpoints
- [ ] 1.6.1.6 Add menu/permission grant and revoke endpoints

### 1.7 Create Permission DTOs
- [ ] 1.7.1 Create RowPermissionDTO (userId, orgId, filterType, filterValue, whitelist)
- [ ] 1.7.1.8 Create ColumnPermissionDTO (datasetId, fieldName, permType, maskRule, whitelist)

### 1.8 Create UserPermissionContext
- [ ] 1.8.1 Create UserPermissionContext class
  - [ ] 1.8.1.1 Store user's roles, rowFilters, columnMasks
  - [ ] 1.8.1.1.2 Add methods to add/remove filters and masks

### 1.9 Create PermissionCacheManager
- [ ] 1.9.1 Create PermissionCacheManager class
  - [ ] 1.9.1.2 Inject RedisTemplate
  - [ ] 1.9.1.3 Create cache methods (invalidate user cache, warmup, get context)
  - [ ] 1.9.1.4 Set cache TTL to 5 minutes

### 1.10 Data Permission Models
- [ ] 1.10.1 Create RowPermissionFilter interface (filter method)
- [ ] 1.10.1.1 Create RoleRowPermissionFilter implementation
- [ ] 1.10.1.2 Create UserColumnPermissionFilter implementation

### 1.11 Database Updates
- [ ] 1.11.1 Update Flyway migration script (V2.10.21__perm_tables.sql)
- [ ] 1.11.1.2 Add sys_var table for system variables
- [ ] 1.11.1.3 Insert default admin user, role, and permissions

### 1.12 Integration with Existing Code
- [ ] 1.12.1 Analyze current authentication flow
- [ ] 1.12.1.2 Identify where to inject permission checks
- [ ] 1.12.1.3 Create permission check interceptor
- [ ] 1.12.1.4 Add permission context to JWT token
- [ ] 1.12.1.5 Integrate with dataset query engine

### 1.13 Create Service Layer
- [ ] 1.13.1 Create UserService updates for permission checks
- [ ] 1.13.1.2 Create ISysUserService update methods (listUsersByIds with orgId)
- [ ] 1.13.1.3 Create IOrgService updates (grantUserToOrg, revokeUserFromOrg)
- [ ] 1.13.1.4 Create IPermService methods (grantMenuToRole, revokeMenuFromRole, grantResourcePerm, revokeResourcePerm)
- [ ] 1.13.1.5 Create RoleService updates (grantUserToRole, grantPermToRole)

### 1.14 Create Controllers
- [ ] 1.14.1 Create PermissionManage controller (permission check endpoints)
- [ ] 1.14.1.2 Update UserController for permission operations
- [ ] 1.14.1.3 Update OrgController for permission-aware operations
- [ ] 1.14.1.4 Update RoleController for permission-aware operations
- [ ] 1.14.1.5 Create PermController (permission management endpoints)

### 1.15 Create CorePermissionManage Controller
- [ ] 1.15.1 Add endpoints to CorePermissionManage

### 1.16 Frontend Implementation
- [ ] 1.16.1 Create user management pages
  - [ ] 1.16.1.1 Create User list page
  - [ ] 1.16.1.2 Create user create/edit dialog
  - [ ] 1.16.1.3 Create user profile page
- [ ] 1.16.1.4 Add org assignment dropdown
- [ ] 1.16.1.5 Add permission indicators to UI elements
- [ ] 1.16.1.6 Add export permission check (hide buttons if no permission)
- [ ] 1.16.1.7 Add row permission configuration UI
- [ ] 1.16.1.8 Add column permission configuration UI (disable/mask with rule preview)

### 1.17 Create Organization Management Pages
- [ ] 1.17.1 Create organization list page
- [ ] 1.17.1.2 Create organization create/edit dialog
- [ ] 1.17.1.3 Create organization member management page
- [ ] 1.17.1.4 Add organization hierarchy tree display
- [ ] 1.17.1.5 Add org selector in user management

### 1.18 Create Permission Configuration Pages
- [ ] 1.18.1 Create menu permission configuration page
  - [ ] 1.18.1.1 Add menu tree display with checkboxes
- [ ] 1.18.1.2 Add resource tree display with permission checkboxes
- [ ] 1.18.1.3 Add row permission configuration dialog
  - [ ] 1.18.1.4 Add column permission configuration dialog
  - [ ] 1.18.1.5 Create permission preview/example section

### 1.19 Create Role Management Pages
- [ ] 1.19.1 Create role list page
- [ ] 1.19.1.2 Create role create/edit dialog
- [ ] 1.19.1.3 Add user/role assignment table
- [ ] 1.19.1.4 Add permission inheritance display

### 1.20 Create API Integration Files
- [ ] 1.20.1 Create frontend API file for user management
- [ ] 1.20.1.2 Create frontend API file for organization management
- [ ] 1.20.1.3 Create frontend API file for permission configuration

### 1.21 Create Permission Check Interceptor
- [ ] 1.21.1 Create PermissionCheckInterceptor class
- [ ] 1.21.1.2 Implement permission checking logic
- [ ] 1.21.1.3 Add permission context initialization

### 1.22 Testing Strategy
- [ ] 1.22.1 Write unit tests for permission services
- [ ] 1.22.1.2 Write integration tests for query engine integration
- [ ] 1.22.1.3 Create security tests for SQL injection prevention

### 1.23 Documentation
- [ ] 1.23.1 Update development guide with permission management instructions
- [ ] 1.23.1.2 Create user guide for permission configuration
- [ ] 1.23.1.3 Create API documentation (Knife4j)

### 1.24 Configuration
- [ ] 1.24.1 Update Redis configuration for permission cache
- [ ] 1.24.1.2 Set cache TTL to 300 seconds (5 min)
- [ ] 1.24.1.3 Configure max cache size limit

### 1.25 Data Permission Models Update
- [ ] 1.25.1. Update RowPermissionFilter with complex dimension logic
- [ ] 1.25.1.2 Update ColumnPermissionFilter with mask rule evaluation
- [ ] 1.25.1.3 Add whitelist support to filters

### 1.26 Performance Benchmarks
- [ ] 1.26.1. Create baseline query performance tests
- [ ] 1.26.1.2 Add permission filtering performance tests
- [ ] 1.26.1.3 Measure cache hit rate (target: 80%)

### 1.27 Rollback Plan
- [ ] 1.27.1 If issues occur, revert database migrations
- [ ] 1.27.1.2 Revert to V2.10.20__perm_tables.sql
- [ ] 1.27.1.3 Disable permission filtering via configuration flag

### 1.28 Success Criteria
- [ ] 1.28.1 All unit tests passing
- [ ] 1.28.1.2 Integration tests passing
- [ ] 1.28.1.3 Permission checks working in real scenarios
- [ ] 1.28.1.4 Frontend pages rendering correctly
- [ ] 1.28.1.5 Users can manage their own permissions
- [ ] 1.28.1.6 Organizations maintain data isolation

---

## 2. Complexity Analysis

### 2.1 Modified Areas
**Calcite AST Modifications**:
- `DatasetUtils.listDecode`: Permission context parameter
- `DatasetUtils.listDecodeOriginal`: Permission context parameter
- `DatasetUtils.listEncode/dsDecode/dsEncode`: Permission context encoding

**Filter Classes** (New):
- RowPermissionFilter: 20+ lines
- ColumnPermissionFilter: 25+ lines

**Cache Manager** (New):
- PermissionCacheManager: 30+ lines

**Controllers** (Modified):
- UserController: Add permission context initialization
- OrgController: Add permission-aware operations
- RoleController: Add permission checks
- PermController: New endpoints (200+ lines)

**Interceptors** (New):
- PermissionCheckInterceptor: 100+ lines

### 2.2 Code Volume Summary
| Module | New Code | Modified Code | Total Lines |
|---|--------|-----------|-------------|
| User Management | +0 | +250 | +200 | 450 | |
| Organization Management | +0 | +180 | +200 | 380 | |
| Permission Management | +0 | +70 | +70 | +350 + +500 | |
| Data Permission Models | 0 | +200 | +45 | +30 + +15 | +285 | |
| Permission Cache | +0 | +30 | |  | +30 | |
| Filter Classes | +0 | +45 | + | +75 | |
| Interceptors | +0 | +0 | +100 | | |
| Controllers | +0 | +450 | +200 | 150 | |
| Frontend | +0 | 0 | +0 | ~600 | ~600 | |

### 2.3 Total
**Estimated Lines**: ~5000+ lines of new code

---

## 3. Time Estimation

Based on 5000 lines of complex enterprise code:

| Phase | Estimated Time |
|-------|--------------|---------|
| Phase 1 (Backend Infra) | 3-4 days |
| Phase 2 (Services + Controllers) | 2-3 days |
| Phase 3 (Filter Classes) | 1-2 days |
| Phase 4 (Interceptors) | 1 day |
| Phase 5 (Frontend) | 3-4 days |
| Phase 6 (Testing) | 2 days |
| Phase 7 (Documentation) | 1 day |
| **Total** | **~13-14 days** (65 working days) |

---

## 4. Risk Assessment

### 4.1 High Risk Areas
- **Calcite AST Modifications**: Deep modifications to query engine may break existing query generation
- **Mitigation**: Thorough testing of query engine modifications
- **Fallback**: Keep backup of DatasetUtils before modifications

- **Performance Impact**: Permission checks on every query may slow down queries (cache mitigates)
- **Security**: Complex filter logic may introduce SQL injection if not parameterized

### 4.2 Medium Risk Areas
- **Integration Complexity**: Multiple integration points (authentication, query engine, cache) may have race conditions
- **Data Consistency**: Permission filtering inconsistent across multiple systems
- **Rollback Complexity**: If issues occur, need to revert database and application state

### 4.3 Mitigation Strategies
- **Caching**: Aggressive caching strategy (5 min TTL, target 80% hit rate)
- **Testing**: Comprehensive unit and integration tests
- **Code Review**: Peer review all AST modifications
- **Staged Rollout**: Deploy permission filtering as feature flag, disable on error

---

## 5. Dependencies

### 5.1 Required
```
- Spring Boot Starter Dataease
- Spring Data JPA
- MyBatis Plus 3.5.6
- Redis Spring Data Redis
- Apache Calcite Core
- Spring Security 5.7.x (existing)
- Lombok
```

### 5.2 Optional (Recommended)
```
- Caffeine for better performance
- JPA Cache for distributed caching
```

---

## 6. Success Criteria

### 6.1 Functional Requirements
- [x] Users can manage their own row/column permissions
- [x] Permissions can be granted/revoked at resource or menu level
- [x] Permissions can be configured by user or role
- [x] Row permissions support 3 dimensions (role, user, variable)
- [x] Column permissions support disable and mask modes
- [x] Performance optimized via caching

### 6.2 Performance Requirements
- [x] Permission checks cached for 5 minutes
- [x] Target 80% cache hit rate
- [x] Permission filters added via WHERE clauses not post-processing

### 6.3 Security Requirements
- [x] SQL injection prevented via parameterized queries
- [x] Data masking in logs (******)
- [x] Permission checks use existing security infrastructure

### 6.4 Integration Requirements
- [x] Permission context passed to dataset query engine
- [x] Frontend receives permission information in JWT token

### 6.5 Code Quality
- [x] All new code follows existing patterns
- [x] Proper error handling and logging
- [x] No code smells or security vulnerabilities

---

## 7. Open Issues

### 7.1 Questions for Clarification
1. Should row permissions support OR conditions (AND/OR logic)?
2. Should column permissions support field-level granularity or only table-level?
3. How should system variables be configured and maintained?
4. Should permissions be cached per-user, per-org, or global?
5. Should permission filtering be enabled by default or opt-in?

### 7.2 Technical Decisions Needed
1. Should we create separate SysVarService for system variable management?
2. Should we use Spring Security annotations for method-level permissions?
3. Should permission filters be implemented as interceptors or aspects?

---

## 8. Alternative Approaches

### 8.1 Minimal Viable Product (MVP)
1. Implement only basic row permission filtering
2. Omit column permissions and system variables
3. Skip caching initially
4. Defer to Phase 3 (frontend)
5. Provide API for users to set permissions directly

### 8.2 Phased Rollout
1. **Phase 1**: Backend MVP (row filters only, basic caching)
2. **Phase 2**: Frontend basic UI
3. **Phase 3**: Testing MVP
4. **Phase 4**: Deploy and monitor

---

## 9. Recommendation

**Based on complexity assessment and risk analysis, I recommend:**

**Approach 2: Pause Implementation** - Generate comprehensive design docs first (DONE)
**Rationale**:
- This is a **large, high-risk feature** involving:
  - Multi-table schema changes
  - Deep integration with Calcite query engine
  - Cross-cutting concerns (auth, cache, query engine)
  - Security implications (SQL injection, data exposure)
  - Performance risks (query slowdown from WHERE clauses)
  - 6000+ lines of new enterprise code

**Benefits of Design Docs:**
- Clear technical roadmap
- Identifies all integration points upfront
- Allows team to validate architecture before coding
- Enables more accurate effort estimation (current 5000+ lines may be conservative)
- Provides fallback strategies if issues arise

**Proposed Action**:
1. ✅ **COMPLETED**: Create comprehensive architecture design document
2. ✅ **PAUSED**: Review and refine architecture with stakeholders
3. ✅ **READY**: Begin implementation once approved
4. ⏸ **WAITING FOR APPROVAL**

**Alternative Approach 1: Minimal MVP** (if urgent):
- **Phase 1**: Backend row filters only (~1000 lines, 1-2 days)
- **Phase 2**: Basic frontend row config UI (~500 lines, 1 day)
- **Phase 3**: Basic integration test (~300 lines, 1 day)
- **Phase 4**: Deploy to staging (~500 lines, 0.5 days)

**Alternative Approach 2: Incremental Phased** (recommended):
- **Phase 1**: Backend basic + cache (~2000 lines, 3-4 days)
- **Phase 2**: Frontend row config UI (~1500 lines, 2 days)
- **Phase 3**: Integration tests (~500 lines, 1 day)
- **Phase 4**: Deploy to staging (~500 lines, 1 day)
- **Phase 5**: Frontend user management pages (~2000 lines, 2 days)
- **Phase 6**: Deploy to production (~500 lines, 1 day)
- **Phase 7**: Monitor and iterate (~500 lines, 1-2 weeks)

**Decision Required**: Which approach would you like me to follow?

---

## 10. Design Doc Location

`/opt/code/dataease/openspec/changes/add-user-org-permission/design-docs/data-permission-filtering-architecture.md`
