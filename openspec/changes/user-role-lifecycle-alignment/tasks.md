## 1. Last-Role Policy Evolution

- [ ] 1.1 Create `governance_policy` repository and service for storing org-scoped last-role policy config (BLOCK / WARN+ALLOW / CASCADE) with default BLOCK. Include DB migration for policy storage table.
- [ ] 1.2 Refactor `role_service.go` UnmountUser to read the configurable last-role policy instead of hardcoded BLOCK. Implement three-mode branching: BLOCK (reject), WARN+ALLOW (proceed + warning audit), CASCADE (proceed + disable user + critical audit).
- [ ] 1.3 Add last-role policy query and update endpoints: `GET /api/governance/last-role-policy` and `PUT /api/governance/last-role-policy`. Validate that only org-level admins can update the policy.
- [ ] 1.4 Write unit tests for all three last-role policy modes covering: policy lookup, rejection under BLOCK, warning audit under WARN+ALLOW, user disable under CASCADE, and default-to-BLOCK behavior.
- [ ] 1.5 Write integration tests for last-role policy scenarios using MySQL test database with `//go:build integration` tag.

## 2. Org-Scoped Transactional Role Assignment

- [ ] 2.1 Add `AssignRolesToUser(orgID, userID, roleIDs)` method to `role_service.go` that validates org context via `requireGovernedOrgContext()`, checks user membership, and creates associations idempotently within a single DB transaction.
- [ ] 2.2 Extend assignment method to call `bindUserToOrgBaseline` when the user is not yet a member of the target org, all within the same transaction.
- [ ] 2.3 Add role-assignment audit entries recording operation type, actor, user, role, org, and timestamp for each successful assignment.
- [ ] 2.4 Write unit tests for idempotent assignment (duplicate binding is no-op), cross-org rejection, and transactional membership+role creation.
- [ ] 2.5 Write integration tests for concurrent role assignments to the same user in the same org verifying no deadlocks and both associations succeed.

## 3. User Org Membership Transfer

- [ ] 3.1 Add `TransferUserOrg(sourceOrgID, targetOrgID, userID)` method to `org_service.go` that validates target org exists and is active, removes source-org role bindings, creates target-org membership, and assigns default built-in role, all in a single transaction.
- [ ] 3.2 Integrate last-role policy evaluation into transfer: if source-org removal hits last-role, apply the source-org's configured policy (BLOCK rejects the transfer, WARN+ALLOW/CASCADE proceed with side effects).
- [ ] 3.3 Add transfer audit entries capturing source org, target org, user, actor, assigned role, and timestamp. Ensure audit entries are queryable by admins of both orgs.
- [ ] 3.4 Add REST endpoint `POST /api/organization/transfer-user` with request validation and governed org-context enforcement.
- [ ] 3.5 Write integration tests for: successful transfer, transfer to inactive org rejected, transfer with BLOCK policy rejected, transfer with CASCADE policy disables source-org membership, and audit entry visibility for both orgs.

## 4. Role-Member Org Consistency Enforcement

- [ ] 4.1 Add org scope validation to `role_service.go` MountUsers: verify that the target role's org scope matches the request org context and that the user is a member of the same org scope.
- [ ] 4.2 Add org scope validation to `role_service.go` UnmountUser: verify the user-role association belongs to the current org scope before removal.
- [ ] 4.3 Add org scope filtering to role member query methods so results only include members within the role's org scope.
- [ ] 4.4 Write unit tests for org scope mismatch rejection on add, remove, and query operations.
- [ ] 4.5 Write integration tests verifying cross-org member operations are rejected and same-org operations succeed.

## 5. Frontend Compat Route Migration Mapping

- [ ] 5.1 Add canonical-path mapping metadata structure to the compat route registration in `compatibility_bridge_handler.go`. For each FRONTEND MIGRATION bucket route, record: legacy path, canonical path, migration status (pending / in-progress / migrated).
- [ ] 5.2 Add read-only admin endpoint `GET /api/admin/compat-route-mappings` that returns the mapping table for frontend migration tracking.
- [ ] 5.3 Write unit tests verifying mapping metadata is correctly populated for all FRONTEND MIGRATION bucket routes and the admin endpoint returns accurate data.
- [ ] 5.4 Verify no PERMANENT SHIM routes are included in the migration mapping (they are frozen as permanent per C1 decision).
