package service

import (
	"encoding/base64"
	"testing"

	"dataease/backend/internal/domain/datasource"
)

func TestDecodeConfig_Base64(t *testing.T) {
	raw := `{"host":"127.0.0.1","port":3306}`
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))

	cfg, err := decodeConfig(encoded)
	if err != nil {
		t.Fatalf("decodeConfig failed: %v", err)
	}
	if cfg.Host != "127.0.0.1" || cfg.Port != 3306 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestDecodeConfig_RawJSON(t *testing.T) {
	cfg, err := decodeConfig(`{"host":"db.local","port":5432}`)
	if err != nil {
		t.Fatalf("decodeConfig failed: %v", err)
	}
	if cfg.Host != "db.local" || cfg.Port != 5432 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestParseHostPort_FromJDBCUrl(t *testing.T) {
	host, port := parseHostPort(&datasource.ConnectionConfig{
		JDBCUrl: "jdbc:mysql://10.0.0.8:3306/dataease",
	})

	if host != "10.0.0.8" || port != 3306 {
		t.Fatalf("unexpected host/port: %s:%d", host, port)
	}
}
