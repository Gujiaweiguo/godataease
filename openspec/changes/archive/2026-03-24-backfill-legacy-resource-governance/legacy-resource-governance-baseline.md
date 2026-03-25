# Legacy Resource Governance Baseline

This document freezes the change-local baseline used by `backfill-legacy-resource-governance` for historical resource identity, parent ownership, organization boundary handling, and anomaly sampling. It is intentionally scoped to what the current implementation and OpenSpec deltas can support safely.

## 1. Stable Resource Identity Rules

| Resource Family | Business Table | Stable Identifier in This Change | Governed Resource Type |
| --- | --- | --- | --- |
| Datasource | `core_datasource` | `core_datasource.id` | `datasource` |
| Dataset | `core_dataset_group` | `core_dataset_group.id` | `dataset` |
| Dashboard | `data_visualization_info` | `data_visualization_info.id` | `dashboard` |
| Screen | `data_visualization_info` | `data_visualization_info.id` + `type in ('dataV', 'screen')` | `screen` |

Frozen rule:

- Historical governed identity is frozen as `(resource_type, business_primary_key)`.
- No extra mapping table is introduced in this change.
- Dataset backfill in this change is scoped to dataset groups (`core_dataset_group`), which are the resource-tree identities exposed in the unified permission center.
- Visualization backfill normalizes `data_visualization_info.type` into governed `dashboard` / `screen` resource types.

## 2. Parent Ownership Rules

| Resource Family | Parent Field | Parent Source | Automatic Backfill Rule |
| --- | --- | --- | --- |
| Datasource | `pid` | `core_datasource.id` | Automatic backfill requires `pid > 0` and a parent that is already governable/governed |
| Dataset | `pid` | `core_dataset_group.id` | Automatic backfill requires `pid > 0` and a parent that is already governable/governed |
| Dashboard | `pid` | `data_visualization_info.id` | Automatic backfill requires `pid > 0` and a parent that is already governable/governed |
| Screen | `pid` | `data_visualization_info.id` | Automatic backfill requires `pid > 0` and a parent that is already governable/governed |

Frozen rule:

- This change does **not** auto-govern top-level historical resources with no positive parent id.
- During backfill, `pid == nil` or `pid <= 0` is treated as `missing_parent` for automatic governance purposes.
- A parent reference is only sufficient when the parent can supply governed inheritance; otherwise the child is skipped as `parent_not_governed`.

## 3. Organization Boundary Rules

| Resource Family | Current Safe Org Boundary | Rollout Support in This Change |
| --- | --- | --- |
| Datasource | No explicit safe org boundary exposed in current resource model | Org-scoped backfill is rejected |
| Dataset | No explicit safe org boundary exposed in current resource model | Org-scoped backfill is rejected |
| Dashboard | `data_visualization_info.org_id` when present | Org-scoped backfill supported through `visualization` backfill |
| Screen | `data_visualization_info.org_id` when present | Org-scoped backfill supported through `visualization` backfill |

Frozen rule:

- `visualization` is the only resource family with an explicit org boundary used by this change's batch execution contract.
- `datasource` and `dataset` do not claim a frozen runtime org derivation rule in this change; requests attempting org-scoped backfill for those resource types are rejected.
- Any datasource/dataset case that requires org-sensitive reconciliation is treated as a follow-up remediation problem, not silently auto-governed.

## 4. Classification Baseline

This section freezes the classification vocabulary required by task 1.2 and implemented by later rollout/reporting work.

| Classification | Minimum Condition | Expected Handling |
| --- | --- | --- |
| 可自动纳管 | Stable identity is known and the resource has a governable parent boundary | Register/reuse `sys_resource`, compute inherited governed permissions |
| 需跳过并记录 | Automatic governance cannot proceed safely in the current run, but the reason is explainable by current data | Return skip classification through backfill report with reason/remediation |
| 需人工修复 | The case requires ownership reconciliation, org-sensitive cleanup, or additional feature/change work | Exclude from automatic governance and track as follow-up remediation |

Current report-level remediation mapping:

| Backfill Signal | Remediation |
| --- | --- |
| `missing_parent` | `data_cleanup` |
| `parent_not_governed` | `govern_parent` |
| `invalid_resource` (or future unsupported anomalies) | `needs_change` |

## 5. Historical Resource Baseline and Anomaly Sample List

This is a change-local baseline inventory, not a production census. It defines the minimum resource families and anomaly samples that must be covered by this change's documentation, tests, and rollout conclusions.

### 5.1 Covered Historical Resource Families

| Baseline Slice | Covered Resource |
| --- | --- |
| DS-1 | Historical datasource folder / datasource entries in `core_datasource` |
| DG-1 | Historical dataset groups in `core_dataset_group` |
| VIZ-1 | Historical dashboards in `data_visualization_info` with dashboard type |
| VIZ-2 | Historical screens in `data_visualization_info` with screen/dataV type |

### 5.2 Required Anomaly Samples

| Sample ID | Scenario | Resource Families | Expected Classification | Follow-up Path |
| --- | --- | --- | --- | --- |
| ANOM-01 | Parent missing or invalid for automatic governance (`pid` empty/non-positive or parent record unavailable for inheritance) | datasource / dataset / dashboard / screen | 需跳过并记录 | `data_cleanup` |
| ANOM-02 | Parent exists but is not yet governed, so inheritance cannot converge | datasource / dataset / dashboard / screen | 需跳过并记录 | `govern_parent` |
| ANOM-03 | Cross-org drift on visualization resources with explicit `org_id` versus parent boundary | dashboard / screen | 需人工修复 | `needs_change` or manual ownership reconciliation |
| ANOM-04 | Resource family needs org-sensitive handling but current model has no safe org boundary (`datasource` / `dataset`) | datasource / dataset | 需人工修复 | separate remediation / follow-up change |
| ANOM-05 | Orphan resource with no safe ownership boundary or no governable parent path | datasource / dataset / dashboard / screen | 需人工修复 | manual reconciliation or separate change |

## 6. Auditable Output Baseline

This change treats the backfill report as the minimum auditable surface.

Required auditable fields from the current implementation:

- request scope: `resourceType`, `afterId`, `limit`, optional `orgId`
- outcome counts: `scanned`, `governed`, `skipped`
- governed identities: `resourceIds`
- skipped identities: `skippedItems[].resourceId`, `resourceType`, `parentId`, `reason`, `remediation`
- rollout semantics: `nextAfterId`, `rollbackBoundary`, `rerunStrategy`

This baseline does **not** claim that a separate persistent migration log table exists in the current change.

## 7. Frozen Decisions

1. Historical governed identity is based on existing business primary keys, not a new mapping table.
2. Automatic backfill depends on a governable parent boundary; top-level historical resources are not auto-governed by default in this change.
3. Org-scoped rollout is only frozen for visualization resources because they expose a safe `org_id` boundary in the current model.
4. Datasource/dataset org-sensitive cases are treated as remediation work instead of guessed automatic governance.
5. The minimum anomaly baseline for this change must cover resource type, missing parent, cross-org drift, and orphan scenarios.

---

Baseline frozen for change `backfill-legacy-resource-governance`. Do not widen these rules without a new explicit change update.
