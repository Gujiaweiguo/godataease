//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestRoleServiceIntegration_Create(t *testing.T) {
	cleanupTables(&role.SysRole{})

	repo := repository.NewRoleRepository(testDB)
	svc := NewRoleService(repo)

	desc := "Test role description"
	req := &role.RoleCreator{
		Name: "TestRole",
		Desc: &desc,
	}

	id, err := svc.CreateRole(req, "tester")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Verify created
	found, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "TestRole", found.RoleName)
	assert.Contains(t, found.RoleCode, "role_")
	assert.Equal(t, role.StatusEnabled, found.Status)
	assert.NotNil(t, found.DataScope)
	assert.Equal(t, role.DataScopeSelf, *found.DataScope)
	assert.NotNil(t, found.CreateBy)
	assert.Equal(t, "tester", *found.CreateBy)
}

func TestRoleServiceIntegration_Edit(t *testing.T) {
	cleanupTables(&role.SysRole{})

	repo := repository.NewRoleRepository(testDB)
	svc := NewRoleService(repo)

	// Create role
	initialDesc := "Initial description"
	id, _ := svc.CreateRole(&role.RoleCreator{Name: "ToEdit", Desc: &initialDesc}, "creator")

	// Edit
	newDesc := "Updated description"
	err := svc.EditRole(&role.RoleEditor{
		ID:   id,
		Name: "EditedRole",
		Desc: &newDesc,
	}, "editor")
	assert.NoError(t, err)

	// Verify
	updated, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "EditedRole", updated.RoleName)
	assert.NotNil(t, updated.RoleDesc)
	assert.Equal(t, newDesc, *updated.RoleDesc)
	assert.NotNil(t, updated.UpdateBy)
	assert.Equal(t, "editor", *updated.UpdateBy)
	assert.NotNil(t, updated.UpdateTime)
}

func TestRoleServiceIntegration_EditNotFound(t *testing.T) {
	cleanupTables(&role.SysRole{})

	repo := repository.NewRoleRepository(testDB)
	svc := NewRoleService(repo)

	err := svc.EditRole(&role.RoleEditor{ID: 9999, Name: "NotFound"}, "editor")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
}

func TestRoleServiceIntegration_Delete(t *testing.T) {
	cleanupTables(&role.SysRole{})

	repo := repository.NewRoleRepository(testDB)
	svc := NewRoleService(repo)

	// Create role
	id, _ := svc.CreateRole(&role.RoleCreator{Name: "ToDelete"}, "tester")

	// Delete
	err := svc.DeleteRole(id)
	assert.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByID(id)
	assert.Error(t, err)
}

func TestRoleServiceIntegration_GetRoleByID(t *testing.T) {
	cleanupTables(&role.SysRole{})

	repo := repository.NewRoleRepository(testDB)
	svc := NewRoleService(repo)

	// Create role
	id, _ := svc.CreateRole(&role.RoleCreator{
		Name: "DetailRole",
		Desc: pStr("Detailed role"),
	}, "tester")

	// Get by ID
	vo, err := svc.GetRoleByID(id)
	assert.NoError(t, err)
	assert.Equal(t, id, vo.ID)
	assert.Equal(t, "DetailRole", vo.Name)
	assert.NotNil(t, vo.Desc)
	assert.Equal(t, "Detailed role", *vo.Desc)
}

func TestRoleServiceIntegration_GetRoleByIDNotFound(t *testing.T) {
	cleanupTables(&role.SysRole{})

	repo := repository.NewRoleRepository(testDB)
	svc := NewRoleService(repo)

	_, err := svc.GetRoleByID(9999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
}

func TestRoleServiceIntegration_QueryRoles(t *testing.T) {
	cleanupTables(&role.SysRole{})

	repo := repository.NewRoleRepository(testDB)
	svc := NewRoleService(repo)

	// Create multiple roles
	zeroParent := int64(0)
	oneParent := int64(1)

	repo.Create(&role.SysRole{RoleName: "Admin", RoleCode: "admin", Status: role.StatusEnabled, ParentID: &zeroParent})
	repo.Create(&role.SysRole{RoleName: "User", RoleCode: "user", Status: role.StatusEnabled, ParentID: &oneParent})
	repo.Create(&role.SysRole{RoleName: "Disabled", RoleCode: "disabled", Status: role.StatusDisabled})

	// Query with keyword
	keyword := "Admin"
	result, err := svc.QueryRoles(&role.RoleQueryRequest{Keyword: &keyword})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Admin", result[0].Name)
	assert.True(t, result[0].Root) // parent_id = 0

	// Query without keyword (should return at least 2 enabled roles)
	allResult, err := svc.QueryRoles(&role.RoleQueryRequest{})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(allResult), 2)

	// Query with non-matching keyword
	emptyResult, err := svc.QueryRoles(&role.RoleQueryRequest{Keyword: pStr("NonExist")})
	assert.NoError(t, err)
	assert.Len(t, emptyResult, 0)
}

func pStr(v string) *string {
	return &v
}
