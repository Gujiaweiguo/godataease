package service

import (
	"bytes"
	"mime/multipart"
	"os"
	"strings"
	"testing"

	"dataease/backend/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestUserImportService_GenerateTemplate(t *testing.T) {
	svc := NewUserImportService(&UserService{})

	content, filename, err := svc.GenerateTemplate()
	require.NoError(t, err)
	assert.Equal(t, "user_import_template.xlsx", filename)
	assert.NotEmpty(t, content)
}

func TestUserImportService_ErrorReportLifecycle(t *testing.T) {
	svc := NewUserImportService(&UserService{})

	key, err := svc.saveErrorReport([]userImportError{{Line: 2, Username: "u1", Reason: "duplicate username"}})
	require.NoError(t, err)
	require.NotEmpty(t, key)

	report, filename, err := svc.GetErrorReport(key)
	require.NoError(t, err)
	assert.NotEmpty(t, report)
	assert.Contains(t, filename, key)

	err = svc.ClearErrorReport(key)
	require.NoError(t, err)

	_, _, err = svc.GetErrorReport(key)
	assert.Error(t, err)
}

func TestUserImportService_ParseRecords_MissingUsernameColumn(t *testing.T) {
	svc := NewUserImportService(&UserService{})

	tmpFile, err := os.CreateTemp("", "user-import-*.csv")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString("email,phone\nuser@example.com,13800000000\n")
	require.NoError(t, err)
	_, err = tmpFile.Seek(0, 0)
	require.NoError(t, err)

	records, err := svc.parseRecords(tmpFile, "invalid.csv")
	assert.Nil(t, records)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required column: username")
}

