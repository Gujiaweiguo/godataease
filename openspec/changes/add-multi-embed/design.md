## Context
DataEase already exposes embedded application management, token initialization, and iframe/DIV embedding flows. The request expands these into a single, explicit multi-dimensional embedding capability that is consistent across dashboards, data screens, datasource/dataset modules, and single charts.

## Goals / Non-Goals
- Goals:
  - Define supported embedding modes: designer, full board/screen, module pages (with tree menu), and single chart.
  - Require bidirectional parameter passing where applicable.
  - Keep authentication based on existing embedded app + JWT token flow.
  - Preserve current iframe and DIV embedding options.
- Non-Goals:
  - Redesign authentication (SSO/simulated login) beyond current embedded token flow.
  - Add new visualization interactions beyond existing drill/jump/linkage/filter features.

## Decisions
- Decision: Use a single capability spec (embedded-bi) with multiple requirements for each embedding mode.
  - Rationale: The change is cohesive and centers on embedding behavior; splitting would add overhead without clear benefit.
- Decision: Require postMessage-based bidirectional parameter passing with origin validation.
  - Rationale: Current iframe/DIV embedding relies on cross-origin messaging and is the safest interoperable mechanism.

## Risks / Trade-offs
- Risk: Broad scope could imply feature gaps for modules (datasource/dataset) if current embed entry points are limited.
  - Mitigation: Scope requirements to the embedding contract (entry, auth, and navigation), not to new module functionality.
- Risk: Bidirectional parameter passing may introduce security concerns.
  - Mitigation: Require origin allowlisting and explicit event names.

## Migration Plan
- No data migrations. Introduce spec requirements and align documentation and UI entry points with them.
- Implementation should validate existing flows first, then extend missing entry points as needed.

## Open Questions
- Whether to include SSO or simulated login as a supported embedding auth mode.
- Whether module-level embedding needs write operations or read-only access.
