## 1. Implementation
- [ ] 1.1 Implement Calcite gRPC client methods for `ParseSQL` and `ValidateSQL`
- [ ] 1.2 Add timeout and retry policy with structured error classification
- [ ] 1.3 Wire Calcite validation into dataset SQL preview workflow
- [ ] 1.4 Wire Calcite validation into compatibility bridge SQL endpoints
- [ ] 1.5 Add unit tests for success, invalid SQL, timeout, and upstream failure

## 2. Verification
- [ ] 2.1 Run Go tests for integration and service layers
- [ ] 2.2 Run contract-diff checks for impacted SQL endpoints
- [ ] 2.3 Update migration matrix status and evidence links
