package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"dataease/backend/internal/domain/datasource"
	"github.com/xuri/excelize/v2"
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

func TestPingTCP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().(*net.TCPAddr)
	if err = pingTCP(addr.IP.String(), addr.Port, time.Second); err != nil {
		t.Fatalf("pingTCP should connect: %v", err)
	}
}

func TestDatasourceService_UploadAndLoadRemoteFile(t *testing.T) {
	svc := NewDatasourceService(nil)

	csvContent := "name,amount\nAlice,100\n"
	tmp, err := os.CreateTemp("", "ds-upload-*.csv")
	if err != nil {
		t.Fatalf("create temp file failed: %v", err)
	}
	defer os.Remove(tmp.Name())
	_, _ = tmp.WriteString(csvContent)
	_, _ = tmp.Seek(0, 0)
	defer tmp.Close()

	header := &multipart.FileHeader{Filename: "ds.csv", Size: int64(len(csvContent))}
	data, err := svc.UploadFile(tmp, header, 1, 0)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}
	if data == nil || len(data.Sheets) == 0 {
		t.Fatal("expected uploaded data with sheets")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(csvContent))
	}))
	defer server.Close()

	remote, err := svc.LoadRemoteFile(server.URL+"/remote.csv", "", "", 1)
	if err != nil {
		t.Fatalf("LoadRemoteFile failed: %v", err)
	}
	if remote == nil || len(remote.Sheets) == 0 {
		t.Fatal("expected remote data with sheets")
	}
}

func TestDatasourceService_ListSyncRecord_InvalidAndNoRepo(t *testing.T) {
	svc := NewDatasourceService(nil)

	_, err := svc.ListSyncRecord(0, 1, 10)
	if err == nil {
		t.Fatal("expected invalid datasource id error")
	}

	_, err = svc.ListSyncRecord(1, 1, 10)
	if err == nil {
		t.Fatal("expected repository unavailable error")
	}
}

func TestDatasourceService_SyncAPIWrappers_ErrorPaths(t *testing.T) {
	svc := NewDatasourceService(nil)

	_, err := svc.SyncAPITable(map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty request")
	}

	_, err = svc.SyncAPITable(map[string]string{"datasourceId": "1"})
	if err == nil {
		t.Fatal("expected seatunnel client error")
	}

	_, err = svc.SyncAPIDs(map[string]string{"datasourceId": "1"})
	if err == nil {
		t.Fatal("expected seatunnel client error")
	}
}

func TestExcelService_ParseExcelFile_XLSX(t *testing.T) {
	svc := NewExcelService()

	f := excelize.NewFile()
	defer f.Close()
	_ = f.SetCellValue("Sheet1", "A1", "name")
	_ = f.SetCellValue("Sheet1", "B1", "amount")
	_ = f.SetCellValue("Sheet1", "A2", "Alice")
	_ = f.SetCellValue("Sheet1", "B2", 100)

	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("write xlsx buffer failed: %v", err)
	}

	sheets, err := svc.parseExcelFile("demo.xlsx", bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("parseExcelFile xlsx failed: %v", err)
	}
	if len(sheets) == 0 {
		t.Fatal("expected parsed sheets from xlsx")
	}
}
