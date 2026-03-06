//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/repository"
	"github.com/stretchr/testify/assert"
)

// Helper function to create OrgService with all dependencies
func newTestOrgService(t *testing.T) *OrgService {
	cleanupTables(&org.SysOrg{})

	orgRepo := repository.NewOrgRepository(testDB)
	userRepo := repository.NewUserRepository(testDB)
	roleRepo := repository.NewRoleRepository(testDB)

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	auditSvc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	return NewOrgService(orgRepo, auditSvc, userRepo, roleRepo)
}

func TestOrgServiceIntegration_CreateRoot(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	req := &org.OrgCreateRequest{OrgName: "Root Org"}
	err := svc.CreateOrg(req)
	assert.NoError(t, err)

	orgs, err := orgRepo.List()
	assert.NoError(t, err)
	assert.Len(t, orgs, 1)

	created := orgs[0]
	assert.Equal(t, int64(org.RootParentID), created.ParentID)
	assert.Equal(t, 1, created.Level)
	assert.Equal(t, org.StatusEnabled, created.Status)
}

func TestOrgServiceIntegration_CreateChild(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	// Create parent
	parent := &org.SysOrg{
		OrgName:  "Parent",
		ParentID: org.RootParentID,
		Level:    1,
		Status:   org.StatusEnabled,
		DelFlag:  org.DelFlagNormal,
	}
	err := orgRepo.Create(parent)
	assert.NoError(t, err)

	// Create child
	req := &org.OrgCreateRequest{OrgName: "Child", ParentID: &parent.OrgID}
	err = svc.CreateOrg(req)
	assert.NoError(t, err)

	children, err := orgRepo.ListByParentID(parent.OrgID)
	assert.NoError(t, err)
	assert.Len(t, children, 1)
	assert.Equal(t, 2, children[0].Level)
}

func TestOrgServiceIntegration_CreateDuplicateName(t *testing.T) {
	svc := newTestOrgService(t)

	// Create first org
	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Duplicate"})
	assert.NoError(t, err)

	// Try to create with same name
	err = svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Duplicate"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization name already exists")
}

func TestOrgServiceIntegration_CreateParentNotFound(t *testing.T) {
	svc := newTestOrgService(t)

	notExistParent := int64(9999)
	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Child", ParentID: &notExistParent})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent organization not found")
}

func TestOrgServiceIntegration_Update(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	// Create org
	existing := &org.SysOrg{
		OrgName:  "Old Name",
		ParentID: org.RootParentID,
		Level:    1,
		Status:   org.StatusEnabled,
		DelFlag:  org.DelFlagNormal,
	}
	err := orgRepo.Create(existing)
	assert.NoError(t, err)

	// Update
	desc := "new description"
	err = svc.UpdateOrg(&org.OrgUpdateRequest{
		OrgID:   existing.OrgID,
		OrgName: "New Name",
		OrgDesc: &desc,
	})
	assert.NoError(t, err)

	// Verify
	updated, err := orgRepo.GetByID(existing.OrgID)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.OrgName)
	assert.NotNil(t, updated.OrgDesc)
	assert.Equal(t, desc, *updated.OrgDesc)
	assert.NotNil(t, updated.UpdateTime)
}

func TestOrgServiceIntegration_UpdateNotFound(t *testing.T) {
	svc := newTestOrgService(t)

	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: 9999, OrgName: "New"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization not found")
}

