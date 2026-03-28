package service

import (
	"testing"

	"dataease/backend/internal/domain/engine"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupEngineServiceRepoTest(t *testing.T) (*EngineService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&engine.Engine{}))

	repo := repository.NewEngineRepository(db)
	return NewEngineService(repo), db
}

func setupClosedEngineServiceRepoTest(t *testing.T) *EngineService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&engine.Engine{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return NewEngineService(repository.NewEngineRepository(db))
}

func TestEngineService_GetEngine(t *testing.T) {
	t.Run("returns nil when repository empty", func(t *testing.T) {
		svc, _ := setupEngineServiceRepoTest(t)

		result, err := svc.GetEngine()
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("maps persisted engine to dto", func(t *testing.T) {
		svc, db := setupEngineServiceRepoTest(t)
		config := "test configuration"
		require.NoError(t, db.Create(&engine.Engine{Name: "Test Engine", Type: "doris", Configuration: &config}).Error)

		result, err := svc.GetEngine()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Test Engine", result.Name)
		assert.Equal(t, "doris", result.Type)
		require.NotNil(t, result.Configuration)
		assert.Equal(t, "test configuration", *result.Configuration)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc := setupClosedEngineServiceRepoTest(t)

		result, err := svc.GetEngine()
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestEngineService_Validate(t *testing.T) {
	svc, _ := setupEngineServiceRepoTest(t)

	result, err := svc.Validate(&engine.ValidateRequest{Type: strPtr("doris")})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Success", result.Status)
	assert.Equal(t, "Engine validation not implemented", result.Message)
}

func TestEngineService_ValidateByID(t *testing.T) {
	svc, _ := setupEngineServiceRepoTest(t)

	result, err := svc.ValidateByID(1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Success", result.Status)
	assert.Equal(t, "Engine validation not implemented", result.Message)
}

func TestEngineService_SupportSetKey(t *testing.T) {
	svc, _ := setupEngineServiceRepoTest(t)

	result, err := svc.SupportSetKey()
	require.NoError(t, err)
	assert.False(t, result)
}
