# T8 Data Permission Scope Note

## Completed in T8
- Dataset preview now applies row-permission field resolution from dataset field IDs to physical column names.
- Dataset preview now applies disabled-column filtering and masking rules to returned rows.
- Chart runtime data now supports permission-aware querying with row filtering, disabled-column filtering, and masking.
- Chart field listing now supports permission-aware filtering, preserving the synthetic count field while hiding disabled fields and marking masked fields as desensitized.
- Compatibility route `datasetField/listWithPermissions/:datasetId` now reflects permission-aware field visibility when a current user is present.
- `whiteList` input is no longer silently ignored; it is explicitly rejected as unsupported in T8.

## Explicitly Deferred in T8
- Generic `RowPermissionMiddleware()` replacement remains deferred because safe runtime enforcement depends on resource-specific dataset/query context.
- Persisted whitelist semantics remain deferred beyond explicit rejection.
- System-variable/runtime parameter semantics remain deferred.
- Broader P2 items such as third-party source metadata and resource-group inheritance remain deferred.

## Why T8 Is Considered Complete
- The highest-value runtime leaks were on real dataset/chart query paths, and those are now governed.
- Unsupported row-permission extensions are explicitly bounded instead of being silently accepted.
- Remaining items are broader P2 or architectural extensions, not missing P0/P1 runtime enforcement on current governed data paths.
