//go:build integration
// +build integration

package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatermarkRepository_FindLatest_ReturnsNilWhenNoRecordExists(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewWatermarkRepository(testDB)
	cleanupTables("visualization_watermark")

	result, err := repo.FindLatest()
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestWatermarkRepository_SaveDefault_UpsertsSingleRecord(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewWatermarkRepository(testDB)
	cleanupTables("visualization_watermark")

	first, err := repo.SaveDefault(`{"enable":true}`, "user1", time.Now().Unix())
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Equal(t, "default", first.ID)
	assert.Equal(t, `{"enable":true}`, first.SettingContent)

	second, err := repo.SaveDefault(`{"enable":false}`, "user2", time.Now().Add(time.Second).Unix())
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.Equal(t, "default", second.ID)
	assert.Equal(t, `{"enable":false}`, second.SettingContent)

	stored, err := repo.FindLatest()
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.Equal(t, "default", stored.ID)
	assert.Equal(t, `{"enable":false}`, stored.SettingContent)
	assert.Equal(t, "user2", stored.CreateBy)

	var count int64
	err = testDB.Table("visualization_watermark").Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
