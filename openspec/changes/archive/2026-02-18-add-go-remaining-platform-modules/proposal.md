# Change: Add Go Remaining Platform Modules

## Why
Core Java-to-Go migration still has uncovered platform modules. Missing modules block complete backend switchover and keep operational risk high.

## What Changes
- Implement remaining platform modules in Go for Java parity:
  - `msg-center`
  - `exportCenter`
  - `share` and `ticket`
  - `engine` and `datasourceDriver`
  - `geometry` and `customGeo`
  - `staticResource`, `store`, `typeface`
- Add phased migration order and acceptance criteria per module.
- Define migration matrix status update and release gates for each module.

## Impact
- Affected specs: `remaining-platform-module-migration`
- Affected code:
  - `backend-go/internal/domain/*`
  - `backend-go/internal/repository/*`
  - `backend-go/internal/service/*`
  - `backend-go/internal/transport/http/handler/*`
  - `backend-go/internal/transport/http/router.go`
