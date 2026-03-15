## Context

The current Go share handler exposes `create`, `validate`, `revoke`, `status`, `detail`, `switcher`, and ticket endpoints, which is enough to create and inspect share state but not enough to support the existing inline edit experience in the frontend share dialog. The Vue share components currently post to `/share/editUuid`, `/share/editExp`, and `/share/editPwd`, then rely on `/share/detail/:resourceId` to reload the resulting share state.

This change has moderate security sensitivity because it alters externally accessible share links, expiration semantics, and password protection behavior. It also needs to preserve frontend expectations around inline validation, especially for UUID edits where the UI currently uses the edit endpoint itself as the validation step.

## Goals / Non-Goals

**Goals:**
- Add Go endpoints for share UUID, expiration, and password editing.
- Preserve compatibility with the current frontend edit/reload interaction pattern.
- Define validation rules for editable UUID and password-protected sharing.
- Keep edited share state consistent with existing detail and switcher flows.

**Non-Goals:**
- Implement `/share/query` list APIs in this change.
- Implement `/share/proxyInfo` external share-access resolution in this change.
- Redesign the frontend share dialog or ticket workflow.
- Change unrelated sharing permissions, audit policy, or public-link routing structure.

## Decisions

### 1. Implement dedicated edit endpoints instead of overloading existing create/switcher flows
The frontend already issues separate edit calls for UUID, expiration, and password updates. Matching that contract in Go avoids forcing frontend rewrites and keeps share editing semantics explicit.

**Alternatives considered:**
- Fold edits into `/share/detail` or `/share/switcher`: rejected because those endpoints have different responsibilities and would blur read/toggle/edit behaviors.
- Change the frontend to use a single bulk update endpoint first: rejected because the immediate gap is backend parity for the existing UI contract.

### 2. Treat UUID edit as a validate-and-apply operation with deterministic response semantics
The frontend currently calls `/share/editUuid` during blur validation and expects response data that can be interpreted as an error message or success. The backend should therefore make UUID edit semantics deterministic: invalid or conflicting UUID values return a clear validation message, while valid values apply the update and return a success-compatible empty result.

**Alternatives considered:**
- Add a separate `/share/validateUuid` endpoint only: rejected because it would not satisfy the current frontend call pattern by itself.

### 3. Keep share detail as the source of truth after edit operations
The frontend reloads `/share/detail/:resourceId` after expiration and password updates. The new edit endpoints should therefore update the same underlying share record fields consumed by the detail response rather than maintaining separate transient state.

**Alternatives considered:**
- Return full detail payload from every edit endpoint and bypass reloads: rejected because it is unnecessary for parity and would increase response-shape coupling.

## Risks / Trade-offs

- **[Risk] UUID edits can break existing distributed share links if rules are too loose** → Mitigation: define uniqueness and format validation explicitly and fail deterministically.
- **[Risk] Password editing can weaken share security semantics if auto-generated and custom passwords are treated inconsistently** → Mitigation: specify how auto-generated versus user-supplied password updates are persisted and surfaced in detail responses.
- **[Risk] Expiration updates may create edge cases around already-expired shares** → Mitigation: require deterministic validation for invalid expiration timestamps and ensure detail reflects the persisted result.
- **[Risk] Scope creep into `/share/query` and `/share/proxyInfo`** → Mitigation: keep those endpoints explicitly out of scope for this change.
