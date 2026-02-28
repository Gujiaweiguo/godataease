## 1. Implementation
- [x] 1.1 Implement SeaTunnel gRPC methods: submit, status, cancel
- [x] 1.2 Define deterministic status mapping (`pending/running/success/failed/cancelled`)
- [x] 1.3 Implement `syncApiTable` and `syncApiDs` using SeaTunnel submit path
- [x] 1.4 Implement `listSyncRecord` with persisted records and pagination
- [x] 1.5 Add retry/circuit-breaker behavior for transient SeaTunnel failures
- [x] 1.6 Add unit and integration tests for task lifecycle paths

## 2. Verification
- [x] 2.1 Run Go tests for integration, service, and handler layers
- [x] 2.2 Validate compatibility contract for sync endpoints
- [x] 2.3 Update migration matrix with implementation evidence
