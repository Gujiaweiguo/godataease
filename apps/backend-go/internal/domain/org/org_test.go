package org

import (
	"testing"
)

func TestOrgConstants(t *testing.T) {
	if StatusEnabled != 1 {
		t.Errorf("StatusEnabled should be 1, got %d", StatusEnabled)
	}
	if StatusDisabled != 0 {
		t.Errorf("StatusDisabled should be 0, got %d", StatusDisabled)
	}
	if DelFlagNormal != 0 {
		t.Errorf("DelFlagNormal should be 0, got %d", DelFlagNormal)
	}
	if DelFlagDeleted != 1 {
		t.Errorf("DelFlagDeleted should be 1, got %d", DelFlagDeleted)
	}
	if RootParentID != 0 {
		t.Errorf("RootParentID should be 0, got %d", RootParentID)
	}
}

func TestSysOrgTableName(t *testing.T) {
	o := SysOrg{}
	if o.TableName() != "sys_org" {
		t.Errorf("TableName should be 'sys_org', got '%s'", o.TableName())
	}
}

func TestSysOrgToTreeNode(t *testing.T) {
	desc := "test description"
	o := SysOrg{
		OrgID:    1,
		OrgName:  "Test Org",
		OrgDesc:  &desc,
		ParentID: 0,
		Level:    1,
		Status:   1,
	}

	node := o.ToTreeNode()

	if node.OrgID != 1 {
		t.Errorf("OrgID should be 1, got %d", node.OrgID)
	}
	if node.OrgName != "Test Org" {
		t.Errorf("OrgName should be 'Test Org', got '%s'", node.OrgName)
	}
	if *node.OrgDesc != "test description" {
		t.Errorf("OrgDesc should be 'test description', got '%s'", *node.OrgDesc)
	}
	if node.ParentID != 0 {
		t.Errorf("ParentID should be 0, got %d", node.ParentID)
	}
	if node.Level != 1 {
		t.Errorf("Level should be 1, got %d", node.Level)
	}
	if node.Status != 1 {
		t.Errorf("Status should be 1, got %d", node.Status)
	}
	if node.Children == nil {
		t.Error("Children should not be nil")
	}
	if len(node.Children) != 0 {
		t.Errorf("Children should be empty, got %d", len(node.Children))
	}
}
