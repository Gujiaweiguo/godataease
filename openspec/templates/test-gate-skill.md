---
name: openspec-test-gate
description: Test gate enforcement for OpenSpec changes. Use when archiving changes to ensure all tests pass.
license: MIT
compatibility: Requires GitHub CLI (gh) and test-gate.yml workflow
metadata:
  author: openspec
  version: "1.0"
---

Enforce test gate before archiving OpenSpec changes.

## Configuration

Add to your `openspec/config.yaml`:

```yaml
test_gate:
  ci:
    - Unit tests
    - Integration tests
    - Lint and type check
  
  archive:
    - All CI tests must pass
    - E2E tests must pass
  
  allow_e2e_skip: true
  e2e_skip_requires_acknowledgment: true
```

## Steps

1. **Check CI status for current branch**

   ```bash
   gh pr list --head <current-branch> --json number,statusCheckRollup
   ```

   If no PR exists, prompt user to create one first.

2. **Verify CI checks passed**

   - If all passed: Continue
   - If any failed: Block archive, tell user to fix failures
   - If pending: Offer to wait or cancel

3. **Trigger E2E tests**

   ```bash
   gh workflow run test-gate.yml \
     -f run_unit=false \
     -f run_integration=false \
     -f run_e2e=true \
     --ref <current-branch>
   ```

4. **Wait for E2E completion**

   ```bash
   gh run watch <run-id> --exit-status
   ```

5. **Handle failures**

   - If E2E fails: Block archive with error details
   - If user wants to skip: Require explicit acknowledgment

## Guardrails

- Test gate is a HARD REQUIREMENT
- Only allow E2E skip with explicit user acknowledgment
- Add test status to archive summary
