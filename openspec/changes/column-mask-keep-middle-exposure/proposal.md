## Why

The permission center frontend currently exposes only a subset of the column desensitization rules already supported by the Go backend. In particular, the existing keep-middle-three rule is implemented in backend runtime masking but cannot be configured from the current frontend column-permission dialog.

## What Changes

- Expose a keep-middle mask option in the permission center column-permission UI.
- Extend the admin-service bridge mapping so the frontend-friendly `maskRule` value round-trips to and from the existing backend keep-middle desensitization rule.
- Add focused backend and frontend regression coverage for this new exposed option.

## Impact

- Closes a concrete frontend exposure gap without redesigning the permission center.
- Reuses existing backend masking capability instead of introducing a new rule type.
- Keeps scope intentionally narrow: no row-permission expression expansion and no broader column-permission UI refactor.
