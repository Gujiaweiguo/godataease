package user

import (
	"testing"
)

func TestUserConstants(t *testing.T) {
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
	if FromLocal != 0 {
		t.Errorf("FromLocal should be 0, got %d", FromLocal)
	}
	if FromThirdParty != 1 {
		t.Errorf("FromThirdParty should be 1, got %d", FromThirdParty)
	}
}

func TestSysUserTableName(t *testing.T) {
	u := SysUser{}
	if u.TableName() != "sys_user" {
		t.Errorf("TableName should be 'sys_user', got '%s'", u.TableName())
	}
}

func TestSysUserRoleTableName(t *testing.T) {
	ur := SysUserRole{}
	if ur.TableName() != "sys_user_role" {
		t.Errorf("TableName should be 'sys_user_role', got '%s'", ur.TableName())
	}
}

func TestSysUserPermTableName(t *testing.T) {
	up := SysUserPerm{}
	if up.TableName() != "sys_user_perm" {
		t.Errorf("TableName should be 'sys_user_perm', got '%s'", up.TableName())
	}
}
