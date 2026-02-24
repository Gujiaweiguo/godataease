## 1. Implementation
- [ ] 1.1 Implement SeaTunnel gRPC methods: submit, status, cancel
- [ ] 1.2 Define deterministic status mapping (`pending/running/success/failed/cancelled`)
- [ ] 1.3 Implement `syncApiTable` and `syncApiDs` using SeaTunnel submit path
- [ ] 1.4 Implement `listSyncRecord` with persisted records and pagination
- [ ] 1.5 Add retry/circuit-breaker behavior for transient SeaTunnel failures
- [ ] 1.6 Add unit and integration tests for task lifecycle paths

## 2. Verification
- [ ] 2.1 Run Go tests for integration, service, and handler layers
- [ ] 2.2 Validate compatibility contract for sync endpoints
- [ ] 2.3 Update migration matrix with implementation evidence
