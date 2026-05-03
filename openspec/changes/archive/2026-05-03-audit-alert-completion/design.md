## Context

The current Go audit backend implements log creation, query, export, retention deletion, and login-failure recording, but it stops short of the settings and alert-management surface already exposed by the frontend audit settings page. The UI currently assumes persisted retention, alert, notification, and export preferences plus operational endpoints for immediate cleanup and test notifications, yet no backend model, storage key, handler, or scheduler job exists for those behaviors.

The repository already has a reusable system-parameter pattern backed by `core_system_param` plus a scheduler foundation that can host recurring jobs. This design uses those two existing foundations rather than introducing a new audit settings table or a standalone notification subsystem.

## Goals / Non-Goals

**Goals:**
- Persist audit settings expected by the frontend, including alert, notification, retention, and export fields.
- Extend the audit service layer with governed alert-detection entry points for failed logins, permission changes, and batch operations.
- Define a notifier abstraction with an immediately shippable in-product implementation that writes alert events back into audit logs.
- Register scheduled cleanup and scheduled alert-check jobs using the existing scheduler foundation.
- Expose backend audit endpoints for reading/saving settings, running cleanup immediately, and sending a test notification.

**Non-Goals:**
- Implementing real email delivery or external webhook dispatch in this change.
- Redesigning the audit settings frontend page.
- Introducing a dedicated `audit_alert_settings` database table.
- Building a generalized event-streaming or SIEM integration framework.

## Decisions

### Decision: Store `AuditAlertSettings` in `core_system_param` as JSON
The backend will define a typed `AuditAlertSettings` domain model, but persistence will reuse the existing system-parameter infrastructure rather than add a new table.

- **Why:** `SystemParamService` already manages JSON-shaped platform settings, provides an established audit trail pattern, and keeps deployment scope minimal.
- **Alternative considered:** create a dedicated audit settings table. Rejected because it adds schema migration and repository overhead for a single singleton-style configuration object that matches current system-param usage.

### Decision: Group all UI-backed fields into one governed settings aggregate
Retention, alert, notification, and export fields will be modeled as a single aggregate retrieved and saved through audit settings APIs.

- **Why:** the frontend works with one page-level settings object, and a single aggregate avoids fragmented persistence or partial defaults across four pseudo-independent forms.
- **Alternative considered:** persist each card separately under unrelated keys. Rejected because it increases drift risk and complicates defaults, validation, and migration.

### Decision: Extend `AuditService` with `DetectAndAlert()` orchestration
Alert evaluation will live in the audit service layer as a new orchestration entry point that consults stored settings and executes rule checks for failed logins, permission changes, and batch operations.

- **Why:** the existing audit service already owns audit-log creation and login-failure logic, so it is the narrowest place to coordinate alert checks without spreading rule logic across handlers or schedulers.
- **Alternative considered:** a completely separate alert service. Rejected because the first implementation still depends heavily on audit repositories and should minimize service sprawl.

### Decision: Use interface-based notifiers and ship `LogNotifier` first
The design introduces an `AlertNotifier` interface and a concrete `LogNotifier` that records alert events as audit log entries. Email remains a placeholder implementation point behind the same interface.

- **Why:** this makes notification dispatch testable and extensible while still delivering an immediately implementable alert channel.
- **Alternative considered:** hard-code alert logging into detection methods. Rejected because it couples rule evaluation to one delivery mechanism and blocks future email/system-notification expansion.

### Decision: Run cleanup and alert evaluation as scheduler jobs plus on-demand endpoints
Retention cleanup and alert checks will both support scheduled execution, while cleanup-now and test-notification endpoints provide operator-triggered control.

- **Why:** the frontend and operations model both require immediate actions and recurring enforcement.
- **Alternative considered:** only expose manual endpoints. Rejected because the settings page explicitly promises automated cleanup, and alerting loses value if no recurring evaluator exists.

### Decision: Keep email notification as declared future capability, not implemented behavior
The settings model will preserve email-related fields and validation requirements, but runtime delivery in this change will remain limited to notifier interfaces and the log-backed implementation.

- **Why:** this matches the current scope, avoids half-built SMTP dependencies, and still preserves the API contract the frontend already expects.
- **Alternative considered:** strip email fields from the contract until delivery exists. Rejected because it would further diverge the backend from the shipped frontend settings surface.

## Risks / Trade-offs

- **[Risk] Singleton JSON settings become loosely governed over time** → **Mitigation:** define a typed domain model, defaults, validation rules, and a stable system-param key owned by the audit module.
- **[Risk] Scheduler-driven alert checks duplicate alerts across repeated scans** → **Mitigation:** define rule windows, deduplication criteria, and alert log metadata in the implementation tasks.
- **[Risk] Log-based notification delivery is weaker than true email/system notices** → **Mitigation:** document `AlertNotifier` as the extension seam and keep email-specific settings as future-ready contract fields.
- **[Risk] Permission-change and batch-operation detection need event inputs not fully captured today** → **Mitigation:** implement the detection entry point and wiring so existing audit events can participate first, with stricter event sourcing left to follow-up work if needed.

## Migration Plan

1. Add the audit settings aggregate, defaults, validation rules, and system-param storage key.
2. Introduce service methods and handlers for get/save settings, cleanup-now, and test notification.
3. Extend audit detection/notifier logic and wire alert-trigger paths into audit operations.
4. Register scheduled cleanup and alert-check jobs in startup using the scheduler foundation.
5. Roll out with safe defaults that preserve current behavior when alerts are disabled or settings are absent.

Rollback remains low risk because the change reuses singleton config storage and additive routes; disabling the new routes/jobs and ignoring the stored settings key returns the backend to its current behavior.

## Open Questions

- Which existing audit operations should count toward the first implementation of batch-operation detection beyond generic create/update/delete bursts?
- Whether system notifications should initially surface only as audit log entries or also integrate with an existing in-app message center in a later follow-up.
