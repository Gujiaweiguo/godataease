## Context

The frontend exposes a standalone `variable.ts` API module that treats system variables as a first-class feature, with separate APIs for variable CRUD and variable-value CRUD. In the Go backend, no `/sysVariable/*` transport surface exists today. The nearest existing capability is `system-parameter-management`, but that spec covers platform-level parameters rather than user-managed variable definitions and enumerated values.

There is also unrelated SQL variable parsing logic inside dataset handling, but that concerns SQL parameter extraction from datasets rather than administrable system variables. This change therefore needs a separate bounded context so the implementation does not overload system parameters or dataset SQL variables with different semantics.

## Goals / Non-Goals

**Goals:**
- Add a dedicated Go module for system variable CRUD and variable-value CRUD.
- Match the existing frontend `/sysVariable/*` contract closely enough to support the current UI without frontend rewrites.
- Keep system variable behavior clearly separated from `sysParameter` and dataset SQL variable parsing.
- Make variable and variable-value flows testable at handler, service, and repository levels.

**Non-Goals:**
- Redesign the frontend variable management UI.
- Merge system variables into `system-parameter-management`.
- Extend dataset SQL variable parsing or query-time substitution behavior.
- Add unrelated configuration features outside the current `/sysVariable/*` surface.

## Decisions

### 1. Introduce a dedicated system variable capability and transport surface
System variables represent a separate management domain with their own CRUD and value-selection workflows. Creating a dedicated capability avoids conflating them with system parameters, which are platform configuration entries with different lifecycle and authorization semantics.

**Alternatives considered:**
- Reuse `system-parameter-management`: rejected because the API surface and domain model are materially different.
- Attach variable management to dataset SQL variable code: rejected because dataset SQL variables are derived query metadata, not managed configuration records.

### 2. Model variable definitions and variable values as separate but linked resources
The frontend calls separate endpoints for variable entities and variable value entities, including paging and batch deletion for values. The backend should preserve that split to keep the implementation aligned with current usage and to avoid ambiguous write semantics.

**Alternatives considered:**
- Store all values inline on the variable object with a single bulk update endpoint: rejected because it would not match the current frontend contract and would complicate incremental editing.

### 3. Keep query/detail/selection flows deterministic for UI reloads
The frontend relies on detail lookup and value-selection endpoints after edits. The backend should therefore make create/edit/delete operations update the same persisted state that query and detail endpoints expose, with stable paging semantics for value lists.

**Alternatives considered:**
- Return ad hoc write-only results without consistent readback behavior: rejected because it makes frontend state refresh brittle.

## Risks / Trade-offs

- **[Risk] The variable domain could drift toward system parameters if boundaries stay vague** → Mitigation: define a separate capability and keep `sysParameter` explicitly out of scope.
- **[Risk] Variable values may need paging and filtering semantics not yet formalized elsewhere** → Mitigation: specify frontend-required selection and query endpoints directly in the new spec.
- **[Risk] Existing data model assumptions are still unknown in Go mainline** → Mitigation: keep the spec focused on externally observable behavior and let implementation choose the minimal schema that supports it.
- **[Risk] Scope creep into dataset SQL variable handling** → Mitigation: treat dataset SQL variable parsing as a separate domain and exclude it from this change.
