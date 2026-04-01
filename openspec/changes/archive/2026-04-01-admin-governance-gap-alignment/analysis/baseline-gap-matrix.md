# Admin Governance Alignment Baseline and Gap Matrix

## T1. Baseline Freeze

### Canonical Baseline
- Baseline date: 2026-03-30
- User and role management: `https://dataease.io/docs/v2/xpack/user_management_user/`
- Role anchor: `https://dataease.io/docs/v2/xpack/user_management_user/#2`
- Organization management: `https://dataease.io/docs/v2/xpack/sys_management_organization/`
- Permission management: `https://dataease.io/docs/v2/xpack/sys_management_permission/`

### Bucket Definitions
- **Implemented but inconsistent**: frontend/backend path exists, but runtime behavior, scope, or enforcement is inconsistent with the frozen baseline.
- **Partially implemented**: main workflow skeleton exists, but a required branch, validation, dual-view consistency, or runtime enforcement is incomplete.
- **Not implemented**: no working path exists, behavior is still a stub/placeholder/501, or no closure evidence exists for a required capability.

### Evidence Format
Every reviewed item in this change MUST carry:
1. manual reference
2. current evidence
3. bucket
4. risk
5. planned action

### Deviation Policy
- Default target is semantic alignment with the frozen baseline.
- If the current fork intentionally keeps a safer or lower-risk behavior that differs from the baseline, that behavior MUST be recorded as an intentional deviation before implementation is considered complete.
- Intentional deviation is allowed only when:
  - the baseline is ambiguous, or
  - strict parity would introduce disproportionate migration or data-safety risk.

## T2. Gap Matrix

| Domain | Capability | Manual Reference | Current Evidence | Bucket | Risk | Planned Action |
|---|---|---|---|---|---|---|
| Organization | Delete org requires child cleanup and resource disposition | org docs: delete org, child-first, resource cleanup | `apps/backend-go/internal/service/org_service.go` has delete flow skeleton; current plan evidence shows soft-delete/audit bias without full resource-disposition proof | Implemented but inconsistent | High | Lock policy in foundation wave, then implement deterministic delete semantics |
| Organization | Tree-driven org administration | org docs: tree structure, unlimited depth | backend tree endpoint exists; frontend org view still mainly uses list flow and client-side assembly | Partially implemented | Medium | Promote tree API to canonical governed source |
| User | Create/edit under explicit organization context | user docs: org selection required | user-management and auth APIs both exist; current governed entry still mixes legacy and canonical route families | Partially implemented | High | Stabilize canonical org-scoped workflow before lifecycle alignment |
| User | Enabled/disabled state governs login | user docs: disabled user cannot login | user enabled flow exists in frontend legacy API surface (`switchEnableApi`) but no confirmed governed end-to-end path in current Go-aligned workflow evidence | Partially implemented | Medium | Audit governed auth + user-management contract together |
| User | Import with partial success and error report | user docs: import + error report + 10MB cap | `user_import_service.go`, `/user/excelTemplate`, `/user/errorRecord/:key`, `/user/defaultPwd` exist | Partially implemented | Medium | Preserve current closure, then extend only documented missing source metadata |
| User | Third-party source metadata | user docs: LDAP/OIDC/飞书/钉钉/企业微信 source fields | no closure evidence in current user model / import template | Not implemented | Medium | Defer as bounded P2 substream under lifecycle alignment |
| Role | Last-role removal semantics | role docs: remove unique role triggers deterministic downstream handling | `role_service.go` supports removal flows; current policy is closer to “block last role removal” per prior evidence | Implemented but inconsistent | High | Record and decide policy in foundation wave before code change |
| Role | Custom role inheritance ceiling | role docs: custom role permissions must not exceed inherited parent | `apps/backend-go/internal/service/role_service.go:506` TODO in `ValidatePermissionInheritance` | Partially implemented | High | Implement validation in lifecycle wave |
| Role | Add org user / external user / remove member | role docs: role member lifecycle | role APIs and UI flows exist in current codebase | Partially implemented | Medium | Preserve existing flow, tighten org-scope and ceiling semantics |
| Permission | Menu permission must be role-bound and runtime-enforced | permission docs: menu auth only binds to roles | `menu_auth.go` `RequireMenuAuth` still returns stubbed forbidden path | Implemented but inconsistent | High | Replace stub with effective authorization enforcement |
| Permission | Row permission runtime enforcement | permission docs: row filters must enforce runtime data visibility | `permission.go` `RowPermissionMiddleware()` is placeholder and only logs warning | Implemented but inconsistent | High | Promote to real runtime enforcement in permission-center wave |
| Permission | Dual-view consistency (by user / by resource) | permission docs: same underlying model, two views | `permission-config` UI and compat APIs exist; target flows still incomplete | Partially implemented | High | Stabilize user/resource semantic convergence before claiming parity |
| Permission | Menu target permission APIs | permission docs: governed target assignment path needed for resource-style view | `menuTargetPerApi` / `saveMenuTargetPerApi` still backed by incomplete path | Not implemented | High | Fill or explicitly retire with compat policy |
| Permission | Row whitelist and system variables | permission docs: whitelist + system variable dimensions | no complete closure evidence found in current implementation review | Not implemented | High | Treat as governed data-permission gap in T8 |
| Permission | Column masking rule coverage | permission docs: disable + masking + custom rule boundary | backend rule support exceeds current frontend exposure | Partially implemented | Medium | Align frontend rule exposure with governed backend semantics |
| Permission | Resource group inheritance | permission docs: new resources inherit group grants | existing spec requires it, but current plan has no direct closure evidence in implementation review | Not implemented | Medium | Keep in permission-center P2-bounded substream |
| Compatibility | Legacy user route family vs canonical Go routes | baseline requires stable governed behavior, not route confusion | `apps/frontend/src/api/user.ts` still calls `/user/pager`, `/user/create`, `/user/edit`, `/user/delete`, `/user/enable`; router evidence does not show canonical registration for those specific legacy paths | Implemented but inconsistent | High | Separate compat policy from lifecycle semantics |
| Compatibility | `/user/org/option` | frontend compatibility need | prior mapping found frontend caller with no confirmed Go route registration | Not implemented | Medium | Decide: add alias, migrate frontend, or retire |
| Compatibility | Compat handler versus canonical permission flow | permission baseline requires working flows, not placeholder aliases | `permission_compat_handler.go` plus `/auth/*` route family exists, but some target paths remain incomplete | Partially implemented | High | Canonicalize or explicitly govern remaining alias set |

## Execution Notes
- This matrix is the only source of truth for bucket decisions in this change.
- Any new gap discovered during implementation must be appended using the same columns before related tasks can be marked complete.
