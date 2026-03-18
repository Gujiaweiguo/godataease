## Why

The interactive parity follow-up closed dashboard and big-screen discovery by making `dataVisualization/interactiveTree` return real visualization resources. The frontend interactive store still handles dataset and datasource discovery through separate direct endpoints (`/datasetTree/tree` and `/datasource/tree`), which means the interactive aggregate view remains split across two different loading paths and governance surfaces.

## What Changes

- Define a governed parity path for dataset and datasource interactive discovery so the interactive aggregate view is consistent across all four BI domains.
- Decide whether dataset/datasource parity is achieved by extending the batched interactive aggregation endpoint or by formalizing equivalent direct-tree behavior behind the same frontend contract.
- Align interactive store loading, tests, and governance evidence so dataset/datasource discovery is no longer a special-case branch outside the interactive parity story.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `dataset-management`: strengthen dataset tree behavior for interactive aggregate discovery.
- `datasource-management`: strengthen datasource tree behavior for interactive aggregate discovery.
- `api-compatibility-bridge`: define how dataset/datasource interactive aggregation is governed alongside dashboard/dataV discovery.

## Impact

- **Frontend**: `src/store/modules/interactive.ts`, dataset/datasource API wrappers, and interactive store tests.
- **Backend Go**: dataset tree and datasource tree handlers/services, plus any compatibility aggregation logic chosen for parity.
- **Governance**: endpoint inventory/whitelist/evidence for dataset and datasource interactive aggregate behavior.
