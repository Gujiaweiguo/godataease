package service

import (
	"testing"

	"dataease/backend/internal/domain/embedded"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupEmbeddedServiceRepoTest(t *testing.T) (*EmbeddedService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&embedded.CoreEmbedded{}))

	repo := repository.NewEmbeddedRepository(db)
	return NewEmbeddedService(repo), db
}

func setupClosedEmbeddedServiceRepoTest(t *testing.T) *EmbeddedService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&embedded.CoreEmbedded{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return NewEmbeddedService(repository.NewEmbeddedRepository(db))
}

func closeEmbeddedDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func TestEmbeddedServiceHelpers_SplitBy(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, splitBy("a..b.c", '.'))
	assert.Empty(t, splitBy("", '.'))
}

func TestEmbeddedServiceHelpers_ParseClaim(t *testing.T) {
	key, val := parseClaim(`"appId":"test123"`)
	assert.Equal(t, "appId", key)
	assert.Equal(t, "test123", val)

	key, val = parseClaim("userId=456")
	assert.Equal(t, "userId", key)
	assert.Equal(t, "456", val)

	key, val = parseClaim("noseparator")
	assert.Equal(t, "", key)
	assert.Equal(t, "", val)
}

func TestEmbeddedServiceHelpers_TrimQuotesAndTrim(t *testing.T) {
	assert.Equal(t, "test", trimQuotes(`"test"`))
	assert.Equal(t, "test", trimQuotes(`'test'`))
	assert.Equal(t, "test", trimQuotes(`"'test'"`))
	assert.Equal(t, "abc", trim("---abc---", '-'))
	assert.Equal(t, "a-b-c", trim("-a-b-c-", '-'))
}

func TestEmbeddedServiceHelpers_DecodeBase64(t *testing.T) {
	assert.Equal(t, "a+b/c", decodeBase64("a-b_c"))
	assert.Equal(t, "abc123", decodeBase64("abc123"))
}

func TestEmbeddedServiceHelpers_SplitToken(t *testing.T) {
	assert.Nil(t, splitToken("nodots"))
	assert.Equal(t, []string{"appId:app123", "userId:1"}, splitToken("header.appId:app123,userId:1.signature"))
}

func TestEmbeddedService_ExtractAppIDFromToken(t *testing.T) {
	svc := &EmbeddedService{}

	appID, err := svc.extractAppIdFromToken("header.appId:app123,userId:1.signature")
	require.NoError(t, err)
	assert.Equal(t, "app123", appID)

	appID, err = svc.extractAppIdFromToken("header.userId:1.signature")
	require.Error(t, err)
	assert.Empty(t, appID)
	assert.Contains(t, err.Error(), "appId claim not found")
}

func TestEmbeddedService_GetTokenArgsAndLimitCount(t *testing.T) {
	svc := &EmbeddedService{}

	args := svc.GetTokenArgs(123, 456)
	require.NotNil(t, args)
	assert.Equal(t, int64(123), args.UserId)
	assert.Equal(t, int64(456), args.OrgId)
	assert.Equal(t, 5, svc.GetLimitCount())
}

func TestEmbeddedService_GetDomainList_Unit(t *testing.T) {
	t.Run("deduplicates parsed domains", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		require.NoError(t, db.Create(&embedded.CoreEmbedded{Name: "A", AppId: "appA", AppSecret: "secret", Domain: "http://a.com,http://b.com", SecretLength: embedded.DefaultSecretLength}).Error)
		require.NoError(t, db.Create(&embedded.CoreEmbedded{Name: "B", AppId: "appB", AppSecret: "secret", Domain: "http://b.com;http://c.com/", SecretLength: embedded.DefaultSecretLength}).Error)

		result, err := svc.GetDomainList()
		require.NoError(t, err)
		assert.Equal(t, []string{"http://a.com", "http://b.com", "http://c.com"}, result)
	})

	t.Run("repo error", func(t *testing.T) {
		svc := setupClosedEmbeddedServiceRepoTest(t)

		result, err := svc.GetDomainList()
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get domain list")
	})
}

