package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoleServiceTest(t *testing.T) (*RoleService, *repository.RoleRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}))

	repo := repository.NewRoleRepository(db)
	return NewRoleService(repo, nil, nil), repo
}

func seedRole(t *testing.T, repo *repository.RoleRepository, name, code string, roleType *string, createTime time.Time) {
	t.Helper()

	require.NoError(t, repo.Create(&role.SysRole{
		RoleName:   name,
		RoleCode:   code,
		RoleType:   roleType,
		Status:     role.StatusEnabled,
		CreateTime: &createTime,
	}))
}

func TestRoleService_QueryRolesPage_DefaultPaging(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	systemType := "system"
	customType := "custom"

	seedRole(t, repo, "Admin", "admin", &systemType, time.Unix(300, 0))
	seedRole(t, repo, "Viewer", "viewer", &customType, time.Unix(200, 0))
	seedRole(t, repo, "Operator", "operator", &systemType, time.Unix(100, 0))

	result, err := svc.QueryRolesPage(&role.RolePageRequest{Current: 0, Size: 0})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(3), result.Total)
	assert.Equal(t, 1, result.Current)
	assert.Equal(t, 10, result.Size)
	assert.Len(t, result.List, 3)
	assert.ElementsMatch(t, []string{"Admin", "Viewer", "Operator"}, []string{result.List[0].Name, result.List[1].Name, result.List[2].Name})
	for _, item := range result.List {
		require.NotNil(t, item.RoleType)
	}
}

func TestRoleService_QueryRolesPage_WithKeywordAndRoleType(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	systemType := "system"
	customType := "custom"

	seedRole(t, repo, "Admin Root", "admin-root", &systemType, time.Unix(300, 0))
	seedRole(t, repo, "Admin Custom", "admin-custom", &customType, time.Unix(200, 0))
	seedRole(t, repo, "Viewer", "viewer", &systemType, time.Unix(100, 0))

	keyword := "Admin"
	result, err := svc.QueryRolesPage(&role.RolePageRequest{Keyword: &keyword, Current: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.List, 2)

	result, err = svc.QueryRolesPage(&role.RolePageRequest{Keyword: &keyword, RoleType: &systemType, Current: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Len(t, result.List, 1)
	assert.Equal(t, "Admin Root", result.List[0].Name)
	require.NotNil(t, result.List[0].RoleType)
	assert.Equal(t, systemType, *result.List[0].RoleType)
}

func TestRoleService_QueryRolesPage_OutOfRangePage(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	systemType := "system"

	seedRole(t, repo, "Admin", "admin", &systemType, time.Unix(300, 0))
	seedRole(t, repo, "Viewer", "viewer", &systemType, time.Unix(200, 0))

	result, err := svc.QueryRolesPage(&role.RolePageRequest{Current: 3, Size: 1})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 3, result.Current)
	assert.Equal(t, 1, result.Size)
	assert.Len(t, result.List, 0)
}
