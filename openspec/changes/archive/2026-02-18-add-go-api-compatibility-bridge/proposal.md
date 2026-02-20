# Change: Add Go API Compatibility Bridge

## Why
Java-to-Go migration is blocked by API contract drift. Several frontend calls still rely on Java route prefixes and payload conventions that are not fully available in Go.

## What Changes
- Add a Go API compatibility bridge for Java-style route prefixes used by frontend and integration clients.
- Define route aliasing and response/error-code parity rules for migration period.
- Prioritize high-traffic prefixes first: `datasource`, `datasetTree`, `datasetData`, `chartData`, `user`, `org`, `msg-center`.
- Add validation checklist to ensure old routes and `/api/*` aliases behave consistently.

## Impact
- Affected specs: `api-compatibility-bridge`
- Affected code:
  - `backend-go/internal/transport/http/router.go`
  - `backend-go/internal/transport/http/handler/*`
  - related domain/service/repository modules for migrated endpoints
