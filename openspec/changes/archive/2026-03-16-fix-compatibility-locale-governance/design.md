## Context

The migrated frontend already sends `Accept-Language` on requests and derives its preferred locale from cached user language, browser locale, and a `zh-CN` default. The Go backend compatibility layer has started localizing menu responses, but the current implementation still keeps locale parsing and title dictionaries inside specific handlers, while `GET /user/info` continues to return a hardcoded `language: "zh-CN"` value.

This creates three cross-cutting problems. First, compatibility endpoints can diverge in locale precedence because each handler is free to interpret request language independently. Second, navigation bootstrap data and user bootstrap data do not currently come from the same effective-locale policy. Third, future migration work can easily reintroduce untranslated keys or fixed-language responses because the locale behavior is not governed as a reusable backend concern.

Stakeholders are the migrated Vue frontend that expects stable locale-aware bootstrap data, backend maintainers extending Java compatibility routes, and QA/release owners who need deterministic parity checks for compatibility endpoints.

## Goals / Non-Goals

**Goals:**
- Define one backend locale resolution path for compatibility-facing responses.
- Keep supported locale output limited to the currently used set: `zh-CN`, `en`, and `tw`.
- Make compatibility navigation endpoints and `/user/info` use the same effective locale and normalized language semantics.
- Reduce future regressions by moving locale parsing and fallback behavior out of ad hoc handler logic into a reusable backend component.
- Preserve existing response envelopes and endpoint shapes so frontend migration parity is not broken.

**Non-Goals:**
- Introduce a full backend i18n framework or dynamic message catalog loading.
- Localize every backend endpoint in this change; the scope is compatibility-facing bootstrap/navigation behavior first.
- Change database schema or add new persisted locale tables.
- Redesign frontend locale storage, browser locale detection, or Vue i18n usage.

## Decisions

### 1. Introduce a shared backend locale resolver for compatibility flows

Create a small reusable locale resolution helper in the Go backend that accepts request context plus optional user language input and returns a normalized effective locale.

- **Chosen approach:** shared resolver used by compatibility handlers and user bootstrap handlers.
- **Rationale:** current behavior is split between handler-local `Accept-Language` parsing and hardcoded profile output; centralizing prevents drift and gives specs a single implementation target.
- **Alternatives considered:**
  - Keep locale parsing in each handler: rejected because it guarantees repeated logic and inconsistent fallback behavior.
  - Add locale extraction into generic HTTP middleware: deferred because most routes do not need locale yet, and middleware would either over-eagerly compute locale or require broader auth/user coupling than this change needs.

The resolver contract will normalize request values such as `en-US` → `en`, `zh-TW`/`zh-HK` → `tw`, and unsupported or absent values → fallback path.

### 2. Govern locale precedence as request header first, normalized user language second, `zh-CN` last

The effective locale for compatibility responses will follow a deterministic precedence chain:
1. Supported `Accept-Language`
2. Supported normalized user language preference
3. Default `zh-CN`

- **Chosen approach:** request header wins over stored user language.
- **Rationale:** the frontend already emits `Accept-Language` on every request and updates it when the user switches locale, so using the request as first priority gives immediate behavior without waiting for persisted user state to change.
- **Alternatives considered:**
  - Stored user language first: rejected because it makes runtime locale switching lag behind the active client state.
  - Browser locale only: rejected because it bypasses explicit user-selected language already cached and sent by the frontend.

### 3. Keep localization data code-based and capability-scoped in this change

Menu title translation and language normalization tables will remain code-based in the backend for the supported locale set, but the design will move them behind reusable helpers instead of keeping them embedded in route handlers.

- **Chosen approach:** code-level dictionaries and normalization maps owned by the compatibility layer.
- **Rationale:** the current migration problem is deterministic compatibility behavior, not content-management scale. This keeps rollout small, avoids introducing a new dependency, and matches the existing frontend locale set.
- **Alternatives considered:**
  - Load frontend locale files directly from the backend: rejected because it couples Go runtime behavior to frontend source structure and build outputs.
  - Introduce an external i18n package or message store: rejected because it adds infrastructure and migration surface far beyond this governance fix.

### 4. Separate “effective locale” from “effective user language” responsibilities

Navigation endpoints will localize response labels using the effective locale, while `/user/info` will expose the normalized language value that the frontend should treat as the user bootstrap language.

- **Chosen approach:** resolve locale for rendering and normalize language for bootstrap, but use the same normalization rules underneath.
- **Rationale:** this keeps the response contract simple: menu payloads render localized text, and user bootstrap returns the locale code the frontend caches.
- **Alternatives considered:**
  - Return only locale codes and let frontend localize all navigation titles: rejected because compatibility endpoints already carry title strings and specs require backend-driven navigation metadata.

### 5. Roll out in a narrow first slice and validate with explicit locale regression coverage

Implementation will first cover:
- `GET /api/roleRouter/query`
- `GET /api/auth/menuResource`
- `GET /user/info`

Regression tests will verify:
- normalization for `zh-CN`, `en-US`, `zh-TW`
- deterministic fallback when header is absent or unsupported
- absence of raw untranslated keys in compatibility menu titles for supported locales
- `/user/info` never returning a universal hardcoded default for all callers once user-language integration is wired

- **Chosen approach:** narrow slice before wider compatibility rollout.
- **Rationale:** these endpoints bootstrap most of the frontend locale/navigation state and are the highest leverage place to establish governance.
- **Alternatives considered:**
  - Expand all compatibility endpoints at once: rejected because it increases review and regression scope before the locale policy is proven.

## Risks / Trade-offs

- **[Risk] Header-first precedence may differ from some legacy expectations** → Mitigation: document precedence explicitly in specs and keep `zh-CN` default to preserve a safe fallback path.
- **[Risk] Code-based translation tables can drift from frontend copy over time** → Mitigation: isolate dictionaries behind a shared backend localization component and add regression checks for key bootstrap endpoints.
- **[Risk] `/user/info` currently lacks real user-language sourcing in the shown implementation** → Mitigation: design the resolver to accept optional user language now and wire actual user preference retrieval as part of this change's implementation tasks.
- **[Risk] Partial rollout may leave other compatibility endpoints with older behavior temporarily** → Mitigation: scope the governance change clearly to bootstrap/navigation endpoints first and leave follow-on adoption as explicit future work.

## Migration Plan

1. Extract locale normalization and fallback logic into a shared backend helper.
2. Refactor compatibility navigation handlers to consume the shared helper instead of handler-local parsing.
3. Refactor compatibility localization dictionaries so they are no longer embedded as ad hoc per-handler logic.
4. Update `/user/info` to return normalized language sourced from request/user context rather than a fixed hardcoded value.
5. Add regression tests for locale normalization, fallback behavior, and supported-locale menu output.
6. Roll out with no API path changes and no envelope changes; rollback is limited to reverting the helper integration and restoring prior hardcoded behavior if parity issues appear.

## Open Questions

- What is the authoritative source for persisted user language during compatibility bootstrap: existing user table data, session materialization, or a later profile API enhancement?
- Should unsupported request locales always fall back to stored user language first, or do some anonymous/bootstrap compatibility flows need direct `zh-CN` fallback without user lookup?
- Do we want a follow-up change to apply the same locale resolver to non-navigation compatibility endpoints once this first slice is stable?
