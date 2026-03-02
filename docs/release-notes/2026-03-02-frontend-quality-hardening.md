# 2026-03-02 Frontend Quality Hardening

## Background

The frontend TypeScript debt had accumulated in chart/editor-related modules and spread into shared UI typing, which increased CI instability and raised regression risk during routine feature work.

## What Was Done

- Reduced and cleared remaining frontend TypeScript errors in `apps/frontend`.
- Fixed related lint issues and stabilized type boundaries in chart/editor/map paths.
- Resolved core regression failures in tests by:
  - hardening `interactive` store behavior for empty input and id normalization
  - adding embedding compatibility utilities:
    - `apps/frontend/src/utils/embeddedParams.ts`
    - `apps/frontend/src/utils/embeddedOriginValidation.ts`
- Merged PR #17 into `main`.

## Verification Results

Local verification after merge:

- `npm run ts:check` passed
- `npm run lint` passed
- `npm run test:core` passed (`19` files, `359` tests)

CI status on PR #17 was green before merge (quality/e2e/contract-diff/typos/llm-code-review).

## Process Improvement Applied

Updated `.github/workflows/frontend.yml` to make quality gates blocking:

- `TypeScript check` is now blocking (removed `continue-on-error`)
- `Test (core)` is now blocking (removed `continue-on-error`)
- `Affected Test Matrix` is now blocking (removed `continue-on-error`)

This prevents silent regressions from passing CI when checks fail.

## Impact

- Improved confidence for ongoing frontend feature delivery.
- Lower chance of type-related regressions re-entering mainline.
- Better CI signal quality by enforcing fail-fast behavior.

## Follow-up

- Monitor next 3-5 frontend PRs for CI stability and false-positive rates.
- If needed, split heavyweight checks by path to keep feedback fast while preserving blocking gates.