func TestEmbeddedService_InitIframe_Unit(t *testing.T) {
	t.Run("empty token", func(t *testing.T) {
		svc, _ := setupEmbeddedServiceRepoTest(t)

		result, err := svc.InitIframe("", "http://allowed.com")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "embedded token cannot be empty")
	})

	t.Run("invalid token", func(t *testing.T) {
		svc, _ := setupEmbeddedServiceRepoTest(t)

		result, err := svc.InitIframe("header.userId:1.signature", "http://allowed.com")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid embedded token")
	})

	t.Run("app not found", func(t *testing.T) {
		svc, _ := setupEmbeddedServiceRepoTest(t)

		result, err := svc.InitIframe("header.appId:missing.signature", "http://allowed.com")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "embedded app not found")
	})

	t.Run("origin not allowed", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		require.NoError(t, db.Create(&embedded.CoreEmbedded{Name: "App", AppId: "app123", AppSecret: "secret", Domain: "http://allowed.com", SecretLength: embedded.DefaultSecretLength}).Error)

		result, err := svc.InitIframe("header.appId:app123.signature", "http://forbidden.com")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "embedded origin not allowed")
	})

	t.Run("valid token and origin", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		require.NoError(t, db.Create(&embedded.CoreEmbedded{Name: "App", AppId: "app123", AppSecret: "secret", Domain: "http://allowed.com,http://other.com", SecretLength: embedded.DefaultSecretLength}).Error)

		result, err := svc.InitIframe("header.appId:app123.signature", "http://allowed.com")
		require.NoError(t, err)
		assert.Equal(t, []string{"http://allowed.com", "http://other.com"}, result)
	})
}

