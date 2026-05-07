package repository

import (
	"context"
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/domain/user"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ==================== Setup helpers ====================

func setupSystemParamRepoTest(t *testing.T) *SystemParamRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&coreSysSetting{}))
	return NewSystemParamRepository(db)
}

func setupRound4TaskRepoTest(t *testing.T) (*TaskRepository, *SubTaskRepository, *SubInstanceRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&datafillingdomain.DataFillingTask{},
		&datafillingdomain.DataFillingSubTask{},
		&datafillingdomain.DataFillingSubInstance{},
	))
	return NewTaskRepository(db), NewSubTaskRepository(db), NewSubInstanceRepository(db), db
}

// ==================== SystemParamRepository tests ====================

func TestRound4Repo_SystemParam_NewRepo(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NotNil(t, repo)
}

func TestRound4Repo_SystemParam_ListBasicSettings_Empty(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	items, err := repo.ListBasicSettings()
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestRound4Repo_SystemParam_SaveAndListBasicSettings(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	items := []system.SettingItem{
		{Pkey: "basic.siteName", Pval: "DataEase", Type: "text", Sort: 1},
		{Pkey: "basic.locale", Pval: "zh-CN", Type: "text", Sort: 2},
	}
	require.NoError(t, repo.SaveBasicSettings(items))

	result, err := repo.ListBasicSettings()
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "basic.siteName", result[0].Pkey)
	assert.Equal(t, "DataEase", result[0].Pval)
	assert.Equal(t, "basic.locale", result[1].Pkey)
}

func TestRound4Repo_SystemParam_SaveBasicSettings_AutoPrefix(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	items := []system.SettingItem{
		{Pkey: "siteName", Pval: "AutoPrefixed", Type: "", Sort: 0},
	}
	require.NoError(t, repo.SaveBasicSettings(items))

	result, err := repo.ListBasicSettings()
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "basic.siteName", result[0].Pkey)
	assert.Equal(t, "text", result[0].Type)
	assert.Equal(t, 1, result[0].Sort)
}

func TestRound4Repo_SystemParam_SaveBasicSettings_Update(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "basic.siteName", Pval: "Old", Type: "text", Sort: 1},
	}))
	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "basic.siteName", Pval: "New", Type: "text", Sort: 1},
	}))
	result, err := repo.ListBasicSettings()
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "New", result[0].Pval)
}

func TestRound4Repo_SystemParam_GetOnlineMap_Default(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	editor, err := repo.GetOnlineMap()
	require.NoError(t, err)
	require.NotNil(t, editor)
	assert.Equal(t, "gaode", editor.MapType)
}

func TestRound4Repo_SystemParam_GetOnlineMapByType_Gaode(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	editor, err := repo.GetOnlineMapByType("gaode")
	require.NoError(t, err)
	assert.Equal(t, "gaode", editor.MapType)
}

func TestRound4Repo_SystemParam_GetOnlineMapByType_Custom(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	editor, err := repo.GetOnlineMapByType("tianditu")
	require.NoError(t, err)
	assert.Equal(t, "tianditu", editor.MapType)
}

func TestRound4Repo_SystemParam_SaveAndGetOnlineMap(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	editor := &system.OnlineMapEditor{MapType: "gaode", Key: "test-key", SecurityCode: "test-code"}
	require.NoError(t, repo.SaveOnlineMap(editor))

	loaded, err := repo.GetOnlineMap()
	require.NoError(t, err)
	assert.Equal(t, "gaode", loaded.MapType)
	assert.Equal(t, "test-key", loaded.Key)
	assert.Equal(t, "test-code", loaded.SecurityCode)
}

func TestRound4Repo_SystemParam_SaveOnlineMap_Nil(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveOnlineMap(nil))
}

func TestRound4Repo_SystemParam_SaveOnlineMap_NonGaode(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	editor := &system.OnlineMapEditor{MapType: "tianditu", Key: "tk", SecurityCode: "sc"}
	require.NoError(t, repo.SaveOnlineMap(editor))

	loaded, err := repo.GetOnlineMapByType("tianditu")
	require.NoError(t, err)
	assert.Equal(t, "tianditu", loaded.MapType)
	assert.Equal(t, "tk", loaded.Key)
}

