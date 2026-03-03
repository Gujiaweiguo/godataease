//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestPermServiceIntegration_CreatePerm_DefaultTypeAndStatus(t *testing.T) {
	cleanupTables(&permission.SysPerm{})

	repo := repository.NewPermRepository(testDB)
	svc := NewPermService(repo)

	permID, err := svc.CreatePerm(&permission.PermCreateRequest{
		PermName: "View Dashboard",
		PermKey:  "dashboard:view",
	})
	assert.NoError(t, err)
	assert.Greater(t, permID, int64(0))

	created, err := repo.GetByID(permID)
	assert.NoError(t, err)
	assert.Equal(t, permission.PermTypeMenu, created.PermType)
	assert.Equal(t, permission.StatusEnabled, created.Status)
	assert.Equal(t, permission.DelFlagNormal, created.DelFlag)
}

func TestPermServiceIntegration_CreatePerm_DuplicateKey(t *testing.T) {
	cleanupTables(&permission.SysPerm{})

	repo := repository.NewPermRepository(testDB)
	svc := NewPermService(repo)

	_, err := svc.CreatePerm(&permission.PermCreateRequest{
		PermName: "First",
		PermKey:  "perm:duplicate",
	})
	assert.NoError(t, err)

	_, err = svc.CreatePerm(&permission.PermCreateRequest{
		PermName: "Second",
		PermKey:  "perm:duplicate",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestPermServiceIntegration_UpdatePerm_Success(t *testing.T) {
	cleanupTables(&permission.SysPerm{})

	repo := repository.NewPermRepository(testDB)
	svc := NewPermService(repo)

	firstID, err := svc.CreatePerm(&permission.PermCreateRequest{PermName: "Old Name", PermKey: "perm:update:old"})
	assert.NoError(t, err)

	desc := "updated description"
	disabled := permission.StatusDisabled
	err = svc.UpdatePerm(&permission.PermUpdateRequest{
		PermID:   firstID,
		PermName: "New Name",
		PermKey:  "perm:update:new",
		PermType: permission.PermTypeButton,
		PermDesc: &desc,
		Status:   &disabled,
	})
	assert.NoError(t, err)

	updated, err := repo.GetByID(firstID)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.PermName)
	assert.Equal(t, "perm:update:new", updated.PermKey)
	assert.Equal(t, permission.PermTypeButton, updated.PermType)
	assert.Equal(t, disabled, updated.Status)
	if assert.NotNil(t, updated.PermDesc) {
		assert.Equal(t, desc, *updated.PermDesc)
	}
	assert.NotNil(t, updated.UpdateTime)
}

func TestPermServiceIntegration_UpdatePerm_DuplicateKey(t *testing.T) {
	cleanupTables(&permission.SysPerm{})

	repo := repository.NewPermRepository(testDB)
	svc := NewPermService(repo)

	firstID, err := svc.CreatePerm(&permission.PermCreateRequest{PermName: "P1", PermKey: "perm:k1"})
	assert.NoError(t, err)

	_, err = svc.CreatePerm(&permission.PermCreateRequest{PermName: "P2", PermKey: "perm:k2"})
	assert.NoError(t, err)

	err = svc.UpdatePerm(&permission.PermUpdateRequest{PermID: firstID, PermKey: "perm:k2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestPermServiceIntegration_DeletePerm_SoftDelete(t *testing.T) {
	cleanupTables(&permission.SysPerm{})

	repo := repository.NewPermRepository(testDB)
	svc := NewPermService(repo)

	permID, err := svc.CreatePerm(&permission.PermCreateRequest{PermName: "Delete", PermKey: "perm:delete"})
	assert.NoError(t, err)

	err = svc.DeletePerm(permID)
	assert.NoError(t, err)

	_, err = svc.GetPermByID(permID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestPermServiceIntegration_ListPerms_PaginationAndDefaults(t *testing.T) {
	cleanupTables(&permission.SysPerm{})

	repo := repository.NewPermRepository(testDB)
	svc := NewPermService(repo)

	for i := 1; i <= 12; i++ {
		_, err := svc.CreatePerm(&permission.PermCreateRequest{
			PermName: "perm",
			PermKey:  "perm:list:" + string(rune('a'+i)),
		})
		assert.NoError(t, err)
	}

	page1, err := svc.ListPerms(&permission.PermQueryRequest{Current: 1, Size: 5})
	assert.NoError(t, err)
	assert.Equal(t, int64(12), page1.Total)
	assert.Equal(t, 1, page1.Current)
	assert.Equal(t, 5, page1.Size)
	assert.Len(t, page1.List.([]*permission.SysPerm), 5)

	defaults, err := svc.ListPerms(&permission.PermQueryRequest{Current: 0, Size: 0})
	assert.NoError(t, err)
	assert.Equal(t, 1, defaults.Current)
	assert.Equal(t, 10, defaults.Size)
	assert.Len(t, defaults.List.([]*permission.SysPerm), 10)
}

func TestPermServiceIntegration_CheckPermKeyExists(t *testing.T) {
	cleanupTables(&permission.SysPerm{})

	repo := repository.NewPermRepository(testDB)
	svc := NewPermService(repo)

	_, err := svc.CreatePerm(&permission.PermCreateRequest{PermName: "Check", PermKey: "perm:exists"})
	assert.NoError(t, err)

	exists, err := svc.CheckPermKeyExists("perm:exists")
	assert.NoError(t, err)
	assert.True(t, exists)

	notExists, err := svc.CheckPermKeyExists("perm:not:exists")
	assert.NoError(t, err)
	assert.False(t, notExists)
}
