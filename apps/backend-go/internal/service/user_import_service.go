package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"dataease/backend/internal/domain/user"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

const (
	MaxUserImportFileSize = 10 * 1024 * 1024
)

var emailRegexp = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type UserImportService struct {
	userService *UserService
	reportDir   string
}

type UserImportResult struct {
	TotalRows   int    `json:"totalRows"`
	SuccessRows int    `json:"successRows"`
	FailedRows  int    `json:"failedRows"`
	ErrorKey    string `json:"errorKey,omitempty"`
	Key         string `json:"key,omitempty"`
}

type userImportRecord struct {
	Line     int
	Username string
	RealName string
	Email    string
	Phone    string
}

type userImportError struct {
	Line     int
	Username string
	Reason   string
}

func NewUserImportService(userService *UserService) *UserImportService {
	reportDir := filepath.Join(os.TempDir(), "dataease", "user-import-errors")
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		reportDir = os.TempDir()
		_ = os.MkdirAll(reportDir, 0o755)
	}

	return &UserImportService{
		userService: userService,
		reportDir:   reportDir,
	}
}

func (s *UserImportService) GenerateTemplate() ([]byte, string, error) {
	f := excelize.NewFile()
	sheetName := f.GetSheetName(0)
	headers := []string{"username", "realName", "email", "phone"}

	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return nil, "", fmt.Errorf("failed to build template cell: %w", err)
		}
		if err = f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, "", fmt.Errorf("failed to set template header: %w", err)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate template: %w", err)
	}

	return buf.Bytes(), "user_import_template.xlsx", nil
}

func (s *UserImportService) ImportUsers(file multipart.File, header *multipart.FileHeader, _ string, orgID int64) (*UserImportResult, error) {
	if s.userService == nil {
		return nil, fmt.Errorf("user import service is not configured")
	}
	if file == nil || header == nil {
		return nil, fmt.Errorf("file is required")
	}
	if header.Size > MaxUserImportFileSize {
		return nil, fmt.Errorf("file size exceeds 10MB limit")
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read import file: %w", err)
	}

	records, err := s.parseRecords(bytes.NewReader(content), header.Filename)
	if err != nil {
		return nil, err
	}

	result := &UserImportResult{}
	if len(records) == 0 {
		return result, nil
	}

	defaultPassword := s.userService.ResolveDefaultPassword()
	importErrors := make([]userImportError, 0)

	for _, record := range records {
		if isBlankRow(record) {
			continue
		}

		result.TotalRows++
		if importErr := s.importRecord(record, defaultPassword, orgID); importErr != nil {
			importErrors = append(importErrors, userImportError{Line: record.Line, Username: record.Username, Reason: importErr.Error()})
			continue
		}

		result.SuccessRows++
	}

	result.FailedRows = len(importErrors)
	if result.FailedRows > 0 {
		key, reportErr := s.saveErrorReport(importErrors)
		if reportErr != nil {
			return nil, reportErr
		}
		result.ErrorKey = key
		result.Key = key
	}

	return result, nil
}

func (s *UserImportService) importRecord(record userImportRecord, defaultPassword string, orgID int64) error {
	if strings.TrimSpace(record.Username) == "" {
		return fmt.Errorf("username is required")
	}
	if record.Email != "" && !emailRegexp.MatchString(record.Email) {
		return fmt.Errorf("invalid email format")
	}

	req := &user.UserCreateRequest{
		Username: record.Username,
		Password: defaultPassword,
		RealName: record.RealName,
		Email:    stringPtr(record.Email),
		Phone:    stringPtr(record.Phone),
	}
	if orgID > 0 {
		req.OrgID = &orgID
	}
	_, err := s.userService.CreateUser(req)
	return err
}

func (s *UserImportService) GetErrorReport(key string) ([]byte, string, error) {
	path := s.errorReportPath(key)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read error report: %w", err)
	}
	return content, filepath.Base(path), nil
}

