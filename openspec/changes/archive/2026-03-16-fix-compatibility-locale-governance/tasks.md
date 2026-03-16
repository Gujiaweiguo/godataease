## 1. Shared locale resolution

- [x] 1.1 Extract request locale normalization and fallback logic into a reusable backend helper for compatibility-facing handlers.
- [x] 1.2 Define the supported locale set (`zh-CN`, `en`, `tw`) and implement deterministic precedence of request header, user language, then default fallback.
- [x] 1.3 Move compatibility-localized title lookup logic behind shared locale-aware helper functions instead of keeping locale parsing embedded in individual handlers.

## 2. Compatibility navigation endpoints

- [x] 2.1 Refactor `GET /api/roleRouter/query` to use the shared locale resolver and return localized `meta.title` values for the effective locale.
- [x] 2.2 Refactor `GET /api/auth/menuResource` to use the shared locale resolver and keep fallback behavior consistent with `GET /api/roleRouter/query`.
- [x] 2.3 Verify compatibility navigation responses no longer expose raw untranslated i18n keys or fixed-language labels for supported locales.

## 3. User language bootstrap

- [x] 3.1 Identify and wire the authoritative user language source used by compatibility bootstrap flows.
- [x] 3.2 Update `GET /user/info` to return a normalized supported `language` value instead of a universal hardcoded default.
- [x] 3.3 Ensure `/user/info` and compatibility navigation endpoints use the same normalization rules even when their response payloads differ.

## 4. Regression coverage and validation

- [x] 4.1 Add automated tests for locale normalization and fallback cases, including `en-US`, `zh-TW`, missing header, and unsupported locale inputs.
- [x] 4.2 Add endpoint-level regression tests for localized responses from `/api/roleRouter/query`, `/api/auth/menuResource`, and `/user/info`.
- [x] 4.3 Run backend verification for the touched code path and document any remaining follow-up scope for non-navigation compatibility endpoints.
