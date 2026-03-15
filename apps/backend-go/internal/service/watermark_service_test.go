package service

import (
	"errors"
	"testing"

	"dataease/backend/internal/domain/visualization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWatermarkRepo struct {
	findLatestFunc  func() (*visualization.Watermark, error)
	saveDefaultFunc func(settingContent string, createBy string, createTime int64) (*visualization.Watermark, error)
}

func (m *mockWatermarkRepo) FindLatest() (*visualization.Watermark, error) {
	if m.findLatestFunc != nil {
		return m.findLatestFunc()
	}
	return nil, nil
}

func (m *mockWatermarkRepo) SaveDefault(settingContent string, createBy string, createTime int64) (*visualization.Watermark, error) {
	if m.saveDefaultFunc != nil {
		return m.saveDefaultFunc(settingContent, createBy, createTime)
	}
	return nil, nil
}

func TestWatermarkService_Find_ReturnsError_WhenRepoIsNil(t *testing.T) {
	svc := NewWatermarkService(nil)
	_, err := svc.Find()
	require.Error(t, err)
	assert.Equal(t, errWatermarkRepoNotReady, err)
}

func TestWatermarkService_Find_ReturnsDefault_WhenNoRecordExists(t *testing.T) {
	mock := &mockWatermarkRepo{
		findLatestFunc: func() (*visualization.Watermark, error) {
			return nil, nil
		},
	}
	svc := NewWatermarkService(mock)
	result, err := svc.Find()
	require.NoError(t, err)
	assert.Equal(t, "default", result.ID)
}

func TestWatermarkService_Find_ReturnsRecord_WhenExists(t *testing.T) {
	expected := &visualization.Watermark{
		ID:             "default",
		Version:        "v1",
		SettingContent: `{"enable":true}`,
	}
	mock := &mockWatermarkRepo{
		findLatestFunc: func() (*visualization.Watermark, error) {
			return expected, nil
		},
	}
	svc := NewWatermarkService(mock)
	result, err := svc.Find()
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestWatermarkService_Find_ReturnsError_WhenRepoFails(t *testing.T) {
	mock := &mockWatermarkRepo{
		findLatestFunc: func() (*visualization.Watermark, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewWatermarkService(mock)
	_, err := svc.Find()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestWatermarkService_Save_ReturnsError_WhenRepoIsNil(t *testing.T) {
	svc := NewWatermarkService(nil)
	_, err := svc.Save(&visualization.WatermarkSaveRequest{SettingContent: "{}"}, "user1")
	require.Error(t, err)
	assert.Equal(t, errWatermarkRepoNotReady, err)
}

func TestWatermarkService_Save_UsesDefault_WhenContentEmpty(t *testing.T) {
	var savedContent string
	mock := &mockWatermarkRepo{
		saveDefaultFunc: func(settingContent string, createBy string, createTime int64) (*visualization.Watermark, error) {
			savedContent = settingContent
			return &visualization.Watermark{ID: "default", SettingContent: settingContent}, nil
		},
	}
	svc := NewWatermarkService(mock)
	result, err := svc.Save(&visualization.WatermarkSaveRequest{SettingContent: ""}, "user1")
	require.NoError(t, err)
	assert.Contains(t, savedContent, "enable")
	assert.Equal(t, "default", result.ID)
}

func TestWatermarkService_Save_SavesProvidedContent(t *testing.T) {
	content := `{"enable":true,"type":"custom"}`
	var savedContent string
	mock := &mockWatermarkRepo{
		saveDefaultFunc: func(settingContent string, createBy string, createTime int64) (*visualization.Watermark, error) {
			savedContent = settingContent
			return &visualization.Watermark{ID: "default", SettingContent: settingContent}, nil
		},
	}
	svc := NewWatermarkService(mock)
	result, err := svc.Save(&visualization.WatermarkSaveRequest{SettingContent: content}, "user1")
	require.NoError(t, err)
	assert.Equal(t, content, savedContent)
	assert.Equal(t, content, result.SettingContent)
}

func TestWatermarkService_Save_ReturnsError_WhenRepoFails(t *testing.T) {
	mock := &mockWatermarkRepo{
		saveDefaultFunc: func(settingContent string, createBy string, createTime int64) (*visualization.Watermark, error) {
			return nil, errors.New("save error")
		},
	}
	svc := NewWatermarkService(mock)
	_, err := svc.Save(&visualization.WatermarkSaveRequest{SettingContent: "{}"}, "user1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save error")
}

func TestDefaultWatermarkSetting_ReturnsValidJSON(t *testing.T) {
	setting := defaultWatermarkSetting()
	assert.Contains(t, setting, "enable")
	assert.Contains(t, setting, "type")
	assert.Contains(t, setting, "content")
}
