//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatermarkServiceIntegration_FindReturnsDefaultWhenNoRecordExists(t *testing.T) {
	cleanupTables(&visualization.Watermark{})

	repo := repository.NewWatermarkRepository(testDB)
	svc := NewWatermarkService(repo)

	result, err := svc.Find()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "default", result.ID)
	assert.Contains(t, result.SettingContent, "enable")
}

func TestWatermarkServiceIntegration_SavePersistsLatestSetting(t *testing.T) {
	cleanupTables(&visualization.Watermark{})

	repo := repository.NewWatermarkRepository(testDB)
	svc := NewWatermarkService(repo)

	content := `{"enable":true,"type":"custom"}`
	result, err := svc.Save(&visualization.WatermarkSaveRequest{SettingContent: content}, "tester")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, content, result.SettingContent)
	assert.Equal(t, "tester", result.CreateBy)

	found, err := svc.Find()
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, content, found.SettingContent)
	assert.Equal(t, "default", found.ID)
}
