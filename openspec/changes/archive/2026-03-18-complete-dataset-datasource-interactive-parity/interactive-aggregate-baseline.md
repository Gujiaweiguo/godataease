# Dataset/Datasource Interactive Aggregate Baseline

This document freezes the current split between dataset/datasource interactive discovery and dashboard/dataV interactive discovery.

## Frontend loading split

`apps/frontend/src/store/modules/interactive.ts` currently uses two different discovery models:

### Batched interactive path

- `loadBusiInteractive()` calls `queryBusiTreeApi(param)` for all enabled BI domains.
- This path is intended to aggregate:
  - `dashboard`
  - `dataV`
  - `dataset`
  - `datasource`

### Direct per-domain path

`setInteractive()` falls back to `apiMap` when no response payload is passed in:

- `dashboard` → `queryTreeApi`
- `dataV` → `queryTreeApi`
- `dataset` → `getDatasetTree`
- `datasource` → `listDatasources`

So the aggregate story is split:

1. dashboard/dataV already have a governed batched interactive path
2. dataset/datasource still rely on direct tree endpoints for normal per-domain loading

## Backend paths currently powering dataset/datasource discovery

### Dataset

- Frontend API wrapper: `apps/frontend/src/api/dataset.ts:getDatasetTree`
- Route used: `POST /datasetTree/tree`
- Compatibility route registration: `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go:444-446`
- Handler: `apps/backend-go/internal/transport/http/handler/dataset_handler.go:21-35`
- Service: `apps/backend-go/internal/service/dataset_service.go:83-133`
- Repository: `apps/backend-go/internal/repository/dataset_repo.go:23-32`

Current dataset tree payload shape is `dataset.TreeNode`:

- `id`
- `name`
- `nodeType`
- `children`

It does **not** natively include the full interactive `BusiTreeNode` fields such as `pid`, `leaf`, `weight`, `extraFlag`, or `extraFlag1`.

### Datasource

- Frontend API wrapper: `apps/frontend/src/api/datasource.ts:listDatasources`
- Route used: `POST /datasource/tree`
- Compatibility tree builder: `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go:1071-1115`
- Service source: `apps/backend-go/internal/service/datasource_service.go:124-126`
- Repository source: `apps/backend-go/internal/repository/datasource_repo.go:149-160`

Datasource direct tree payload already looks much closer to `BusiTreeNode`:

- `id`
- `name`
- `pid`
- `leaf`
- `weight`
- `extraFlag`
- `children`

But it still sits outside the governed batched interactive parity path.

## Current governance/evidence gap

- The previous `complete-interactive-tree-parity` change upgraded only `dataVisualization/interactiveTree` to full parity.
- Dataset and datasource discovery are still not described as part of the same interactive aggregate governance story.
- There is no dedicated evidence document yet showing dataset/datasource aggregate discovery is coherent with dashboard/dataV aggregate discovery.

## Chosen implementation direction for this change

Use the smallest coherent architecture:

1. extend backend `dataVisualization/interactiveTree` so it can also return real dataset and datasource trees for batched aggregate loading
2. preserve existing direct tree APIs (`/datasetTree/tree`, `/datasource/tree`) for pages that still call them directly
3. converge the frontend interactive aggregate flow onto the batched path so the aggregate view is no longer split across mismatched discovery models

## Why this is the lowest-risk path

- It preserves existing page-level APIs.
- It avoids redesigning dataset/datasource CRUD behavior.
- It aligns all four BI discovery domains under one interactive aggregate contract without breaking direct endpoints that other pages may still use.
