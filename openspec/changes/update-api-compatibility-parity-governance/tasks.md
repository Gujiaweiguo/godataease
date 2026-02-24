## 1. Governance Rules
- [ ] 1.1 Define compatibility endpoint status taxonomy (`full/partial/stub/missing`) and code-behavior criteria
- [ ] 1.2 Define explicit prohibition of placeholder success for migration-scoped endpoints
- [ ] 1.3 Define required error semantics for unavailable features

## 2. Runtime and Metadata Alignment
- [ ] 2.1 Inventory current compatibility endpoints with actual behavior evidence
- [ ] 2.2 Update migration matrix and whitelist statuses to match runtime reality
- [ ] 2.3 Add drift checks in CI for route status and contract semantics

## 3. Verification
- [ ] 3.1 Add/extend compatibility tests for placeholder-success detection
- [ ] 3.2 Validate contract-diff reports include status drift findings
- [ ] 3.3 Document waiver process for temporary stubs with expiry
