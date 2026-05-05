//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/license"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicenseRepository_Load(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_sys_setting")

	repo := NewLicenseRepository(testDB)
	result, raw, err := repo.Load()
	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "", raw)

	seed := []coreSysSetting{
		{Pkey: "license.status", Pval: "valid", Type: "text", Sort: 1},
		{Pkey: "license.message", Pval: "ok", Type: "text", Sort: 2},
		{Pkey: "license.corporation", Pval: "DataEase", Type: "text", Sort: 3},
		{Pkey: "license.expired", Pval: "2026-12-31", Type: "text", Sort: 4},
		{Pkey: "license.count", Pval: "128", Type: "text", Sort: 5},
		{Pkey: "license.version", Pval: "v1", Type: "text", Sort: 6},
		{Pkey: "license.edition", Pval: "enterprise", Type: "text", Sort: 7},
		{Pkey: "license.serialNo", Pval: "SN-001", Type: "text", Sort: 8},
		{Pkey: "license.remark", Pval: "remark", Type: "text", Sort: 9},
		{Pkey: "license.isv", Pval: "FIT2CLOUD", Type: "text", Sort: 10},
		{Pkey: "license.raw", Pval: "raw-license-data", Type: "text", Sort: 11},
	}
	for i := range seed {
		require.NoError(t, testDB.Create(&seed[i]).Error)
	}

	result, raw, err = repo.Load()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "valid", result.Status)
	assert.Equal(t, "ok", result.Message)
	require.NotNil(t, result.License)
	assert.Equal(t, "DataEase", result.License.Corporation)
	assert.Equal(t, "2026-12-31", result.License.Expired)
	assert.Equal(t, int64(128), result.License.Count)
	assert.Equal(t, "v1", result.License.Version)
	assert.Equal(t, "enterprise", result.License.Edition)
	assert.Equal(t, "SN-001", result.License.SerialNo)
	assert.Equal(t, "remark", result.License.Remark)
	assert.Equal(t, "FIT2CLOUD", result.License.ISV)
	assert.Equal(t, "raw-license-data", raw)
}

func TestLicenseRepository_Save(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_sys_setting")

	repo := NewLicenseRepository(testDB)
	result := &license.ValidateResult{
		Status:  "valid",
		Message: "saved",
		License: &license.LicenseInfo{
			Corporation: "DataEase",
			Expired:     "2026-12-31",
			Count:       128,
			Version:     "v1",
			Edition:     "enterprise",
			SerialNo:    "SN-001",
			Remark:      "remark",
			ISV:         "FIT2CLOUD",
		},
	}
	require.NoError(t, repo.Save(result, "raw-license-data"))

	var rows []coreSysSetting
	require.NoError(t, testDB.Order("sort ASC").Find(&rows).Error)
	require.Equal(t, 11, len(rows))
	assert.Equal(t, "license.status", rows[0].Pkey)
	assert.Equal(t, "valid", rows[0].Pval)
	assert.Equal(t, "license.raw", rows[10].Pkey)
	assert.Equal(t, "raw-license-data", rows[10].Pval)

	result = &license.ValidateResult{Status: "invalid", Message: "no-license", License: nil}
	require.NoError(t, repo.Save(result, "raw-empty-license"))

	require.NoError(t, testDB.Order("sort ASC").Find(&rows).Error)
	require.Equal(t, 3, len(rows))
	assert.Equal(t, "license.status", rows[0].Pkey)
	assert.Equal(t, "invalid", rows[0].Pval)
	assert.Equal(t, "license.message", rows[1].Pkey)
	assert.Equal(t, "no-license", rows[1].Pval)
	assert.Equal(t, "license.raw", rows[2].Pkey)
	assert.Equal(t, "raw-empty-license", rows[2].Pval)
}

func TestLicenseRepository_Clear(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_sys_setting")

	seed := []coreSysSetting{
		{Pkey: "basic.companyName", Pval: "DataEase", Type: "text", Sort: 1},
		{Pkey: "license.status", Pval: "valid", Type: "text", Sort: 2},
		{Pkey: "license.raw", Pval: "raw-license-data", Type: "text", Sort: 3},
	}
	for i := range seed {
		require.NoError(t, testDB.Create(&seed[i]).Error)
	}

	repo := NewLicenseRepository(testDB)
	require.NoError(t, repo.Clear())

	var rows []coreSysSetting
	require.NoError(t, testDB.Order("sort ASC").Find(&rows).Error)
	require.Equal(t, 1, len(rows))
	assert.Equal(t, "basic.companyName", rows[0].Pkey)
	assert.Equal(t, "DataEase", rows[0].Pval)
}
