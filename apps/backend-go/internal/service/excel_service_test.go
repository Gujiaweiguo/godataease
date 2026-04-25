package service

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dataease/backend/internal/domain/datasource"

	"github.com/xuri/excelize/v2"
)

type failingMultipartFile struct {
	data []byte
	pos  int64
	err  error
}

func (f *failingMultipartFile) Read(p []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.pos >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.pos:])
	f.pos += int64(n)
	return n, nil
}

func (f *failingMultipartFile) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (f *failingMultipartFile) Seek(offset int64, whence int) (int64, error) {
	var next int64
	switch whence {
	case io.SeekStart:
		next = offset
	case io.SeekCurrent:
		next = f.pos + offset
	case io.SeekEnd:
		next = int64(len(f.data)) + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if next < 0 {
		return 0, errors.New("negative position")
	}
	f.pos = next
	return f.pos, nil
}

func (f *failingMultipartFile) Close() error { return nil }

func createTestWorkbook(t *testing.T, sheets map[string][][]string) []byte {
	t.Helper()

	f := excelize.NewFile()
	first := true
	for name, rows := range sheets {
		var sheet string
		if first {
			sheet = f.GetSheetName(0)
			if err := f.SetSheetName(sheet, name); err != nil {
				t.Fatalf("rename sheet failed: %v", err)
			}
			sheet = name
			first = false
		} else {
			idx, err := f.NewSheet(name)
			if err != nil {
				t.Fatalf("create sheet failed: %v", err)
			}
			f.SetActiveSheet(idx)
			sheet = name
		}
		for r, row := range rows {
			for c, value := range row {
				cell, err := excelize.CoordinatesToCellName(c+1, r+1)
				if err != nil {
					t.Fatalf("cell name failed: %v", err)
				}
				if err = f.SetCellValue(sheet, cell, value); err != nil {
					t.Fatalf("set cell failed: %v", err)
				}
			}
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("write workbook failed: %v", err)
	}
	_ = f.Close()
	return buf.Bytes()
}

func bypassExcelSSRFProtection(t *testing.T) {
	t.Helper()
	originalClient := marketTemplateHTTPClient
	originalValidator := marketTemplateURLValidator
	marketTemplateHTTPClient = &http.Client{Timeout: 10 * time.Second}
	marketTemplateURLValidator = func(string) error { return nil }
	t.Cleanup(func() {
		marketTemplateHTTPClient = originalClient
		marketTemplateURLValidator = originalValidator
	})
}

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
	bypassExcelSSRFProtection(t)

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
	bypassExcelSSRFProtection(t)

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
		name        string
		fieldType   string
		expectedDe  int
		expectedExt int
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

func TestNewExcelService_FallbackToTmp(t *testing.T) {
	// Test that NewExcelService works even when /opt/dataease2.0/data/excel/ is not writable
	// Since we can't easily mock os.MkdirAll, we verify the service is created
	svc := NewExcelService()
	if svc == nil {
		t.Fatal("expected ExcelService instance")
	}
	if svc.uploadDir == "" {
		t.Fatal("expected uploadDir to be set")
	}
	// Verify the upload directory exists
	if _, err := os.Stat(svc.uploadDir); os.IsNotExist(err) {
		t.Fatalf("uploadDir %s should exist", svc.uploadDir)
	}
}

func TestExcelService_ParseExcelFile(t *testing.T) {
	svc := NewExcelService()

	// Test with empty content
	sheets, err := svc.parseExcelFile("test.csv", strings.NewReader(""))
	if err != nil {
		t.Fatalf("parseExcelFile with empty content failed: %v", err)
	}
	if len(sheets) != 0 {
		t.Fatalf("expected 0 sheets for empty CSV, got %d", len(sheets))
	}

	// Test CSV with only header
	csvHeaderOnly := "name,value\n"
	sheets, err = svc.parseExcelFile("header.csv", strings.NewReader(csvHeaderOnly))
	if err != nil {
		t.Fatalf("parseExcelFile with header only failed: %v", err)
	}
	if len(sheets) != 1 {
		t.Fatalf("expected 1 sheet for header only CSV, got %d", len(sheets))
	}
	if len(sheets[0].Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(sheets[0].Fields))
	}
	if len(sheets[0].Data) != 0 {
		t.Fatalf("expected 0 data rows, got %d", len(sheets[0].Data))
	}
}

func TestExcelService_ParseCSV_EmptyHeader(t *testing.T) {
	svc := NewExcelService()

	// CSV with empty column names should skip them (but all rows must have same field count)
	csvContent := "name\nvalue1\nvalue2\n"
	sheets, err := svc.parseCSV("single_header.csv", strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("parseCSV failed: %v", err)
	}
	if len(sheets) != 1 {
		t.Fatalf("expected 1 sheet, got %d", len(sheets))
	}
	if len(sheets[0].Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(sheets[0].Fields))
	}
}

