## Context

`dataVisualization/tree` already returns a real visualization resource tree for dashboard and big-screen management flows, while `dataVisualization/interactiveTree` currently returns only a synthetic root when a corresponding menu is authorized. The frontend interactive store tolerates this shape after the prior stabilization change, but that makes `interactiveTree` a permission-aware placeholder rather than a parity-complete visualization discovery endpoint.

The follow-up change exists to close that gap without reopening the broader migration-stability scope. The target is not “more compatibility code”; the target is to make `interactiveTree` a governed, authorization-filtered view of actual visualization resources for dashboard and screen workflows.

## Goals / Non-Goals

**Goals**
- Make `interactiveTree` return real dashboard and big-screen resource nodes rather than synthetic authorization roots.
- Preserve authorization filtering while matching the frontend’s required node contract.
- Define evidence and gate criteria for moving `interactiveTree` from `partial` to `full` in governed metadata.

**Non-Goals**
- Reworking dataset or datasource interactive loading.
- Replacing `dataVisualization/tree` or redesigning all visualization APIs.
- Expanding this change into xpack or unrelated compatibility endpoints.

## Decisions

### Decision: Build `interactiveTree` from visualization resources, not menu placeholders
The endpoint should derive its data from persisted visualization resources scoped by `busiFlag` and authorization rules, not from menu visibility alone.

### Decision: Keep the existing frontend tree contract shape
The endpoint should keep returning nodes with `id`, `pid`, `name`, `leaf`, `weight`, and `children`, because the frontend interactive store and downstream consumers already normalize that shape.

### Decision: Upgrade governance only after evidence is in place
Whitelist/governance metadata should move from `partial` to `full` only after handler, service, authorization, and smoke evidence all exist.

## Risks / Trade-offs

- **Authorization drift**: resource-tree parity could accidentally expose nodes hidden by menu rules. Mitigation: test both menu visibility and resource authorization together.
- **Contract drift**: real nodes may not carry all fields expected by the frontend. Mitigation: preserve current contract shape and add tree-shape tests.
- **Scope creep**: this could expand into full visualization listing refactors. Mitigation: constrain work to `interactiveTree` only.

## Migration Plan

1. Freeze current `interactiveTree` gaps and frontend consumer expectations.
2. Add failing backend and frontend tests that distinguish synthetic-root behavior from real resource-tree behavior.
3. Implement authorization-aware resource-tree assembly for dashboard and screen flags.
4. Update governance metadata and smoke evidence once parity is demonstrated.

## Open Questions

- Whether dataset and datasource interactive aggregation should eventually move to the same unified resource-tree assembly pattern.
- Whether `interactiveTree` should expose extra node metadata beyond the currently normalized fields once parity is complete.