func TestUserImportService_ImportUsers_EarlyReturns(t *testing.T) {
	svc := NewUserImportService(nil)
	_, err := svc.ImportUsers(nil, nil, "", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	svc = NewUserImportService(&UserService{})
	_, err = svc.ImportUsers(nil, nil, "", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file is required")

	file, err := os.CreateTemp("", "user-import-too-large-*.csv")
	require.NoError(t, err)
	defer func() { _ = os.Remove(file.Name()) }()
	_, _ = file.WriteString("username\nuser1\n")
	_, _ = file.Seek(0, 0)
	defer func() { _ = file.Close() }()

	_, err = svc.ImportUsers(file, &multipart.FileHeader{Filename: "users.csv", Size: MaxUserImportFileSize + 1}, "", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "10MB")
}

func TestUserImportService_ParseRecords_CSVSuccessAndHelpers(t *testing.T) {
	svc := NewUserImportService(&UserService{})
	records, err := svc.parseRecords(strings.NewReader("username,realName,email,phone\n alice , Alice ,alice@example.com, 13800138000 \n"), "users.csv")
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "alice", records[0].Username)
	assert.Equal(t, "Alice", records[0].RealName)
	assert.Equal(t, "alice@example.com", records[0].Email)
	assert.Equal(t, "13800138000", records[0].Phone)

	path := svc.errorReportPath(" a-b_c ")
	assert.Contains(t, path, "user_import_error_abc.xlsx")

	headers := buildHeaderIndex([]string{" User Name ", "Real Name", "email"})
	assert.Equal(t, 0, headers["username"])
	assert.Equal(t, 1, pickIndex(headers, []string{"realname", "nickname"}))
	assert.Equal(t, -1, pickIndex(headers, []string{"phone"}))
	assert.Equal(t, "v", readCell([]string{" v "}, 0))
	assert.Equal(t, "", readCell([]string{"v"}, 2))
	assert.True(t, isBlankRow(userImportRecord{}))
	assert.False(t, isBlankRow(userImportRecord{Username: "u"}))
	assert.Nil(t, stringPtr("  "))
	assert.Equal(t, "x", *stringPtr(" x "))
}

func TestUserImportService_ParseRecords_UnsupportedAndEmpty(t *testing.T) {
	svc := NewUserImportService(&UserService{})

	_, err := svc.parseRecords(strings.NewReader("abc"), "users.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file type")

	records, err := svc.parseRecords(strings.NewReader(""), "users.csv")
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestUserImportService_ParseRecords_XLSXAliasHeaders(t *testing.T) {
	svc := NewUserImportService(&UserService{})
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	require.NoError(t, f.SetSheetRow(sheet, "A1", &[]string{"username", "nick_name", "email", "phone"}))
	require.NoError(t, f.SetSheetRow(sheet, "A2", &[]string{"alice", "Alice", "alice@example.com", "13800138000"}))

	buf, err := f.WriteToBuffer()
	require.NoError(t, err)

	records, err := svc.parseRecords(bytes.NewReader(buf.Bytes()), "users.xlsx")
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "alice", records[0].Username)
	assert.Equal(t, "Alice", records[0].RealName)
	assert.Equal(t, "alice@example.com", records[0].Email)
	assert.Equal(t, "13800138000", records[0].Phone)
}

func TestUserImportService_ParseRecords_InvalidXLSX(t *testing.T) {
	svc := NewUserImportService(&UserService{})

	records, err := svc.parseRecords(strings.NewReader("not an xlsx"), "users.xlsx")
	assert.Nil(t, records)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse excel file")
}

func TestUserImportService_SaveErrorReportEmptyAndClearMissing(t *testing.T) {
	svc := NewUserImportService(&UserService{})

	key, err := svc.saveErrorReport(nil)
	require.NoError(t, err)
	assert.Equal(t, "", key)

	err = svc.ClearErrorReport("missing-key")
	assert.Error(t, err)
}

func TestUserImportService_SaveErrorReport_WritesWorkbookRows(t *testing.T) {
	svc := NewUserImportService(&UserService{})

	key, err := svc.saveErrorReport([]userImportError{{Line: 2, Username: "u1", Reason: "duplicate username"}, {Line: 3, Username: "u2", Reason: "invalid email"}})
	require.NoError(t, err)
	require.NotEmpty(t, key)
	defer func() { _ = svc.ClearErrorReport(key) }()

	content, _, err := svc.GetErrorReport(key)
	require.NoError(t, err)

	f, err := excelize.OpenReader(bytes.NewReader(content))
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows(f.GetSheetName(0))
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, []string{"line", "username", "reason"}, rows[0])
	assert.Equal(t, []string{"2", "u1", "duplicate username"}, rows[1])
	assert.Equal(t, []string{"3", "u2", "invalid email"}, rows[2])
}

func TestUserImportService_ImportRecordValidation(t *testing.T) {
	svc := NewUserImportService(&UserService{})

	err := svc.importRecord(userImportRecord{}, "pwd", 0)
	assert.EqualError(t, err, "username is required")

	err = svc.importRecord(userImportRecord{Username: "u", Email: "bad-email"}, "pwd", 0)
	assert.EqualError(t, err, "invalid email format")
}

func TestUserImportService_ImportUsers_ZeroOrBlankRows(t *testing.T) {
	t.Run("header only returns zero result", func(t *testing.T) {
		svc := NewUserImportService(&UserService{})
		file, err := os.CreateTemp("", "user-import-empty-*.csv")
		require.NoError(t, err)
		defer func() { _ = os.Remove(file.Name()) }()
		defer func() { _ = file.Close() }()
		_, err = file.WriteString("username,realName,email,phone\n")
		require.NoError(t, err)
		_, err = file.Seek(0, 0)
		require.NoError(t, err)

		result, err := svc.ImportUsers(file, &multipart.FileHeader{Filename: "users.csv", Size: 30}, "", 0)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Zero(t, result.TotalRows)
		assert.Zero(t, result.SuccessRows)
		assert.Zero(t, result.FailedRows)
	})

	t.Run("blank rows are skipped from totals", func(t *testing.T) {
		svc := NewUserImportService(&UserService{})
		file, err := os.CreateTemp("", "user-import-blank-*.csv")
		require.NoError(t, err)
		defer func() { _ = os.Remove(file.Name()) }()
		defer func() { _ = file.Close() }()
		_, err = file.WriteString("username,realName,email,phone\n , , , \n\t,\t,\t,\t\n")
		require.NoError(t, err)
		_, err = file.Seek(0, 0)
		require.NoError(t, err)

		result, err := svc.ImportUsers(file, &multipart.FileHeader{Filename: "users.csv", Size: 64}, "", 0)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Zero(t, result.TotalRows)
		assert.Zero(t, result.SuccessRows)
		assert.Zero(t, result.FailedRows)
	})
}

