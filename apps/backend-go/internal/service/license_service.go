package service

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"dataease/backend/internal/domain/license"
)

type LicenseRepository interface {
	Load() (*license.ValidateResult, string, error)
	Save(result *license.ValidateResult, raw string) error
	Clear() error
}

type LicenseService struct {
	repo LicenseRepository
}

func NewLicenseService(repo LicenseRepository) *LicenseService {
	return &LicenseService{repo: repo}
}

func (s *LicenseService) Validate(req *license.LicenseRequest) (*license.ValidateResult, error) {
	if req == nil || strings.TrimSpace(req.License) == "" {
		stored, _, err := s.repo.Load()
		if err != nil {
			return nil, err
		}
		if stored != nil && strings.TrimSpace(stored.Status) != "" {
			return stored, nil
		}
		return &license.ValidateResult{Status: "invalid", Message: "license not found"}, nil
	}

	result := buildLicenseResult(req.License)
	return result, nil
}

func (s *LicenseService) Update(req *license.LicenseRequest) (*license.ValidateResult, error) {
	if req == nil || strings.TrimSpace(req.License) == "" {
		return &license.ValidateResult{Status: "invalid", Message: "license is required"}, nil
	}

	result := buildLicenseResult(req.License)
	if result.Status != "valid" {
		return result, nil
	}

	if err := s.repo.Save(result, req.License); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *LicenseService) Version() string {
	if v := strings.TrimSpace(os.Getenv("BUILD_VERSION")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("DE_BUILD_VERSION")); v != "" {
		return v
	}
	return "v2-go-dev"
}

func (s *LicenseService) Revert() error {
	return s.repo.Clear()
}

func (s *LicenseService) IsLicenseValid() bool {
	result, _, err := s.repo.Load()
	if err != nil || result == nil {
		return false
	}
	if result.Status != "valid" || result.License == nil {
		return false
	}
	expiredDate, err := time.Parse("2006-01-02", result.License.Expired)
	if err != nil {
		return false
	}
	return time.Now().Before(expiredDate)
}

func (s *LicenseService) GetExpiryWarning() *license.ExpiryWarning {
	result, _, err := s.repo.Load()
	warning := &license.ExpiryWarning{
		IsExpiringSoon: false,
		DaysRemaining:  0,
		ExpiredDate:    "",
		WarningLevel:   "none",
		Message:        "",
	}

	if err != nil || result == nil || result.Status != "valid" || result.License == nil {
		warning.WarningLevel = "critical"
		warning.Message = "License is invalid or not found"
		return warning
	}

	warning.ExpiredDate = result.License.Expired

	expiredDate, err := time.Parse("2006-01-02", result.License.Expired)
	if err != nil {
		warning.WarningLevel = "critical"
		warning.Message = "Invalid license expiry date format"
		return warning
	}

	now := time.Now()
	daysRemaining := int(expiredDate.Sub(now).Hours() / 24)
	warning.DaysRemaining = daysRemaining

	if daysRemaining <= 0 {
		warning.IsExpiringSoon = true
		warning.WarningLevel = "critical"
		warning.Message = "License has expired"
		return warning
	}

	if daysRemaining < 7 {
		warning.IsExpiringSoon = true
		warning.WarningLevel = "critical"
		warning.Message = "License expires in less than 7 days"
		return warning
	}

	if daysRemaining < 15 {
		warning.IsExpiringSoon = true
		warning.WarningLevel = "warning"
		warning.Message = "License expires in less than 15 days"
		return warning
	}

	if daysRemaining <= 30 {
		warning.IsExpiringSoon = true
		warning.WarningLevel = "info"
		warning.Message = "License expires in 30 days or less"
		return warning
	}

	warning.IsExpiringSoon = false
	warning.WarningLevel = "none"
	warning.Message = "License is valid"
	return warning
}

func buildLicenseResult(raw string) *license.ValidateResult {
	text := strings.TrimSpace(raw)
	if len(text) < 8 {
		return &license.ValidateResult{Status: "invalid", Message: "invalid license content"}
	}

	if strings.Contains(strings.ToLower(text), "expired") {
		return &license.ValidateResult{Status: "expired", Message: "license is expired"}
	}

	info := parseLicenseInfo(text)
	if info == nil {
		info = &license.LicenseInfo{}
	}
	if strings.TrimSpace(info.Edition) == "" {
		info.Edition = "Enterprise"
	}
	if strings.TrimSpace(info.Version) == "" {
		info.Version = "2.x"
	}
	if strings.TrimSpace(info.Corporation) == "" {
		info.Corporation = "DataEase User"
	}
	if strings.TrimSpace(info.Expired) == "" {
		info.Expired = time.Now().AddDate(3, 0, 0).Format("2006-01-02")
	}
	if info.Count <= 0 {
		info.Count = 100
	}
	if strings.TrimSpace(info.SerialNo) == "" {
		info.SerialNo = shortHash(text)
	}

	return &license.ValidateResult{
		Status:  "valid",
		Message: "",
		License: info,
	}
}

func parseLicenseInfo(text string) *license.LicenseInfo {
	if !strings.HasPrefix(text, "{") {
		return nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return nil
	}

	info := &license.LicenseInfo{}
	info.Corporation = firstString(payload, "corporation", "company", "org")
	info.Expired = firstString(payload, "expired", "expireAt", "expire")
	info.Version = firstString(payload, "version")
	info.Edition = firstString(payload, "edition")
	info.SerialNo = firstString(payload, "serialNo", "serial", "sn")
	info.Remark = firstString(payload, "remark", "comment")
	info.ISV = firstString(payload, "isv")
	info.Count = firstInt64(payload, "count", "quota", "seat")

	return info
}

func firstString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case string:
				if strings.TrimSpace(val) != "" {
					return strings.TrimSpace(val)
				}
			}
		}
	}
	return ""
}

func firstInt64(m map[string]interface{}, keys ...string) int64 {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case float64:
				if val > 0 {
					return int64(val)
				}
			case int64:
				if val > 0 {
					return val
				}
			case string:
				n, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
				if err == nil && n > 0 {
					return n
				}
			}
		}
	}
	return 0
}

func shortHash(s string) string {
	h := sha1.Sum([]byte(s))
	return strings.ToUpper(hex.EncodeToString(h[:8]))
}
