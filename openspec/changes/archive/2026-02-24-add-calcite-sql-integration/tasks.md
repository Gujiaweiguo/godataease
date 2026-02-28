## 1. Implementation
- [x] 1.1 Implement Calcite gRPC client methods for `ParseSQL` and `ValidateSQL`
- [x] 1.2 Add timeout and retry policy with structured error classification
- [x] 1.3 Wire Calcite validation into dataset SQL preview workflow
- [x] 1.4 Wire Calcite validation into compatibility bridge SQL endpoints
- [x] 1.5 Add unit tests for success, invalid SQL, timeout, and upstream failure

## 2. Verification
- [x] 2.1 Run Go tests for integration and service layers
- [x] 2.2 Run contract-diff checks for impacted SQL endpoints
- [x] 2.3 Update migration matrix status and evidence links