func TestUserImportService_ImportUsers_SuccessAndFailures(t *testing.T) {
	t.Run("successful import creates users", func(t *testing.T) {
		userSvc, db := setupUserServiceRepoTest(t)
		svc := NewUserImportService(userSvc)
		file, err := os.CreateTemp("", "user-import-success-*.csv")
		require.NoError(t, err)
		defer func() { _ = os.Remove(file.Name()) }()
		defer func() { _ = file.Close() }()
		_, err = file.WriteString("username,realName,email,phone\nalice,Alice,alice@example.com,13800138000\nbob,Bob,bob@example.com,13800138001\n")
		require.NoError(t, err)
		_, err = file.Seek(0, 0)
		require.NoError(t, err)

		result, err := svc.ImportUsers(file, &multipart.FileHeader{Filename: "users.csv", Size: 128}, "", 0)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 2, result.TotalRows)
		assert.Equal(t, 2, result.SuccessRows)
		assert.Zero(t, result.FailedRows)
		assert.Empty(t, result.ErrorKey)

		var count int64
		require.NoError(t, db.Model(&user.SysUser{}).Count(&count).Error)
		assert.Equal(t, int64(2), count)
	})

	t.Run("duplicate username writes error report", func(t *testing.T) {
		userSvc, _ := setupUserServiceRepoTest(t)
		svc := NewUserImportService(userSvc)
		file, err := os.CreateTemp("", "user-import-duplicate-*.csv")
		require.NoError(t, err)
		defer func() { _ = os.Remove(file.Name()) }()
		defer func() { _ = file.Close() }()
		_, err = file.WriteString("username,realName,email,phone\nalice,Alice,alice@example.com,13800138000\nalice,Alice2,alice2@example.com,13800138001\n")
		require.NoError(t, err)
		_, err = file.Seek(0, 0)
		require.NoError(t, err)

		result, err := svc.ImportUsers(file, &multipart.FileHeader{Filename: "users.csv", Size: 140}, "", 0)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 2, result.TotalRows)
		assert.Equal(t, 1, result.SuccessRows)
		assert.Equal(t, 1, result.FailedRows)
		require.NotEmpty(t, result.ErrorKey)
		assert.Equal(t, result.ErrorKey, result.Key)
		defer func() { _ = svc.ClearErrorReport(result.ErrorKey) }()

		report, _, err := svc.GetErrorReport(result.ErrorKey)
		require.NoError(t, err)
		workbook, err := excelize.OpenReader(bytes.NewReader(report))
		require.NoError(t, err)
		defer func() { _ = workbook.Close() }()
		rows, err := workbook.GetRows(workbook.GetSheetName(0))
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, []string{"3", "alice", "username already exists"}, rows[1])
	})

	t.Run("parse error from unsupported file type propagates", func(t *testing.T) {
		userSvc, _ := setupUserServiceRepoTest(t)
		svc := NewUserImportService(userSvc)
		file, err := os.CreateTemp("", "user-import-unsupported-*.txt")
		require.NoError(t, err)
		defer func() { _ = os.Remove(file.Name()) }()
		defer func() { _ = file.Close() }()
		_, err = file.WriteString("username\nalice\n")
		require.NoError(t, err)
		_, err = file.Seek(0, 0)
		require.NoError(t, err)

		result, err := svc.ImportUsers(file, &multipart.FileHeader{Filename: "users.txt", Size: 16}, "", 0)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unsupported file type")
	})
}
