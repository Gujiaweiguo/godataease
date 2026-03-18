## Context

The current interactive aggregate view in the frontend store is only partially unified. Dashboard and screen now use the governed batched interactive path, while dataset and datasource still fall back to dedicated endpoints through `getDatasetTree` and `listDatasources`. The result works, but it leaves the aggregate discovery story split between two different backend behaviors and two different governance models.

This change exists to close that split without undoing the already completed visualization parity work. The goal is not necessarily to force all domains onto one endpoint if that harms clarity; the goal is to make dataset and datasource interactive discovery equally governed, parity-backed, and contract-consistent with the rest of the interactive aggregate flow.

## Goals / Non-Goals

**Goals**
- Make dataset and datasource interactive discovery parity explicit and testable.
- Preserve the `BusiTreeNode` contract already consumed by the interactive store.
- Reduce special-case branching in the aggregate interactive loading path where doing so improves correctness and governance.

**Non-Goals**
- Redesigning dataset or datasource CRUD APIs beyond tree/discovery behavior.
- Reopening dashboard/dataV parity work that is already complete.
- Broadly refactoring all BI store logic unrelated to interactive discovery.

## Decisions

### Decision: Preserve frontend contract first
Whether dataset/datasource parity is achieved by batched aggregation or direct endpoints, the frontend store must continue consuming the same normalized `BusiTreeNode` shape.

### Decision: Treat governance parity as important as implementation parity
The work is not complete when the store can load the tree; it is complete only when the route behavior, test evidence, and governed metadata all describe the same reality.

### Decision: Prefer the smallest architecture that removes the special-case gap
If extending the batched interactive endpoint is simpler and safer, use it. If formalizing equivalent direct-tree behavior is clearer, allow that — but only if the aggregate loading path remains contract-consistent and verifiable.

## Risks / Trade-offs

- **Mixed-path complexity**: keeping multiple loading paths may preserve hidden divergence. Mitigation: require explicit baseline and evidence no matter which path remains.
- **Over-unification risk**: forcing dataset/datasource into the visualization-style interactive endpoint may complicate ownership boundaries. Mitigation: decide after freezing the current gap inventory.
- **Frontend regression risk**: dataset/datasource tree nodes use different optional semantics in some pages. Mitigation: preserve the interactive store contract and update tests before behavior changes.

## Migration Plan

1. Freeze the current dataset/datasource interactive aggregate loading split.
2. Add failing tests that expose the remaining parity gap.
3. Implement the chosen parity path with minimal frontend contract change.
4. Update evidence and governance so the aggregate discovery model is coherent across all BI domains.

## Open Questions

- Whether the right end-state is one batched interactive endpoint for all BI domains, or two equivalent governed tree endpoints plus a stable frontend aggregate loader.
- Whether datasource and dataset tree node optional fields should be normalized further in the interactive aggregate path.
