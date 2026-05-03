## 1. Settings model and persistence

- [ ] 1.1 Add the `AuditAlertSettings` domain model, defaults, validation rules, and serialization helpers covering retention, alert, notification, and export fields.
- [ ] 1.2 Add system-parameter repository/service support for loading and saving the governed audit settings JSON key in `core_system_param`.

## 2. Audit settings service and API surface

- [ ] 2.1 Implement audit settings service methods for get/save flows and for translating validation failures into explicit API responses.
- [ ] 2.2 Add audit settings handler endpoints for querying settings, saving settings groups, triggering cleanup now, and sending a test notification.
- [ ] 2.3 Wire the new audit settings routes into the audit router/startup path while preserving the existing `/api/audit` response envelope.

## 3. Alert detection and notification pipeline

- [ ] 3.1 Extend `AuditService` with `DetectAndAlert()` orchestration that evaluates failed login, permission change, and batch operation rules against persisted settings.
- [ ] 3.2 Implement the `AlertNotifier` abstraction and ship a `LogNotifier` that records alert dispatches as audit log entries, leaving email delivery as a future implementation seam.

## 4. Scheduled operations and runtime wiring

- [ ] 4.1 Register the scheduled retention-cleanup job so it uses persisted `retentionDays` and `cleanupFrequency` settings during recurring execution.
- [ ] 4.2 Register the scheduled alert-check job and connect it to the governed audit detection flow without duplicating existing login-failure persistence logic.

## 5. Verification

- [ ] 5.1 Add backend unit tests for settings serialization/validation, audit settings service behavior, and notifier dispatch behavior.
- [ ] 5.2 Add backend handler or integration tests covering settings save/load, cleanup-now execution, test notification, and alert-trigger rule paths.
- [ ] 5.3 Run the required backend verification for the change scope (`make test`, plus integration coverage if repository persistence paths are added or changed).