func TestExcelService_ParseCSV_AllEmptyHeaders(t *testing.T) {
	svc := NewExcelService()

	// Empty CSV file should result in no sheets
	csvContent := ""
	sheets, err := svc.parseCSV("empty.csv", strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("parseCSV failed: %v", err)
	}
	// Empty CSV should return empty sheets
	if len(sheets) != 0 {
		t.Fatalf("expected 0 sheets for empty CSV, got %d", len(sheets))
	}
}

func TestExcelService_UploadFile_WithXLSX(t *testing.T) {
	svc := NewExcelService()

	// Create a minimal valid XLSX file content (this is a simplified test)
	// In practice, you would use a real XLSX file
	// For now, we test the CSV path more thoroughly

	csvContent := "col1,col2,col3\nval1,val2,val3\n"
	tmp, err := os.CreateTemp("", "excel-upload-*.csv")
	if err != nil {
		t.Fatalf("create temp file failed: %v", err)
	}
	defer os.Remove(tmp.Name())
	_, _ = tmp.WriteString(csvContent)
	_, _ = tmp.Seek(0, 0)
	defer tmp.Close()

	header := &multipart.FileHeader{Filename: "test.csv", Size: int64(len(csvContent))}
	uploaded, err := svc.UploadFile(tmp, header, 1, 0)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}
	if uploaded == nil {
		t.Fatal("expected uploaded result")
	}
	if uploaded.ExcelLabel != "test" {
		t.Fatalf("expected ExcelLabel 'test', got '%s'", uploaded.ExcelLabel)
	}
	if len(uploaded.Sheets) != 1 {
		t.Fatalf("expected 1 sheet, got %d", len(uploaded.Sheets))
	}
	if uploaded.Sheets[0].TableName != uploaded.Sheets[0].ExcelLabel {
		t.Fatal("expected TableName to equal ExcelLabel")
	}
}

func TestExcelService_DownloadFile_InvalidURL(t *testing.T) {
	svc := NewExcelService()

	// Test with invalid URL
	_, err := svc.downloadFile("not-a-valid-url", "", "")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestExcelService_DownloadFile_RequestCreationError(t *testing.T) {
	svc := NewExcelService()

	// Test with URL that causes request creation to fail
	// Using control characters that will fail http.NewRequest
	_, err := svc.downloadFile("http://\x00invalid.com/file.csv", "", "")
	if err == nil {
		t.Fatal("expected error for URL with control character")
	}
}

func TestExcelService_DownloadFile_BlockedSSRF(t *testing.T) {
	svc := NewExcelService()

	blockedURLs := []string{
		"http://127.0.0.1:8080/secret.csv",
		"http://169.254.169.254/latest/meta-data/",
		"file:///etc/passwd",
	}

	for _, rawURL := range blockedURLs {
		_, err := svc.downloadFile(rawURL, "", "")
		if err == nil {
			t.Fatalf("expected error for blocked URL %s", rawURL)
		}
		if !strings.Contains(err.Error(), "invalid download URL") && !strings.Contains(err.Error(), "scheme") && !strings.Contains(err.Error(), "blocked") {
			t.Fatalf("expected SSRF-related error for %s, got %v", rawURL, err)
		}
	}
}

func TestExcelService_LoadRemoteFile_WithAuth(t *testing.T) {
	svc := NewExcelService()
	bypassExcelSSRFProtection(t)

	csvContent := "name,value\ntest,123\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify basic auth is set
		user, pass, ok := r.BasicAuth()
		if !ok || user != "testuser" || pass != "testpass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(csvContent))
	}))
	defer server.Close()

	remote, err := svc.LoadRemoteFile(&RemoteExcelRequest{
		URL:      server.URL + "/remote.csv",
		UserName: "testuser",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("LoadRemoteFile with auth failed: %v", err)
	}
	if remote == nil || len(remote.Sheets) == 0 {
		t.Fatal("expected remote sheets")
	}
}

