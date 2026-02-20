# Contract Diff Whitelist Governance

## Purpose

The critical whitelist (`critical-whitelist.yaml`) excludes specific API changes from automatic CI blocking. This document defines how whitelist entries are proposed, reviewed, and maintained.

## Change Process

### Adding an Entry

1. Open a PR modifying `critical-whitelist.yaml`
2. Add entry with required fields:
   - `id`: unique identifier (format: `WL-{category}-{number}`)
   - `path`: API path pattern
   - `reason`: documented justification
   - `added_at`: date in ISO format (YYYY-MM-DD)
3. Link to related issue/PR explaining the API change
4. Wait for approval before merge

### Removing an Entry

1. Create PR with entry removal
2. Document why the exclusion is no longer needed
3. Verify no active PRs depend on this entry

### Modifying an Entry

Follow the same process as adding, with explanation of what changed and why.

## Review Requirements

| Change Type | Required Approvers |
|-------------|-------------------|
| Add entry | API Owner + Tech Lead |
| Remove entry | Tech Lead |
| Modify entry | API Owner + Tech Lead |

### Reviewer Responsibilities

- **API Owner**: Verify the exclusion is justified for their domain
- **Tech Lead**: Ensure no security or compatibility risks

## Priority Definitions

| Priority | Meaning |
|----------|---------|
| P0 | Critical - blocks all deployments immediately |
| P1 | High - blocks within 24 hours if unresolved |
| P2 | Normal - flags for review, does not auto-block |

## Blocking Level Definitions

| Level | Behavior |
|-------|----------|
| critical | CI fails immediately, requires whitelist or fix |
| high | CI warns, fails after grace period |
| normal | CI logs warning only |

## Versioning

- Whitelist changes are tracked via git history
- Each entry includes `added_at` date for audit trail
- Major changes should update this governance document
- Review whitelist quarterly to remove stale entries

## Effective Date Rules

- New entries take effect immediately upon PR merge
- Entry removals take effect immediately - ensure no active changes depend on the entry
- Emergency additions require retrospective review within 48 hours

---

*Last updated: 2026-02-18*
