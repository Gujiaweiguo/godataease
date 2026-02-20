package license

import (
	"testing"
)

func TestLicenseRequest_Fields(t *testing.T) {
	req := LicenseRequest{License: "license-key-123"}
	if req.License != "license-key-123" {
		t.Errorf("Expected License 'license-key-123', got '%s'", req.License)
	}
}

func TestLicenseInfo_Fields(t *testing.T) {
	info := LicenseInfo{
		Corporation: "Test Corp",
		Expired:     "2025-12-31",
		Count:       100,
		Version:     "2.0",
		Edition:     "Enterprise",
		SerialNo:    "SN-12345",
		Remark:      "Test license",
		ISV:         "DataEase",
	}

	if info.Corporation != "Test Corp" {
		t.Errorf("Expected Corporation 'Test Corp', got '%s'", info.Corporation)
	}
	if info.Count != 100 {
		t.Errorf("Expected Count 100, got %d", info.Count)
	}
	if info.Edition != "Enterprise" {
		t.Errorf("Expected Edition 'Enterprise', got '%s'", info.Edition)
	}
}

func TestValidateResult_Fields(t *testing.T) {
	license := &LicenseInfo{
		Corporation: "Test Corp",
		Expired:     "2025-12-31",
	}

	result := ValidateResult{
		Status:  "valid",
		Message: "License is valid",
		License: license,
	}

	if result.Status != "valid" {
		t.Errorf("Expected Status 'valid', got '%s'", result.Status)
	}
	if result.License == nil {
		t.Error("Expected License to be non-nil")
	}
}

func TestValidateResult_NoLicense(t *testing.T) {
	result := ValidateResult{
		Status:  "invalid",
		Message: "License not found",
		License: nil,
	}

	if result.Status != "invalid" {
		t.Errorf("Expected Status 'invalid', got '%s'", result.Status)
	}
	if result.License != nil {
		t.Error("Expected License to be nil")
	}
}

func TestExpiryWarning_Fields(t *testing.T) {
	warning := ExpiryWarning{
		IsExpiringSoon: true,
		DaysRemaining:  30,
		ExpiredDate:    "2025-12-31",
		WarningLevel:   "warning",
		Message:        "License expires in 30 days",
	}

	if !warning.IsExpiringSoon {
		t.Error("Expected IsExpiringSoon to be true")
	}
	if warning.DaysRemaining != 30 {
		t.Errorf("Expected DaysRemaining 30, got %d", warning.DaysRemaining)
	}
	if warning.WarningLevel != "warning" {
		t.Errorf("Expected WarningLevel 'warning', got '%s'", warning.WarningLevel)
	}
}

func TestExpiryWarning_NoWarning(t *testing.T) {
	warning := ExpiryWarning{
		IsExpiringSoon: false,
		DaysRemaining:  365,
		WarningLevel:   "none",
		Message:        "",
	}

	if warning.IsExpiringSoon {
		t.Error("Expected IsExpiringSoon to be false")
	}
	if warning.WarningLevel != "none" {
		t.Errorf("Expected WarningLevel 'none', got '%s'", warning.WarningLevel)
	}
}
