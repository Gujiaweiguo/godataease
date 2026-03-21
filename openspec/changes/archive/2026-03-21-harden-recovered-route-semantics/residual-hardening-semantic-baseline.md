# Residual Hardening Semantic Baseline

This document freezes the semantic expectations for the remaining hardening work after the archived recovery changes.

## Semantic targets

| Condition | Expected semantic | Must not degrade into |
|---|---|---|
| Unauthenticated | explicit auth-missing non-success at the governed boundary | silent success, HTML fallback, not-found, or generic business failure |
| Forbidden | explicit forbidden non-success for authenticated-but-unauthorized callers | unauthenticated, missing-resource, or generic internal error |
| Missing route/resource | explicit not-found or unavailable response at the governed boundary | forbidden, unauthenticated, or placeholder success |
| Business/dependency failure | explicit non-success with diagnosable message | misleading empty-success state or swallowed error |

## Residual hardening semantic matrix

| Flow family | Governed path | Current proof | Remaining semantic hardening target |
|---|---|---|---|
| Datasource | datasource list aliases and related list-entry callers | unauthenticated alias proof exists; datasource view middleware has 401/403 proof | add direct forbidden proof on datasource list aliases so list-path denial is frozen at the same boundary as unauthenticated behavior |
| Dashboard | dashboard detail paths using `findById()` | payload-consumption and unauthenticated proof exist; missing-resource is strongest at service level | move missing-resource semantic proof closer to handler/route boundary |
| Big-screen | big-screen detail/edit paths beyond preview | preview/discovery auth and payload-consumption proof exist | add explicit detail/edit semantics so missing-resource and detail-path behavior are not only inferred from preview routes |
| Export-center | `/api/exportCenter/exportTasks/*` and download paths | unauthenticated query semantics are explicit; page-init/query coverage exists | freeze download-path semantics so auth/not-found/business-state failures are explicit at the route boundary |
| Audit | `/api/audit/list`, `/api/audit/:id`, audit page init | unauthenticated list/detail proof exists; detail not-found and query-init non-success are explicit | add route-level page-entry/detail smoke to prove these semantics survive beyond unit and handler coverage |

## Current interpretation

The semantics listed here are already mostly explicit because of the archived recovery work. This hardening change exists to close the last boundary gaps where the proof is still indirect, incomplete, or only present at lower layers than the real caller path.
