package repository

import (
	"testing"

	"dataease/backend/internal/domain/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupEngineRepoTest(t *testing.T) (*EngineRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&engine.Engine{}))
	return NewEngineRepository(db), db
}

func TestEngineRepository_Get_Empty(t *testing.T) {
	repo, _ := setupEngineRepoTest(t)

	_, err := repo.Get()
	require.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestEngineRepository_Get_Success(t *testing.T) {
	repo, db := setupEngineRepoTest(t)

	config := "mysql://localhost:3306"
	createBy := "admin"
	createTime := int64(1700000000)
	require.NoError(t, db.Create(&engine.Engine{
		Name:          "default-engine",
		Type:          "mysql",
		Configuration: &config,
		CreateBy:      &createBy,
		CreateTime:    &createTime,
	}).Error)

	eng, err := repo.Get()
	require.NoError(t, err)
	assert.Equal(t, "default-engine", eng.Name)
	assert.Equal(t, "mysql", eng.Type)
	require.NotNil(t, eng.Configuration)
	assert.Equal(t, "mysql://localhost:3306", *eng.Configuration)
}

func TestEngineRepository_Get_MultipleRowsReturnsFirst(t *testing.T) {
	repo, db := setupEngineRepoTest(t)

	require.NoError(t, db.Create(&engine.Engine{Name: "engine-1", Type: "mysql"}).Error)
	require.NoError(t, db.Create(&engine.Engine{Name: "engine-2", Type: "clickhouse"}).Error)

	eng, err := repo.Get()
	require.NoError(t, err)
	assert.Equal(t, "engine-1", eng.Name)
}