func (s *UserImportService) ClearErrorReport(key string) error {
	path := s.errorReportPath(key)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to clear error report: %w", err)
	}
	return nil
}

func (s *UserImportService) parseRecords(reader io.Reader, filename string) ([]userImportRecord, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	var rows [][]string

	switch ext {
	case ".csv":
		csvReader := csv.NewReader(reader)
		parsedRows, err := csvReader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to parse csv: %w", err)
		}
		rows = parsedRows
	case ".xlsx", ".xlsm":
		f, err := excelize.OpenReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse excel file: %w", err)
		}
		defer f.Close()

		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("excel has no sheets")
		}

		parsedRows, err := f.GetRows(sheets[0])
		if err != nil {
			return nil, fmt.Errorf("failed to read excel rows: %w", err)
		}
		rows = parsedRows
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	if len(rows) == 0 {
		return []userImportRecord{}, nil
	}

	headerIndex := buildHeaderIndex(rows[0])
	usernameIdx, ok := headerIndex["username"]
	if !ok {
		return nil, fmt.Errorf("missing required column: username")
	}

	realNameIdx := pickIndex(headerIndex, []string{"realname", "nickname", "nick_name"})
	emailIdx := pickIndex(headerIndex, []string{"email"})
	phoneIdx := pickIndex(headerIndex, []string{"phone"})

	records := make([]userImportRecord, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		records = append(records, userImportRecord{
			Line:     i + 1,
			Username: readCell(row, usernameIdx),
			RealName: readCell(row, realNameIdx),
			Email:    readCell(row, emailIdx),
			Phone:    readCell(row, phoneIdx),
		})
	}

	return records, nil
}

func (s *UserImportService) saveErrorReport(errors []userImportError) (string, error) {
	if len(errors) == 0 {
		return "", nil
	}

	key := strings.ReplaceAll(uuid.NewString(), "-", "")
	fileName := "user_import_error_" + key + ".xlsx"
	path := filepath.Join(s.reportDir, fileName)

	f := excelize.NewFile()
	sheetName := f.GetSheetName(0)

	headers := []string{"line", "username", "reason"}
	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return "", fmt.Errorf("failed to build error report header cell: %w", err)
		}
		if err = f.SetCellValue(sheetName, cell, header); err != nil {
			return "", fmt.Errorf("failed to write error report header: %w", err)
		}
	}

	for i, row := range errors {
		excelRow := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", excelRow), row.Line); err != nil {
			return "", fmt.Errorf("failed to write error line: %w", err)
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", excelRow), row.Username); err != nil {
			return "", fmt.Errorf("failed to write error username: %w", err)
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", excelRow), row.Reason); err != nil {
			return "", fmt.Errorf("failed to write error reason: %w", err)
		}
	}

	if err := f.SaveAs(path); err != nil {
		return "", fmt.Errorf("failed to save error report: %w", err)
	}

	return key, nil
}

func (s *UserImportService) errorReportPath(key string) string {
	clean := strings.TrimSpace(key)
	clean = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, clean)

	return filepath.Join(s.reportDir, "user_import_error_"+clean+".xlsx")
}

func buildHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, name := range header {
		normalized := strings.ToLower(strings.TrimSpace(name))
		normalized = strings.ReplaceAll(normalized, " ", "")
		if normalized == "" {
			continue
		}
		index[normalized] = i
	}
	return index
}

func pickIndex(headerIndex map[string]int, keys []string) int {
	for _, key := range keys {
		if idx, ok := headerIndex[key]; ok {
			return idx
		}
	}
	return -1
}

func readCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func isBlankRow(row userImportRecord) bool {
	return strings.TrimSpace(row.Username) == "" &&
		strings.TrimSpace(row.RealName) == "" &&
		strings.TrimSpace(row.Email) == "" &&
		strings.TrimSpace(row.Phone) == ""
}

func stringPtr(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}
