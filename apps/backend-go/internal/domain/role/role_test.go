package role

import (
	"testing"
	"time"
)

func TestStatusConstants(t *testing.T) {
	if StatusEnabled != 1 {
		t.Errorf("Expected StatusEnabled = 1, got %d", StatusEnabled)
	}
	if StatusDisabled != 0 {
		t.Errorf("Expected StatusDisabled = 0, got %d", StatusDisabled)
	}
}

func TestDataScopeConstants(t *testing.T) {
	scopes := []string{DataScopeAll, DataScopeCustom, DataScopeDept, DataScopeDeptAndChild, DataScopeSelf}
	expected := []string{"all", "custom", "dept", "dept_and_child", "self"}

	for i, scope := range scopes {
		if scope != expected[i] {
			t.Errorf("Expected scope '%s', got '%s'", expected[i], scope)
		}
	}
}

func TestSysRole_TableName(t *testing.T) {
	role := SysRole{}
	if role.TableName() != "sys_role" {
		t.Errorf("Expected table name 'sys_role', got '%s'", role.TableName())
	}
}

func TestSysRole_Fields(t *testing.T) {
	now := time.Now()
	desc := "test role"
	parentID := int64(1)
	level := 2
	dataScope := "all"
	createBy := "admin"
	updateBy := "admin"

	role := SysRole{
		RoleID:     1,
		RoleName:   "Admin",
		RoleCode:   "ROLE_ADMIN",
		RoleDesc:   &desc,
		ParentID:   &parentID,
		Level:      &level,
		DataScope:  &dataScope,
		Status:     StatusEnabled,
		CreateBy:   &createBy,
		CreateTime: &now,
		UpdateBy:   &updateBy,
		UpdateTime: &now,
	}

	if role.RoleID != 1 {
		t.Errorf("Expected RoleID 1, got %d", role.RoleID)
	}
	if role.RoleName != "Admin" {
		t.Errorf("Expected RoleName 'Admin', got '%s'", role.RoleName)
	}
	if role.RoleCode != "ROLE_ADMIN" {
		t.Errorf("Expected RoleCode 'ROLE_ADMIN', got '%s'", role.RoleCode)
	}
	if *role.RoleDesc != "test role" {
		t.Errorf("Expected RoleDesc 'test role', got '%s'", *role.RoleDesc)
	}
	if role.Status != StatusEnabled {
		t.Errorf("Expected Status %d, got %d", StatusEnabled, role.Status)
	}
}

func TestSysRole_NilFields(t *testing.T) {
	role := SysRole{
		RoleID:   1,
		RoleName: "Test",
		Status:   StatusEnabled,
	}

	if role.RoleDesc != nil {
		t.Error("Expected RoleDesc to be nil")
	}
	if role.ParentID != nil {
		t.Error("Expected ParentID to be nil")
	}
	if role.Level != nil {
		t.Error("Expected Level to be nil")
	}
	if role.DataScope != nil {
		t.Error("Expected DataScope to be nil")
	}
}

func TestRoleCreator_Fields(t *testing.T) {
	desc := "test role"
	creator := RoleCreator{
		Name:     "TestRole",
		TypeCode: 1,
		Desc:     &desc,
	}

	if creator.Name != "TestRole" {
		t.Errorf("Expected Name 'TestRole', got '%s'", creator.Name)
	}
	if creator.TypeCode != 1 {
		t.Errorf("Expected TypeCode 1, got %d", creator.TypeCode)
	}
	if *creator.Desc != "test role" {
		t.Errorf("Expected Desc 'test role', got '%s'", *creator.Desc)
	}
}

func TestRoleEditor_Fields(t *testing.T) {
	desc := "updated role"
	editor := RoleEditor{
		ID:   1,
		Name: "UpdatedRole",
		Desc: &desc,
	}

	if editor.ID != 1 {
		t.Errorf("Expected ID 1, got %d", editor.ID)
	}
	if editor.Name != "UpdatedRole" {
		t.Errorf("Expected Name 'UpdatedRole', got '%s'", editor.Name)
	}
}

func TestRoleVO_Fields(t *testing.T) {
	vo := RoleVO{
		ID:       1,
		Name:     "Admin",
		ReadOnly: true,
		Root:     true,
	}

	if vo.ID != 1 {
		t.Errorf("Expected ID 1, got %d", vo.ID)
	}
	if !vo.ReadOnly {
		t.Error("Expected ReadOnly to be true")
	}
	if !vo.Root {
		t.Error("Expected Root to be true")
	}
}

func TestRoleDetailVO_Fields(t *testing.T) {
	desc := "detail desc"
	parentID := int64(1)
	level := 2
	dataScope := "all"

	vo := RoleDetailVO{
		ID:        1,
		Name:      "Admin",
		Code:      "ROLE_ADMIN",
		Desc:      &desc,
		ParentID:  &parentID,
		Level:     &level,
		DataScope: &dataScope,
		Status:    StatusEnabled,
	}

	if vo.ID != 1 {
		t.Errorf("Expected ID 1, got %d", vo.ID)
	}
	if vo.Code != "ROLE_ADMIN" {
		t.Errorf("Expected Code 'ROLE_ADMIN', got '%s'", vo.Code)
	}
}

func TestRoleQueryRequest_Fields(t *testing.T) {
	keyword := "admin"
	req := RoleQueryRequest{Keyword: &keyword}

	if *req.Keyword != "admin" {
		t.Errorf("Expected Keyword 'admin', got '%s'", *req.Keyword)
	}
}

func TestRoleQueryRequest_NilKeyword(t *testing.T) {
	req := RoleQueryRequest{}

	if req.Keyword != nil {
		t.Error("Expected Keyword to be nil")
	}
}

func TestMountUserRequest_Fields(t *testing.T) {
	req := MountUserRequest{
		Rid:   1,
		Uids:  []int64{1, 2, 3},
		OrgId: 10,
		Over:  true,
	}

	if req.Rid != 1 {
		t.Errorf("Expected Rid 1, got %d", req.Rid)
	}
	if len(req.Uids) != 3 {
		t.Errorf("Expected 3 Uids, got %d", len(req.Uids))
	}
	if req.OrgId != 10 {
		t.Errorf("Expected OrgId 10, got %d", req.OrgId)
	}
	if !req.Over {
		t.Error("Expected Over to be true")
	}
}
