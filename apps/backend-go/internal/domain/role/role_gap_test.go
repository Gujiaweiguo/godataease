package role

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRoleTypeAndBuiltinConstants(t *testing.T) {
	assert.Equal(t, "system", RoleTypeSystem)
	assert.Equal(t, "custom", RoleTypeCustom)
	assert.Equal(t, "org", RoleTypeOrganization)
	assert.Equal(t, "ROLE_ORG_DEFAULT_USER", BuiltInOrgUserRoleCode)
	assert.Equal(t, "普通用户", BuiltInOrgUserRoleName)
}

func TestRoleMenu_TableNameAndFields(t *testing.T) {
	now := time.Now()
	relation := RoleMenu{
		ID:        10,
		RoleID:    20,
		MenuID:    30,
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, "sys_role_menu", relation.TableName())
	assert.Equal(t, int64(20), relation.RoleID)
	assert.Equal(t, int64(30), relation.MenuID)
	assert.True(t, relation.CreatedAt.Equal(now))
	assert.True(t, relation.UpdatedAt.Equal(now))
}

func TestRoleAuxiliaryRequestsAndResults(t *testing.T) {
	keyword := "admin"
	roleType := RoleTypeSystem
	uid := int64(42)
	email := "admin@example.com"
	phone := "13800138000"
	request := RoleRequest{Keyword: &keyword, Uid: &uid}

	assert.Equal(t, keyword, *request.Keyword)
	assert.Equal(t, uid, *request.Uid)

	unmount := UnmountUserRequest{Rid: 1, Uid: 2, OrgId: 3}
	external := MountExternalUserRequest{Rid: 4, Uid: 5}
	user := ExternalUserVO{Uid: 6, Account: "tester", Name: "Tester", Email: &email, Phone: &phone}
	page := RolePageRequest{Keyword: &keyword, RoleType: &roleType, Current: 2, Size: 50}
	result := RolePageResult{List: []*RoleVO{{ID: 7, Name: "Admin", Code: "ROLE_ADMIN", Status: StatusEnabled}}, Total: 1, Current: 2, Size: 50}

	assert.Equal(t, int64(1), unmount.Rid)
	assert.Equal(t, int64(2), unmount.Uid)
	assert.Equal(t, int64(3), unmount.OrgId)
	assert.Equal(t, int64(4), external.Rid)
	assert.Equal(t, int64(5), external.Uid)
	assert.Equal(t, int64(6), user.Uid)
	assert.Equal(t, "tester", user.Account)
	assert.Equal(t, "Tester", user.Name)
	assert.Equal(t, email, *user.Email)
	assert.Equal(t, phone, *user.Phone)
	assert.Equal(t, keyword, *page.Keyword)
	assert.Equal(t, roleType, *page.RoleType)
	assert.Equal(t, 2, page.Current)
	assert.Equal(t, 50, page.Size)
	assert.Len(t, result.List, 1)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "ROLE_ADMIN", result.List[0].Code)
	assert.Equal(t, 2, result.Current)
	assert.Equal(t, 50, result.Size)
}
