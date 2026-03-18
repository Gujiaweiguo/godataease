## 1. Current-state inventory

- [x] 1.1 Document how `interactiveStore` currently loads dataset and datasource trees versus dashboard/dataV trees.
- [x] 1.2 Map the backend endpoints and handlers that currently power dataset and datasource interactive discovery.
- [x] 1.3 Freeze the current governance/evidence gap for dataset and datasource interactive aggregate behavior.

## 2. Regression-first baseline

- [x] 2.1 Add failing tests that capture the current split between interactive aggregate loading and dataset/datasource direct-tree loading.
- [x] 2.2 Add failing contract tests for dataset tree nodes in the interactive aggregate path.
- [x] 2.3 Add failing contract tests for datasource tree nodes in the interactive aggregate path.

## 3. Parity implementation

- [x] 3.1 Implement the chosen dataset parity path for interactive aggregate loading.
- [x] 3.2 Implement the chosen datasource parity path for interactive aggregate loading.
- [x] 3.3 Ensure the interactive aggregate path preserves the stable `BusiTreeNode` contract for dataset and datasource nodes.
- [x] 3.4 Ensure failure or unavailable behavior remains deterministic and does not silently break aggregate loading.

## 4. Frontend convergence

- [x] 4.1 Update the interactive store only where necessary to remove dataset/datasource special-case divergence.
- [x] 4.2 Add or update unit tests covering dataset/datasource interactive aggregate loading.
- [x] 4.3 Verify current consumers do not regress when dataset/datasource nodes are loaded through the chosen parity path.

## 5. Governance and verification

- [x] 5.1 Update governance/evidence metadata for dataset and datasource interactive discovery.
- [x] 5.2 Run backend tests covering dataset/datasource interactive parity behavior.
- [x] 5.3 Run frontend tests covering aggregate interactive loading.
- [x] 5.4 Document remaining gaps before marking the change implementation-ready.
