## 1. Domain Layer

- [x] 1.1 Add `RoleType` field to `RoleVO` struct
- [x] 1.2 Update `QueryRoles` method to return `[]*RoleVO, with pagination support

- [x] 1.3 Add `QueryWithPage` method to role repository

## 2. Service Layer

- [x] 2.1 Add `QueryRolesPage` method to role service
- [x] 2.2 Implement pagination logic (filter, count, query)
- [x] 2.3 Add unit tests for pagination scenarios

## 3. HTTP Handler

- [x] 3.1 Add `Page` handler
- [x] 3.2 Add handler unit tests

## 4. Router Integration
- [x] 4.1 Register `/role/page` route in router.go
- [x] 4.2 Run unit tests
- [x] 4.3 Run integration tests
- [x] 4.4 Run build and lint
  - [x] `make build` passed
  - [x] `golangci-lint run` passed (all blockers fixed)
- [ ] 4.5 Create PR
