package service

import (
	"mime/multipart"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	defer os.Remove(tmpFile.Name())

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
	_, err := svc.ImportUsers(nil, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	svc = NewUserImportService(&UserService{})
	_, err = svc.ImportUsers(nil, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file is required")

	file, err := os.CreateTemp("", "user-import-too-large-*.csv")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	_, _ = file.WriteString("username\nuser1\n")
	_, _ = file.Seek(0, 0)
	defer file.Close()

	_, err = svc.ImportUsers(file, &multipart.FileHeader{Filename: "users.csv", Size: MaxUserImportFileSize + 1}, "")
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

func TestUserImportService_SaveErrorReportEmptyAndClearMissing(t *testing.T) {
	svc := NewUserImportService(&UserService{})

	key, err := svc.saveErrorReport(nil)
	require.NoError(t, err)
	assert.Equal(t, "", key)

	err = svc.ClearErrorReport("missing-key")
	assert.Error(t, err)
}

func TestUserImportService_ImportRecordValidation(t *testing.T) {
	svc := NewUserImportService(&UserService{})

	err := svc.importRecord(userImportRecord{}, "pwd")
	assert.EqualError(t, err, "username is required")

	err = svc.importRecord(userImportRecord{Username: "u", Email: "bad-email"}, "pwd")
	assert.EqualError(t, err, "invalid email format")
}