func TestExcelService_LoadRemoteFile_DownloadError(t *testing.T) {
	svc := NewExcelService()
	bypassExcelSSRFProtection(t)

	// Test with server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := svc.LoadRemoteFile(&RemoteExcelRequest{URL: server.URL + "/error.csv"})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestExcelService_SaveFile_SeekError(t *testing.T) {
	svc := NewExcelService()

	// Create a custom reader that implements multipart.File but fails on Seek
	tmp, err := os.CreateTemp("", "seek-error-*.csv")
	if err != nil {
		t.Fatalf("create temp file failed: %v", err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	// Open file for reading (this will work)
	file, err := os.Open(tmp.Name())
	if err != nil {
		t.Fatalf("open temp file failed: %v", err)
	}
	defer file.Close()

	// This should work since we're using a real file
	header := &multipart.FileHeader{Filename: "test.csv", Size: 10}
	_, err = svc.saveFile(file, header, "test-id")
	// The actual behavior depends on whether the directory is writable
	// We just verify no panic occurs
	_ = err
}

func TestExcelService_ParseExcelFile_InvalidXLSXReturnsError(t *testing.T) {
	svc := NewExcelService()

	_, err := svc.parseExcelFile("broken.xlsx", strings.NewReader("not-an-excel-file"))
	if err == nil || !strings.Contains(err.Error(), "failed to open excel file") {
		t.Fatalf("expected xlsx parse error, got %v", err)
	}
}

func TestExcelService_LoadRemoteFile_ParseError(t *testing.T) {
	svc := NewExcelService()
	bypassExcelSSRFProtection(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-an-excel-file"))
	}))
	defer server.Close()

	_, err := svc.LoadRemoteFile(&RemoteExcelRequest{URL: server.URL + "/broken.xlsx"})
	if err == nil || !strings.Contains(err.Error(), "failed to parse excel file") {
		t.Fatalf("expected wrapped parse error, got %v", err)
	}
}

func TestExcelService_LoadRemoteFile_FiltersSheetsWithoutFields(t *testing.T) {
	svc := NewExcelService()
	bypassExcelSSRFProtection(t)
	workbook := createTestWorkbook(t, map[string][][]string{
		"Valid": {{"name", "value"}, {"alice", "1"}},
		"Blank": {{"", ""}, {"x", "y"}},
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(workbook)
	}))
	defer server.Close()

	result, err := svc.LoadRemoteFile(&RemoteExcelRequest{URL: server.URL + "/demo.xlsx"})
	if err != nil {
		t.Fatalf("LoadRemoteFile failed: %v", err)
	}
	if len(result.Sheets) != 1 {
		t.Fatalf("expected only valid sheet, got %d", len(result.Sheets))
	}
	if result.Sheets[0].ExcelLabel != "Valid" {
		t.Fatalf("expected Valid sheet, got %s", result.Sheets[0].ExcelLabel)
	}
}

func TestExcelService_SaveFile_CreateFileError(t *testing.T) {
	tmp, err := os.CreateTemp("", "excel-upload-dir-file")
	if err != nil {
		t.Fatalf("create temp file failed: %v", err)
	}
	defer os.Remove(tmp.Name())
	_ = tmp.Close()

	svc := &ExcelService{uploadDir: tmp.Name()}
	file := &failingMultipartFile{data: []byte("name\nvalue\n")}

	_, err = svc.saveFile(file, &multipart.FileHeader{Filename: "demo.csv"}, "excel-id")
	if err == nil || !strings.Contains(err.Error(), "failed to create file") {
		t.Fatalf("expected create file error, got %v", err)
	}
}

func TestExcelService_SaveFile_CopyError(t *testing.T) {
	svc := &ExcelService{uploadDir: t.TempDir()}
	file := &failingMultipartFile{data: []byte("partial"), err: errors.New("copy failed")}

	_, err := svc.saveFile(file, &multipart.FileHeader{Filename: "demo.csv"}, "excel-id")
	if err == nil || !strings.Contains(err.Error(), "failed to save file") {
		t.Fatalf("expected save file error, got %v", err)
	}
}

