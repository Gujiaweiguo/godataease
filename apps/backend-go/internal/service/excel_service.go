package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dataease/backend/internal/domain/datasource"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

const (
	FieldTypeText     = "TEXT"
	FieldTypeLong     = "LONG"
	FieldTypeDateTime = "DATETIME"
	FieldTypeDouble   = "DOUBLE"
)

type ExcelService struct {
	uploadDir string
}

func NewExcelService() *ExcelService {
	uploadDir := "/opt/dataease2.0/data/excel/"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		uploadDir = "/tmp/dataease/excel/"
		_ = os.MkdirAll(uploadDir, 0755)
	}
	return &ExcelService{uploadDir: uploadDir}
}

func (s *ExcelService) UploadFile(file multipart.File, header *multipart.FileHeader, datasourceID int64, editType int) (*datasource.ExcelFileData, error) {
	filename := header.Filename
	sheets, err := s.parseExcelFile(filename, file)
	if err != nil {
		return nil, err
	}

	excelID := uuid.New().String()
	filePath, err := s.saveFile(file, header, excelID)
	if err != nil {
		return nil, err
	}

	filteredSheets := make([]*datasource.ExcelSheetData, 0)
	for _, sheet := range sheets {
		if len(sheet.Fields) > 0 {
			filteredSheets = append(filteredSheets, sheet)
		}
	}

	now := time.Now().UnixMilli()
	fileSize := header.Size
	sizeStr := s.formatFileSize(fileSize)

	for _, sheet := range filteredSheets {
		sheet.LastUpdateTime = now
		sheet.TableName = sheet.ExcelLabel
		sheet.DeTableName = fmt.Sprintf("excel_%s_%s", sheet.ExcelLabel, uuid.New().String()[:10])
		sheet.Path = filePath
		sheet.SheetID = uuid.New().String()
		sheet.SheetExcelID = excelID
		sheet.FileName = filename
		sheet.Size = sizeStr

		for _, field := range sheet.Fields {
			s.setFieldType(field)
		}
	}

	return &datasource.ExcelFileData{
		ID:         excelID,
		ExcelLabel: strings.TrimSuffix(filename, filepath.Ext(filename)),
		Sheets:     filteredSheets,
		Path:       filePath,
		IsSheet:    false,
	}, nil
}

func (s *ExcelService) parseExcelFile(filename string, reader io.Reader) ([]*datasource.ExcelSheetData, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	if ext == ".csv" {
		return s.parseCSV(filename, reader)
	}

	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	result := make([]*datasource.ExcelSheetData, 0, len(sheetList))

	for _, sheetName := range sheetList {
		rows, err := f.GetRows(sheetName)
		if err != nil || len(rows) == 0 {
			continue
		}

		header := rows[0]
		fields := make([]*datasource.ExcelTableField, 0, len(header))
		for _, h := range header {
			h = strings.TrimSpace(h)
			if h == "" {
				continue
			}
			fields = append(fields, &datasource.ExcelTableField{
				OriginName: h,
				Name:       h,
				FieldType:  FieldTypeText,
			})
		}

		if len(fields) == 0 {
			continue
		}

		data := make([][]string, 0)
		for i := 1; i < len(rows) && i < 101; i++ {
			data = append(data, rows[i])
		}

		result = append(result, &datasource.ExcelSheetData{
			ExcelLabel: sheetName,
			Data:       data,
			Fields:     fields,
			IsSheet:    true,
		})
	}

	return result, nil
}

func (s *ExcelService) parseCSV(filename string, reader io.Reader) ([]*datasource.ExcelSheetData, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return []*datasource.ExcelSheetData{}, nil
	}

	header := records[0]
	fields := make([]*datasource.ExcelTableField, 0, len(header))
	for _, h := range header {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		fields = append(fields, &datasource.ExcelTableField{
			OriginName: h,
			Name:       h,
			FieldType:  FieldTypeText,
		})
	}

	data := make([][]string, 0)
	for i := 1; i < len(records) && i < 101; i++ {
		data = append(data, records[i])
	}

	sheetName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	return []*datasource.ExcelSheetData{
		{
			ExcelLabel: sheetName,
			Data:       data,
			Fields:     fields,
			IsSheet:    true,
		},
	}, nil
}

func (s *ExcelService) saveFile(file multipart.File, header *multipart.FileHeader, excelID string) (string, error) {
	ext := filepath.Ext(header.Filename)
	fileName := excelID + ext
	filePath := filepath.Join(s.uploadDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err = file.Seek(0, 0); err != nil {
		return "", err
	}

	if _, err = io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filePath, nil
}

func (s *ExcelService) formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%d KB", size/1024)
	}
	return fmt.Sprintf("%d MB", size/1024/1024)
}

func (s *ExcelService) setFieldType(field *datasource.ExcelTableField) {
	switch field.FieldType {
	case FieldTypeText:
		field.DeType = 0
		field.DeExtractType = 0
	case FieldTypeDateTime:
		field.DeType = 1
		field.DeExtractType = 1
	case FieldTypeLong:
		field.DeType = 2
		field.DeExtractType = 2
	case FieldTypeDouble:
		field.DeType = 3
		field.DeExtractType = 3
	default:
		field.DeType = 0
		field.DeExtractType = 0
	}
}

// RemoteExcelRequest represents remote file request
type RemoteExcelRequest struct {
	URL          string `json:"url"`
	UserName     string `json:"userName"`
	Password     string `json:"passwd"`
	DatasourceID int64  `json:"datasourceId"`
}

// LoadRemoteFile loads and parses a remote Excel file
func (s *ExcelService) LoadRemoteFile(req *RemoteExcelRequest) (*datasource.ExcelFileData, error) {
	// Download the file
	resp, err := s.downloadFile(req.URL, req.UserName, req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to download remote file: %w", err)
	}
	defer resp.Body.Close()

	// Parse the file
	filename := filepath.Base(req.URL)
	sheets, err := s.parseExcelFile(filename, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse excel file: %w", err)
	}

	// Filter sheets with fields
	filteredSheets := make([]*datasource.ExcelSheetData, 0)
	for _, sheet := range sheets {
		if len(sheet.Fields) > 0 {
			filteredSheets = append(filteredSheets, sheet)
		}
	}

	now := time.Now().UnixMilli()
	for _, sheet := range filteredSheets {
		sheet.LastUpdateTime = now
		sheet.TableName = sheet.ExcelLabel
		sheet.DeTableName = fmt.Sprintf("excel_%s_%s", sheet.ExcelLabel, uuid.New().String()[:10])
		sheet.FileName = filename
		sheet.Size = "N/A"

		for _, field := range sheet.Fields {
			s.setFieldType(field)
		}
	}

	return &datasource.ExcelFileData{
		ID:         uuid.New().String(),
		ExcelLabel: strings.TrimSuffix(filename, filepath.Ext(filename)),
		Sheets:     filteredSheets,
		IsSheet:    false,
	}, nil
}

// downloadFile downloads a file from remote URL
func (s *ExcelService) downloadFile(url, userName, password string) (*http.Response, error) {
	if err := marketTemplateURLValidator(url); err != nil {
		return nil, fmt.Errorf("invalid download URL: %w", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if userName != "" || password != "" {
		req.SetBasicAuth(userName, password)
	}

	client := &http.Client{
		Timeout:       30 * time.Second,
		Transport:     marketTemplateHTTPClient.Transport,
		CheckRedirect: marketTemplateHTTPClient.CheckRedirect,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	resp.Body = io.NopCloser(io.LimitReader(resp.Body, marketTemplateMaxResponseBytes+1))

	return resp, nil
}
