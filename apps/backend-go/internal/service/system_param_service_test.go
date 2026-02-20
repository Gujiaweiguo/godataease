package service

import (
	"errors"
	"testing"

	"dataease/backend/internal/domain/system"
)

// MockSystemParamRepository implements SystemParamRepository for testing
type MockSystemParamRepository struct {
	basicSettings    []system.SettingItem
	onlineMap        *system.OnlineMapEditor
	sqlBotConfig     *system.SQLBotConfig
	saveBasicErr     error
	saveOnlineMapErr error
	saveSQLBotErr    error
}

func NewMockSystemParamRepository() *MockSystemParamRepository {
	return &MockSystemParamRepository{
		basicSettings: []system.SettingItem{
			{Pkey: "test.key", Pval: "test.value", Type: "basic", Sort: 1},
		},
	}
}

func (m *MockSystemParamRepository) ListBasicSettings() ([]system.SettingItem, error) {
	return m.basicSettings, nil
}

func (m *MockSystemParamRepository) SaveBasicSettings(items []system.SettingItem) error {
	if m.saveBasicErr != nil {
		return m.saveBasicErr
	}
	m.basicSettings = items
	return nil
}

func (m *MockSystemParamRepository) GetOnlineMap() (*system.OnlineMapEditor, error) {
	if m.onlineMap == nil {
		return &system.OnlineMapEditor{MapType: "gaode", Key: "test-key"}, nil
	}
	return m.onlineMap, nil
}

func (m *MockSystemParamRepository) GetOnlineMapByType(mapType string) (*system.OnlineMapEditor, error) {
	return m.GetOnlineMap()
}

func (m *MockSystemParamRepository) SaveOnlineMap(editor *system.OnlineMapEditor) error {
	if m.saveOnlineMapErr != nil {
		return m.saveOnlineMapErr
	}
	m.onlineMap = editor
	return nil
}

func (m *MockSystemParamRepository) GetSQLBotConfig() (*system.SQLBotConfig, error) {
	if m.sqlBotConfig == nil {
		return &system.SQLBotConfig{Domain: "test.domain", Enabled: true}, nil
	}
	return m.sqlBotConfig, nil
}

func (m *MockSystemParamRepository) SaveSQLBotConfig(cfg *system.SQLBotConfig) error {
	if m.saveSQLBotErr != nil {
		return m.saveSQLBotErr
	}
	m.sqlBotConfig = cfg
	return nil
}

func (m *MockSystemParamRepository) GetShareBase() (*system.ShareBase, error) {
	return &system.ShareBase{Disable: false, PERequire: true}, nil
}

func (m *MockSystemParamRepository) GetRequestTimeOut() (int, error) {
	return 30, nil
}

func (m *MockSystemParamRepository) GetDefaultSettings() (map[string]interface{}, error) {
	return map[string]interface{}{"key": "value"}, nil
}

func (m *MockSystemParamRepository) GetUI() ([]interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockSystemParamRepository) GetDefaultLogin() (int, error) {
	return 0, nil
}

func (m *MockSystemParamRepository) GetI18nOptions() (map[string]string, error) {
	return map[string]string{"en": "English", "zh": "中文"}, nil
}

func setupSystemParamService() *SystemParamService {
	mockRepo := NewMockSystemParamRepository()
	return NewSystemParamService(mockRepo, nil)
}

func TestSystemParam_QueryBasic(t *testing.T) {
	svc := setupSystemParamService()

	items, err := svc.QueryBasic()
	if err != nil {
		t.Fatalf("QueryBasic failed: %v", err)
	}
	if len(items) == 0 {
		t.Error("Expected non-empty settings list")
	}
}

func TestSystemParam_SaveBasic_Success(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	svc := NewSystemParamService(mockRepo, nil)

	items := []system.SettingItem{
		{Pkey: "new.key", Pval: "new.value", Type: "basic", Sort: 1},
	}

	err := svc.SaveBasic(items)
	if err != nil {
		t.Fatalf("SaveBasic failed: %v", err)
	}
}

func TestSystemParam_SaveOnlineMap_Success(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	svc := NewSystemParamService(mockRepo, nil)

	editor := &system.OnlineMapEditor{
		MapType:      "gaode",
		Key:          "test-key",
		SecurityCode: "test-code",
	}

	err := svc.SaveOnlineMap(editor)
	if err != nil {
		t.Fatalf("SaveOnlineMap failed: %v", err)
	}
}

func TestSystemParam_SaveSQLBot_Success(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	svc := NewSystemParamService(mockRepo, nil)

	cfg := &system.SQLBotConfig{
		Domain:  "test.domain",
		ID:      "test-id",
		Enabled: true,
		Valid:   true,
	}

	err := svc.SaveSQLBot(cfg)
	if err != nil {
		t.Fatalf("SaveSQLBot failed: %v", err)
	}
}

func TestSystemParam_RepoNotReady(t *testing.T) {
	// Service with nil repo
	svc := &SystemParamService{repo: nil, auditService: nil}

	_, err := svc.QueryBasic()
	if err == nil {
		t.Error("Expected error when repo is nil")
	}
	if !errors.Is(err, errSystemParamRepoNotReady) {
		t.Errorf("Expected errSystemParamRepoNotReady, got %v", err)
	}

	err = svc.SaveBasic([]system.SettingItem{})
	if err == nil {
		t.Error("Expected error when repo is nil for SaveBasic")
	}

	_, err = svc.QueryOnlineMap()
	if err == nil {
		t.Error("Expected error when repo is nil for QueryOnlineMap")
	}
}

func TestSystemParam_SaveBasic_WithoutAudit(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	// Service with nil audit service
	svc := NewSystemParamService(mockRepo, nil)

	items := []system.SettingItem{
		{Pkey: "key", Pval: "value", Type: "basic", Sort: 1},
	}

	err := svc.SaveBasic(items)
	if err != nil {
		t.Fatalf("SaveBasic failed: %v", err)
	}
	// Should succeed without audit
}

func TestSystemParam_QueryOnlineMapByType(t *testing.T) {
	svc := setupSystemParamService()

	// Test with empty type
	result, err := svc.QueryOnlineMapByType("")
	if err != nil {
		t.Fatalf("QueryOnlineMapByType with empty type failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}

	// Test with specific type
	result, err = svc.QueryOnlineMapByType("gaode")
	if err != nil {
		t.Fatalf("QueryOnlineMapByType failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestSystemParam_ShareBase(t *testing.T) {
	svc := setupSystemParamService()

	result, err := svc.ShareBase()
	if err != nil {
		t.Fatalf("ShareBase failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil ShareBase")
	}
}

func TestSystemParam_RequestTimeOut(t *testing.T) {
	svc := setupSystemParamService()

	timeout, err := svc.RequestTimeOut()
	if err != nil {
		t.Fatalf("RequestTimeOut failed: %v", err)
	}
	if timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", timeout)
	}
}

func TestSystemParam_DefaultSettings(t *testing.T) {
	svc := setupSystemParamService()

	settings, err := svc.DefaultSettings()
	if err != nil {
		t.Fatalf("DefaultSettings failed: %v", err)
	}
	if settings == nil {
		t.Error("Expected non-nil settings")
	}
}

func TestSystemParam_I18nOptions(t *testing.T) {
	svc := setupSystemParamService()

	options, err := svc.I18nOptions()
	if err != nil {
		t.Fatalf("I18nOptions failed: %v", err)
	}
	if len(options) == 0 {
		t.Error("Expected non-empty i18n options")
	}
}
