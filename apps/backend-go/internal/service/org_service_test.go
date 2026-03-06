package service

import (
	"strings"
	"testing"

	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockOrgRepository struct {
	repo *repository.OrgRepository
	db   *gorm.DB
}

func setupOrgService(t *testing.T) (*OrgService, *mockOrgRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&org.SysOrg{}); err != nil {
		t.Fatalf("migrate sys_org failed: %v", err)
	}

	orgRepo := repository.NewOrgRepository(db)
	// Pass nil for auditSvc, userRepo, roleRepo as they are optional for most tests
	return NewOrgService(orgRepo, nil, nil, nil), &mockOrgRepository{repo: orgRepo, db: db}
}

func createSeedOrg(t *testing.T, mockRepo *mockOrgRepository, name string, parentID int64, level int) *org.SysOrg {
	t.Helper()

	o := &org.SysOrg{
		OrgName:  name,
		ParentID: parentID,
		Level:    level,
		Status:   org.StatusEnabled,
		DelFlag:  org.DelFlagNormal,
	}
	if err := mockRepo.repo.Create(o); err != nil {
		t.Fatalf("seed org failed: %v", err)
	}

	return o
}

func TestOrgCreateOrg_SuccessRoot(t *testing.T) {
	svc, mockRepo := setupOrgService(t)

	req := &org.OrgCreateRequest{OrgName: "Root Org"}
	if err := svc.CreateOrg(req); err != nil {
		t.Fatalf("CreateOrg failed: %v", err)
	}

	orgs, err := mockRepo.repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(orgs) != 1 {
		t.Fatalf("expected 1 org, got %d", len(orgs))
	}

	created := orgs[0]
	if created.ParentID != org.RootParentID {
		t.Fatalf("expected root parent %d, got %d", org.RootParentID, created.ParentID)
	}
	if created.Level != 1 {
		t.Fatalf("expected level 1, got %d", created.Level)
	}
	if created.Status != org.StatusEnabled {
		t.Fatalf("expected status enabled, got %d", created.Status)
	}
}

func TestOrgCreateOrg_SuccessChild(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	parent := createSeedOrg(t, mockRepo, "Parent", org.RootParentID, 1)

	req := &org.OrgCreateRequest{OrgName: "Child", ParentID: &parent.OrgID}
	if err := svc.CreateOrg(req); err != nil {
		t.Fatalf("CreateOrg child failed: %v", err)
	}

	children, err := mockRepo.repo.ListByParentID(parent.OrgID)
	if err != nil {
		t.Fatalf("ListByParentID failed: %v", err)
	}
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0].Level != parent.Level+1 {
		t.Fatalf("expected level %d, got %d", parent.Level+1, children[0].Level)
	}
}

func TestOrgCreateOrg_DuplicateName(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	createSeedOrg(t, mockRepo, "Dup", org.RootParentID, 1)

	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Dup"})
	if err == nil {
		t.Fatal("expected duplicate name error, got nil")
	}
	if !strings.Contains(err.Error(), "organization name already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgCreateOrg_ParentNotFound(t *testing.T) {
	svc, _ := setupOrgService(t)
	notExistParent := int64(999)

	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Child", ParentID: &notExistParent})
	if err == nil {
		t.Fatal("expected parent not found error, got nil")
	}
	if !strings.Contains(err.Error(), "parent organization not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgUpdateOrg_Success(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "Old", org.RootParentID, 1)
	desc := "new desc"

	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: existing.OrgID, OrgName: "New", OrgDesc: &desc})
	if err != nil {
		t.Fatalf("UpdateOrg failed: %v", err)
	}

	updated, err := mockRepo.repo.GetByID(existing.OrgID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if updated.OrgName != "New" {
		t.Fatalf("expected org name New, got %s", updated.OrgName)
	}
	if updated.OrgDesc == nil || *updated.OrgDesc != desc {
		t.Fatalf("expected org desc %s", desc)
	}
	if updated.UpdateTime == nil {
		t.Fatal("expected update time to be set")
	}
}

func TestOrgUpdateOrg_NotFound(t *testing.T) {
	svc, _ := setupOrgService(t)
	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: 1000, OrgName: "New"})
	if err == nil {
		t.Fatal("expected organization not found error, got nil")
	}
	if !strings.Contains(err.Error(), "organization not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgUpdateOrg_DuplicateName(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	target := createSeedOrg(t, mockRepo, "Target", org.RootParentID, 1)
	createSeedOrg(t, mockRepo, "Duplicated", org.RootParentID, 1)

	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: target.OrgID, OrgName: "Duplicated"})
	if err == nil {
		t.Fatal("expected duplicate name error, got nil")
	}
	if !strings.Contains(err.Error(), "organization name already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgDeleteOrg_Success(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "ToDelete", org.RootParentID, 1)

	err := svc.DeleteOrg(existing.OrgID, 1, "test-user", "127.0.0.1")
	if err != nil {
		t.Fatalf("DeleteOrg failed: %v", err)
	}

	_, err = mockRepo.repo.GetByID(existing.OrgID)
	if err == nil {
		t.Fatal("expected deleted org not found")
	}
}