func TestRound4Repo_SystemParam_SQLBot_Empty(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	cfg, err := repo.GetSQLBotConfig()
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestRound4Repo_SystemParam_SaveAndGetSQLBotConfig(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	cfg := &system.SQLBotConfig{Domain: "http://bot.test", ID: "abc", Enabled: true, Valid: false}
	require.NoError(t, repo.SaveSQLBotConfig(cfg))

	loaded, err := repo.GetSQLBotConfig()
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, "http://bot.test", loaded.Domain)
	assert.Equal(t, "abc", loaded.ID)
	assert.True(t, loaded.Enabled)
	assert.False(t, loaded.Valid)
}

func TestRound4Repo_SystemParam_SaveSQLBotConfig_Nil(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveSQLBotConfig(nil))
}

func TestRound4Repo_SystemParam_SaveSQLBotConfig_Update(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveSQLBotConfig(&system.SQLBotConfig{Domain: "old", ID: "1", Enabled: false, Valid: false}))
	require.NoError(t, repo.SaveSQLBotConfig(&system.SQLBotConfig{Domain: "new", ID: "2", Enabled: true, Valid: true}))

	loaded, err := repo.GetSQLBotConfig()
	require.NoError(t, err)
	assert.Equal(t, "new", loaded.Domain)
	assert.Equal(t, "2", loaded.ID)
}

func TestRound4Repo_SystemParam_GetShareBase_Default(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	sb, err := repo.GetShareBase()
	require.NoError(t, err)
	assert.NotNil(t, sb)
	assert.False(t, sb.Disable)
	assert.False(t, sb.PERequire)
}

func TestRound4Repo_SystemParam_GetShareBase_WithValues(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "basic.shareDisable", Pval: "true", Type: "text", Sort: 1},
		{Pkey: "basic.sharePeRequire", Pval: "true", Type: "text", Sort: 2},
	}))
	sb, err := repo.GetShareBase()
	require.NoError(t, err)
	assert.True(t, sb.Disable)
	assert.True(t, sb.PERequire)
}

func TestRound4Repo_SystemParam_GetRequestTimeOut_Default(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	timeout, err := repo.GetRequestTimeOut()
	require.NoError(t, err)
	assert.Equal(t, 60, timeout)
}

func TestRound4Repo_SystemParam_GetRequestTimeOut_Custom(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "basic.frontTimeOut", Pval: "120", Type: "text", Sort: 1},
	}))
	timeout, err := repo.GetRequestTimeOut()
	require.NoError(t, err)
	assert.Equal(t, 120, timeout)
}

func TestRound4Repo_SystemParam_GetRequestTimeOut_Invalid(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "basic.frontTimeOut", Pval: "not-a-number", Type: "text", Sort: 1},
	}))
	timeout, err := repo.GetRequestTimeOut()
	require.NoError(t, err)
	assert.Equal(t, 60, timeout)
}

func TestRound4Repo_SystemParam_GetRequestTimeOut_Negative(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "basic.frontTimeOut", Pval: "-5", Type: "text", Sort: 1},
	}))
	timeout, err := repo.GetRequestTimeOut()
	require.NoError(t, err)
	assert.Equal(t, 60, timeout)
}

func TestRound4Repo_SystemParam_GetDefaultSettings(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	settings, err := repo.GetDefaultSettings()
	require.NoError(t, err)
	assert.Equal(t, "1", settings["defaultSort"])
}

func TestRound4Repo_SystemParam_GetDefaultSettings_Custom(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "basic.defaultSort", Pval: "2", Type: "text", Sort: 1},
		{Pkey: "basic.defaultOpen", Pval: "dashboard", Type: "text", Sort: 2},
	}))
	settings, err := repo.GetDefaultSettings()
	require.NoError(t, err)
	assert.Equal(t, "2", settings["defaultSort"])
	assert.Equal(t, "dashboard", settings["defaultOpen"])
}

func TestRound4Repo_SystemParam_GetUI(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	ui, err := repo.GetUI()
	require.NoError(t, err)
	assert.Len(t, ui, 3)
}

func TestRound4Repo_SystemParam_GetDefaultLogin(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	login, err := repo.GetDefaultLogin()
	require.NoError(t, err)
	assert.Equal(t, 0, login)
}

