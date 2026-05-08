package service

import (
	"strings"
	"testing"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/governance"
	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	if err := db.AutoMigrate(&org.SysOrg{}, &role.SysRole{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}); err != nil {
		t.Fatalf("migrate sys_org failed: %v", err)
	}

	orgRepo := repository.NewOrgRepository(db)
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	governancePolicySvc := NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), nil)
	return NewOrgService(orgRepo, nil, userRepo, roleRepo, userRoleRepo, governancePolicySvc), &mockOrgRepository{repo: orgRepo, db: db}
}

func setupOrgServiceWithAudit(t *testing.T) (*OrgService, *mockOrgRepository, *repository.AuditLogRepository, *repository.UserRepository, *repository.UserRoleRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&org.SysOrg{}, &audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, &user.SysUser{}, &user.SysUserRole{}, &role.SysRole{}, &governance.SysGovernancePolicy{}); err != nil {
		t.Fatalf("migrate audit-aware org fixtures failed: %v", err)
	}

	orgRepo := repository.NewOrgRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	loginFailureRepo := repository.NewLoginFailureRepository(db)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(db)
	auditSvc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	governancePolicySvc := NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), auditSvc)

	return NewOrgService(orgRepo, auditSvc, userRepo, roleRepo, userRoleRepo, governancePolicySvc), &mockOrgRepository{repo: orgRepo, db: db}, auditLogRepo, userRepo, userRoleRepo
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
	if err := svc.CreateOrg(req, 1); err != nil {
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
	if err := svc.CreateOrg(req, 1); err != nil {
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

	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Dup"}, 1)
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

	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Child", ParentID: &notExistParent}, 1)
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

	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: existing.OrgID, OrgName: "New", OrgDesc: &desc}, 1)
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
	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: 1000, OrgName: "New"}, 1)
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

	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: target.OrgID, OrgName: "Duplicated"}, 1)
	if err == nil {
		t.Fatal("expected duplicate name error, got nil")
	}
	if !strings.Contains(err.Error(), "organization name already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgUpdateOrg_UnchangedNameSkipsDuplicateCheckAndNilDescKeepsExisting(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	desc := "old desc"
	existing := createSeedOrg(t, mockRepo, "SameName", org.RootParentID, 1)
	existing.OrgDesc = &desc
	if err := mockRepo.repo.Update(existing); err != nil {
		t.Fatalf("seed update failed: %v", err)
	}
	createSeedOrg(t, mockRepo, "OtherOrg", org.RootParentID, 1)

	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: existing.OrgID, OrgName: "SameName", OrgDesc: nil}, 1)
	if err != nil {
		t.Fatalf("UpdateOrg failed: %v", err)
	}

	updated, err := mockRepo.repo.GetByID(existing.OrgID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if updated.OrgName != "SameName" {
		t.Fatalf("expected unchanged org name, got %s", updated.OrgName)
	}
	if updated.OrgDesc == nil || *updated.OrgDesc != desc {
		t.Fatalf("expected existing description to be preserved, got %#v", updated.OrgDesc)
	}
	if updated.UpdateTime == nil {
		t.Fatal("expected update time to be set")
	}
}

func TestOrgUpdateOrg_MoveToNewParent(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	parent := createSeedOrg(t, mockRepo, "Parent Org", org.RootParentID, 1)
	child := createSeedOrg(t, mockRepo, "Child Org", org.RootParentID, 1)

	newParentID := parent.OrgID
	require.NoError(t, svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: child.OrgID, ParentID: &newParentID}, 1))

	updated, err := mockRepo.repo.GetByID(child.OrgID)
	require.NoError(t, err)
	assert.Equal(t, parent.OrgID, updated.ParentID)
	assert.Equal(t, parent.Level+1, updated.Level)
}

func TestOrgUpdateOrg_MoveToRoot(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	parent := createSeedOrg(t, mockRepo, "Parent Org", org.RootParentID, 1)
	child := createSeedOrg(t, mockRepo, "Child Org", parent.OrgID, parent.Level+1)

	rootID := int64(0)
	require.NoError(t, svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: child.OrgID, ParentID: &rootID}, 1))

	updated, err := mockRepo.repo.GetByID(child.OrgID)
	require.NoError(t, err)
	assert.Equal(t, org.RootParentID, updated.ParentID)
	assert.Equal(t, 1, updated.Level)
}