func TestOrgDeleteOrg_HasChildren(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	parent := createSeedOrg(t, mockRepo, "Parent", org.RootParentID, 1)
	createSeedOrg(t, mockRepo, "Child", parent.OrgID, 2)

	err := svc.DeleteOrg(parent.OrgID, 1, "test-user", "127.0.0.1")
	if err == nil {
		t.Fatal("expected children exists error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot delete organization with") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgGetOrgByID_SuccessAndFail(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "QueryOrg", org.RootParentID, 1)

	found, err := svc.GetOrgByID(existing.OrgID)
	if err != nil {
		t.Fatalf("GetOrgByID failed: %v", err)
	}
	if found.OrgName != "QueryOrg" {
		t.Fatalf("expected QueryOrg, got %s", found.OrgName)
	}

	_, err = svc.GetOrgByID(404)
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestOrgListOrgs(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	createSeedOrg(t, mockRepo, "A", org.RootParentID, 1)
	createSeedOrg(t, mockRepo, "B", org.RootParentID, 1)

	orgs, err := svc.ListOrgs()
	if err != nil {
		t.Fatalf("ListOrgs failed: %v", err)
	}
	if len(orgs) != 2 {
		t.Fatalf("expected 2 orgs, got %d", len(orgs))
	}
}

func TestOrgListByParentID(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	parent := createSeedOrg(t, mockRepo, "Parent", org.RootParentID, 1)
	createSeedOrg(t, mockRepo, "C1", parent.OrgID, 2)
	createSeedOrg(t, mockRepo, "C2", parent.OrgID, 2)

	orgs, err := svc.ListByParentID(parent.OrgID)
	if err != nil {
		t.Fatalf("ListByParentID failed: %v", err)
	}
	if len(orgs) != 2 {
		t.Fatalf("expected 2 orgs, got %d", len(orgs))
	}
}

func TestOrgGetOrgTree_MultiLevel(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	root1 := createSeedOrg(t, mockRepo, "Root1", org.RootParentID, 1)
	root2 := createSeedOrg(t, mockRepo, "Root2", org.RootParentID, 1)
	child := createSeedOrg(t, mockRepo, "Child1", root1.OrgID, 2)
	createSeedOrg(t, mockRepo, "GrandChild", child.OrgID, 3)

	tree, err := svc.GetOrgTree()
	if err != nil {
		t.Fatalf("GetOrgTree failed: %v", err)
	}
	if len(tree) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(tree))
	}

	var root1Node *org.OrgTreeNode
	for _, n := range tree {
		if n.OrgID == root1.OrgID {
			root1Node = n
			break
		}
	}
	if root1Node == nil {
		t.Fatal("root1 not found in tree")
	}
	if len(root1Node.Children) != 1 {
		t.Fatalf("expected root1 has 1 child, got %d", len(root1Node.Children))
	}
	if len(root1Node.Children[0].Children) != 1 {
		t.Fatalf("expected child has 1 grandchild, got %d", len(root1Node.Children[0].Children))
	}

	_ = root2
}

func TestOrgUpdateOrgStatus(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "NeedDisable", org.RootParentID, 1)

	err := svc.UpdateOrgStatus(existing.OrgID, org.StatusDisabled)
	if err != nil {
		t.Fatalf("UpdateOrgStatus failed: %v", err)
	}

	updated, err := mockRepo.repo.GetByID(existing.OrgID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if updated.Status != org.StatusDisabled {
		t.Fatalf("expected status disabled, got %d", updated.Status)
	}
	if updated.UpdateTime == nil {
		t.Fatal("expected update time to be set")
	}
}

