//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemParamRepository_BasicSettingsAndShareSettings(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_sys_setting")

	repo := NewSystemParamRepository(testDB)
	seed := []coreSysSetting{
		{Pkey: "basic.companyName", Pval: "DataEase", Type: "text", Sort: 2},
		{Pkey: "basic.shareDisable", Pval: "true", Type: "text", Sort: 3},
		{Pkey: "basic.sharePeRequire", Pval: "false", Type: "text", Sort: 4},
	}
	for i := range seed {
		require.NoError(t, testDB.Create(&seed[i]).Error)
	}

	basicList, err := repo.ListBasicSettings()
	require.NoError(t, err)
	require.Len(t, basicList, 3)
	assert.Equal(t, "basic.companyName", basicList[0].Pkey)

	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "frontTimeOut", Pval: "120", Type: "", Sort: 0},
		{Pkey: "basic.companyName", Pval: "DataEase Pro", Type: "textarea", Sort: 9},
	}))

	companyName, err := repo.GetSettingValueByKey("basic.companyName")
	require.NoError(t, err)
	assert.Equal(t, "DataEase Pro", companyName)

	frontTimeoutVal, err := repo.GetSettingValueByKey("basic.frontTimeOut")
	require.NoError(t, err)
	assert.Equal(t, "120", frontTimeoutVal)

	shareBase, err := repo.GetShareBase()
	require.NoError(t, err)
	assert.True(t, shareBase.Disable)
	assert.False(t, shareBase.PERequire)

	requestTimeout, err := repo.GetRequestTimeOut()
	require.NoError(t, err)
	assert.Equal(t, 120, requestTimeout)

	require.NoError(t, repo.SaveSettingValueByKey("basic.defaultLogin", "2"))
	defaultLogin, err := repo.GetDefaultLogin()
	require.NoError(t, err)
	assert.Equal(t, 2, defaultLogin)

	settings, err := repo.GetDefaultSettings()
	require.NoError(t, err)
	assert.Equal(t, "1", settings["defaultSort"])

	require.NoError(t, repo.SaveSettingValueByKey("basic.defaultSort", "3"))
	require.NoError(t, repo.SaveSettingValueByKey("basic.defaultOpen", "tab"))

	settings, err = repo.GetDefaultSettings()
	require.NoError(t, err)
	assert.Equal(t, "3", settings["defaultSort"])
	assert.Equal(t, "tab", settings["defaultOpen"])

	missing, err := repo.GetSettingValueByKey("basic.missing")
	require.NoError(t, err)
	assert.Empty(t, missing)

	require.NoError(t, repo.SaveSettingValueByKey("", "ignored"))
	unchanged, err := repo.GetSettingValueByKey("basic.companyName")
	require.NoError(t, err)
	assert.Equal(t, "DataEase Pro", unchanged)
}

func TestSystemParamRepository_OnlineMapAndSQLBot(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_sys_setting")

	repo := NewSystemParamRepository(testDB)

	defaultMap, err := repo.GetOnlineMap()
	require.NoError(t, err)
	assert.Equal(t, "gaode", defaultMap.MapType)
	assert.Empty(t, defaultMap.Key)

	gaodeMap, err := repo.GetOnlineMapByType("gaode")
	require.NoError(t, err)
	assert.Equal(t, "gaode", gaodeMap.MapType)

	require.NoError(t, repo.SaveOnlineMap(&system.OnlineMapEditor{MapType: "gaode", Key: "gaode-key", SecurityCode: "gaode-sec"}))
	require.NoError(t, repo.SaveOnlineMap(&system.OnlineMapEditor{MapType: "custom", Key: "custom-key", SecurityCode: "custom-sec"}))
	require.NoError(t, repo.SaveSettingValueByKey("map.mapType", "custom"))
	require.NoError(t, repo.SaveOnlineMap(nil))

	currentMap, err := repo.GetOnlineMap()
	require.NoError(t, err)
	assert.Equal(t, "custom", currentMap.MapType)
	assert.Equal(t, "custom-key", currentMap.Key)
	assert.Equal(t, "custom-sec", currentMap.SecurityCode)

	customMap, err := repo.GetOnlineMapByType("custom")
	require.NoError(t, err)
	assert.Equal(t, "custom-key", customMap.Key)

	sqlbotEmpty, err := repo.GetSQLBotConfig()
	require.NoError(t, err)
	assert.Nil(t, sqlbotEmpty)

	require.NoError(t, repo.SaveSQLBotConfig(&system.SQLBotConfig{Domain: "https://sqlbot.example.com", ID: "bot-1", Enabled: true, Valid: false}))
	require.NoError(t, repo.SaveSQLBotConfig(nil))

	sqlbotCfg, err := repo.GetSQLBotConfig()
	require.NoError(t, err)
	require.NotNil(t, sqlbotCfg)
	assert.Equal(t, "https://sqlbot.example.com", sqlbotCfg.Domain)
	assert.Equal(t, "bot-1", sqlbotCfg.ID)
	assert.True(t, sqlbotCfg.Enabled)
	assert.False(t, sqlbotCfg.Valid)

	var sqlbotRows []coreSysSetting
	require.NoError(t, testDB.Where("pkey LIKE ?", "sqlbot.%").Find(&sqlbotRows).Error)
	assert.Len(t, sqlbotRows, 4)
}

func TestSystemParamRepository_DefaultMethods(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_sys_setting")

	repo := NewSystemParamRepository(testDB)

	shareBase, err := repo.GetShareBase()
	require.NoError(t, err)
	assert.False(t, shareBase.Disable)
	assert.False(t, shareBase.PERequire)

	requestTimeout, err := repo.GetRequestTimeOut()
	require.NoError(t, err)
	assert.Equal(t, 60, requestTimeout)

	defaultSettings, err := repo.GetDefaultSettings()
	require.NoError(t, err)
	assert.Equal(t, "1", defaultSettings["defaultSort"])
	_, hasDefaultOpen := defaultSettings["defaultOpen"]
	assert.False(t, hasDefaultOpen)

	defaultLogin, err := repo.GetDefaultLogin()
	require.NoError(t, err)
	assert.Equal(t, 0, defaultLogin)

	ui, err := repo.GetUI()
	require.NoError(t, err)
	require.Len(t, ui, 3)
	firstUI, ok := ui[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "community", firstUI["pkey"])
	assert.Equal(t, true, firstUI["pval"])

	i18n, err := repo.GetI18nOptions()
	require.NoError(t, err)
	assert.Empty(t, i18n)
}
