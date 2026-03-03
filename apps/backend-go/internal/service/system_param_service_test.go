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
	getOnlineMapErr  error
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
	if m.getOnlineMapErr != nil {
		return nil, m.getOnlineMapErr
	}
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

func TestSystemParam_QueryOnlineMap(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	svc := NewSystemParamService(mockRepo, nil)

	result, err := svc.QueryOnlineMap()
	if err != nil {
		t.Fatalf("QueryOnlineMap failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	mockRepo.getOnlineMapErr = errors.New("query online map error")
	_, err = svc.QueryOnlineMap()
	if err == nil {
		t.Error("Expected error when repository GetOnlineMap fails")
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

func TestSystemParam_QuerySQLBotAndUIAndDefaultLogin(t *testing.T) {
	svc := setupSystemParamService()

	sqlbot, err := svc.QuerySQLBot()
	if err != nil {
		t.Fatalf("QuerySQLBot failed: %v", err)
	}
	if sqlbot == nil {
		t.Fatal("expected non-nil sqlbot config")
	}

	ui, err := svc.UI()
	if err != nil {
		t.Fatalf("UI failed: %v", err)
	}
	if len(ui) != 0 {
		t.Fatalf("expected empty ui list, got %d", len(ui))
	}

	login, err := svc.DefaultLogin()
	if err != nil {
		t.Fatalf("DefaultLogin failed: %v", err)
	}
	if login != 0 {
		t.Fatalf("expected default login 0, got %d", login)
	}
}

func TestSystemParam_SaveOnlineMap_WithAudit(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	svc := NewSystemParamService(mockRepo, nil)

	editor := &system.OnlineMapEditor{
		MapType:      "gaode",
		Key:          "test-key",
		SecurityCode: "test-code",
	}

	err := svc.SaveOnlineMap(editor)
	if err != nil {
		t.Fatalf("SaveOnlineMap with audit failed: %v", err)
	}
}

func TestSystemParam_SaveSQLBot_WithAudit(t *testing.T) {
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
		t.Fatalf("SaveSQLBot with audit failed: %v", err)
	}
}

func TestSystemParam_SaveBasic_RepoError(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	mockRepo.saveBasicErr = errors.New("save basic error")
	svc := NewSystemParamService(mockRepo, nil)

	err := svc.SaveBasic([]system.SettingItem{{Pkey: "test", Pval: "value"}})
	if err == nil {
		t.Error("Expected error from SaveBasic")
	}
}

func TestSystemParam_SaveOnlineMap_RepoError(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	mockRepo.saveOnlineMapErr = errors.New("save online map error")
	svc := NewSystemParamService(mockRepo, nil)

	editor := &system.OnlineMapEditor{MapType: "gaode"}
	err := svc.SaveOnlineMap(editor)
	if err == nil {
		t.Error("Expected error from SaveOnlineMap")
	}
}

func TestSystemParam_SaveSQLBot_RepoError(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	mockRepo.saveSQLBotErr = errors.New("save sql bot error")
	svc := NewSystemParamService(mockRepo, nil)

	cfg := &system.SQLBotConfig{Domain: "test.domain"}
	err := svc.SaveSQLBot(cfg)
	if err == nil {
		t.Error("Expected error from SaveSQLBot")
	}
}

func TestSystemParam_RepoNotReady_AllMethods(t *testing.T) {
	// Service with nil repo
	svc := &SystemParamService{repo: nil, auditService: nil}

	// Test ShareBase
	_, err := svc.ShareBase()
	if err == nil {
		t.Error("Expected error when repo is nil for ShareBase")
	}

	// Test RequestTimeOut
	_, err = svc.RequestTimeOut()
	if err == nil {
		t.Error("Expected error when repo is nil for RequestTimeOut")
	}

	// Test DefaultSettings
	_, err = svc.DefaultSettings()
	if err == nil {
		t.Error("Expected error when repo is nil for DefaultSettings")
	}

	// Test UI
	_, err = svc.UI()
	if err == nil {
		t.Error("Expected error when repo is nil for UI")
	}

	// Test DefaultLogin
	_, err = svc.DefaultLogin()
	if err == nil {
		t.Error("Expected error when repo is nil for DefaultLogin")
	}

	// Test I18nOptions
	_, err = svc.I18nOptions()
	if err == nil {
		t.Error("Expected error when repo is nil for I18nOptions")
	}

	// Test QuerySQLBot
	_, err = svc.QuerySQLBot()
	if err == nil {
		t.Error("Expected error when repo is nil for QuerySQLBot")
	}
}