func TestRound4Repo_SystemParam_GetDefaultLogin_Custom(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "basic.defaultLogin", Pval: "2", Type: "text", Sort: 1},
	}))
	login, err := repo.GetDefaultLogin()
	require.NoError(t, err)
	assert.Equal(t, 2, login)
}

func TestRound4Repo_SystemParam_GetDefaultLogin_Invalid(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "basic.defaultLogin", Pval: "abc", Type: "text", Sort: 1},
	}))
	login, err := repo.GetDefaultLogin()
	require.NoError(t, err)
	assert.Equal(t, 0, login)
}

func TestRound4Repo_SystemParam_GetI18nOptions(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	opts, err := repo.GetI18nOptions()
	require.NoError(t, err)
	assert.Empty(t, opts)
}

func TestRound4Repo_SystemParam_GetSettingValueByKey(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveBasicSettings([]system.SettingItem{
		{Pkey: "basic.testKey", Pval: "hello", Type: "text", Sort: 1},
	}))
	val, err := repo.GetSettingValueByKey("basic.testKey")
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestRound4Repo_SystemParam_GetSettingValueByKey_NotFound(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	val, err := repo.GetSettingValueByKey("basic.nonexistent")
	require.NoError(t, err)
	assert.Equal(t, "", val)
}

func TestRound4Repo_SystemParam_SaveSettingValueByKey(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveSettingValueByKey("custom.key", "value1"))
	val, err := repo.GetSettingValueByKey("custom.key")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestRound4Repo_SystemParam_SaveSettingValueByKey_Update(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveSettingValueByKey("custom.key", "v1"))
	require.NoError(t, repo.SaveSettingValueByKey("custom.key", "v2"))
	val, err := repo.GetSettingValueByKey("custom.key")
	require.NoError(t, err)
	assert.Equal(t, "v2", val)
}

func TestRound4Repo_SystemParam_SaveSettingValueByKey_EmptyKey(t *testing.T) {
	repo := setupSystemParamRepoTest(t)
	require.NoError(t, repo.SaveSettingValueByKey("", "value"))
	require.NoError(t, repo.SaveSettingValueByKey("   ", "value"))
}

// ==================== Private helper tests ====================

func TestRound4Repo_SystemParam_ParseBool(t *testing.T) {
	assert.True(t, parseBool("true"))
	assert.True(t, parseBool("TRUE"))
	assert.True(t, parseBool(" True "))
	assert.False(t, parseBool("false"))
	assert.False(t, parseBool(""))
	assert.False(t, parseBool("1"))
}

func TestRound4Repo_SystemParam_MapKeyPrefixByType(t *testing.T) {
	assert.Equal(t, "map.", mapKeyPrefixByType("gaode"))
	assert.Equal(t, "tianditu.map.", mapKeyPrefixByType("tianditu"))
}

// ==================== ResourcePerm additional edge cases ====================

