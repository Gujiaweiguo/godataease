## 1. Datasource list identity discovery

- [x] 1.1 Inventory all current datasource list callers and compatibility aliases.
- [x] 1.2 Determine whether any existing datasource list request already carries a stable governing scope usable for permission decisions.
- [x] 1.3 Record whether current runtime semantics are best described as filtered list, scoped forbidden, or auth-only list behavior.

## 2. Option selection

- [x] 2.1 Compare filtered list, explicit scoped list, and detail-only forbidden strategies against current callers.
- [x] 2.2 Select one design direction and document why the other options were rejected.

## 3. Implementation planning

- [x] 3.1 Define the minimal backend changes required for the selected design.
- [x] 3.2 Define the minimal frontend caller or test changes required for the selected design.
- [x] 3.3 Define the regression coverage needed to prove the selected design at runtime.
