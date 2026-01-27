## 1. Implementation

### 1.1 Database Schema
- [x] 1.1.1 Create row permission table (data_perm_row)
- [x] 1.1.2 Create Flyway migration script for row permissions (V2.10.21__data_perm_row.sql)

### 1.2 Backend Implementation
- [x] 1.2.1 Implement DataPermRow entity, service, and controller
- [x] 1.2.2 Implement row-level permission filtering in query engine (RowPermissionFilter)

### 1.3 Frontend Implementation
- [x] 1.3.1 Integrate row permission configuration UI (existing XpackComponent)
- [x] 1.3.2 Add dataset.ts APIs for row permissions

### 1.4 Verification
- [x] 1.4.1 Maven compile success
