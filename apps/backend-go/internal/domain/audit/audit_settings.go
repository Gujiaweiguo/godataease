package audit

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
)

const (
	cleanupFrequencyDaily   = "daily"
	cleanupFrequencyWeekly  = "weekly"
	cleanupFrequencyMonthly = "monthly"

	exportFormatCSV  = "csv"
	exportFormatJSON = "json"
)

type AuditAlertSettings struct {
	RetentionDays            int    `json:"retentionDays"`
	CleanupFrequency         string `json:"cleanupFrequency"`
	EnableAlerts             bool   `json:"enableAlerts"`
	FailedLoginThreshold     int    `json:"failedLoginThreshold"`
	AlertOnPermissionChange  bool   `json:"alertOnPermissionChange"`
	AlertOnSensitiveAccess   bool   `json:"alertOnSensitiveAccess"`
	BatchOperationThreshold  int    `json:"batchOperationThreshold"`
	EnableEmailNotification  bool   `json:"enableEmailNotification"`
	NotificationEmail        string `json:"notificationEmail"`
	EnableSystemNotification bool   `json:"enableSystemNotification"`
	DefaultExportFormat      string `json:"defaultExportFormat"`
	ExportLimit              int    `json:"exportLimit"`
}

func DefaultAuditAlertSettings() *AuditAlertSettings {
	return &AuditAlertSettings{
		RetentionDays:            90,
		CleanupFrequency:         cleanupFrequencyWeekly,
		EnableAlerts:             true,
		FailedLoginThreshold:     5,
		AlertOnPermissionChange:  true,
		AlertOnSensitiveAccess:   false,
		BatchOperationThreshold:  50,
		EnableEmailNotification:  false,
		NotificationEmail:        "",
		EnableSystemNotification: true,
		DefaultExportFormat:      exportFormatCSV,
		ExportLimit:              1000,
	}
}

func (s *AuditAlertSettings) Validate() error {
	if s == nil {
		return fmt.Errorf("audit alert settings is nil")
	}
	if s.RetentionDays < 7 || s.RetentionDays > 365 {
		return fmt.Errorf("retentionDays must be between 7 and 365")
	}
	if !isSupportedCleanupFrequency(s.CleanupFrequency) {
		return fmt.Errorf("cleanupFrequency must be one of daily, weekly, monthly")
	}
	if s.FailedLoginThreshold < 1 || s.FailedLoginThreshold > 100 {
		return fmt.Errorf("failedLoginThreshold must be between 1 and 100")
	}
	if s.BatchOperationThreshold < 1 || s.BatchOperationThreshold > 10000 {
		return fmt.Errorf("batchOperationThreshold must be between 1 and 10000")
	}
	if !isSupportedExportFormat(s.DefaultExportFormat) {
		return fmt.Errorf("defaultExportFormat must be one of csv, json")
	}
	if s.ExportLimit < 1 || s.ExportLimit > 100000 {
		return fmt.Errorf("exportLimit must be between 1 and 100000")
	}
	if s.EnableEmailNotification {
		email := strings.TrimSpace(s.NotificationEmail)
		if email == "" {
			return fmt.Errorf("notificationEmail is required when email notification is enabled")
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return fmt.Errorf("notificationEmail is invalid: %w", err)
		}
	}
	return nil
}

func (s *AuditAlertSettings) ToJSON() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(s)
}

func AuditAlertSettingsFromJSON(data []byte) (*AuditAlertSettings, error) {
	if strings.TrimSpace(string(data)) == "" || strings.TrimSpace(string(data)) == "null" {
		return DefaultAuditAlertSettings(), nil
	}
	settings := DefaultAuditAlertSettings()
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}
	if err := settings.Validate(); err != nil {
		return nil, err
	}
	return settings, nil
}

func isSupportedCleanupFrequency(v string) bool {
	switch strings.TrimSpace(v) {
	case cleanupFrequencyDaily, cleanupFrequencyWeekly, cleanupFrequencyMonthly:
		return true
	default:
		return false
	}
}

func isSupportedExportFormat(v string) bool {
	switch strings.TrimSpace(v) {
	case exportFormatCSV, exportFormatJSON:
		return true
	default:
		return false
	}
}
