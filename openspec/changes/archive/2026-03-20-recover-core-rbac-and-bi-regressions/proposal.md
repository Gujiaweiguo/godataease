## Why

After the recent stabilization, compatibility, and frontend runtime changes, the user reports that previously repaired core capabilities now appear lost again. The most urgent impact is concentrated in the system-management domain: user, role, organization, menu, permission, and the menu-driven route reachability chain. Repository evidence suggests the backend feature handlers still exist, but the frontend login, dynamic route generation, permission refresh, and authorized-menu bootstrap path may now be causing broad inaccessibility that looks like feature loss.

## What Changes

- Establish a recovery matrix for the system-management domain to classify each regression as route loss, permission mismatch, API mismatch, page-init failure, or real implementation gap.
- Recover the “total gate” path first: login bootstrap, current-user initialization, authorized menu loading, dynamic route generation, route validation, and unauthorized vs. missing-route behavior.
- Recover RBAC administration workflows next: user, role, organization, menu, and permission pages plus their critical frontend/backend compatibility APIs.
- Reconcile menu reachability with RBAC expectations so first-level and second-level menus remain visible, reachable, and semantically correct after login.
- Add regression gates so the same system-management feature-loss symptom cannot silently recur.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `login-management`
- `menu-access-governance`
- `navigation-rendering`
- `user-management`
- `role-management`
- `organization-management`
- `menu-management`
- `permission-config`

## Impact

- **Frontend**: login flow, user bootstrap, permission refresh, route generation, runtime menu rendering, system-management pages, and RBAC entry paths.
- **Backend Go**: auth, user, role, org, menu, permission compatibility handlers, related middleware, and the route/permission plumbing required for system-management reachability.
- **Verification**: route/access matrix, RBAC functional smoke coverage, menu reachability coverage, and regression gates for login → menu → route → page-init health.
