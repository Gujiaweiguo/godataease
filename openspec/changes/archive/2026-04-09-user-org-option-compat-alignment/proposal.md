## Why

The legacy compatibility endpoint `/user/org/option` currently diverges from expected user-option semantics when an org handler is present. In that branch it returns organization list data (`org.ListOrgs`) instead of user options (`user.GetUserOptions`), creating contract inconsistency for clients still depending on the legacy path.

## What Changes

- Make `/user/org/option` consistently route to `user.GetUserOptions` in compatibility bridge registration.
- Add focused compatibility bridge regression tests that lock this behavior and prevent regression to org-list wiring.
- Keep scope limited to this legacy endpoint contract decision; do not expand into `/org/mounted` response-shape migration in this slice.

## Impact

- Restores deterministic legacy endpoint semantics for `/user/org/option`.
- Reduces compatibility drift between legacy and canonical user-option behavior.
- No intended changes to canonical `/api/system/user/options` behavior or organization list endpoints.
