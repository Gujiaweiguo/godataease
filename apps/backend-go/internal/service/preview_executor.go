package service

import (
	"fmt"
	"strings"

	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/repository"

	mysqlconfig "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type PreviewExecutor interface {
	PreviewSQL(rawSQL string, limit int) ([]map[string]interface{}, error)
	Close() error
}

type PreviewExecutorFactory func(ds *datasource.CoreDatasource, cfg *datasource.ConnectionConfig) (PreviewExecutor, error)

type localPreviewExecutor struct {
	repo *repository.DatasetRepository
}

func (e *localPreviewExecutor) PreviewSQL(rawSQL string, limit int) ([]map[string]interface{}, error) {
	if e == nil || e.repo == nil {
		return nil, fmt.Errorf("dataset repository is unavailable")
	}
	return e.repo.PreviewSQL(rawSQL, limit)
}

func (e *localPreviewExecutor) Close() error {
	return nil
}

type mysqlPreviewExecutor struct {
	db *gorm.DB
}

func (e *mysqlPreviewExecutor) PreviewSQL(rawSQL string, limit int) ([]map[string]interface{}, error) {
	if e == nil || e.db == nil {
		return nil, fmt.Errorf("mysql preview executor is unavailable")
	}
	if limit < 1 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	rows := make([]map[string]interface{}, 0)
	query := fmt.Sprintf("SELECT * FROM (%s) AS de_preview LIMIT ?", rawSQL)
	if err := e.db.Raw(query, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (e *mysqlPreviewExecutor) Close() error {
	if e == nil || e.db == nil {
		return nil
	}
	sqlDB, err := e.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func buildMySQLPreviewConfig(cfg *datasource.ConnectionConfig) (*mysqlconfig.Config, error) {
	if cfg == nil {
		return nil, fmt.Errorf("datasource configuration is required")
	}
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		return nil, fmt.Errorf("datasource host is required")
	}
	if cfg.Port <= 0 {
		return nil, fmt.Errorf("datasource port is required")
	}
	databaseName := strings.TrimSpace(cfg.Database)
	if databaseName == "" {
		return nil, fmt.Errorf("datasource database is required")
	}
	username := strings.TrimSpace(cfg.Username)
	if username == "" {
		return nil, fmt.Errorf("datasource username is required")
	}
	config := mysqlconfig.NewConfig()
	config.User = username
	config.Passwd = cfg.Password
	config.Net = "tcp"
	config.Addr = fmt.Sprintf("%s:%d", host, cfg.Port)
	config.DBName = databaseName
	config.Params = map[string]string{
		"charset":   "utf8mb4",
		"parseTime": "True",
		"loc":       "Local",
	}

	params := strings.TrimSpace(cfg.ExtraParams)
	if params != "" {
		trimmed := strings.TrimPrefix(params, "?")
		if trimmed != "" {
			for _, part := range strings.Split(trimmed, "&") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				pieces := strings.SplitN(part, "=", 2)
				if len(pieces) != 2 {
					continue
				}
				key := strings.TrimSpace(pieces[0])
				value := strings.TrimSpace(pieces[1])
				if key == "" {
					continue
				}
				config.Params[key] = value
			}
		}
	}

	return config, nil
}

func defaultPreviewExecutorFactory(ds *datasource.CoreDatasource, cfg *datasource.ConnectionConfig) (PreviewExecutor, error) {
	if ds == nil {
		return nil, fmt.Errorf("datasource is required")
	}
	if !strings.EqualFold(strings.TrimSpace(ds.Type), "mysql") {
		return nil, ErrPreviewSQLExternalDatasourceUnsupported
	}
	mysqlConfig, err := buildMySQLPreviewConfig(cfg)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(mysql.Open(mysqlConfig.FormatDSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect datasource preview")
	}
	return &mysqlPreviewExecutor{db: db}, nil
}
