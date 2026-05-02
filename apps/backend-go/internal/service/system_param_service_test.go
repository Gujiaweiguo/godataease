package service

import (
	"errors"
	"testing"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSystemParamRepository implements SystemParamRepository for testing
type MockSystemParamRepository struct {
	basicSettings         []system.SettingItem
	onlineMap             *system.OnlineMapEditor
	onlineMapByType       *system.OnlineMapEditor
	sqlBotConfig          *system.SQLBotConfig
	settingsByKey         map[string]string
	shareBase             *system.ShareBase
	requestTimeOut        int
	defaultSettings       map[string]interface{}
	ui                    []interface{}
	defaultLogin          int
	i18nOptions           map[string]string
	saveBasicErr          error
	saveOnlineMapErr      error
	saveSQLBotErr         error
	saveSettingByKeyErr   error
	getOnlineMapErr       error
	getOnlineMapByTypeErr error
	getSQLBotErr          error
	getSettingByKeyErr    error
	getShareBaseErr       error
	getRequestTimeOutErr  error
	getDefaultSettingsErr error
	getUIErr              error
	getDefaultLoginErr    error
	getI18nOptionsErr     error
	lastMapType           string
}

func NewMockSystemParamRepository() *MockSystemParamRepository {
	return &MockSystemParamRepository{
		basicSettings: []system.SettingItem{
			{Pkey: "test.key", Pval: "test.value", Type: "basic", Sort: 1},
		},
		settingsByKey:   map[string]string{},
		shareBase:       &system.ShareBase{Disable: false, PERequire: true},
		requestTimeOut:  30,
		defaultSettings: map[string]interface{}{"key": "value"},
		ui:              []interface{}{},
		defaultLogin:    0,
		i18nOptions:     map[string]string{"en": "English", "zh": "中文"},
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
	m.lastMapType = mapType
	if m.getOnlineMapByTypeErr != nil {
		return nil, m.getOnlineMapByTypeErr
	}
	if m.onlineMapByType != nil {
		return m.onlineMapByType, nil
	}
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
	if m.getSQLBotErr != nil {
		return nil, m.getSQLBotErr
	}
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

func (m *MockSystemParamRepository) GetSettingValueByKey(key string) (string, error) {
	if m.getSettingByKeyErr != nil {
		return "", m.getSettingByKeyErr
	}
	return m.settingsByKey[key], nil
}

func (m *MockSystemParamRepository) SaveSettingValueByKey(key, value string) error {
	if m.saveSettingByKeyErr != nil {
		return m.saveSettingByKeyErr
	}
	m.settingsByKey[key] = value
	return nil
}

func (m *MockSystemParamRepository) GetShareBase() (*system.ShareBase, error) {
	if m.getShareBaseErr != nil {
		return nil, m.getShareBaseErr
	}
	if m.shareBase == nil {
		return &system.ShareBase{Disable: false, PERequire: true}, nil
	}
	return m.shareBase, nil
}

func (m *MockSystemParamRepository) GetRequestTimeOut() (int, error) {
	if m.getRequestTimeOutErr != nil {
		return 0, m.getRequestTimeOutErr
	}
	return m.requestTimeOut, nil
}

func (m *MockSystemParamRepository) GetDefaultSettings() (map[string]interface{}, error) {
	if m.getDefaultSettingsErr != nil {
		return nil, m.getDefaultSettingsErr
	}
	return m.defaultSettings, nil
}

func (m *MockSystemParamRepository) GetUI() ([]interface{}, error) {
	if m.getUIErr != nil {
		return nil, m.getUIErr
	}
	return m.ui, nil
}

func (m *MockSystemParamRepository) GetDefaultLogin() (int, error) {
	if m.getDefaultLoginErr != nil {
		return 0, m.getDefaultLoginErr
	}
	return m.defaultLogin, nil
}

func (m *MockSystemParamRepository) GetI18nOptions() (map[string]string, error) {
	if m.getI18nOptionsErr != nil {
		return nil, m.getI18nOptionsErr
	}
	return m.i18nOptions, nil
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

func TestSystemParam_QueryAuditAlertSettings_DefaultWhenMissing(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	svc := NewSystemParamService(mockRepo, nil)

	settings, err := svc.QueryAuditAlertSettings()
	require.NoError(t, err)
	assert.Equal(t, audit.DefaultAuditAlertSettings(), settings)
}

func TestSystemParam_QueryAuditAlertSettings_Success(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	stored := audit.DefaultAuditAlertSettings()
	stored.EnableEmailNotification = true
	stored.NotificationEmail = "audit@example.com"
	data, err := stored.ToJSON()
	require.NoError(t, err)
	mockRepo.settingsByKey[auditAlertSettingsKey] = string(data)

	svc := NewSystemParamService(mockRepo, nil)
	settings, err := svc.QueryAuditAlertSettings()
	require.NoError(t, err)
	assert.Equal(t, stored, settings)
}

func TestSystemParam_SaveAuditAlertSettings_Success(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	svc := NewSystemParamService(mockRepo, nil)

	settings := audit.DefaultAuditAlertSettings()
	settings.EnableEmailNotification = true
	settings.NotificationEmail = "audit@example.com"

	err := svc.SaveAuditAlertSettings(settings)
	require.NoError(t, err)

	stored := mockRepo.settingsByKey[auditAlertSettingsKey]
	require.NotEmpty(t, stored)
	decoded, err := audit.AuditAlertSettingsFromJSON([]byte(stored))
	require.NoError(t, err)
	assert.Equal(t, settings, decoded)
}

func TestSystemParam_SaveAuditAlertSettings_ValidateFirst(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	svc := NewSystemParamService(mockRepo, nil)

	settings := audit.DefaultAuditAlertSettings()
	settings.RetentionDays = 1

	err := svc.SaveAuditAlertSettings(settings)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.settingsByKey[auditAlertSettingsKey])
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
	mockRepo := NewMockSystemParamRepository()
	mockRepo.onlineMapByType = &system.OnlineMapEditor{MapType: "gaode", Key: "typed-key"}
	svc := NewSystemParamService(mockRepo, nil)

	// Test with empty type
	result, err := svc.QueryOnlineMapByType("")
	if err != nil {
		t.Fatalf("QueryOnlineMapByType with empty type failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}

	// Test with specific type
	result, err = svc.QueryOnlineMapByType("  gaode  ")
	if err != nil {
		t.Fatalf("QueryOnlineMapByType failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Key != "typed-key" {
		t.Fatalf("expected typed lookup result, got %#v", result)
	}
	if mockRepo.lastMapType != "gaode" {
		t.Fatalf("expected trimmed map type 'gaode', got %q", mockRepo.lastMapType)
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

func TestSystemParam_Getters_RepoErrors(t *testing.T) {
	t.Run("query sql bot", func(t *testing.T) {
		mockRepo := NewMockSystemParamRepository()
		mockRepo.getSQLBotErr = errors.New("sqlbot error")
		svc := NewSystemParamService(mockRepo, nil)

		cfg, err := svc.QuerySQLBot()
		if err == nil {
			t.Fatal("expected QuerySQLBot error")
		}
		if cfg != nil {
			t.Fatalf("expected nil config, got %#v", cfg)
		}
	})

	t.Run("share base", func(t *testing.T) {
		mockRepo := NewMockSystemParamRepository()
		mockRepo.getShareBaseErr = errors.New("share base error")
		svc := NewSystemParamService(mockRepo, nil)

		cfg, err := svc.ShareBase()
		if err == nil {
			t.Fatal("expected ShareBase error")
		}
		if cfg != nil {
			t.Fatalf("expected nil share base, got %#v", cfg)
		}
	})

	t.Run("request timeout", func(t *testing.T) {
		mockRepo := NewMockSystemParamRepository()
		mockRepo.getRequestTimeOutErr = errors.New("timeout error")
		svc := NewSystemParamService(mockRepo, nil)

		timeout, err := svc.RequestTimeOut()
		if err == nil {
			t.Fatal("expected RequestTimeOut error")
		}
		if timeout != 0 {
			t.Fatalf("expected zero timeout on error, got %d", timeout)
		}
	})

	t.Run("default settings", func(t *testing.T) {
		mockRepo := NewMockSystemParamRepository()
		mockRepo.getDefaultSettingsErr = errors.New("default settings error")
		svc := NewSystemParamService(mockRepo, nil)

		settings, err := svc.DefaultSettings()
		if err == nil {
			t.Fatal("expected DefaultSettings error")
		}
		if settings != nil {
			t.Fatalf("expected nil settings, got %#v", settings)
		}
	})

	t.Run("ui", func(t *testing.T) {
		mockRepo := NewMockSystemParamRepository()
		mockRepo.getUIErr = errors.New("ui error")
		svc := NewSystemParamService(mockRepo, nil)

		ui, err := svc.UI()
		if err == nil {
			t.Fatal("expected UI error")
		}
		if ui != nil {
			t.Fatalf("expected nil ui, got %#v", ui)
		}
	})

	t.Run("default login", func(t *testing.T) {
		mockRepo := NewMockSystemParamRepository()
		mockRepo.getDefaultLoginErr = errors.New("default login error")
		svc := NewSystemParamService(mockRepo, nil)

		login, err := svc.DefaultLogin()
		if err == nil {
			t.Fatal("expected DefaultLogin error")
		}
		if login != 0 {
			t.Fatalf("expected zero default login on error, got %d", login)
		}
	})

	t.Run("i18n options", func(t *testing.T) {
		mockRepo := NewMockSystemParamRepository()
		mockRepo.getI18nOptionsErr = errors.New("i18n error")
		svc := NewSystemParamService(mockRepo, nil)

		options, err := svc.I18nOptions()
		if err == nil {
			t.Fatal("expected I18nOptions error")
		}
		if options != nil {
			t.Fatalf("expected nil options, got %#v", options)
		}
	})
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
	auditSvc, db := setupAuditServiceRepoTest(t)
	svc := NewSystemParamService(mockRepo, auditSvc)

	editor := &system.OnlineMapEditor{
		MapType:      "gaode",
		Key:          "test-key",
		SecurityCode: "test-code",
	}

	err := svc.SaveOnlineMap(editor)
	if err != nil {
		t.Fatalf("SaveOnlineMap with audit failed: %v", err)
	}
	var count int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestSystemParam_SaveSQLBot_WithAudit(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	auditSvc, db := setupAuditServiceRepoTest(t)
	svc := NewSystemParamService(mockRepo, auditSvc)

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
	var count int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestSystemParam_SaveBasic_WithAudit(t *testing.T) {
	mockRepo := NewMockSystemParamRepository()
	auditSvc, db := setupAuditServiceRepoTest(t)
	svc := NewSystemParamService(mockRepo, auditSvc)
	items := []system.SettingItem{{Pkey: "audit.key", Pval: "audit.value", Type: "basic", Sort: 1}}

	require.NoError(t, svc.SaveBasic(items))
	assert.Equal(t, items, mockRepo.basicSettings)

	var count int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestSystemParam_SaveMethods_AuditFailureDoesNotBreakSave(t *testing.T) {
	t.Run("save basic", func(t *testing.T) {
		mockRepo := NewMockSystemParamRepository()
		auditSvc, db := setupAuditServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_audit_insert_basic BEFORE INSERT ON de_audit_log BEGIN SELECT RAISE(FAIL, 'deny audit insert'); END;").Error)
		svc := NewSystemParamService(mockRepo, auditSvc)
		items := []system.SettingItem{{Pkey: "after.key", Pval: "after.value", Type: "basic", Sort: 2}}

		require.NoError(t, svc.SaveBasic(items))
		assert.Equal(t, items, mockRepo.basicSettings)
	})

	t.Run("save online map", func(t *testing.T) {
		mockRepo := NewMockSystemParamRepository()
		auditSvc, db := setupAuditServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_audit_insert_map BEFORE INSERT ON de_audit_log BEGIN SELECT RAISE(FAIL, 'deny audit insert'); END;").Error)
		svc := NewSystemParamService(mockRepo, auditSvc)
		editor := &system.OnlineMapEditor{MapType: "gaode", Key: "map-key", SecurityCode: "map-code"}

		require.NoError(t, svc.SaveOnlineMap(editor))
		require.NotNil(t, mockRepo.onlineMap)
		assert.Equal(t, editor, mockRepo.onlineMap)
	})

	t.Run("save sql bot", func(t *testing.T) {
		mockRepo := NewMockSystemParamRepository()
		auditSvc, db := setupAuditServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_audit_insert_sqlbot BEFORE INSERT ON de_audit_log BEGIN SELECT RAISE(FAIL, 'deny audit insert'); END;").Error)
		svc := NewSystemParamService(mockRepo, auditSvc)
		cfg := &system.SQLBotConfig{Domain: "sqlbot.domain", ID: "sqlbot-id", Enabled: true, Valid: true}

		require.NoError(t, svc.SaveSQLBot(cfg))
		require.NotNil(t, mockRepo.sqlBotConfig)
		assert.Equal(t, cfg, mockRepo.sqlBotConfig)
	})
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
