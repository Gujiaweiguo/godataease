package service

import (
	"errors"
	"testing"
	"time"

	"dataease/backend/internal/domain/license"
)

// MockLicenseRepository implements LicenseRepository for testing
type MockLicenseRepository struct {
	storedResult *license.ValidateResult
	storedRaw    string
	loadErr      error
	saveErr      error
	clearErr     error
}

func NewMockLicenseRepository() *MockLicenseRepository {
	return &MockLicenseRepository{}
}

func (m *MockLicenseRepository) Load() (*license.ValidateResult, string, error) {
	if m.loadErr != nil {
		return nil, "", m.loadErr
	}
	return m.storedResult, m.storedRaw, nil
}

func (m *MockLicenseRepository) Save(result *license.ValidateResult, raw string) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.storedResult = result
	m.storedRaw = raw
	return nil
}

func (m *MockLicenseRepository) Clear() error {
	if m.clearErr != nil {
		return m.clearErr
	}
	m.storedResult = nil
	m.storedRaw = ""
	return nil
}

func setupLicenseService() *LicenseService {
	mockRepo := NewMockLicenseRepository()
	return NewLicenseService(mockRepo)
}

func TestLicense_Validate_EmptyLicense(t *testing.T) {
	svc := setupLicenseService()

	// Test with nil request
	result, err := svc.Validate(nil)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if result.Status != "invalid" {
		t.Errorf("Expected status 'invalid', got '%s'", result.Status)
	}
	if result.Message != "license not found" {
		t.Errorf("Expected message 'license not found', got '%s'", result.Message)
	}

	// Test with empty license
	result, err = svc.Validate(&license.LicenseRequest{License: ""})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if result.Status != "invalid" {
		t.Errorf("Expected status 'invalid', got '%s'", result.Status)
	}
}

func TestLicense_Validate_WithValidLicense(t *testing.T) {
	svc := setupLicenseService()

	// Test with valid license content
	result, err := svc.Validate(&license.LicenseRequest{License: "valid-license-content"})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if result.Status != "valid" {
		t.Errorf("Expected status 'valid', got '%s'", result.Status)
	}
	if result.License == nil {
		t.Fatal("Expected license info to be non-nil")
	}
}

func TestLicense_Validate_WithExpiredLicense(t *testing.T) {
	svc := setupLicenseService()

	// Test with expired license content
	result, err := svc.Validate(&license.LicenseRequest{License: "this license is expired"})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if result.Status != "expired" {
		t.Errorf("Expected status 'expired', got '%s'", result.Status)
	}
}

func TestLicense_Validate_WithShortLicense(t *testing.T) {
	svc := setupLicenseService()

	// Test with too short license content
	result, err := svc.Validate(&license.LicenseRequest{License: "short"})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if result.Status != "invalid" {
		t.Errorf("Expected status 'invalid', got '%s'", result.Status)
	}
	if result.Message != "invalid license content" {
		t.Errorf("Expected message 'invalid license content', got '%s'", result.Message)
	}
}

func TestLicense_Validate_StoredLicense(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.storedResult = &license.ValidateResult{
		Status: "valid",
		License: &license.LicenseInfo{
			Corporation: "Test Corp",
			Expired:     time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
		},
	}
	svc := NewLicenseService(mockRepo)

	// Validate with empty request should return stored license
	result, err := svc.Validate(&license.LicenseRequest{License: ""})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if result.Status != "valid" {
		t.Errorf("Expected status 'valid', got '%s'", result.Status)
	}
}

func TestLicense_Version_Default(t *testing.T) {
	svc := setupLicenseService()

	version := svc.Version()
	if version == "" {
		t.Error("Expected non-empty version")
	}
	// Default version when no env vars set
	if version != "v2-go-dev" {
		t.Logf("Version: %s (may be set by env var)", version)
	}
}