func TestRound4Repo_ResourcePerm_ListPerms_EmptyType(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	desc := "test"
	perm1 := &permission.SysPerm{PermName: "P1", PermKey: "dataset:view", PermType: permission.PermTypeData, PermDesc: &desc, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	perm2 := &permission.SysPerm{PermName: "P2", PermKey: "screen:view", PermType: "menu", PermDesc: &desc, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.CreatePerm(perm1))
	require.NoError(t, repo.CreatePerm(perm2))

	perms, total, err := repo.ListPerms("", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, perms, 2)

	permsData, totalData, err := repo.ListPerms(permission.PermTypeData, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalData)
	require.Len(t, permsData, 1)
	assert.Equal(t, "dataset:view", permsData[0].PermKey)
}

func TestRound4Repo_ResourcePerm_GetUserPerms_NoRows(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	permIDs, err := repo.GetUserPerms(999)
	require.NoError(t, err)
	assert.Empty(t, permIDs)
}

func TestRound4Repo_ResourcePerm_GetRolePerms_NoRows(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	permIDs, err := repo.GetRolePerms(999)
	require.NoError(t, err)
	assert.Empty(t, permIDs)
}

func TestRound4Repo_ResourcePerm_GetUserRoleIDs_NoRows(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	roleIDs, err := repo.GetUserRoleIDs(999)
	require.NoError(t, err)
	assert.Empty(t, roleIDs)
}

func TestRound4Repo_ResourcePerm_GrantPermToUser_ZeroOrg(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	err := repo.GrantPermToUser(501, 601, "admin")
	require.NoError(t, err)
	hasPerm, err := repo.CheckUserPermission(501, 601)
	require.NoError(t, err)
	assert.True(t, hasPerm)
}

func TestRound4Repo_ResourcePerm_ReplaceResourcePermissions_InvalidInput(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	err := repo.ReplaceResourcePermissions(0, permission.ResourceTypeDataset, []int64{1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resource id and type are required")

	err = repo.ReplaceResourcePermissions(1, "", []int64{1})
	require.Error(t, err)
}

func TestRound4Repo_ResourcePerm_RegisterResource_UpdateOnlyName(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	require.NoError(t, repo.RegisterResource(100, "Initial", permission.ResourceTypeDashboard, nil))
	require.NoError(t, repo.RegisterResource(100, "", permission.ResourceTypeDashboard, nil))
}

func TestRound4Repo_ResourcePerm_GetUserResources_NoPerms(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	results, err := repo.GetUserResources(999, permission.ResourceTypeDataset)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestRound4Repo_ResourcePerm_GetResourceUsers_NotGovernedNoPerms(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	results, err := repo.GetResourceUsers(999, "unknown_type")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestRound4Repo_ResourcePerm_CheckPermissionConsistency_EmptyDB(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	result, err := repo.CheckPermissionConsistency()
	require.NoError(t, err)
	assert.True(t, result.Consistent)
	assert.Zero(t, result.UserCount)
	assert.Zero(t, result.ResourceCount)
}

func TestRound4Repo_ResourcePerm_CheckPermissionConsistencyByOrg_Zero(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	result, err := repo.CheckPermissionConsistencyByOrg(0)
	require.NoError(t, err)
	assert.True(t, result.Consistent)
}

func TestRound4Repo_ResourcePerm_ParseInconsistency(t *testing.T) {
	inc := parseInconsistency("123:dataset:view", "granted", "missing", resourceRow{}, "user %d has %s")
	require.NotNil(t, inc)
	assert.Equal(t, int64(123), inc.UserID)
	assert.Equal(t, "dataset", inc.ResourceType)
	assert.Equal(t, "granted", inc.UserView)
	assert.Equal(t, "missing", inc.ResourceView)

	inc = parseInconsistency("bad", "granted", "missing", resourceRow{}, "user %d has %s")
	assert.Nil(t, inc)

	inc = parseInconsistency("abc:foo", "granted", "missing", resourceRow{}, "user %d has %s")
	assert.Nil(t, inc)
}

// ==================== Task additional edge cases ====================

func TestRound4Repo_Task_NewConstructors(t *testing.T) {
	taskRepo, subTaskRepo, subInstanceRepo, _ := setupRound4TaskRepoTest(t)
	require.NotNil(t, taskRepo)
	require.NotNil(t, subTaskRepo)
	require.NotNil(t, subInstanceRepo)
}

func TestRound4Repo_Task_DeleteTasksByIDs_Empty(t *testing.T) {
	ctx := context.Background()
	repo, _, _, _ := setupRound4TaskRepoTest(t)
	require.NoError(t, repo.DeleteTasksByIDs(ctx, nil))
	require.NoError(t, repo.DeleteTasksByIDs(ctx, []int64{}))
}

func TestRound4Repo_Task_ListTasksByFormID_PageBeyondData(t *testing.T) {
	ctx := context.Background()
	repo, _, _, _ := setupRound4TaskRepoTest(t)
	require.NoError(t, repo.CreateTask(ctx, &datafillingdomain.DataFillingTask{FormID: 1, Name: "T1", Status: datafillingdomain.TaskStatusStarted, FillType: 1}))

	rows, total, err := repo.ListTasksByFormID(ctx, 1, 99, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Empty(t, rows)
}

func TestRound4Repo_Task_GetStartedTasks_NoneStarted(t *testing.T) {
	ctx := context.Background()
	repo, _, _, _ := setupRound4TaskRepoTest(t)
	require.NoError(t, repo.CreateTask(ctx, &datafillingdomain.DataFillingTask{FormID: 1, Name: "Stopped", Status: datafillingdomain.TaskStatusStopped, FillType: 1}))
	rows, err := repo.GetStartedTasks(ctx)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound4Repo_SubTask_DeleteSubTasksByIDs_Empty(t *testing.T) {
	ctx := context.Background()
	_, repo, _, _ := setupRound4TaskRepoTest(t)
	require.NoError(t, repo.DeleteSubTasksByIDs(ctx, nil))
	require.NoError(t, repo.DeleteSubTasksByIDs(ctx, []int64{}))
}

func TestRound4Repo_SubTask_ListSubTasksByTaskID_Empty(t *testing.T) {
	ctx := context.Background()
	_, repo, _, _ := setupRound4TaskRepoTest(t)
	rows, total, err := repo.ListSubTasksByTaskID(ctx, 999, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, rows)
}

func TestRound4Repo_SubTask_ListSubTaskIDsByTaskIDs_Empty(t *testing.T) {
	ctx := context.Background()
	_, repo, _, _ := setupRound4TaskRepoTest(t)
	ids, err := repo.ListSubTaskIDsByTaskIDs(ctx, []int64{999})
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestRound4Repo_SubInstance_BatchCreate_Empty(t *testing.T) {
	ctx := context.Background()
	_, _, repo, _ := setupRound4TaskRepoTest(t)
	require.NoError(t, repo.BatchCreateSubInstances(ctx, nil))
	require.NoError(t, repo.BatchCreateSubInstances(ctx, []*datafillingdomain.DataFillingSubInstance{}))
}

func TestRound4Repo_SubInstance_DeleteByPIDs_Empty(t *testing.T) {
	ctx := context.Background()
	_, _, repo, _ := setupRound4TaskRepoTest(t)
	require.NoError(t, repo.DeleteSubInstancesByPIDs(ctx, nil))
	require.NoError(t, repo.DeleteSubInstancesByPIDs(ctx, []int64{}))
}

func TestRound4Repo_SubInstance_DeleteByTaskIDs_Empty(t *testing.T) {
	ctx := context.Background()
	_, _, repo, _ := setupRound4TaskRepoTest(t)
	require.NoError(t, repo.DeleteSubInstancesByTaskIDs(ctx, nil))
	require.NoError(t, repo.DeleteSubInstancesByTaskIDs(ctx, []int64{}))
}

func TestRound4Repo_SubInstance_ListByPID_NoMatch(t *testing.T) {
	ctx := context.Background()
	_, _, repo, _ := setupRound4TaskRepoTest(t)
	rows, err := repo.ListSubInstancesByPID(ctx, 999, nil)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound4Repo_SubInstance_CountOpen_NoMatch(t *testing.T) {
	ctx := context.Background()
	_, _, repo, _ := setupRound4TaskRepoTest(t)
	count, err := repo.CountOpenSubInstancesByUID(ctx, 999)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRound4Repo_SubInstance_GetByPIDAndUID_NoMatch(t *testing.T) {
	ctx := context.Background()
	_, _, repo, _ := setupRound4TaskRepoTest(t)
	rows, err := repo.GetSubInstanceByPIDAndUID(ctx, 999, 999)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound4Repo_SubInstance_ListByUID_NoMatch(t *testing.T) {
	ctx := context.Background()
	_, _, repo, _ := setupRound4TaskRepoTest(t)
	rows, total, err := repo.ListSubInstancesByUID(ctx, 999, 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, rows)
}

func TestRound4Repo_ResourcePerm_GetUserResourcesByOrg_ZeroOrgDelegates(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	datasetView := &permission.SysPerm{PermName: "DV", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(datasetView).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 1101, Username: "test-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 1101, PermID: datasetView.PermID, Status: 1, DelFlag: 0}).Error)

	results, err := repo.GetUserResourcesByOrg(1101, permission.ResourceTypeDataset, 0)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "direct", results[0].SourceType)
}

func TestRound4Repo_ResourcePerm_GetResourceUsersByOrg_ZeroOrgDelegates(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	require.NoError(t, repo.RegisterResource(200, "test-resource", permission.ResourceTypeDataset, nil))

	results, err := repo.GetResourceUsersByOrg(200, permission.ResourceTypeDataset, 0)
	require.NoError(t, err)
	assert.Empty(t, results)
}
