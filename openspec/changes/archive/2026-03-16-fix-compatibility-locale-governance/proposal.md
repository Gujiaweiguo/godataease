## Why

Current Go compatibility endpoints still mix hardcoded Chinese labels, untranslated i18n keys, and fixed language defaults. This causes navigation and profile responses to drift from the frontend's active locale, makes migration parity harder to verify, and leaves each compatibility endpoint free to invent its own locale fallback behavior.

## What Changes

- Define a governed locale resolution policy for compatibility-facing responses, including supported locale normalization, request-language handling, and deterministic fallback behavior.
- Require compatibility navigation endpoints to return localized menu titles instead of hardcoded Chinese strings or raw i18n keys.
- Align user profile compatibility responses to expose normalized language values instead of fixed defaults so frontend locale state reflects backend truth.
- Add regression coverage for locale-sensitive compatibility endpoints and fallback scenarios so future migration work cannot silently reintroduce untranslated or mismatched titles.

## Capabilities

### New Capabilities
- `locale-resolution`: Defines supported locale normalization, precedence, and fallback rules for compatibility responses consumed by the migrated frontend.

### Modified Capabilities
- `api-compatibility-bridge`: Compatibility endpoints must preserve locale-aware response semantics for migrated frontend routes.
- `navigation-rendering`: Backend-driven navigation responses must provide localized menu metadata consistently across top and side navigation.
- `user-management`: User profile responses must expose the effective language preference using normalized locale semantics instead of hardcoded defaults.

## Impact

- Affected specs:
  - New: `locale-resolution`
  - Modified: `api-compatibility-bridge`, `navigation-rendering`, `user-management`
- Affected APIs and responses:
  - `GET /api/roleRouter/query`
  - `GET /api/auth/menuResource`
  - `GET /user/info`
- Affected code:
  - `apps/backend-go/internal/transport/http/handler/frontend_compat_handler.go`
  - `apps/backend-go/internal/transport/http/handler/user_handler.go`
  - related compatibility and handler regression tests
- Affected systems:
  - frontend locale propagation via `Accept-Language`
  - backend compatibility response shaping for navigation and user profile bootstrap data
