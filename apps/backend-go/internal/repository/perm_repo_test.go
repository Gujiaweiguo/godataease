package repository

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/permission"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPermRepositoryTest(t *testing.T) (*PermRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&permission.SysPerm{}))
	return NewPermRepository(db), db
}

func seedPerm(t *testing.T, db *gorm.DB, perm *permission.SysPerm) {
	t.Helper()
	require.NoError(t, db.Create(perm).Error)
}

func TestPermRepository_CreateUpdateGetAndDelete(t *testing.T) {
	repo, db := setupPermRepositoryTest(t)
	desc := "initial desc"
	createBy := "tester"
	perm := &permission.SysPerm{
		PermName:   "Dataset View",
		PermKey:    "dataset:view",
		PermType:   permission.PermTypeData,
		PermDesc:   &desc,
		Status:     permission.StatusEnabled,
		CreateBy:   &createBy,
		CreateTime: time.Now(),
		DelFlag:    permission.DelFlagNormal,
	}

	require.NoError(t, repo.Create(perm))
	require.Positive(t, perm.PermID)

	byID, err := repo.GetByID(perm.PermID)
	require.NoError(t, err)
	assert.Equal(t, "Dataset View", byID.PermName)

	byKey, err := repo.GetByKey("dataset:view")
	require.NoError(t, err)
	assert.Equal(t, perm.PermID, byKey.PermID)

	updatedDesc := "updated desc"
	perm.PermName = "Dataset Manage"
	perm.PermDesc = &updatedDesc
	perm.Status = permission.StatusDisabled
	require.NoError(t, repo.Update(perm))

	updated, err := repo.GetByID(perm.PermID)
	require.NoError(t, err)
	assert.Equal(t, "Dataset Manage", updated.PermName)
	require.NotNil(t, updated.PermDesc)
	assert.Equal(t, updatedDesc, *updated.PermDesc)
	assert.Equal(t, permission.StatusDisabled, updated.Status)

	require.NoError(t, repo.Delete(perm.PermID))

	_, err = repo.GetByID(perm.PermID)
	require.Error(t, err)
	_, err = repo.GetByKey("dataset:view")
	require.Error(t, err)

	var raw permission.SysPerm
	require.NoError(t, db.Unscoped().Where("perm_id = ?", perm.PermID).First(&raw).Error)
	assert.Equal(t, permission.DelFlagDeleted, raw.DelFlag)
}

func TestPermRepository_ListGetByTypeAndCheckKeyExists(t *testing.T) {
	repo, db := setupPermRepositoryTest(t)
	now := time.Now()
	menu1 := &permission.SysPerm{PermName: "Menu 1", PermKey: "menu:1", PermType: permission.PermTypeMenu, Status: permission.StatusEnabled, CreateTime: now.Add(-2 * time.Minute), DelFlag: permission.DelFlagNormal}
	menu2 := &permission.SysPerm{PermName: "Menu 2", PermKey: "menu:2", PermType: permission.PermTypeMenu, Status: permission.StatusEnabled, CreateTime: now.Add(-1 * time.Minute), DelFlag: permission.DelFlagNormal}
	dataPerm := &permission.SysPerm{PermName: "Data 1", PermKey: "data:1", PermType: permission.PermTypeData, Status: permission.StatusEnabled, CreateTime: now, DelFlag: permission.DelFlagNormal}
	deleted := &permission.SysPerm{PermName: "Deleted", PermKey: "deleted:key", PermType: permission.PermTypeButton, Status: permission.StatusEnabled, CreateTime: now.Add(-3 * time.Minute), DelFlag: permission.DelFlagDeleted}
	seedPerm(t, db, menu1)
	seedPerm(t, db, menu2)
	seedPerm(t, db, dataPerm)
	seedPerm(t, db, deleted)

	perms, err := repo.List()
	require.NoError(t, err)
	require.Len(t, perms, 3)
	assert.Equal(t, dataPerm.PermID, perms[0].PermID)
	assert.Equal(t, menu2.PermID, perms[1].PermID)
	assert.Equal(t, menu1.PermID, perms[2].PermID)

	menuPerms, err := repo.GetByType(permission.PermTypeMenu)
	require.NoError(t, err)
	require.Len(t, menuPerms, 2)
	assert.Equal(t, menu2.PermID, menuPerms[0].PermID)
	assert.Equal(t, menu1.PermID, menuPerms[1].PermID)

	count, err := repo.CheckKeyExists("menu:1", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = repo.CheckKeyExists("menu:1", menu1.PermID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	count, err = repo.CheckKeyExists("deleted:key", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
