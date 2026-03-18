## Why

The previous stabilization change intentionally left `dataVisualization/interactiveTree` in `partial` status because the current Go implementation only derives synthetic root nodes from authorized menus. That is enough for basic compatibility fallback, but it is not full parity with the frontend workflows that expect interactive tree data to behave like a real visualization resource tree for dashboard and big-screen discovery.

## What Changes

- Replace the current synthetic `interactiveTree` behavior with governed resource-tree behavior aligned with actual visualization resources.
- Define the expected contract for `interactiveTree` payloads, including node identity, hierarchy, authorization filtering, and dashboard/screen scope semantics.
- Align frontend consumers of `queryBusiTreeApi` with the stronger tree contract so interactive BI navigation can rely on real resource nodes instead of menu-derived placeholders.
- Promote `interactiveTree` governance metadata from `partial` toward `full` once route behavior and evidence satisfy the new parity scope.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `visualization-management`: strengthen interactive resource-tree requirements for dashboard and big-screen discovery flows.
- `api-compatibility-bridge`: change `interactiveTree` governance from partial compatibility fallback to parity-backed route behavior.

## Impact

- **Backend Go**: `frontend_compat_handler.go`, visualization service/repository integration, authorization-aware resource-tree assembly, compatibility governance metadata.
- **Frontend**: `src/store/modules/interactive.ts`, `src/api/visualization/dataVisualization.ts`, and consumers relying on `queryBusiTreeApi`.
- **Testing/Gates**: backend handler/integration tests, frontend unit tests, and smoke/e2e coverage for interactive tree consumers.
