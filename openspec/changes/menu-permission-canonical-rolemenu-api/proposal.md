## Why

The permission-center menu tab still calls compatibility endpoints (`/auth/menuPermission`, `/auth/saveMenuPer`) even though canonical role-menu APIs are already available and tested (`/roleMenu/auth/:roleId`, `/roleMenu/auth`).

## What Changes

- Migrate permission-center `MenuPermission.vue` role-scoped load/save flows to canonical role-menu APIs.
- Keep initial tree loading behavior intact using existing menu tree query, while removing role-scoped dependency on compatibility endpoints.
- Add focused frontend regression coverage for canonical menu auth API usage.

## Impact

- Reduces compatibility-route coupling in permission-center menu authorization.
- Moves menu auth rollout one step closer to canonical API-only usage with a minimal independent slice.
- No backend contract changes required.