func TestOrgUpdateOrg_RejectSelfAsParent(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "Self Parent", org.RootParentID, 1)

	selfID := existing.OrgID
	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: existing.OrgID, ParentID: &selfID}, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be its own parent")
}

func TestOrgUpdateOrg_RejectDescendantAsParent(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	parent := createSeedOrg(t, mockRepo, "Grandparent", org.RootParentID, 1)
	child := createSeedOrg(t, mockRepo, "Child", parent.OrgID, parent.Level+1)
	grandchild := createSeedOrg(t, mockRepo, "Grandchild", child.OrgID, child.Level+1)

	descendantID := grandchild.OrgID
	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: parent.OrgID, ParentID: &descendantID}, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "own descendant")
}

func TestOrgUpdateOrg_RejectNonexistentParent(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "MoveToMissing", org.RootParentID, 1)

	missingID := int64(9999)
	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: existing.OrgID, ParentID: &missingID}, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent organization not found")
}

func TestOrgDeleteOrg_Success(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "ToDelete", org.RootParentID, 1)

	err := svc.DeleteOrg(existing.OrgID, 1, 1, "test-user", "127.0.0.1")
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

	err := svc.DeleteOrg(parent.OrgID, 1, 1, "test-user", "127.0.0.1")
	if err == nil {
		t.Fatal("expected children exists error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot delete organization with") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgDeleteOrg_HasChildren_RecordsFailedAuditLog(t *testing.T) {
	svc, mockRepo, auditLogRepo, _, _ := setupOrgServiceWithAudit(t)
	parent := createSeedOrg(t, mockRepo, "Parent", org.RootParentID, 1)
	createSeedOrg(t, mockRepo, "Child", parent.OrgID, 2)

	err := svc.DeleteOrg(parent.OrgID, 7, 7, "operator", "127.0.0.1")
	if err == nil {
		t.Fatal("expected children exists error, got nil")
	}

	logs, total, logErr := auditLogRepo.Query(&audit.AuditLogQuery{Page: 1, PageSize: 10})
	if logErr != nil {
		t.Fatalf("query audit logs failed: %v", logErr)
	}
	if total != 1 || len(logs) != 1 {
		t.Fatalf("expected one audit log, got total=%d logs=%d", total, len(logs))
	}
	log := logs[0]
	if log.Status != audit.StatusFailed || log.ActionName != "删除组织" || log.Operation != audit.OperationDelete {
		t.Fatalf("unexpected failed audit log: %+v", log)
	}
	if log.ResourceName == nil || *log.ResourceName != "Parent" {
		t.Fatalf("expected resource name Parent, got %+v", log.ResourceName)
	}
	if log.OrganizationID == nil || *log.OrganizationID != parent.OrgID {
		t.Fatalf("expected organization id %d, got %+v", parent.OrgID, log.OrganizationID)
	}
	if log.FailureReason == nil || !strings.Contains(*log.FailureReason, "无法删除") {
		t.Fatalf("expected failure reason to mention child orgs, got %+v", log.FailureReason)
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

func TestOrgListOrgs_Empty(t *testing.T) {
	svc, _ := setupOrgService(t)

	orgs, err := svc.ListOrgs()
	if err != nil {
		t.Fatalf("ListOrgs failed: %v", err)
	}
	if orgs == nil || len(orgs) != 0 {
		t.Fatalf("expected empty non-nil org slice, got %#v", orgs)
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

func TestOrgListByParentID_Empty(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	parent := createSeedOrg(t, mockRepo, "Parent", org.RootParentID, 1)

	orgs, err := svc.ListByParentID(parent.OrgID)
	if err != nil {
		t.Fatalf("ListByParentID failed: %v", err)
	}
	if orgs == nil || len(orgs) != 0 {
		t.Fatalf("expected empty non-nil child slice, got %#v", orgs)
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

func TestOrgGetOrgTree_OrphanNodeSkipped(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	createSeedOrg(t, mockRepo, "Root", org.RootParentID, 1)
	createSeedOrg(t, mockRepo, "Orphan", 99999, 2)

	tree, err := svc.GetOrgTree()
	if err != nil {
		t.Fatalf("GetOrgTree failed: %v", err)
	}
	if len(tree) != 1 {
		t.Fatalf("expected only rooted node in tree, got %d", len(tree))
	}
	if len(tree[0].Children) != 0 {
		t.Fatalf("expected orphan to be skipped, got children %+v", tree[0].Children)
	}
}

func TestOrgGetOrgTree_OnlyOrphansReturnsEmptyRoots(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	createSeedOrg(t, mockRepo, "Orphan1", 99998, 2)
	createSeedOrg(t, mockRepo, "Orphan2", 99999, 2)

	tree, err := svc.GetOrgTree()
	if err != nil {
		t.Fatalf("GetOrgTree failed: %v", err)
	}
	if len(tree) != 0 {
		t.Fatalf("expected no rooted nodes, got %#v", tree)
	}
}

func TestOrgUpdateOrgStatus(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "NeedDisable", org.RootParentID, 1)

	err := svc.UpdateOrgStatus(existing.OrgID, org.StatusDisabled, 1)
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

func TestOrgDeleteOrg_NotFound(t *testing.T) {
	svc, _ := setupOrgService(t)
	err := svc.DeleteOrg(404, 1, 1, "test-user", "127.0.0.1")
	if err == nil {
		t.Fatal("expected organization not found error")
	}
	if !strings.Contains(err.Error(), "organization not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgDeleteOrg_Success_WithAffectedUsers_RecordsSuccessAuditLog(t *testing.T) {
	svc, mockRepo, auditLogRepo, userRepo, userRoleRepo := setupOrgServiceWithAudit(t)
	existing := createSeedOrg(t, mockRepo, "AuditedDelete", org.RootParentID, 1)

	u := &user.SysUser{Username: "user-a", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}
	if err := userRepo.Create(u); err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	if err := userRoleRepo.Create(&user.SysUserRole{UserID: u.UserID, RoleID: 1, OrgID: existing.OrgID}); err != nil {
		t.Fatalf("seed user role failed: %v", err)
	}

	err := svc.DeleteOrg(existing.OrgID, 9, 9, "operator", "127.0.0.1")
	if err != nil {
		t.Fatalf("DeleteOrg failed: %v", err)
	}

	logs, total, logErr := auditLogRepo.Query(&audit.AuditLogQuery{Page: 1, PageSize: 10})
	if logErr != nil {
		t.Fatalf("query audit logs failed: %v", logErr)
	}
	if total != 1 || len(logs) != 1 {
		t.Fatalf("expected one audit log, got total=%d logs=%d", total, len(logs))
	}
	log := logs[0]
	if log.Status != audit.StatusSuccess {
		t.Fatalf("expected success audit log, got %+v", log)
	}
	if log.OrganizationID == nil || *log.OrganizationID != existing.OrgID {
		t.Fatalf("expected organization id %d, got %+v", existing.OrgID, log.OrganizationID)
	}
	if log.AfterValue == nil || !strings.Contains(*log.AfterValue, "disposition=soft-delete") || !strings.Contains(*log.AfterValue, "affected_users=1") {
		t.Fatalf("expected affected users in after value, got %+v", log.AfterValue)
	}
	if log.BeforeValue == nil || !strings.Contains(*log.BeforeValue, "AuditedDelete") {
		t.Fatalf("expected before value to mention org name, got %+v", log.BeforeValue)
	}
}

func TestOrgDeleteOrg_Success_UserCountLookupFailureStillDeletes(t *testing.T) {
	svc, mockRepo, auditLogRepo, _, _ := setupOrgServiceWithAudit(t)
	existing := createSeedOrg(t, mockRepo, "CountErrorDelete", org.RootParentID, 1)
	if err := mockRepo.db.Exec("DROP TABLE sys_user_role").Error; err != nil {
		t.Fatalf("drop user role table failed: %v", err)
	}

	err := svc.DeleteOrg(existing.OrgID, 10, 10, "operator", "127.0.0.1")
	if err != nil {
		t.Fatalf("DeleteOrg should ignore user count lookup error, got %v", err)
	}

	_, err = mockRepo.repo.GetByID(existing.OrgID)
	if err == nil {
		t.Fatal("expected org to be deleted despite user count lookup failure")
	}

	logs, total, logErr := auditLogRepo.Query(&audit.AuditLogQuery{Page: 1, PageSize: 10})
	if logErr != nil {
		t.Fatalf("query audit logs failed: %v", logErr)
	}
	if total != 1 || len(logs) != 1 {
		t.Fatalf("expected one audit log, got total=%d logs=%d", total, len(logs))
	}
	if logs[0].AfterValue == nil || !strings.Contains(*logs[0].AfterValue, "affected_users=unknown") {
		t.Fatalf("expected unknown affected user count in after value, got %+v", logs[0].AfterValue)
	}
}

func TestOrgDeleteOrg_Success_AuditFailureDoesNotBlockDelete(t *testing.T) {
	svc, mockRepo, _, _, _ := setupOrgServiceWithAudit(t)
	existing := createSeedOrg(t, mockRepo, "AuditFailureDelete", org.RootParentID, 1)
	if err := mockRepo.db.Exec("CREATE TRIGGER deny_audit_insert BEFORE INSERT ON de_audit_log BEGIN SELECT RAISE(FAIL, 'deny audit insert'); END;").Error; err != nil {
		t.Fatalf("create audit trigger failed: %v", err)
	}

	err := svc.DeleteOrg(existing.OrgID, 11, 11, "operator", "127.0.0.1")
	if err != nil {
		t.Fatalf("DeleteOrg should ignore audit failure, got %v", err)
	}

	_, err = mockRepo.repo.GetByID(existing.OrgID)
	if err == nil {
		t.Fatal("expected org to be deleted even when audit insert fails")
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

	err := svc.DeleteOrg(1, 1, 1, "test-user", "127.0.0.1")
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
	err := svc.UpdateOrgStatus(1234, org.StatusDisabled, 1)
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

	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Any"}, 1)
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

	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Any"}, 1)
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

	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: existing.OrgID, OrgDesc: strPtr("desc")}, 1)
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

	err := svc.DeleteOrg(existing.OrgID, 1, 1, "test-user", "127.0.0.1")
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

	err := svc.UpdateOrgStatus(existing.OrgID, org.StatusDisabled, 1)
	if err == nil {
		t.Fatal("expected update organization status error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update organization status") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgPtrHelpers(t *testing.T) {
	if got := ptrStr("hello"); got == nil || *got != "hello" {
		t.Fatalf("expected ptrStr to return pointer to value, got %#v", got)
	}
	if got := ptrStatus(audit.StatusSuccess); got == nil || *got != audit.StatusSuccess {
		t.Fatalf("expected ptrStatus to return pointer to status, got %#v", got)
	}
}

func TestOrgService_RejectsInvalidOrgContext(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	existing := createSeedOrg(t, mockRepo, "Scoped", org.RootParentID, 1)

	err := svc.CreateOrg(&org.OrgCreateRequest{OrgName: "Child"}, 0)
	require.ErrorIs(t, err, ErrInvalidOrgContext)

	err = svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: existing.OrgID, OrgName: "Rename"}, 0)
	require.ErrorIs(t, err, ErrInvalidOrgContext)

	err = svc.DeleteOrg(existing.OrgID, 0, 1, "tester", "127.0.0.1")
	require.ErrorIs(t, err, ErrInvalidOrgContext)

	err = svc.UpdateOrgStatus(existing.OrgID, org.StatusDisabled, 0)
	require.ErrorIs(t, err, ErrInvalidOrgContext)
}

func TestOrgTransferUserOrg_Success(t *testing.T) {
	svc, mockRepo, auditLogRepo, userRepo, _ := setupOrgServiceWithAudit(t)
	source := createSeedOrg(t, mockRepo, "Source", org.RootParentID, 1)
	target := createSeedOrg(t, mockRepo, "Target", org.RootParentID, 1)

	require.NoError(t, repository.NewGovernancePolicyRepository(mockRepo.db).SetLastRolePolicy(source.OrgID, governance.LastRolePolicyWarnAllow, "tester"))
	require.NoError(t, mockRepo.db.Create(&role.SysRole{RoleName: "Source Role", RoleCode: "source-role", Status: role.StatusEnabled}).Error)
	usr := &user.SysUser{Username: "transfer-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}
	require.NoError(t, userRepo.Create(usr))
	require.NoError(t, mockRepo.db.Create(&user.SysUserRole{UserID: usr.UserID, RoleID: 1, OrgID: source.OrgID}).Error)

	require.NoError(t, svc.TransferUserOrg(source.OrgID, target.OrgID, usr.UserID, 88))

	var sourceCount int64
	require.NoError(t, mockRepo.db.Model(&user.SysUserRole{}).Where("user_id = ? AND org_id = ?", usr.UserID, source.OrgID).Count(&sourceCount).Error)
	assert.Equal(t, int64(0), sourceCount)

	var targetBindings []user.SysUserRole
	require.NoError(t, mockRepo.db.Where("user_id = ? AND org_id = ?", usr.UserID, target.OrgID).Find(&targetBindings).Error)
	require.Len(t, targetBindings, 1)
	defaultRole, err := repository.NewRoleRepository(mockRepo.db).GetByRoleCode(role.BuiltInOrgUserRoleCode)
	require.NoError(t, err)
	assert.Equal(t, defaultRole.RoleID, targetBindings[0].RoleID)

	logs, total, err := auditLogRepo.Query(&audit.AuditLogQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, logs, 2)
}

func TestOrgTransferUserOrg_RejectsDisabledTarget(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	source := createSeedOrg(t, mockRepo, "Source", org.RootParentID, 1)
	target := createSeedOrg(t, mockRepo, "Target", org.RootParentID, 1)
	target.Status = org.StatusDisabled
	require.NoError(t, mockRepo.repo.Update(target))

	require.NoError(t, mockRepo.db.Create(&role.SysRole{RoleName: "Source Role", RoleCode: "source-role-disabled", Status: role.StatusEnabled}).Error)
	require.NoError(t, mockRepo.db.Create(&user.SysUser{UserID: 9, Username: "u9", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, mockRepo.db.Create(&user.SysUserRole{UserID: 9, RoleID: 1, OrgID: source.OrgID}).Error)

	err := svc.TransferUserOrg(source.OrgID, target.OrgID, 9, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target organization is disabled")
}

func TestOrgTransferUserOrg_RejectsBlockedLastRolePolicy(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	source := createSeedOrg(t, mockRepo, "Source", org.RootParentID, 1)
	target := createSeedOrg(t, mockRepo, "Target", org.RootParentID, 1)

	require.NoError(t, repository.NewGovernancePolicyRepository(mockRepo.db).SetLastRolePolicy(source.OrgID, governance.LastRolePolicyBlock, "tester"))
	require.NoError(t, mockRepo.db.Create(&role.SysRole{RoleName: "Only Role", RoleCode: "only-role", Status: role.StatusEnabled}).Error)
	require.NoError(t, mockRepo.db.Create(&user.SysUser{UserID: 10, Username: "u10", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, mockRepo.db.Create(&user.SysUserRole{UserID: 10, RoleID: 1, OrgID: source.OrgID}).Error)

	err := svc.TransferUserOrg(source.OrgID, target.OrgID, 10, 2)
	require.ErrorIs(t, err, ErrLastRoleRemovalBlocked)

	var count int64
	require.NoError(t, mockRepo.db.Model(&user.SysUserRole{}).Where("user_id = ? AND org_id = ?", 10, source.OrgID).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

// --- Batch B: TransferUserOrg input validation, source org not found, user not in source org, WARN_ALLOW, UpdateOrg branches ---

func TestOrgTransferUserOrg_RejectsInvalidSourceOrgID(t *testing.T) {
	svc, _ := setupOrgService(t)
	err := svc.TransferUserOrg(0, 5, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source org, target org, and user id are required")
}

func TestOrgTransferUserOrg_RejectsInvalidTargetOrgID(t *testing.T) {
	svc, _ := setupOrgService(t)
	err := svc.TransferUserOrg(5, 0, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source org, target org, and user id are required")
}

func TestOrgTransferUserOrg_RejectsInvalidUserID(t *testing.T) {
	svc, _ := setupOrgService(t)
	err := svc.TransferUserOrg(5, 7, 0, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source org, target org, and user id are required")
}

func TestOrgTransferUserOrg_RejectsSameOrg(t *testing.T) {
	svc, _ := setupOrgService(t)
	err := svc.TransferUserOrg(5, 5, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source and target organizations must be different")
}

func TestOrgTransferUserOrg_SourceOrgNotFound(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	target := createSeedOrg(t, mockRepo, "Target", org.RootParentID, 1)

	err := svc.TransferUserOrg(99999, target.OrgID, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source organization not found")
}

func TestOrgTransferUserOrg_TargetOrgNotFound(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	source := createSeedOrg(t, mockRepo, "Source", org.RootParentID, 1)
	require.NoError(t, mockRepo.db.Create(&user.SysUserRole{UserID: 20, RoleID: 1, OrgID: source.OrgID}).Error)

	err := svc.TransferUserOrg(source.OrgID, 99999, 20, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target organization not found")
}

func TestOrgTransferUserOrg_UserNotInSourceOrg(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	source := createSeedOrg(t, mockRepo, "Source", org.RootParentID, 1)
	target := createSeedOrg(t, mockRepo, "Target", org.RootParentID, 1)
	require.NoError(t, mockRepo.db.Create(&user.SysUser{UserID: 30, Username: "not-in-source", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

	err := svc.TransferUserOrg(source.OrgID, target.OrgID, 30, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user is not a member of source organization")
}

func TestOrgTransferUserOrg_WarnAllowPolicy_SingleRole(t *testing.T) {
	svc, mockRepo, _, userRepo, _ := setupOrgServiceWithAudit(t)
	source := createSeedOrg(t, mockRepo, "WarnSource", org.RootParentID, 1)
	target := createSeedOrg(t, mockRepo, "WarnTarget", org.RootParentID, 1)

	require.NoError(t, repository.NewGovernancePolicyRepository(mockRepo.db).SetLastRolePolicy(source.OrgID, governance.LastRolePolicyWarnAllow, "tester"))
	require.NoError(t, mockRepo.db.Create(&role.SysRole{RoleName: "Warn Role", RoleCode: "warn-role", Status: role.StatusEnabled}).Error)
	usr := &user.SysUser{Username: "warn-transfer-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}
	require.NoError(t, userRepo.Create(usr))
	require.NoError(t, mockRepo.db.Create(&user.SysUserRole{UserID: usr.UserID, RoleID: 1, OrgID: source.OrgID}).Error)

	err := svc.TransferUserOrg(source.OrgID, target.OrgID, usr.UserID, 88)
	require.NoError(t, err)

	var sourceCount int64
	require.NoError(t, mockRepo.db.Model(&user.SysUserRole{}).Where("user_id = ? AND org_id = ?", usr.UserID, source.OrgID).Count(&sourceCount).Error)
	assert.Equal(t, int64(0), sourceCount)

	var targetBindings []user.SysUserRole
	require.NoError(t, mockRepo.db.Where("user_id = ? AND org_id = ?", usr.UserID, target.OrgID).Find(&targetBindings).Error)
	require.Len(t, targetBindings, 1)
}

func TestOrgTransferUserOrg_MultipleRoles_SkipsLastRoleCheck(t *testing.T) {
	svc, mockRepo, _, userRepo, _ := setupOrgServiceWithAudit(t)
	source := createSeedOrg(t, mockRepo, "MultiSource", org.RootParentID, 1)
	target := createSeedOrg(t, mockRepo, "MultiTarget", org.RootParentID, 1)

	require.NoError(t, mockRepo.db.Create(&role.SysRole{RoleName: "Multi Role 1", RoleCode: "multi-role-1", Status: role.StatusEnabled}).Error)
	require.NoError(t, mockRepo.db.Create(&role.SysRole{RoleName: "Multi Role 2", RoleCode: "multi-role-2", Status: role.StatusEnabled}).Error)
	usr := &user.SysUser{Username: "multi-transfer-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}
	require.NoError(t, userRepo.Create(usr))
	require.NoError(t, mockRepo.db.Create(&user.SysUserRole{UserID: usr.UserID, RoleID: 1, OrgID: source.OrgID}).Error)
	require.NoError(t, mockRepo.db.Create(&user.SysUserRole{UserID: usr.UserID, RoleID: 2, OrgID: source.OrgID}).Error)

	err := svc.TransferUserOrg(source.OrgID, target.OrgID, usr.UserID, 99)
	require.NoError(t, err)

	var sourceCount int64
	require.NoError(t, mockRepo.db.Model(&user.SysUserRole{}).Where("user_id = ? AND org_id = ?", usr.UserID, source.OrgID).Count(&sourceCount).Error)
	assert.Equal(t, int64(0), sourceCount)
}

func TestOrgTransferUserOrg_RecordsAuditForSystemActor(t *testing.T) {
	svc, mockRepo, auditLogRepo, userRepo, _ := setupOrgServiceWithAudit(t)
	source := createSeedOrg(t, mockRepo, "AuditSource", org.RootParentID, 1)
	target := createSeedOrg(t, mockRepo, "AuditTarget", org.RootParentID, 1)

	require.NoError(t, repository.NewGovernancePolicyRepository(mockRepo.db).SetLastRolePolicy(source.OrgID, governance.LastRolePolicyWarnAllow, "tester"))
	require.NoError(t, mockRepo.db.Create(&role.SysRole{RoleName: "Audit Role", RoleCode: "audit-role", Status: role.StatusEnabled}).Error)
	usr := &user.SysUser{Username: "audit-transfer-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}
	require.NoError(t, userRepo.Create(usr))
	require.NoError(t, mockRepo.db.Create(&user.SysUserRole{UserID: usr.UserID, RoleID: 1, OrgID: source.OrgID}).Error)

	err := svc.TransferUserOrg(source.OrgID, target.OrgID, usr.UserID, 0)
	require.NoError(t, err)

	logs, total, logErr := auditLogRepo.Query(&audit.AuditLogQuery{Page: 1, PageSize: 10})
	require.NoError(t, logErr)
	require.Equal(t, int64(2), total)
	for _, l := range logs {
		require.NotNil(t, l.Username)
		assert.Equal(t, "system", *l.Username)
	}
}

func TestOrgTransferUserOrg_NilOrgRepo(t *testing.T) {
	svc := NewOrgService(nil, nil, nil, nil, nil, nil)
	err := svc.TransferUserOrg(5, 7, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organization repository is not configured")
}

func TestOrgTransferUserOrg_NilRoleRepo(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&org.SysOrg{}, &role.SysRole{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}))
	orgRepo := repository.NewOrgRepository(db)
	svc := NewOrgService(orgRepo, nil, nil, nil, nil, nil)

	err = svc.TransferUserOrg(5, 7, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "role repository is not configured")
}

func TestOrgTransferUserOrg_NilUserRoleRepo(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&org.SysOrg{}, &role.SysRole{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}))
	orgRepo := repository.NewOrgRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	svc := NewOrgService(orgRepo, nil, nil, roleRepo, nil, nil)

	err = svc.TransferUserOrg(5, 7, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user role repository is not configured")
}

func TestOrgUpdateOrg_DisabledParent(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	parent := createSeedOrg(t, mockRepo, "ParentDisabled", org.RootParentID, 1)
	parent.Status = org.StatusDisabled
	require.NoError(t, mockRepo.repo.Update(parent))
	child := createSeedOrg(t, mockRepo, "ChildToMove", org.RootParentID, 1)

	parentID := parent.OrgID
	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: child.OrgID, ParentID: &parentID}, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent organization is disabled")
}

func TestOrgUpdateOrg_SameParentEarlyReturn(t *testing.T) {
	svc, mockRepo := setupOrgService(t)
	parent := createSeedOrg(t, mockRepo, "SameParent", org.RootParentID, 1)
	child := createSeedOrg(t, mockRepo, "SameParentChild", parent.OrgID, parent.Level+1)

	sameParent := parent.OrgID
	err := svc.UpdateOrg(&org.OrgUpdateRequest{OrgID: child.OrgID, ParentID: &sameParent}, 1)
	require.NoError(t, err)

	updated, getErr := mockRepo.repo.GetByID(child.OrgID)
	require.NoError(t, getErr)
	assert.Equal(t, parent.OrgID, updated.ParentID)
	assert.Equal(t, parent.Level+1, updated.Level)
}
