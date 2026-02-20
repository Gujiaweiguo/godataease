package embedded

import (
	"strings"
	"testing"
)

func TestConstants(t *testing.T) {
	if DefaultSecretLength != 16 {
		t.Errorf("Expected DefaultSecretLength 16, got %d", DefaultSecretLength)
	}
	if TokenExpireTime != 86400000 {
		t.Errorf("Expected TokenExpireTime 86400000, got %d", TokenExpireTime)
	}
}

func TestCoreEmbedded_TableName(t *testing.T) {
	e := CoreEmbedded{}
	if e.TableName() != "core_embedded" {
		t.Errorf("Expected table name 'core_embedded', got '%s'", e.TableName())
	}
}

func TestCoreEmbedded_Fields(t *testing.T) {
	e := CoreEmbedded{
		ID:           1,
		Name:         "Test App",
		AppId:        "app_123",
		AppSecret:    "secret123",
		Domain:       "https://example.com",
		SecretLength: 16,
		CreateTime:   1700000000,
		UpdateBy:     "admin",
		UpdateTime:   1700001000,
	}

	if e.ID != 1 {
		t.Errorf("Expected ID 1, got %d", e.ID)
	}
	if e.Name != "Test App" {
		t.Errorf("Expected Name 'Test App', got '%s'", e.Name)
	}
	if e.SecretLength != 16 {
		t.Errorf("Expected SecretLength 16, got %d", e.SecretLength)
	}
}

func TestEmbeddedCreator_Fields(t *testing.T) {
	secretLength := 32
	creator := EmbeddedCreator{
		Name:         "New App",
		Domain:       "https://newapp.com",
		SecretLength: &secretLength,
	}

	if creator.Name != "New App" {
		t.Errorf("Expected Name 'New App', got '%s'", creator.Name)
	}
	if *creator.SecretLength != 32 {
		t.Errorf("Expected SecretLength 32, got %d", *creator.SecretLength)
	}
}

func TestEmbeddedEditor_Fields(t *testing.T) {
	domain := "https://updated.com"
	secretLength := 24

	editor := EmbeddedEditor{
		ID:           1,
		Name:         "Updated App",
		Domain:       &domain,
		SecretLength: &secretLength,
	}

	if editor.ID != 1 {
		t.Errorf("Expected ID 1, got %d", editor.ID)
	}
	if *editor.Domain != "https://updated.com" {
		t.Errorf("Expected Domain 'https://updated.com', got '%s'", *editor.Domain)
	}
}

func TestEmbeddedGridVO_Fields(t *testing.T) {
	vo := EmbeddedGridVO{
		ID:           1,
		Name:         "Test App",
		AppId:        "app_123",
		AppSecret:    "secret123",
		Domain:       "https://example.com",
		SecretLength: 16,
	}

	if vo.ID != 1 {
		t.Errorf("Expected ID 1, got %d", vo.ID)
	}
	if vo.AppId != "app_123" {
		t.Errorf("Expected AppId 'app_123', got '%s'", vo.AppId)
	}
}

func TestEmbeddedPagerResponse_Fields(t *testing.T) {
	list := []EmbeddedGridVO{
		{ID: 1, Name: "App 1"},
		{ID: 2, Name: "App 2"},
	}

	resp := EmbeddedPagerResponse{
		List:    list,
		Total:   2,
		Current: 1,
		Size:    10,
	}

	if len(resp.List) != 2 {
		t.Errorf("Expected 2 items, got %d", len(resp.List))
	}
	if resp.Total != 2 {
		t.Errorf("Expected Total 2, got %d", resp.Total)
	}
}

func TestGenerateAppId(t *testing.T) {
	appId := GenerateAppId()

	if !strings.HasPrefix(appId, "app_") {
		t.Errorf("Expected AppId to start with 'app_', got '%s'", appId)
	}
	if len(appId) <= 4 {
		t.Error("Expected AppId to have content after 'app_' prefix")
	}
}

