## 1. Scope freeze and gap inventory

- [x] 1.1 Document the current synthetic-root behavior of `dataVisualization/interactiveTree` with concrete backend and frontend references.
- [x] 1.2 Map all frontend consumers of `queryBusiTreeApi` and the exact node fields they require.
- [x] 1.3 Identify how dashboard and big-screen resources are currently loaded through `dataVisualization/tree` and where that differs from `interactiveTree`.

## 2. Backend regression first

- [x] 2.1 Add failing handler or integration coverage proving `interactiveTree` currently returns synthetic roots instead of real dashboard nodes.
- [x] 2.2 Add failing coverage for `dataV` interactive tree parity.
- [x] 2.3 Add failing authorization coverage for filtered visualization nodes in interactive tree responses.
- [x] 2.4 Add failing contract-shape coverage for `id`, `pid`, `leaf`, `weight`, and `children` on interactive tree nodes.

## 3. Backend parity implementation

- [x] 3.1 Refactor interactive tree assembly to use visualization resource data instead of menu-derived placeholders.
- [x] 3.2 Scope returned nodes correctly for dashboard vs. big-screen requests.
- [x] 3.3 Preserve authorization filtering without corrupting parent/child relationships.
- [x] 3.4 Ensure unauthorized or empty scopes return deterministic empty-tree behavior instead of synthetic success placeholders.

## 4. Frontend compatibility convergence

- [x] 4.1 Update frontend assumptions around interactive tree payloads only where real-resource parity changes the behavior.
- [x] 4.2 Add or update unit tests for interactive store consumers that rely on `queryBusiTreeApi`.
- [x] 4.3 Verify no frontend caller still depends on synthetic-root-only semantics.

## 5. Governance and evidence

- [x] 5.1 Update whitelist/matrix metadata for `dataVisualization/interactiveTree` once parity is achieved.
- [x] 5.2 Add regression evidence showing dashboard and big-screen interactive trees now return real resource nodes.
- [x] 5.3 Re-run relevant compatibility governance checks and drift checks.

## 6. Final verification

- [x] 6.1 Run backend tests covering interactive tree handler/service behavior.
- [x] 6.2 Run frontend unit tests covering interactive tree consumers.
- [x] 6.3 Run targeted smoke/e2e verification for dashboard and big-screen interactive entry paths.
- [x] 6.4 Document any remaining gaps before marking the change implementation-ready.