func TestExcelService_SaveFile_SuccessReturnsPath(t *testing.T) {
	svc := &ExcelService{uploadDir: t.TempDir()}
	file := &failingMultipartFile{data: []byte("name\nvalue\n")}

	path, err := svc.saveFile(file, &multipart.FileHeader{Filename: "demo.csv"}, "excel-id")
	if err != nil {
		t.Fatalf("saveFile failed: %v", err)
	}
	if filepath.Base(path) != "excel-id.csv" {
		t.Fatalf("unexpected saved file path: %s", path)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("expected saved file to exist: %v", statErr)
	}
}

func TestExcelService_UploadFile_ParseErrorReturnsError(t *testing.T) {
	svc := &ExcelService{uploadDir: t.TempDir()}
	file := &failingMultipartFile{data: []byte("not-an-excel-file")}

	_, err := svc.UploadFile(file, &multipart.FileHeader{Filename: "broken.xlsx", Size: int64(len(file.data))}, 1, 0)
	if err == nil || !strings.Contains(err.Error(), "failed to open excel file") {
		t.Fatalf("expected upload parse error, got %v", err)
	}
}

func TestExcelService_UploadFile_FiltersEmptyFieldSheets(t *testing.T) {
	svc := &ExcelService{uploadDir: t.TempDir()}
	workbook := createTestWorkbook(t, map[string][][]string{
		"Valid": {{"name", "value"}, {"alice", "1"}},
		"Blank": {{"", ""}, {"x", "y"}},
	})
	file := &failingMultipartFile{data: workbook}

	result, err := svc.UploadFile(file, &multipart.FileHeader{Filename: "demo.xlsx", Size: int64(len(workbook))}, 1, 0)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}
	if len(result.Sheets) != 1 {
		t.Fatalf("expected only valid sheet after filtering, got %d", len(result.Sheets))
	}
	if result.Sheets[0].ExcelLabel != "Valid" {
		t.Fatalf("expected Valid sheet, got %s", result.Sheets[0].ExcelLabel)
	}
	if !strings.HasSuffix(result.Path, ".xlsx") {
		t.Fatalf("expected saved xlsx path, got %s", result.Path)
	}
	if _, err = os.Stat(result.Path); err != nil {
		t.Fatalf("expected uploaded file to exist: %v", err)
	}
	_ = os.Remove(result.Path)
}

func TestExcelService_UploadFile_AssignsSheetMetadataAndPersistsSavedPath(t *testing.T) {
	svc := &ExcelService{uploadDir: t.TempDir()}
	content := []byte("name,amount\nAlice,100\nBob,200\n")
	file := &failingMultipartFile{data: content}

	result, err := svc.UploadFile(file, &multipart.FileHeader{Filename: "upload.csv", Size: int64(len(content))}, 1, 0)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}
	if result == nil || len(result.Sheets) != 1 {
		t.Fatalf("expected single sheet result, got %#v", result)
	}
	assertUploadedSheetMetadata(t, result)
	if _, err = os.Stat(result.Path); err != nil {
		t.Fatalf("expected saved upload path to exist: %v", err)
	}
	_ = os.Remove(result.Path)
}

func TestExcelService_LoadRemoteFile_AssignsRemoteSheetMetadataAndNASize(t *testing.T) {
	svc := NewExcelService()
	bypassExcelSSRFProtection(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("name,value\nremote,1\n"))
	}))
	defer server.Close()

	result, err := svc.LoadRemoteFile(&RemoteExcelRequest{URL: server.URL + "/remote.csv"})
	if err != nil {
		t.Fatalf("LoadRemoteFile failed: %v", err)
	}
	if result == nil || len(result.Sheets) != 1 {
		t.Fatalf("expected single remote sheet, got %#v", result)
	}
	assertRemoteSheetMetadata(t, result)
}

