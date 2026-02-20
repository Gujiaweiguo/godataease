package app

import (
	"dataease/backend/internal/pkg/logger"
)

// Application 应用实例
type Application struct {
	Name    string
	Version string
	Config  *Config
}

// Init 初始化应用
func Init() (*Application, error) {
	// 加载配置
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// 初始化日志
	if err := logger.Init(&logger.Config{
		Level:  config.Log.Level,
		Format: config.Log.Format,
	}); err != nil {
		return nil, err
	}

	return &Application{
		Name:    "dataease-backend",
		Version: "1.0.0",
		Config:  config,
	}, nil
}
