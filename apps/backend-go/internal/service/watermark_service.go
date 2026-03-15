package service

import (
	"encoding/json"
	"errors"
	"time"

	"dataease/backend/internal/domain/visualization"
)

var errWatermarkRepoNotReady = errors.New("watermark repository not initialized")

type WatermarkRepository interface {
	FindLatest() (*visualization.Watermark, error)
	SaveDefault(settingContent string, createBy string, createTime int64) (*visualization.Watermark, error)
}

type WatermarkService struct {
	repo WatermarkRepository
}

func NewWatermarkService(repo WatermarkRepository) *WatermarkService {
	return &WatermarkService{repo: repo}
}

func (s *WatermarkService) Find() (*visualization.Watermark, error) {
	if s.repo == nil {
		return nil, errWatermarkRepoNotReady
	}
	record, err := s.repo.FindLatest()
	if err != nil {
		return nil, err
	}
	if record != nil && record.SettingContent != "" {
		return record, nil
	}
	return &visualization.Watermark{
		ID:             "default",
		Version:        "v1",
		SettingContent: defaultWatermarkSetting(),
	}, nil
}

func (s *WatermarkService) Save(req *visualization.WatermarkSaveRequest, createBy string) (*visualization.Watermark, error) {
	if s.repo == nil {
		return nil, errWatermarkRepoNotReady
	}
	setting := ""
	if req != nil {
		setting = req.SettingContent
	}
	if setting == "" {
		setting = defaultWatermarkSetting()
	}
	now := time.Now().Unix()
	return s.repo.SaveDefault(setting, createBy, now)
}

func defaultWatermarkSetting() string {
	payload := map[string]interface{}{
		"enable":             false,
		"enablePanelCustom":  false,
		"type":               "userName",
		"content":            "${time}-${ip}-${nickName}",
		"watermark_color":    "#999999",
		"watermark_x_space":  100,
		"watermark_y_space":  100,
		"watermark_fontsize": 20,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(buf)
}
