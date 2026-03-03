package service

import (
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"dataease/backend/internal/domain/datasource"
)

func TestExcelService_ParseCSVAndFieldType(t *testing.T) {
	svc := NewExcelService()

	csvContent := "name,amount,created_at\nAlice,100,2026-03-03\nBob,200,2026-03-04\n"
	sheets, err := svc.parseCSV("demo.csv", strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("parseCSV failed: %v", err)
	}
	if len(sheets) != 1 {
		t.Fatalf("expected 1 sheet, got %d", len(sheets))
	}
	if sheets[0].ExcelLabel != "demo" {
		t.Fatalf("expected sheet label demo, got %s", sheets[0].ExcelLabel)
	}
	if len(sheets[0].Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(sheets[0].Fields))
	}
	if len(sheets[0].Data) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(sheets[0].Data))
	}

	field := &datasource.ExcelTableField{FieldType: FieldTypeDateTime}
	svc.setFieldType(field)
	if field.DeType != 1 || field.DeExtractType != 1 {
		t.Fatalf("expected datetime detype 1/1, got %d/%d", field.DeType, field.DeExtractType)
	}

	field.FieldType = FieldTypeLong
	svc.setFieldType(field)
	if field.DeType != 2 || field.DeExtractType != 2 {
		t.Fatalf("expected long detype 2/2, got %d/%d", field.DeType, field.DeExtractType)
	}
}

func TestExcelService_FormatFileSize(t *testing.T) {
	svc := NewExcelService()

	if got := svc.formatFileSize(512); got != "512 B" {
		t.Fatalf("unexpected bytes format: %s", got)
	}
	if got := svc.formatFileSize(2048); got != "2 KB" {
		t.Fatalf("unexpected kb format: %s", got)
	}
	if got := svc.formatFileSize(3 * 1024 * 1024); got != "3 MB" {
		t.Fatalf("unexpected mb format: %s", got)
	}
}

func TestExcelService_DownloadFile(t *testing.T) {
	svc := NewExcelService()

	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer okServer.Close()

	resp, err := svc.downloadFile(okServer.URL, "", "")
	if err != nil {
		t.Fatalf("downloadFile success path failed: %v", err)
	}
	_ = resp.Body.Close()

	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer errServer.Close()

	_, err = svc.downloadFile(errServer.URL, "user", "pwd")
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestExcelService_UploadFileAndLoadRemoteFile(t *testing.T) {
	svc := NewExcelService()

	csvContent := "name,amount\nAlice,100\nBob,200\n"
	tmp, err := os.CreateTemp("", "excel-upload-*.csv")
	if err != nil {
		t.Fatalf("create temp file failed: %v", err)
	}
	defer os.Remove(tmp.Name())
	_, _ = tmp.WriteString(csvContent)
	_, _ = tmp.Seek(0, 0)
	defer tmp.Close()

	header := &multipart.FileHeader{Filename: "upload.csv", Size: int64(len(csvContent))}
	uploaded, err := svc.UploadFile(tmp, header, 1, 0)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}
	if uploaded == nil || len(uploaded.Sheets) == 0 {
		t.Fatal("expected uploaded sheets")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(csvContent))
	}))
	defer server.Close()

	remote, err := svc.LoadRemoteFile(&RemoteExcelRequest{URL: server.URL + "/remote.csv"})
	if err != nil {
		t.Fatalf("LoadRemoteFile failed: %v", err)
	}
	if remote == nil || len(remote.Sheets) == 0 {
		t.Fatal("expected remote sheets")
	}
}

func TestExcelService_SetFieldType_AllTypes(t *testing.T) {
	svc := NewExcelService()

	tests := []struct {
		name         string
		fieldType    string
		expectedDe   int
		expectedExt  int
	}{
		{"text type", FieldTypeText, 0, 0},
		{"datetime type", FieldTypeDateTime, 1, 1},
		{"long type", FieldTypeLong, 2, 2},
		{"double type", FieldTypeDouble, 3, 3},
		{"unknown type", "unknown", 0, 0},
		{"empty type", "", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &datasource.ExcelTableField{FieldType: tt.fieldType}
			svc.setFieldType(field)
			if field.DeType != tt.expectedDe {
				t.Errorf("expected DeType %d, got %d", tt.expectedDe, field.DeType)
			}
			if field.DeExtractType != tt.expectedExt {
				t.Errorf("expected DeExtractType %d, got %d", tt.expectedExt, field.DeExtractType)
			}
		})
	}
}
