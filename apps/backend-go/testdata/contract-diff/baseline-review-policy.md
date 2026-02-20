# Baseline Review Policy

This document defines the review requirements, checklist, and No-Go conditions for API baseline management in the contract-diff testing framework.

## Review Requirements

### Required Reviewers

| Role | Responsibility |
|------|---------------|
| API Owner | Verifies baseline accuracy against API contract |
| Tech Lead | Validates technical implementation and structure |
| QA | Confirms test coverage and expected behavior |

### Scope of Review

All baseline-related changes require review:

- **New Baselines**: Initial fixture files for new API endpoints
- **Baseline Updates**: Modifications to existing baseline files
- **Baseline Removals**: Deletion of obsolete baseline files

## Review Checklist

Before approving a baseline change, reviewers must verify:

### 1. Baseline-Whitelist Consistency

- [ ] Baseline filename matches the corresponding `whitelist.json` entry
- [ ] HTTP method and path are consistent between baseline and whitelist
- [ ] Baseline description clearly explains the expected behavior

### 2. Response Structure Validation

- [ ] Response body matches actual API response structure
- [ ] HTTP status code is correct and appropriate
- [ ] Headers are accurate (Content-Type, etc.)

### 3. Required Fields

- [ ] `meta.version` reflects current API version
- [ ] `meta.generated_at` timestamp is recent
- [ ] `request` section includes all relevant parameters
- [ ] `response` section includes complete data structure

### 4. Version Metadata

- [ ] Version number matches the API version being tested
- [ ] Changelog entries are updated if behavior changes

## No-Go Conditions

Baseline changes MUST NOT be merged if any of the following conditions exist:

| Condition | Reason |
|-----------|--------|
| Unapproved baseline changes | Requires sign-off from all required reviewers |
| Missing owner sign-off | API owner must verify accuracy |
| Baseline used to bypass gate failure | Baselines validate expected behavior, not mask failures |
| Inconsistent with whitelist | Creates confusion in test interpretation |
| Missing required fields | Incomplete baselines cause test failures |

### Anti-Pattern: Bypass via Baseline

**WRONG**: Adding a baseline to silence a legitimate API contract violation.

```
# Do NOT do this
API returns 500 → Add baseline with 500 → Test passes (masking bug)
```

**RIGHT**: Fix the API issue, then add baseline for correct behavior.

```
# Correct approach
API returns 500 → Fix API to return 200 → Add baseline with 200 → Test validates fix
```

## Approval Process

### PR Template Requirements

All baseline PRs must include:

1. **Change Type**: `[ ] New [ ] Update [ ] Removal`
2. **Affected Endpoint**: HTTP method and path
3. **Justification**: Why this baseline change is needed
4. **Verification Steps**: How reviewers can verify accuracy

### Sign-off Format

Approved PRs must have sign-off comments from required reviewers:

```
Reviewed-by: API Owner <email>
Tech-Lead-Approved: @username
QA-Verified: @username
```

### Merge Requirements

- [ ] All required reviewers have signed off
- [ ] No outstanding review comments
- [ ] CI/CD pipeline passes
- [ ] Baseline files pass schema validation
- [ ] Whitelist entries are synchronized

## Enforcement

This policy is enforced by:

1. **CODEOWNERS**: Baseline directory requires approval from designated reviewers
2. **CI Pipeline**: Automated validation of baseline structure
3. **Pre-commit Hooks**: Basic format validation before commit

---

*Last updated: 2026-02-18*
*Policy version: 1.0*
