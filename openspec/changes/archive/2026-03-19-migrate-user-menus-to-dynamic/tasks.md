# Tasks: Migrate User Menus to Dynamic

## 1. Database Migration

- [x] 1.1 Create database migration script to add menu_location, menu_type, action_config fields
- [x] 1.2 Insert initial menu data (about, change-password, system-settings, help-docs, forum, blog, enterprise-trial)

## 2. Backend Implementation

- [x] 2.1 Extend CoreMenu and MenuVO domain models with new fields
- [x] 2.2 Extend MenuService to support filtering by menu_location
- [x] 2.3 Extend FrontendCompatHandler.GetRoleRouters to return new fields
- [x] 2.4 Run backend integration tests

## 3. Frontend Implementation

- [x] 3.1 Create menu-actions.ts for event handlers (open-about-dialog, user-logout, etc.)
- [x] 3.2 Extend permissionStore for user menu state management
- [x] 3.3 Refactor AccountOperator.vue to load dynamic menus
- [x] 3.4 Refactor MoreMenu.vue to load dynamic help links

## 4. Testing

- [x] 4.1 Create E2E tests for user menu
- [x] 4.2 Create E2E tests for help menu
- [x] 4.3 Run regression tests
- [x] 4.4 Rebuild frontend and verify in container

## 5. Documentation

- [x] 5.1 Update menu management UI to support new fields (optional)
