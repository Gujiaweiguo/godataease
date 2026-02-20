package service

import (
	"encoding/json"
	"errors"
	"strings"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/system"
)

var errSystemParamRepoNotReady = errors.New("system parameter repository not initialized")

type SystemParamRepository interface {
	ListBasicSettings() ([]system.SettingItem, error)
	SaveBasicSettings(items []system.SettingItem) error

	GetOnlineMap() (*system.OnlineMapEditor, error)
	GetOnlineMapByType(mapType string) (*system.OnlineMapEditor, error)
	SaveOnlineMap(editor *system.OnlineMapEditor) error

	GetSQLBotConfig() (*system.SQLBotConfig, error)
	SaveSQLBotConfig(cfg *system.SQLBotConfig) error

	GetShareBase() (*system.ShareBase, error)

	GetRequestTimeOut() (int, error)
	GetDefaultSettings() (map[string]interface{}, error)
	GetUI() ([]interface{}, error)
	GetDefaultLogin() (int, error)
	GetI18nOptions() (map[string]string, error)
}

type SystemParamService struct {
	repo         SystemParamRepository
	auditService *AuditService
}

func NewSystemParamService(repo SystemParamRepository, auditService *AuditService) *SystemParamService {
	return &SystemParamService{repo: repo, auditService: auditService}
}

func (s *SystemParamService) QueryBasic() ([]system.SettingItem, error) {
	if s.repo == nil {
		return nil, errSystemParamRepoNotReady
	}
	return s.repo.ListBasicSettings()
}

func (s *SystemParamService) SaveBasic(items []system.SettingItem) error {
	if s.repo == nil {
		return errSystemParamRepoNotReady
	}
	if err := s.repo.SaveBasicSettings(items); err != nil {
		return err
	}
	if s.auditService != nil {
		afterValue, _ := json.Marshal(items)
		afterValueStr := string(afterValue)
		_, _ = s.auditService.CreateAuditLog(&audit.AuditLogCreateRequest{
			ActionType: audit.ActionTypeSystemConfig,
			ActionName: "保存基础设置",
			Operation:  audit.OperationUpdate,
			AfterValue: &afterValueStr,
		})
	}
	return nil
}

func (s *SystemParamService) QueryOnlineMap() (*system.OnlineMapEditor, error) {
	if s.repo == nil {
		return nil, errSystemParamRepoNotReady
	}
	return s.repo.GetOnlineMap()
}

func (s *SystemParamService) QueryOnlineMapByType(mapType string) (*system.OnlineMapEditor, error) {
	if s.repo == nil {
		return nil, errSystemParamRepoNotReady
	}
	mt := strings.TrimSpace(mapType)
	if mt == "" {
		return s.repo.GetOnlineMap()
	}
	return s.repo.GetOnlineMapByType(mt)
}

func (s *SystemParamService) SaveOnlineMap(editor *system.OnlineMapEditor) error {
	if s.repo == nil {
		return errSystemParamRepoNotReady
	}
	if err := s.repo.SaveOnlineMap(editor); err != nil {
		return err
	}
	if s.auditService != nil {
		afterValue, _ := json.Marshal(editor)
		afterValueStr := string(afterValue)
		_, _ = s.auditService.CreateAuditLog(&audit.AuditLogCreateRequest{
			ActionType: audit.ActionTypeSystemConfig,
			ActionName: "保存在线地图配置",
			Operation:  audit.OperationUpdate,
			AfterValue: &afterValueStr,
		})
	}
	return nil
}

func (s *SystemParamService) QuerySQLBot() (*system.SQLBotConfig, error) {
	if s.repo == nil {
		return nil, errSystemParamRepoNotReady
	}
	return s.repo.GetSQLBotConfig()
}

func (s *SystemParamService) SaveSQLBot(cfg *system.SQLBotConfig) error {
	if s.repo == nil {
		return errSystemParamRepoNotReady
	}
	if err := s.repo.SaveSQLBotConfig(cfg); err != nil {
		return err
	}
	if s.auditService != nil {
		afterValue, _ := json.Marshal(cfg)
		afterValueStr := string(afterValue)
		_, _ = s.auditService.CreateAuditLog(&audit.AuditLogCreateRequest{
			ActionType: audit.ActionTypeSystemConfig,
			ActionName: "保存SQLBot配置",
			Operation:  audit.OperationUpdate,
			AfterValue: &afterValueStr,
		})
	}
	return nil
}

func (s *SystemParamService) ShareBase() (*system.ShareBase, error) {
	if s.repo == nil {
		return nil, errSystemParamRepoNotReady
	}
	return s.repo.GetShareBase()
}

func (s *SystemParamService) RequestTimeOut() (int, error) {
	if s.repo == nil {
		return 0, errSystemParamRepoNotReady
	}
	return s.repo.GetRequestTimeOut()
}

func (s *SystemParamService) DefaultSettings() (map[string]interface{}, error) {
	if s.repo == nil {
		return nil, errSystemParamRepoNotReady
	}
	return s.repo.GetDefaultSettings()
}

func (s *SystemParamService) UI() ([]interface{}, error) {
	if s.repo == nil {
		return nil, errSystemParamRepoNotReady
	}
	return s.repo.GetUI()
}

func (s *SystemParamService) DefaultLogin() (int, error) {
	if s.repo == nil {
		return 0, errSystemParamRepoNotReady
	}
	return s.repo.GetDefaultLogin()
}

func (s *SystemParamService) I18nOptions() (map[string]string, error) {
	if s.repo == nil {
		return nil, errSystemParamRepoNotReady
	}
	return s.repo.GetI18nOptions()
}
