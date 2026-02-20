package permission

import (
	"testing"
)

func TestPermissionConstants(t *testing.T) {
	if PermTypeMenu != "menu" {
		t.Errorf("PermTypeMenu should be 'menu', got '%s'", PermTypeMenu)
	}
	if PermTypeButton != "button" {
		t.Errorf("PermTypeButton should be 'button', got '%s'", PermTypeButton)
	}
	if PermTypeData != "data" {
		t.Errorf("PermTypeData should be 'data', got '%s'", PermTypeData)
	}
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
}

func TestSysPermTableName(t *testing.T) {
	p := SysPerm{}
	if p.TableName() != "sys_perm" {
		t.Errorf("TableName should be 'sys_perm', got '%s'", p.TableName())
	}
}