func TestOrgCheckOrgNameExists(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	createSeedOrg(t, mockRepo, "ExistMe", org.RootParentID, 1)

	exists, err := svc.CheckOrgNameExists("ExistMe")
	if err != nil {
		t.Fatalf("CheckOrgNameExists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected name to exist")
	}

	notExists, err := svc.CheckOrgNameExists("Nope")
	if err != nil {
		t.Fatalf("CheckOrgNameExists failed: %v", err)
	}
	if notExists {
		t.Fatal("expected name to not exist")
	}
}

func TestOrgGetOrgTree_ListFailed(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	if err := mockRepo.db.Exec("DROP TABLE sys_org").Error; err != nil {
		t.Fatalf("drop table failed: %v", err)
	}

	_, err := svc.GetOrgTree()
	if err == nil {
		t.Fatal("expected list organizations error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to list organizations") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgDeleteOrg_CountChildrenFailed(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	if err := mockRepo.db.Exec("DROP TABLE sys_org").Error; err != nil {
		t.Fatalf("drop table failed: %v", err)
	}

	err := svc.DeleteOrg(1, 1, "test-user", "127.0.0.1")
	if err == nil {
		t.Fatal("expected database error, got nil")
	}
	// After adding audit logging, the first operation is GetByID which fails with "organization not found"
	if !strings.Contains(err.Error(), "organization not found") && !strings.Contains(err.Error(), "failed to check children") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgCheckOrgNameExists_CheckFailed(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	if err := mockRepo.db.Exec("DROP TABLE sys_org").Error; err != nil {
		t.Fatalf("drop table failed: %v", err)
	}

	_, err := svc.CheckOrgNameExists("Any")
	if err == nil {
		t.Fatal("expected check org name error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to check org name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgUpdateOrgStatus_NotFound(t *testing.T) {
	svc, _ := setupOrgService(t)
	err := svc.UpdateOrgStatus(1234, org.StatusDisabled)
	if err == nil {
		t.Fatal("expected organization not found error, got nil")
	}
	if !strings.Contains(err.Error(), "organization not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgCreateOrg_CheckNameFailed(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	if err := mockRepo.db.Exec("DROP TABLE sys_org").Error; err != nil {
		t.Fatalf("drop table failed: %v", err)
	}

	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Any"})
	if err == nil {
		t.Fatal("expected check name error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to check org name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgCreateOrg_CreateFailed(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	if err := mockRepo.db.Exec("CREATE TRIGGER deny_org_insert BEFORE INSERT ON sys_org BEGIN SELECT RAISE(FAIL, 'deny insert'); END;").Error; err != nil {
		t.Fatalf("create trigger failed: %v", err)
	}

	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Any"})
	if err == nil {
		t.Fatal("expected create organization error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create organization") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgUpdateOrg_UpdateFailed(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "NeedUpdate", org.RootParentID, 1)
	if err := mockRepo.db.Exec("CREATE TRIGGER deny_org_update BEFORE UPDATE ON sys_org BEGIN SELECT RAISE(FAIL, 'deny update'); END;").Error; err != nil {
		t.Fatalf("create trigger failed: %v", err)
	}

	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: existing.OrgID, OrgDesc: strPtr("desc")})
	if err == nil {
		t.Fatal("expected update organization error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update organization") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgDeleteOrg_DeleteFailed(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "NeedDelete", org.RootParentID, 1)
	if err := mockRepo.db.Exec("CREATE TRIGGER deny_org_delete BEFORE UPDATE ON sys_org BEGIN SELECT RAISE(FAIL, 'deny delete'); END;").Error; err != nil {
		t.Fatalf("create trigger failed: %v", err)
	}

	err := svc.DeleteOrg(existing.OrgID, 1, "test-user", "127.0.0.1")
	if err == nil {
		t.Fatal("expected delete organization error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to delete organization") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgUpdateOrgStatus_UpdateFailed(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "NeedStatus", org.RootParentID, 1)
	if err := mockRepo.db.Exec("CREATE TRIGGER deny_org_status_update BEFORE UPDATE ON sys_org BEGIN SELECT RAISE(FAIL, 'deny status update'); END;").Error; err != nil {
		t.Fatalf("create trigger failed: %v", err)
	}

	err := svc.UpdateOrgStatus(existing.OrgID, org.StatusDisabled)
	if err == nil {
		t.Fatal("expected update organization status error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update organization status") {
		t.Fatalf("unexpected error: %v", err)
	}
}
