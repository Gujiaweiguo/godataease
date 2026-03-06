# API Contract Baseline Mapping

## 1. User Import Endpoints

| Endpoint | Method | Baseline File | Status | Priority |
|----------|--------|---------------|--------|----------|
| `/de2api/user/batchImport` | POST | user/user_batchImport_POST.json | MISSING | P0 |
| `/de2api/user/excelTemplate` | POST | user/user_excelTemplate_POST.json | MISSING | P1 |
| `/de2api/user/errorRecord/{key}` | GET | user/user_errorRecord_key_GET.json | MISSING | P1 |
| `/de2api/user/clearErrorRecord/{key}` | GET | user/user_clearErrorRecord_key_GET.json | MISSING | P2 |

## 2. Reset Password Endpoints

| Endpoint | Method | Baseline File | Status | Priority |
|----------|--------|---------------|--------|----------|
| `/de2api/user/resetPwd/{uid}` | POST | user/user_resetPwd_uid_POST.json | MISSING | P0 |
| `/de2api/user/defaultPwd` | GET | user/user_defaultPwd_GET.json | MISSING | P1 |

## 3. Role Member Management Endpoints

| Endpoint | Method | Baseline File | Status | Priority |
|----------|--------|---------------|--------|----------|
| `/de2api/role/mountUser` | POST | role/role_mountUser_POST.json | MISSING | P0 |
| `/de2api/role/mountExternalUser` | POST | role/role_mountExternalUser_POST.json | MISSING | P0 |
| `/de2api/role/unMountUser` | POST | role/role_unMountUser_POST.json | MISSING | P0 |
| `/de2api/role/beforeUnmountInfo` | POST | role/role_beforeUnmountInfo_POST.json | MISSING | P1 |
| `/de2api/role/searchExternalUser/{keyword}` | GET | role/role_searchExternalUser_keyword_GET.json | MISSING | P1 |

## 4. Role CRUD Endpoints

| Endpoint | Method | Baseline File | Status | Priority |
|----------|--------|---------------|--------|----------|
| `/de2api/role/create` | POST | role/role_create_POST.json | MISSING | P0 |
| `/de2api/role/edit` | POST | role/role_edit_POST.json | MISSING | P0 |
| `/de2api/role/delete/{id}` | POST | role/role_delete_id_POST.json | MISSING | P0 |
| `/de2api/role/query` | POST | role/role_query_POST.json | MISSING | P1 |
| `/de2api/role/detail/{id}` | GET | role/role_detail_id_GET.json | MISSING | P1 |

## 5. Permission Configuration Endpoints

| Endpoint | Method | Baseline File | Status | Priority |
|----------|--------|---------------|--------|----------|
| `/de2api/system/permission/list` | POST | permission/permission_list_POST.json | MISSING | P0 |
| `/de2api/system/permission/create` | POST | permission/permission_create_POST.json | MISSING | P0 |
| `/de2api/system/permission/update` | POST | permission/permission_update_POST.json | MISSING | P0 |
| `/de2api/system/permission/delete/{permId}` | POST | permission/permission_delete_permId_POST.json | MISSING | P0 |
| `/de2api/system/role/permission/save` | POST | permission/role_permission_save_POST.json | MISSING | P0 |

## 6. Organization Endpoints

| Endpoint | Method | Baseline File | Status | Priority |
|----------|--------|---------------|--------|----------|
| `/de2api/org/tree` | GET | org/org_tree_GET.json | MISSING | P0 |
| `/de2api/org/create` | POST | org/org_create_POST.json | MISSING | P0 |
| `/de2api/org/update` | POST | org/org_update_POST.json | MISSING | P0 |
| `/de2api/org/delete/{orgId}` | POST | org/org_delete_orgId_POST.json | MISSING | P0 |

## 7. Test Coverage Matrix

| Feature | Unit Tests | Integration Tests | Contract Tests | Audit Tests |
|---------|------------|-------------------|----------------|-------------|
| User Import (partial success) | EXISTS | EXISTS | MISSING | MISSING |
| User Import (10MB limit) | EXISTS | MISSING | MISSING | N/A |
| Reset Password | N/A | EXISTS | MISSING | MISSING |
| Add Role Member (org user) | MISSING | MISSING | MISSING | MISSING |
| Add Role Member (external) | MISSING | MISSING | MISSING | MISSING |
| Remove Role Member | MISSING | MISSING | MISSING | MISSING |
| Last Role Safety | EXISTS | MISSING | MISSING | MISSING |
| Role Inheritance | EXISTS | MISSING | MISSING | N/A |
| Permission Dual-View | EXISTS | MISSING | MISSING | N/A |
| Organization Delete | EXISTS | EXISTS | MISSING | EXISTS |

## 8. Field Mapping: Frontend <-> Backend

### Role DTO Mapping

| Frontend Field | Backend RoleCreator | Backend RoleEditor | Notes |
|----------------|---------------------|-------------------|-------|
| `roleId` | - | `id` | Rename required |
| `roleName` | `name` | `name` | Rename required |
| `roleKey` | `typeCode` | - | Different semantics |
| `roleDesc` | `desc` | `desc` | Rename required |
| `status` | - | - | Not supported in DTO |

### Permission DTO Mapping

| Frontend Field | Backend PermCreateRequest | Backend PermUpdateRequest | Notes |
|----------------|---------------------------|---------------------------|-------|
| `permId` | - | `permId` | OK |
| `permName` | `permName` | `permName` | OK |
| `permKey` | `permKey` | `permKey` | OK |
| `permType` | `permType` | `permType` | OK |
| `permDesc` | `permDesc` | `permDesc` | OK |
| `parentId` | - | - | NOT SUPPORTED |
| `status` | `status` | `status` | OK |

## 9. Route Mapping: Frontend <-> Backend

### Role Routes

| Frontend API | Frontend Path | Backend Path | Status |
|--------------|---------------|--------------|--------|
| `roleCreateApi` | `/system/role/create` | `/role/create` | MISMATCH |
| `roleUpdateApi` | `/system/role/update` | `/role/edit` | MISMATCH |
| `roleDeleteApi` | `/system/role/delete/:roleId` | `/role/delete/:id` | MISMATCH |

### Permission Routes

| Frontend API | Frontend Path | Backend Path | Status |
|--------------|---------------|--------------|--------|
| `permListApi` | `/system/permission/list` | `/system/permission/list` | OK |
| `permCreateApi` | `/system/permission/create` | `/system/permission/create` | OK |
| `permUpdateApi` | `/system/permission/update` | `/system/permission/update` | OK |
| `permDeleteApi` | `/system/permission/delete/:id` | `/system/permission/delete/:permId` | MISMATCH |
