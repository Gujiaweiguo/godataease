package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAuditAlertSettings(t *testing.T) {
	settings := DefaultAuditAlertSettings()
	require.NotNil(t, settings)

	assert.Equal(t, 90, settings.RetentionDays)
	assert.Equal(t, "weekly", settings.CleanupFrequency)
	assert.True(t, settings.EnableAlerts)
	assert.Equal(t, 5, settings.FailedLoginThreshold)
	assert.True(t, settings.AlertOnPermissionChange)
	assert.False(t, settings.AlertOnSensitiveAccess)
	assert.Equal(t, 50, settings.BatchOperationThreshold)
	assert.False(t, settings.EnableEmailNotification)
	assert.Empty(t, settings.NotificationEmail)
	assert.True(t, settings.EnableSystemNotification)
	assert.Equal(t, "csv", settings.DefaultExportFormat)
	assert.Equal(t, 1000, settings.ExportLimit)
	assert.NoError(t, settings.Validate())
}

func TestAuditAlertSettingsValidate_Success(t *testing.T) {
	settings := DefaultAuditAlertSettings()
	settings.CleanupFrequency = "daily"
	settings.DefaultExportFormat = "json"
	settings.EnableEmailNotification = true
	settings.NotificationEmail = "audit@example.com"
	settings.RetentionDays = 365
	settings.FailedLoginThreshold = 100
	settings.BatchOperationThreshold = 10000
	settings.ExportLimit = 100000

	assert.NoError(t, settings.Validate())
}

func TestAuditAlertSettingsValidate_RetentionDays(t *testing.T) {
	for _, retentionDays := range []int{6, 366} {
		settings := DefaultAuditAlertSettings()
		settings.RetentionDays = retentionDays
		assert.Error(t, settings.Validate())
	}
}

func TestAuditAlertSettingsValidate_InvalidCleanupFrequency(t *testing.T) {
	settings := DefaultAuditAlertSettings()
	settings.CleanupFrequency = "hourly"

	assert.Error(t, settings.Validate())
}

func TestAuditAlertSettingsValidate_InvalidExportFormat(t *testing.T) {
	settings := DefaultAuditAlertSettings()
	settings.DefaultExportFormat = "xml"

	assert.Error(t, settings.Validate())
}

func TestAuditAlertSettingsValidate_EmailValidation(t *testing.T) {
	t.Run("missing email", func(t *testing.T) {
		settings := DefaultAuditAlertSettings()
		settings.EnableEmailNotification = true
		settings.NotificationEmail = ""

		assert.Error(t, settings.Validate())
	})

	t.Run("invalid email", func(t *testing.T) {
		settings := DefaultAuditAlertSettings()
		settings.EnableEmailNotification = true
		settings.NotificationEmail = "invalid-email"

		assert.Error(t, settings.Validate())
	})

	t.Run("email ignored when disabled", func(t *testing.T) {
		settings := DefaultAuditAlertSettings()
		settings.NotificationEmail = "invalid-email"

		assert.NoError(t, settings.Validate())
	})
}

func TestAuditAlertSettingsValidate_ThresholdRanges(t *testing.T) {
	t.Run("failed login threshold", func(t *testing.T) {
		for _, threshold := range []int{0, 101} {
			settings := DefaultAuditAlertSettings()
			settings.FailedLoginThreshold = threshold
			assert.Error(t, settings.Validate())
		}
	})

	t.Run("batch operation threshold", func(t *testing.T) {
		for _, threshold := range []int{0, 10001} {
			settings := DefaultAuditAlertSettings()
			settings.BatchOperationThreshold = threshold
			assert.Error(t, settings.Validate())
		}
	})

	t.Run("export limit", func(t *testing.T) {
		for _, limit := range []int{0, 100001} {
			settings := DefaultAuditAlertSettings()
			settings.ExportLimit = limit
			assert.Error(t, settings.Validate())
		}
	})
}

func TestAuditAlertSettingsJSONRoundTrip(t *testing.T) {
	settings := DefaultAuditAlertSettings()
	settings.EnableEmailNotification = true
	settings.NotificationEmail = "audit@example.com"
	settings.DefaultExportFormat = "json"

	data, err := settings.ToJSON()
	require.NoError(t, err)

	decoded, err := AuditAlertSettingsFromJSON(data)
	require.NoError(t, err)
	assert.Equal(t, settings, decoded)
}

func TestAuditAlertSettingsFromJSON_EmptyUsesDefaults(t *testing.T) {
	settings, err := AuditAlertSettingsFromJSON(nil)
	require.NoError(t, err)
	assert.Equal(t, DefaultAuditAlertSettings(), settings)
}