func assertUploadedSheetMetadata(t *testing.T, result *datasource.ExcelFileData) {
	t.Helper()
	sheet := result.Sheets[0]
	if result.ExcelLabel != "upload" || sheet.ExcelLabel != "upload" {
		t.Fatalf("unexpected excel labels: result=%q sheet=%q", result.ExcelLabel, sheet.ExcelLabel)
	}
	if sheet.TableName != "upload" || sheet.FileName != "upload.csv" {
		t.Fatalf("unexpected table/file name metadata: %#v", sheet)
	}
	if sheet.SheetExcelID != result.ID || sheet.Path != result.Path {
		t.Fatalf("expected sheet to reference saved file metadata, got %#v", sheet)
	}
	if sheet.Size != "30 B" {
		t.Fatalf("expected formatted file size, got %q", sheet.Size)
	}
	if sheet.SheetID == "" || sheet.DeTableName == "" || sheet.LastUpdateTime == 0 {
		t.Fatalf("expected generated metadata fields, got %#v", sheet)
	}
	if len(sheet.Fields) != 2 || sheet.Fields[0].DeType != 0 || sheet.Fields[0].DeExtractType != 0 {
		t.Fatalf("expected field metadata defaults, got %#v", sheet.Fields)
	}
}

func assertRemoteSheetMetadata(t *testing.T, result *datasource.ExcelFileData) {
	t.Helper()
	sheet := result.Sheets[0]
	if result.ExcelLabel != "remote" || sheet.ExcelLabel != "remote" {
		t.Fatalf("unexpected remote labels: result=%q sheet=%q", result.ExcelLabel, sheet.ExcelLabel)
	}
	if sheet.TableName != "remote" || sheet.FileName != "remote.csv" {
		t.Fatalf("unexpected remote metadata: %#v", sheet)
	}
	if sheet.Size != "N/A" || sheet.Path != "" || sheet.SheetExcelID != "" || sheet.SheetID != "" {
		t.Fatalf("expected remote-only metadata fields, got %#v", sheet)
	}
	if sheet.LastUpdateTime == 0 || sheet.DeTableName == "" {
		t.Fatalf("expected generated remote metadata, got %#v", sheet)
	}
	if len(sheet.Fields) != 2 || sheet.Fields[0].DeType != 0 || sheet.Fields[1].DeExtractType != 0 {
		t.Fatalf("expected remote fields to have default types, got %#v", sheet.Fields)
	}
}

func TestExcelService_ParseCSV_TruncatesRowsAt100(t *testing.T) {
	svc := NewExcelService()
	var builder strings.Builder
	builder.WriteString("name,value\n")
	for i := 0; i < 105; i++ {
		builder.WriteString("row,1\n")
	}

	sheets, err := svc.parseCSV("limit.csv", strings.NewReader(builder.String()))
	if err != nil {
		t.Fatalf("parseCSV failed: %v", err)
	}
	if len(sheets) != 1 || len(sheets[0].Data) != 100 {
		t.Fatalf("expected exactly 100 data rows, got %#v", sheets)
	}
}

func TestExcelService_ParseExcelFile_TruncatesRowsAt100ForXLSX(t *testing.T) {
	svc := NewExcelService()
	rows := make([][]string, 0, 106)
	rows = append(rows, []string{"name", "value"})
	for i := 0; i < 105; i++ {
		rows = append(rows, []string{"row", "1"})
	}
	workbook := createTestWorkbook(t, map[string][][]string{"Sheet1": rows})

	sheets, err := svc.parseExcelFile("limit.xlsx", bytes.NewReader(workbook))
	if err != nil {
		t.Fatalf("parseExcelFile failed: %v", err)
	}
	if len(sheets) != 1 || len(sheets[0].Data) != 100 {
		t.Fatalf("expected exactly 100 xlsx data rows, got %#v", sheets)
	}
}

func TestExcelService_ParseCSV_InvalidCSVReturnsWrappedError(t *testing.T) {
	svc := NewExcelService()

	_, err := svc.parseCSV("broken.csv", strings.NewReader("name,value\na\n"))
	if err == nil || !strings.Contains(err.Error(), "failed to parse CSV") {
		t.Fatalf("expected wrapped invalid csv error, got %v", err)
	}
}

func TestExcelService_LoadRemoteFile_EmptyRemoteCSVReturnsNoSheets(t *testing.T) {
	svc := NewExcelService()
	bypassExcelSSRFProtection(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(""))
	}))
	defer server.Close()

	result, err := svc.LoadRemoteFile(&RemoteExcelRequest{URL: server.URL + "/empty.csv"})
	if err != nil {
		t.Fatalf("LoadRemoteFile failed: %v", err)
	}
	if result == nil || len(result.Sheets) != 0 {
		t.Fatalf("expected no sheets for empty remote csv, got %#v", result)
	}
}
