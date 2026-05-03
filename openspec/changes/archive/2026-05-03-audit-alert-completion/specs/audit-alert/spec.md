## ADDED Requirements

### Requirement: Audit alert settings SHALL be persisted as a governed backend configuration
The system SHALL persist audit alert settings as a typed backend configuration stored through the existing `core_system_param` infrastructure, with defaults covering retention, alert, notification, and export fields required by the audit settings UI.

#### Scenario: Read default audit alert settings
- **WHEN** no persisted audit alert settings exist
- **THEN** the system SHALL return a complete `AuditAlertSettings` payload with backend defaults for `enableAlerts`, `failedLoginThreshold`, `alertOnPermissionChange`, `alertOnSensitiveAccess`, `batchOperationThreshold`, `enableEmailNotification`, `notificationEmail`, `enableSystemNotification`, `retentionDays`, `cleanupFrequency`, `defaultExportFormat`, and `exportLimit`

#### Scenario: Save audit alert settings
- **WHEN** an authorized client saves audit alert settings
- **THEN** the system SHALL validate the payload and persist the full settings aggregate through the governed system-parameter storage key

#### Scenario: Reject invalid audit alert settings
- **WHEN** an authorized client submits an invalid audit alert settings payload such as an out-of-range threshold, unsupported cleanup frequency, or unsupported export format
- **THEN** the system SHALL reject the request with an explicit validation failure

### Requirement: Audit alert rules SHALL evaluate governed suspicious activity types
The system SHALL evaluate audit alert rules for failed login attempts, permission changes, and high-volume batch operations by consulting persisted audit alert settings before triggering notifications.

#### Scenario: Failed login threshold triggers alert
- **WHEN** failed login attempts for a username meet or exceed the configured threshold within the governed detection window and alerts are enabled
- **THEN** the system SHALL create an alert event for the failed login rule

#### Scenario: Permission change alert respects settings
- **WHEN** a permission change audit event is processed and `alertOnPermissionChange` is enabled
- **THEN** the system SHALL create an alert event for that permission change

#### Scenario: Batch operation alert respects settings
- **WHEN** audit events for the same actor meet or exceed the configured batch-operation threshold within the governed batch window and alerts are enabled
- **THEN** the system SHALL create an alert event for the batch-operation rule

#### Scenario: Disabled alerts suppress notifications
- **WHEN** alert rules match but `enableAlerts` is disabled
- **THEN** the system SHALL record no outbound notification dispatch for that evaluation cycle

### Requirement: Audit alerts SHALL dispatch through notifier implementations
The system SHALL dispatch audit alerts through an `AlertNotifier` abstraction and MUST provide a log-backed notifier implementation in this change.

#### Scenario: Log notifier records alert delivery
- **WHEN** an alert is dispatched through the shipped notifier implementation
- **THEN** the system SHALL write a corresponding audit log entry describing the alert type, detection context, and delivery outcome

#### Scenario: Email delivery remains a future implementation
- **WHEN** email notification is enabled in settings
- **THEN** the system SHALL preserve the email-related settings fields and notifier interface contract without requiring SMTP delivery in this change

### Requirement: Audit alert operations SHALL support test notification and immediate cleanup actions
The system SHALL expose backend operations for sending a test notification and for triggering audit cleanup immediately using persisted retention settings.

#### Scenario: Send test notification
- **WHEN** an authorized client requests a test notification
- **THEN** the system SHALL invoke the notifier pipeline with a test alert payload and return an explicit success or failure result

#### Scenario: Trigger cleanup now
- **WHEN** an authorized client requests immediate cleanup
- **THEN** the system SHALL execute retention cleanup using the configured retention settings and return the cleanup result summary
