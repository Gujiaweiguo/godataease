package system

import (
	"testing"
)

func TestSettingItem_Fields(t *testing.T) {
	item := SettingItem{
		Pkey: "ui.logo",
		Pval: "https://example.com/logo.png",
		Type: "ui",
		Sort: 1,
	}

	if item.Pkey != "ui.logo" {
		t.Errorf("Expected Pkey 'ui.logo', got '%s'", item.Pkey)
	}
	if item.Pval != "https://example.com/logo.png" {
		t.Errorf("Expected Pval 'https://example.com/logo.png', got '%s'", item.Pval)
	}
	if item.Type != "ui" {
		t.Errorf("Expected Type 'ui', got '%s'", item.Type)
	}
	if item.Sort != 1 {
		t.Errorf("Expected Sort 1, got %d", item.Sort)
	}
}

func TestOnlineMapEditor_Fields(t *testing.T) {
	editor := OnlineMapEditor{
		MapType:      "gaode",
		Key:          "your-api-key",
		SecurityCode: "your-security-code",
	}

	if editor.MapType != "gaode" {
		t.Errorf("Expected MapType 'gaode', got '%s'", editor.MapType)
	}
	if editor.Key != "your-api-key" {
		t.Errorf("Expected Key 'your-api-key', got '%s'", editor.Key)
	}
}

func TestSQLBotConfig_Fields(t *testing.T) {
	cfg := SQLBotConfig{
		Domain:  "https://bot.example.com",
		ID:      "bot-123",
		Enabled: true,
		Valid:   true,
	}

	if cfg.Domain != "https://bot.example.com" {
		t.Errorf("Expected Domain 'https://bot.example.com', got '%s'", cfg.Domain)
	}
	if cfg.ID != "bot-123" {
		t.Errorf("Expected ID 'bot-123', got '%s'", cfg.ID)
	}
	if !cfg.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if !cfg.Valid {
		t.Error("Expected Valid to be true")
	}
}

func TestShareBase_Fields(t *testing.T) {
	share := ShareBase{
		Disable:   false,
		PERequire: true,
	}

	if share.Disable {
		t.Error("Expected Disable to be false")
	}
	if !share.PERequire {
		t.Error("Expected PERequire to be true")
	}
}