func TestOrgServiceIntegration_UpdateDuplicateName(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	// Create two orgs
	org1 := &org.SysOrg{OrgName: "Org1", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	org2 := &org.SysOrg{OrgName: "Org2", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	orgRepo.Create(org1)
	orgRepo.Create(org2)

	// Try to rename org2 to org1
	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: org2.OrgID, OrgName: "Org1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization name already exists")
}

func TestOrgServiceIntegration_Delete(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	// Create org
	existing := &org.SysOrg{
		OrgName:  "ToDelete",
		ParentID: org.RootParentID,
		Level:    1,
		Status:   org.StatusEnabled,
		DelFlag:  org.DelFlagNormal,
	}
	err := orgRepo.Create(existing)
	assert.NoError(t, err)

	// Delete
	err = svc.DeleteOrg(existing.OrgID, 1, "test-user", "127.0.0.1")
	assert.NoError(t, err)

	// Verify deleted
	_, err = orgRepo.GetByID(existing.OrgID)
	assert.Error(t, err)
}

func TestOrgServiceIntegration_DeleteWithChildren(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	// Create parent and child
	parent := &org.SysOrg{OrgName: "Parent", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	err := orgRepo.Create(parent)
	assert.NoError(t, err)

	child := &org.SysOrg{OrgName: "Child", ParentID: parent.OrgID, Level: 2, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	err = orgRepo.Create(child)
	assert.NoError(t, err)

	// Try to delete parent
	err = svc.DeleteOrg(parent.OrgID, 1, "test-user", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete organization with 1 child organizations")
}

func TestOrgServiceIntegration_GetOrgByID(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	// Create org
	existing := &org.SysOrg{
		OrgName:  "QueryTest",
		ParentID: org.RootParentID,
		Level:    1,
		Status:   org.StatusEnabled,
		DelFlag:  org.DelFlagNormal,
	}
	err := orgRepo.Create(existing)
	assert.NoError(t, err)

	// Get
	found, err := svc.GetOrgByID(existing.OrgID)
	assert.NoError(t, err)
	assert.Equal(t, "QueryTest", found.OrgName)

	// Not found
	_, err = svc.GetOrgByID(9999)
	assert.Error(t, err)
}

func TestOrgServiceIntegration_ListOrgs(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	// Create multiple orgs
	for _, name := range []string{"A", "B", "C"} {
		err := orgRepo.Create(&org.SysOrg{
			OrgName:  name,
			ParentID: org.RootParentID,
			Level:    1,
			Status:   org.StatusEnabled,
			DelFlag:  org.DelFlagNormal,
		})
		assert.NoError(t, err)
	}

	// List
	orgs, err := svc.ListOrgs()
	assert.NoError(t, err)
	assert.Len(t, orgs, 3)
}

func TestOrgServiceIntegration_GetOrgTree(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	// Create tree structure
	root1 := &org.SysOrg{OrgName: "Root1", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	root2 := &org.SysOrg{OrgName: "Root2", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	orgRepo.Create(root1)
	orgRepo.Create(root2)

	child1 := &org.SysOrg{OrgName: "Child1", ParentID: root1.OrgID, Level: 2, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	orgRepo.Create(child1)

	grandchild := &org.SysOrg{OrgName: "GrandChild", ParentID: child1.OrgID, Level: 3, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	orgRepo.Create(grandchild)

	// Get tree
	tree, err := svc.GetOrgTree()
	assert.NoError(t, err)
	assert.Len(t, tree, 2) // 2 roots

	// Find root1 and verify children
	var root1Node *org.OrgTreeNode
	for _, n := range tree {
		if n.OrgID == root1.OrgID {
			root1Node = n
			break
		}
	}
	assert.NotNil(t, root1Node)
	assert.Len(t, root1Node.Children, 1)
	assert.Len(t, root1Node.Children[0].Children, 1) // grandchild
}

func TestOrgServiceIntegration_UpdateStatus(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	// Create org
	existing := &org.SysOrg{
		OrgName:  "StatusTest",
		ParentID: org.RootParentID,
		Level:    1,
		Status:   org.StatusEnabled,
		DelFlag:  org.DelFlagNormal,
	}
	err := orgRepo.Create(existing)
	assert.NoError(t, err)

	// Update status
	err = svc.UpdateOrgStatus(existing.OrgID, org.StatusDisabled)
	assert.NoError(t, err)

	// Verify
	updated, err := orgRepo.GetByID(existing.OrgID)
	assert.NoError(t, err)
	assert.Equal(t, org.StatusDisabled, updated.Status)
	assert.NotNil(t, updated.UpdateTime)
}

func TestOrgServiceIntegration_CheckOrgNameExists(t *testing.T) {
	svc := newTestOrgService(t)
	orgRepo := repository.NewOrgRepository(testDB)

	// Create org
	err := orgRepo.Create(&org.SysOrg{
		OrgName:  "Existing",
		ParentID: org.RootParentID,
		Level:    1,
		Status:   org.StatusEnabled,
		DelFlag:  org.DelFlagNormal,
	})
	assert.NoError(t, err)

	// Check existing
	exists, err := svc.CheckOrgNameExists("Existing")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existing
	notExists, err := svc.CheckOrgNameExists("NotExisting")
	assert.NoError(t, err)
	assert.False(t, notExists)
}
