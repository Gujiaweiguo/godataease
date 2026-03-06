package service

import (
	"os"
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
