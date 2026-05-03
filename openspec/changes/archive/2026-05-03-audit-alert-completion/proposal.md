## Why

`apps/frontend/src/views/audit/settings.vue` already exposes a full audit settings experience, but the Go backend has no persistence model for those settings, no alert detection pipeline beyond a boolean login-failure helper, and no notification dispatch path. As a result, every save or test action in the UI is a ghost interaction that reports success without changing backend state.

This change closes that product gap by defining the missing audit alert capability and extending the existing Go audit capability so audit settings, scheduled cleanup, alert detection, and notification behaviors become explicit, implementable, and verifiable.

## What Changes

- Add a new `audit-alert` capability covering audit settings persistence, suspicious-activity detection rules, notification dispatch contracts, and test-notification behavior.
- Extend the Go audit capability to support persisted retention/export settings, scheduled cleanup execution, and integration points for alert checks.
- Define backend API expectations for saving audit settings groups, triggering cleanup immediately, and sending a test alert.
- Document the affected backend layers: audit domain models, system parameter storage, audit service/handler/router wiring, and scheduler job registration.

## Capabilities

### New Capabilities
- `audit-alert`: persist audit alert/settings state, evaluate governed alert rules, dispatch alert notifications through notifier implementations, and support test notification flows.

### Modified Capabilities
- `audit-go`: expand Go audit behavior to include persisted settings-backed cleanup/export configuration, scheduled cleanup execution, and alert-service integration points.

## Impact

- **Backend domain**: add `AuditAlertSettings` and related request/response contracts while reusing `core_system_param` JSON storage conventions.
- **Backend services**: extend `AuditService`, add alert-settings accessors, notifier abstractions, and scheduler-driven cleanup/alert orchestration.
- **Backend API**: add audit settings CRUD-style endpoints plus immediate cleanup and test-notification endpoints under the audit surface.
- **Operations**: register recurring cleanup and alert-check jobs using the scheduler foundation already available in the repository.
- **Frontend compatibility**: align the existing audit settings page with real backend persistence and execution semantics without requiring frontend redesign.