func TestGenerateAppSecret(t *testing.T) {
	secret := GenerateAppSecret(16)

	if len(secret) != 16 {
		t.Errorf("Expected secret length 16, got %d", len(secret))
	}

	secret32 := GenerateAppSecret(32)
	if len(secret32) != 32 {
		t.Errorf("Expected secret length 32, got %d", len(secret32))
	}
}

func TestGenerateAppSecret_DefaultLength(t *testing.T) {
	secret := GenerateAppSecret(0)
	if len(secret) != DefaultSecretLength {
		t.Errorf("Expected default secret length %d, got %d", DefaultSecretLength, len(secret))
	}

	secretNeg := GenerateAppSecret(-5)
	if len(secretNeg) != DefaultSecretLength {
		t.Errorf("Expected default secret length %d for negative input, got %d", DefaultSecretLength, len(secretNeg))
	}
}

func TestMaskAppSecret(t *testing.T) {
	tests := []struct {
		secret   string
		expected string
	}{
		{"", ""},
		{"short", "********"},
		{"12345678", "********"},
		{"1234567890", "1234****7890"},
		{"abcdefghijklmnopqrstuvwxyz", "abcd****wxyz"},
	}

	for _, tt := range tests {
		result := MaskAppSecret(tt.secret)
		if result != tt.expected {
			t.Errorf("MaskAppSecret(%q) = %q, expected %q", tt.secret, result, tt.expected)
		}
	}
}

func TestNormalizeOrigin(t *testing.T) {
	tests := []struct {
		origin   string
		expected string
	}{
		{"", ""},
		{"https://example.com", "https://example.com"},
		{"https://example.com/", "https://example.com"},
		{"https://example.com///", "https://example.com"},
		{"  https://example.com/  ", "https://example.com"},
	}

	for _, tt := range tests {
		result := NormalizeOrigin(tt.origin)
		if result != tt.expected {
			t.Errorf("NormalizeOrigin(%q) = %q, expected %q", tt.origin, result, tt.expected)
		}
	}
}

func TestParseDomains(t *testing.T) {
	tests := []struct {
		domainList string
		expected   int
	}{
		{"", 0},
		{"https://example.com", 1},
		{"https://a.com,https://b.com", 2},
		{"https://a.com;https://b.com", 2},
		{"https://a.com https://b.com", 2},
		{"https://a.com/, https://b.com/", 2},
	}

	for _, tt := range tests {
		result := ParseDomains(tt.domainList)
		if len(result) != tt.expected {
			t.Errorf("ParseDomains(%q) returned %d domains, expected %d", tt.domainList, len(result), tt.expected)
		}
	}
}

func TestIsOriginAllowed(t *testing.T) {
	domainList := "https://allowed.com,https://trusted.com"

	if !IsOriginAllowed("https://allowed.com", domainList) {
		t.Error("Expected https://allowed.com to be allowed")
	}
	if !IsOriginAllowed("https://allowed.com/", domainList) {
		t.Error("Expected https://allowed.com/ (with trailing slash) to be allowed")
	}
	if IsOriginAllowed("https://blocked.com", domainList) {
		t.Error("Expected https://blocked.com to be blocked")
	}
	if IsOriginAllowed("", domainList) {
		t.Error("Expected empty origin to be blocked")
	}
	if IsOriginAllowed("https://test.com", "") {
		t.Error("Expected any origin to be blocked when domainList is empty")
	}
}

func TestTokenArgsResponse_Fields(t *testing.T) {
	resp := TokenArgsResponse{
		UserId: 1,
		OrgId:  10,
	}

	if resp.UserId != 1 {
		t.Errorf("Expected UserId 1, got %d", resp.UserId)
	}
	if resp.OrgId != 10 {
		t.Errorf("Expected OrgId 10, got %d", resp.OrgId)
	}
}
