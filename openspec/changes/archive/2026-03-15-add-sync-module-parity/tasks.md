## 1. Sync module skeleton

- [x] 1.1 Add Go transport registration for `/sync/datasource/*`, `/sync/task/*`, `/sync/task/log/*`, and `/sync/summary/*`
- [x] 1.2 Define request/response DTOs and service interfaces needed by sync datasource, task, log, and summary handlers
- [x] 1.3 Align sync status and paging response shapes with current frontend expectations

## 2. Sync datasource parity

- [x] 2.1 Implement sync datasource pager, get, save, update, delete, and batch delete flows
- [x] 2.2 Implement sync datasource validate, validate-by-id, schema, fields, and latest-use endpoints
- [x] 2.3 Implement datasource option and table-list endpoints used by the sync task editor

## 3. Sync task lifecycle and observability

- [x] 3.1 Implement sync task pager, detail, add, update, remove, and batch delete flows
- [x] 3.2 Implement sync task execute, start, and stop flows using deterministic lifecycle status mapping
- [x] 3.3 Implement sync task log pager, detail, delete, clear, and termination endpoints
- [x] 3.4 Implement sync summary resource count and log chart data endpoints

## 4. Integration alignment and verification

- [x] 4.1 Reuse or extend SeaTunnel orchestration and persisted sync record paths so `/sync/*` and existing datasource sync entrypoints stay consistent
- [x] 4.2 Add handler/service/repository tests for sync datasource, task, log, and summary flows
- [x] 4.3 Add compatibility coverage for the new `/sync/*` routes and document any required contract whitelist updates