func TestLicense_Update_Success(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	svc := NewLicenseService(mockRepo)

	result, err := svc.Update(&license.LicenseRequest{License: "valid-license-content"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result.Status != "valid" {
		t.Errorf("Expected status 'valid', got '%s'", result.Status)
	}
	if mockRepo.storedResult == nil {
		t.Error("Expected license to be saved")
	}
}

func TestLicense_Update_EmptyLicense(t *testing.T) {
	svc := setupLicenseService()

	result, err := svc.Update(&license.LicenseRequest{License: ""})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result.Status != "invalid" {
		t.Errorf("Expected status 'invalid', got '%s'", result.Status)
	}
	if result.Message != "license is required" {
		t.Errorf("Expected message 'license is required', got '%s'", result.Message)
	}
}

func TestLicense_Update_ExpiredLicense(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	svc := NewLicenseService(mockRepo)

	result, err := svc.Update(&license.LicenseRequest{License: "this license is expired"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result.Status != "expired" {
		t.Errorf("Expected status 'expired', got '%s'", result.Status)
	}
	// Expired license should not be saved
	if mockRepo.storedResult != nil {
		t.Error("Expired license should not be saved")
	}
}

func TestLicense_Revert(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	svc := NewLicenseService(mockRepo)

	// First store a license
	_, _ = svc.Update(&license.LicenseRequest{License: "valid-license-content"})

	err := svc.Revert()
	if err != nil {
		t.Fatalf("Revert failed: %v", err)
	}
	if mockRepo.storedResult != nil {
		t.Error("Expected license to be cleared")
	}
}

func TestLicense_IsLicenseValid_NoLicense(t *testing.T) {
	svc := setupLicenseService()

	valid := svc.IsLicenseValid()
	if valid {
		t.Error("Expected IsLicenseValid to return false when no license")
	}
}

func TestLicense_IsLicenseValid_WithValidLicense(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.storedResult = &license.ValidateResult{
		Status: "valid",
		License: &license.LicenseInfo{
			Expired: time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
		},
	}
	svc := NewLicenseService(mockRepo)

	valid := svc.IsLicenseValid()
	if !valid {
		t.Error("Expected IsLicenseValid to return true for valid license")
	}
}

func TestLicense_IsLicenseValid_ExpiredLicense(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.storedResult = &license.ValidateResult{
		Status: "valid",
		License: &license.LicenseInfo{
			Expired: time.Now().AddDate(-1, 0, 0).Format("2006-01-02"),
		},
	}
	svc := NewLicenseService(mockRepo)

	valid := svc.IsLicenseValid()
	if valid {
		t.Error("Expected IsLicenseValid to return false for expired license")
	}
}

func TestLicense_IsLicenseValid_InvalidStatus(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.storedResult = &license.ValidateResult{
		Status:  "invalid",
		License: &license.LicenseInfo{},
	}
	svc := NewLicenseService(mockRepo)

	valid := svc.IsLicenseValid()
	if valid {
		t.Error("Expected IsLicenseValid to return false for invalid status")
	}
}

func TestLicense_GetExpiryWarning_NoLicense(t *testing.T) {
	svc := setupLicenseService()

	warning := svc.GetExpiryWarning()
	if warning == nil {
		t.Fatal("Expected non-nil warning")
	}
	if warning.WarningLevel != "critical" {
		t.Errorf("Expected warning level 'critical', got '%s'", warning.WarningLevel)
	}
	if warning.Message != "License is invalid or not found" {
		t.Errorf("Unexpected message: %s", warning.Message)
	}
}

func TestLicense_GetExpiryWarning_ExpiredLicense(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.storedResult = &license.ValidateResult{
		Status: "valid",
		License: &license.LicenseInfo{
			Expired: time.Now().AddDate(-1, 0, 0).Format("2006-01-02"),
		},
	}
	svc := NewLicenseService(mockRepo)

	warning := svc.GetExpiryWarning()
	if warning == nil {
		t.Fatal("Expected non-nil warning")
	}
	if warning.WarningLevel != "critical" {
		t.Errorf("Expected warning level 'critical', got '%s'", warning.WarningLevel)
	}
	if warning.Message != "License has expired" {
		t.Errorf("Unexpected message: %s", warning.Message)
	}
	if !warning.IsExpiringSoon {
		t.Error("Expected IsExpiringSoon to be true")
	}
}

func TestLicense_GetExpiryWarning_ExpiresIn5Days(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.storedResult = &license.ValidateResult{
		Status: "valid",
		License: &license.LicenseInfo{
			Expired: time.Now().AddDate(0, 0, 5).Format("2006-01-02"),
		},
	}
	svc := NewLicenseService(mockRepo)

	warning := svc.GetExpiryWarning()
	if warning.WarningLevel != "critical" {
		t.Errorf("Expected warning level 'critical', got '%s'", warning.WarningLevel)
	}
	if warning.Message != "License expires in less than 7 days" {
		t.Errorf("Unexpected message: %s", warning.Message)
	}
}

func TestLicense_GetExpiryWarning_ExpiresIn10Days(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.storedResult = &license.ValidateResult{
		Status: "valid",
		License: &license.LicenseInfo{
			Expired: time.Now().AddDate(0, 0, 10).Format("2006-01-02"),
		},
	}
	svc := NewLicenseService(mockRepo)

	warning := svc.GetExpiryWarning()
	if warning.WarningLevel != "warning" {
		t.Errorf("Expected warning level 'warning', got '%s'", warning.WarningLevel)
	}
}

func TestLicense_GetExpiryWarning_ExpiresIn25Days(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.storedResult = &license.ValidateResult{
		Status: "valid",
		License: &license.LicenseInfo{
			Expired: time.Now().AddDate(0, 0, 25).Format("2006-01-02"),
		},
	}
	svc := NewLicenseService(mockRepo)

	warning := svc.GetExpiryWarning()
	if warning.WarningLevel != "info" {
		t.Errorf("Expected warning level 'info', got '%s'", warning.WarningLevel)
	}
}

func TestLicense_GetExpiryWarning_ExpiresIn60Days(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.storedResult = &license.ValidateResult{
		Status: "valid",
		License: &license.LicenseInfo{
			Expired: time.Now().AddDate(0, 0, 60).Format("2006-01-02"),
		},
	}
	svc := NewLicenseService(mockRepo)

	warning := svc.GetExpiryWarning()
	if warning.WarningLevel != "none" {
		t.Errorf("Expected warning level 'none', got '%s'", warning.WarningLevel)
	}
	if warning.Message != "License is valid" {
		t.Errorf("Unexpected message: %s", warning.Message)
	}
}

func TestLicense_GetExpiryWarning_InvalidDateFormat(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.storedResult = &license.ValidateResult{
		Status: "valid",
		License: &license.LicenseInfo{
			Expired: "invalid-date",
		},
	}
	svc := NewLicenseService(mockRepo)

	warning := svc.GetExpiryWarning()
	if warning.WarningLevel != "critical" {
		t.Errorf("Expected warning level 'critical', got '%s'", warning.WarningLevel)
	}
	if warning.Message != "Invalid license expiry date format" {
		t.Errorf("Unexpected message: %s", warning.Message)
	}
}

func TestLicense_LoadError(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.loadErr = errors.New("database error")
	svc := NewLicenseService(mockRepo)

	result, err := svc.Validate(nil)
	if err == nil {
		t.Error("Expected error when load fails")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
}

func TestLicense_Update_SaveError(t *testing.T) {
	mockRepo := NewMockLicenseRepository()
	mockRepo.saveErr = errors.New("save failed")
	svc := NewLicenseService(mockRepo)

	_, err := svc.Update(&license.LicenseRequest{License: "valid-license-content"})
	if err == nil {
		t.Error("Expected error when save fails")
	}
}

func TestLicense_BuildLicenseResult_WithJSON(t *testing.T) {
	jsonLicense := `{"corporation":"Test Corp","exp":"2026-12-31","count":50}`
	result := buildLicenseResult(jsonLicense)

	if result.Status != "valid" {
		t.Errorf("Expected status 'valid', got '%s'", result.Status)
	}
	if result.License == nil {
		t.Fatal("Expected license info")
	}
	if result.License.Corporation != "Test Corp" {
		t.Errorf("Expected corporation 'Test Corp', got '%s'", result.License.Corporation)
	}
	if result.License.Count != 50 {
		t.Errorf("Expected count 50, got %d", result.License.Count)
	}
}

func TestLicense_BuildLicenseResult_ContainsExpired(t *testing.T) {
	jsonLicense := `{"corporation":"Test Corp","expired":"2020-01-01","count":50}`
	result := buildLicenseResult(jsonLicense)

	if result.Status != "expired" {
		t.Errorf("Expected status 'expired' when contains 'expired', got '%s'", result.Status)
	}
}
