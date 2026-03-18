## Why

The Go migration already covers the main BI domains, but the current system is still unstable along the core flow from datasource to dataset to dashboard and big-screen rendering. Compatibility handlers, permission semantics, and frontend fallback logic exist, yet they are not governed as a single release-readiness surface, which makes regressions hard to detect before merge or cutover.

## What Changes

- Define a stabilization baseline for the four critical BI flow families: datasource, dataset, dashboard, and big-screen.
- Tighten Java-to-Go compatibility expectations for canonical routes, compatibility aliases, and frontend-facing compatibility endpoints used by those BI flows.
- Clarify permission behavior for core BI flows so unauthorized access, missing resources, and partially migrated features do not collapse into the same error shape.
- Govern visualization resource-tree behavior used by dashboard and big-screen workflows so tree payload regressions and route gaps are treated as release blockers.
- Convert current migration hardening from scattered fixes into an explicit spec-backed change with finer-grained execution tasks and verification gates.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `datasource-management`: strengthen datasource list and validation requirements with release-readiness and parity expectations for critical BI flows.
- `dataset-management`: strengthen dataset tree, field metadata, and preview requirements with compatibility and permission-aware stability expectations.
- `visualization-management`: tighten dashboard and big-screen tree, detail, and resource operation compatibility requirements.
- `api-compatibility-bridge`: narrow the allowed compatibility surface for core BI flows and require stricter parity, non-placeholder behavior, and release gating.
- `permission-config`: clarify permission semantics for BI routes and resource trees, especially `401`/`403`/`404` distinctions during migration.

## Impact

- **Backend Go**: router aliases, compatibility handlers, permission middleware, datasource service, dataset service, visualization resource-tree handlers, and related tests.
- **Frontend**: BI API wrappers, compatibility handling branches, and pages consuming datasource, dataset, dashboard, and big-screen resource payloads.
- **Release governance**: strict compatibility gates, contract diff expectations, shadow validation scope, and regression evidence for core BI flows.
- **OpenSpec artifacts**: delta specs for the modified capabilities above, plus design and tasks that define a low-granularity stabilization plan.
