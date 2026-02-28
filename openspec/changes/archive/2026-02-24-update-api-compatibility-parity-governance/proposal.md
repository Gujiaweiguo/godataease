# Change: Update API Compatibility Parity Governance

## Why
Current migration contains compatibility endpoints that return placeholder success payloads or static stubs. This creates false parity signals and weakens release-gate confidence.

## What Changes
- Enforce governance rules that prohibit placeholder success semantics for migration-scoped compatibility endpoints.
- Require matrix/whitelist status to stay synchronized with runtime implementation status.
- Add CI checks to detect route status drift between code behavior and migration metadata.

## Impact
- Affected specs: `api-compatibility-bridge`, `remaining-platform-module-migration`
- Affected code:
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
  - `apps/backend-go/internal/transport/http/handler/template_handler.go`
  - `apps/backend-go/testdata/contract-diff/critical-whitelist.yaml`
  - `openspec/changes/archive/*/compatibility-matrix.md`
