## 1. Governance Rules
- [x] 1.1 Define compatibility endpoint status taxonomy (`full/partial/stub/missing`) and code-behavior criteria
- [x] 1.2 Define explicit prohibition of placeholder success for migration-scoped endpoints
- [x] 1.3 Define required error semantics for unavailable features

## 2. Runtime and Metadata Alignment
- [x] 2.1 Inventory current compatibility endpoints with actual behavior evidence
- [x] 2.2 Update migration matrix and whitelist statuses to match runtime reality
- [x] 2.3 Add drift checks in CI for route status and contract semantics

## 3. Verification
- [x] 3.1 Add/extend compatibility tests for placeholder-success detection
- [x] 3.2 Validate contract-diff reports include status drift findings
- [x] 3.3 Document waiver process for temporary stubs with expiry
