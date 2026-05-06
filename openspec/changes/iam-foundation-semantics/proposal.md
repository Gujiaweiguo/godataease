## Why

The governance plan has already frozen the core IAM foundation policies, but those policies are not yet represented as an executable OpenSpec change. Organization-scoped administration, built-in role baselines, last-role safety semantics, organization delete policy, and legacy compatibility routing currently span multiple modules without a single implementation contract, which creates drift risk for all downstream user/role and permission work.

## What Changes

- Lock the IAM foundation semantics for governed administration flows before lifecycle and permission-center work begins.
- Make organization context a mandatory contract for governed role and user administration write paths.
- Record the intentional last-role policy deviation: block removal of a user's last role instead of cascading user deletion.
- Define the governed organization delete policy as child rejection + soft delete + auditable deferred resource disposition.
- Freeze the compatibility-routing policy for legacy frontend/backend contracts so later changes can distinguish permanent shims, migration targets, and dual-support routes.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `organization-management`: tighten governed organization isolation semantics and make delete-policy behavior deterministic and auditable.
- `role-management`: freeze built-in role baseline, last-role safety semantics, and organization-scoped role workflows.
- `user-management`: align governed user workflows with organization-scoped membership baseline and legacy `/user/*` compatibility policy.
- `api-compatibility-bridge`: classify legacy IAM-related routes into permanent shim, frontend migration, and dual-support transition buckets.

## Impact

- Affected backend modules: `org_service.go`, `role_service.go`, `auth_service.go`, compatibility bridge handlers, and related route registration.
- Affected contracts: organization-scoped admin writes, last-role member removal behavior, organization delete workflow, and legacy `/user/*` / permission compatibility routes.
- Downstream dependency: this change becomes the implementation prerequisite for `user-role-lifecycle-alignment` and `permission-center-semantic-alignment`.
- Rollback strategy: keep enforcement changes feature-flagged or compat-layered where possible so high-risk policy rollout can fall back without losing audit and verification assets.