func TestEmbeddedService_CRUDAndQueryGrid(t *testing.T) {
	t.Run("create uses default secret length", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)

		id, err := svc.Create(&embedded.EmbeddedCreator{Name: "Default App", Domain: "http://a.com"}, "tester")
		require.NoError(t, err)
		assert.NotZero(t, id)

		var item embedded.CoreEmbedded
		require.NoError(t, db.First(&item, id).Error)
		assert.Equal(t, embedded.DefaultSecretLength, item.SecretLength)
		assert.Equal(t, "Default App", item.Name)
		assert.Equal(t, "http://a.com", item.Domain)
		assert.NotEmpty(t, item.AppId)
		assert.NotEmpty(t, item.AppSecret)
		assert.Equal(t, "tester", item.UpdateBy)
	})

	t.Run("create uses provided secret length", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		length := 24

		id, err := svc.Create(&embedded.EmbeddedCreator{Name: "Custom App", Domain: "http://b.com", SecretLength: &length}, "tester")
		require.NoError(t, err)

		var item embedded.CoreEmbedded
		require.NoError(t, db.First(&item, id).Error)
		assert.Equal(t, 24, item.SecretLength)
		assert.Len(t, item.AppSecret, 24)
	})

	t.Run("edit not found returns wrapped error", func(t *testing.T) {
		svc, _ := setupEmbeddedServiceRepoTest(t)

		err := svc.Edit(&embedded.EmbeddedEditor{ID: 999, Name: "missing"}, "tester")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "embedded app not found")
	})

	t.Run("edit updates only provided fields", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		length := 20
		id, err := svc.Create(&embedded.EmbeddedCreator{Name: "Editable", Domain: "http://before.com", SecretLength: &length}, "creator")
		require.NoError(t, err)

		newDomain := "http://after.com"
		require.NoError(t, svc.Edit(&embedded.EmbeddedEditor{ID: id, Domain: &newDomain}, "editor"))

		var item embedded.CoreEmbedded
		require.NoError(t, db.First(&item, id).Error)
		assert.Equal(t, "Editable", item.Name)
		assert.Equal(t, "http://after.com", item.Domain)
		assert.Equal(t, 20, item.SecretLength)
		assert.Equal(t, "editor", item.UpdateBy)
	})

	t.Run("delete success and wrapped error", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		id, err := svc.Create(&embedded.EmbeddedCreator{Name: "Delete Me"}, "tester")
		require.NoError(t, err)

		require.NoError(t, svc.Delete(id))
		var count int64
		require.NoError(t, db.Model(&embedded.CoreEmbedded{}).Where("id = ?", id).Count(&count).Error)
		assert.Zero(t, count)

		closeEmbeddedDB(t, db)
		err = svc.Delete(123)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete embedded app")
	})

	t.Run("batch delete rejects empty and succeeds", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		err := svc.BatchDelete(nil)
		require.Error(t, err)
		assert.EqualError(t, err, "ids list cannot be empty")

		id1, err := svc.Create(&embedded.EmbeddedCreator{Name: "Batch A"}, "tester")
		require.NoError(t, err)
		id2, err := svc.Create(&embedded.EmbeddedCreator{Name: "Batch B"}, "tester")
		require.NoError(t, err)

		require.NoError(t, svc.BatchDelete([]int64{id1, id2}))
		var count int64
		require.NoError(t, db.Model(&embedded.CoreEmbedded{}).Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("reset secret not found provided and generated", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		err := svc.ResetSecret(&embedded.EmbeddedResetRequest{ID: 999}, "tester")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "embedded app not found")

		length := 18
		id, err := svc.Create(&embedded.EmbeddedCreator{Name: "Reset Me", SecretLength: &length}, "creator")
		require.NoError(t, err)

		customSecret := "manual-secret"
		require.NoError(t, svc.ResetSecret(&embedded.EmbeddedResetRequest{ID: id, AppSecret: &customSecret}, "resetter"))
		var item embedded.CoreEmbedded
		require.NoError(t, db.First(&item, id).Error)
		assert.Equal(t, "manual-secret", item.AppSecret)
		assert.Equal(t, "resetter", item.UpdateBy)

		previous := item.AppSecret
		require.NoError(t, svc.ResetSecret(&embedded.EmbeddedResetRequest{ID: id}, "resetter2"))
		require.NoError(t, db.First(&item, id).Error)
		assert.Len(t, item.AppSecret, 18)
		assert.NotEqual(t, previous, item.AppSecret)
		assert.Equal(t, "resetter2", item.UpdateBy)
	})

	t.Run("query grid masks secrets and wraps repo error", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		length := 12
		id, err := svc.Create(&embedded.EmbeddedCreator{Name: "Query Me", Domain: "http://q.com", SecretLength: &length}, "tester")
		require.NoError(t, err)

		var item embedded.CoreEmbedded
		require.NoError(t, db.First(&item, id).Error)

		resp, err := svc.QueryGrid("Query", 1, 10)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Total)
		assert.Equal(t, 1, resp.Current)
		assert.Equal(t, 10, resp.Size)
		require.Len(t, resp.List, 1)
		assert.Equal(t, embedded.MaskAppSecret(item.AppSecret), resp.List[0].AppSecret)
		assert.Equal(t, item.AppId, resp.List[0].AppId)
		assert.Equal(t, item.Domain, resp.List[0].Domain)

		emptyResp, err := svc.QueryGrid("missing", 1, 10)
		require.NoError(t, err)
		require.NotNil(t, emptyResp)
		assert.Empty(t, emptyResp.List)
		assert.Zero(t, emptyResp.Total)

		closeEmbeddedDB(t, db)
		resp, err = svc.QueryGrid("Query", 1, 10)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to query embedded apps")
	})

	t.Run("create edit batch delete reset secret and init iframe wrap repo errors", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		closeEmbeddedDB(t, db)

		id, err := svc.Create(&embedded.EmbeddedCreator{Name: "Broken Create"}, "tester")
		require.Error(t, err)
		assert.Zero(t, id)
		assert.Contains(t, err.Error(), "failed to create embedded app")

		err = svc.BatchDelete([]int64{1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to batch delete embedded apps")

		result, err := svc.InitIframe("header.appId:app123.signature", "http://allowed.com")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "embedded app not found")
	})

	t.Run("edit update error wraps repo failure", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		id, err := svc.Create(&embedded.EmbeddedCreator{Name: "Update Error", Domain: "http://before.com"}, "creator")
		require.NoError(t, err)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_embedded_update BEFORE UPDATE ON core_embedded BEGIN SELECT RAISE(FAIL, 'deny update'); END;").Error)

		newDomain := "http://after.com"
		err = svc.Edit(&embedded.EmbeddedEditor{ID: id, Domain: &newDomain}, "editor")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update embedded app")
	})

	t.Run("reset secret update error wraps repo failure", func(t *testing.T) {
		svc, db := setupEmbeddedServiceRepoTest(t)
		id, err := svc.Create(&embedded.EmbeddedCreator{Name: "Reset Error"}, "creator")
		require.NoError(t, err)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_embedded_secret_update BEFORE UPDATE ON core_embedded BEGIN SELECT RAISE(FAIL, 'deny secret update'); END;").Error)

		err = svc.ResetSecret(&embedded.EmbeddedResetRequest{ID: id}, "resetter")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to reset secret")
	})
}

func TestEmbeddedService_GetDomainList_EmptyReturnsNoDomains(t *testing.T) {
	svc, _ := setupEmbeddedServiceRepoTest(t)

	result, err := svc.GetDomainList()
	require.NoError(t, err)
	assert.Len(t, result, 0)
}
