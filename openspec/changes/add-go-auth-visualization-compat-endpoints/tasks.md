## 1. Compatibility Route Coverage

- [x] 1.1 Add auth/permission compatibility handler and register `/auth/menuPermission`, `/auth/busiPermission`, `/auth/saveMenuPer`, `/auth/saveBusiPer`, `/system/role/permission/save`
- [x] 1.2 Add legacy `/system/role/*` aliases mapped to canonical role handlers (`create`, `update->edit`, `delete`)
- [x] 1.3 Ensure new routes are reachable through `/de2api/*` bridge and protected by existing middleware policy

## 2. Visualization Tree Parity

- [x] 2.1 Implement `POST /dataVisualization/tree` in Go visualization module with contract-compatible payload
- [x] 2.2 Add response validation/error semantics to avoid placeholder-success for unavailable branches
- [x] 2.3 Verify dashboard/screen resource-tree dependent pages no longer hit `404`

## 3. Verification and Governance

- [x] 3.1 Build endpoint coverage inventory for this change (critical frontend flows only)
- [x] 3.2 Add scripted regression checks for non-404 + `code/data/msg` envelope on covered endpoints
- [x] 3.3 Execute frontend regression for role-menu management and dashboard tree workflows, record evidence
