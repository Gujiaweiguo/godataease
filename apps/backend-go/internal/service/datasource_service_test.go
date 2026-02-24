package service

import (
	"encoding/base64"
	"encoding/json"
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

func TestDecodeMaybeBase64JSONMap(t *testing.T) {
	raw := map[string]interface{}{"name": "api_foo"}
	body, _ := json.Marshal(raw)
	encoded := base64.StdEncoding.EncodeToString(body)

	parsed, err := decodeMaybeBase64JSONMap(encoded)
	if err != nil {
		t.Fatalf("decodeMaybeBase64JSONMap failed: %v", err)
	}
	if parsed["name"] != "api_foo" {
		t.Fatalf("unexpected parsed payload: %#v", parsed)
	}
}

func TestParseDatasourceID(t *testing.T) {
	id, err := parseDatasourceID(map[string]string{"datasourceId": "12"})
	if err != nil {
		t.Fatalf("parseDatasourceID failed: %v", err)
	}
	if id != 12 {
		t.Fatalf("unexpected id: %d", id)
	}
}

func TestCheckAPIDatasource(t *testing.T) {
	raw := base64.StdEncoding.EncodeToString([]byte(`{"url":"https://example.com/api"}`))
	svc := &DatasourceService{}

	result, err := svc.CheckAPIDatasource(map[string]string{"data": raw, "type": "apiStructure"})
	if err != nil {
		t.Fatalf("CheckAPIDatasource failed: %v", err)
	}
	if result["type"] != "table" {
		t.Fatalf("expected type=table, got %#v", result["type"])
	}
	if result["showApiStructure"] != true {
		t.Fatalf("expected showApiStructure=true, got %#v", result["showApiStructure"])
	}
}
