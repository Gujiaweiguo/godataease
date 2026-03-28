package service

import (
	"encoding/json"
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
	assert.Equal(t, "v1", result.Version)
	assert.Equal(t, defaultWatermarkSetting(), result.SettingContent)
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

func TestWatermarkService_Find_ReturnsDefault_WhenSettingContentEmpty(t *testing.T) {
	mock := &mockWatermarkRepo{
		findLatestFunc: func() (*visualization.Watermark, error) {
			return &visualization.Watermark{ID: "existing", Version: "v1", SettingContent: ""}, nil
		},
	}
	svc := NewWatermarkService(mock)

	result, err := svc.Find()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "default", result.ID)
	assert.Equal(t, "v1", result.Version)
	assert.Equal(t, defaultWatermarkSetting(), result.SettingContent)
	assert.Contains(t, result.SettingContent, "enable")
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

func TestWatermarkService_Save_UsesDefault_WhenRequestNil(t *testing.T) {
	var savedContent string
	mock := &mockWatermarkRepo{
		saveDefaultFunc: func(settingContent string, createBy string, createTime int64) (*visualization.Watermark, error) {
			savedContent = settingContent
			return &visualization.Watermark{ID: "default", SettingContent: settingContent}, nil
		},
	}
	svc := NewWatermarkService(mock)

	result, err := svc.Save(nil, "user1")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, savedContent, "enable")
	assert.Equal(t, savedContent, result.SettingContent)
}

func TestWatermarkService_Save_PassesCreateByAndTimestamp(t *testing.T) {
	var gotCreateBy string
	var gotCreateTime int64
	mock := &mockWatermarkRepo{
		saveDefaultFunc: func(settingContent string, createBy string, createTime int64) (*visualization.Watermark, error) {
			gotCreateBy = createBy
			gotCreateTime = createTime
			return &visualization.Watermark{ID: "default", SettingContent: settingContent}, nil
		},
	}
	svc := NewWatermarkService(mock)

	_, err := svc.Save(&visualization.WatermarkSaveRequest{SettingContent: "{}"}, "tester")
	require.NoError(t, err)
	assert.Equal(t, "tester", gotCreateBy)
	assert.Positive(t, gotCreateTime)
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
	assert.NotEqual(t, defaultWatermarkSetting(), savedContent)
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

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(setting), &payload))
	assert.Equal(t, false, payload["enable"])
	assert.Equal(t, false, payload["enablePanelCustom"])
	assert.Equal(t, "userName", payload["type"])
	assert.Equal(t, "${time}-${ip}-${nickName}", payload["content"])
	assert.Equal(t, "#999999", payload["watermark_color"])
	assert.Equal(t, float64(100), payload["watermark_x_space"])
	assert.Equal(t, float64(100), payload["watermark_y_space"])
	assert.Equal(t, float64(20), payload["watermark_fontsize"])
}
