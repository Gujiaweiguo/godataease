package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/share"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupShareServiceRepoTest(t *testing.T) (*ShareService, *repository.ShareRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE core_share (id INTEGER PRIMARY KEY AUTOINCREMENT, creator INTEGER, resource_id INTEGER, resource_type TEXT, time DATETIME, exp INTEGER, uuid TEXT UNIQUE, pwd TEXT, auto_pwd BOOLEAN, ticket_require BOOLEAN)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE core_share_ticket (id INTEGER PRIMARY KEY AUTOINCREMENT, uuid TEXT, ticket TEXT UNIQUE, exp INTEGER, args TEXT, access_time DATETIME)`).Error)

	repo := repository.NewShareRepository(db)
	return NewShareService(repo), repo, db
}

func closeShareDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func TestShareServiceHelpers_PasswordAndExpiry(t *testing.T) {
	pwd, err := generatePassword(8)
	assert.NoError(t, err)
	assert.Len(t, pwd, 8)

	assert.False(t, isShareExpired(0))
	assert.True(t, isShareExpired(time.Now().Add(-time.Hour).UnixMilli()))
	assert.False(t, isShareExpired(time.Now().Add(time.Hour).UnixMilli()))
	assert.True(t, isShareExpired(time.Now().Add(-time.Hour).Unix()))

	assert.True(t, isValidSharePassword("Ab1!"))
	assert.False(t, isValidSharePassword("abc"))
	assert.False(t, isValidSharePassword("abcdef"))
	assert.False(t, isValidSharePassword("123456"))
	assert.False(t, isValidSharePassword("Abcdef"))
	assert.False(t, isValidSharePassword("Ab1中文"))
}

func TestShareServiceHelpers_EnsureShareOwner(t *testing.T) {
	svc := &ShareService{}
	assert.Equal(t, gorm.ErrRecordNotFound, svc.ensureShareOwner(nil, 1))
	assert.EqualError(t, svc.ensureShareOwner(&share.Share{Creator: 1}, 2), "forbidden")
	assert.NoError(t, svc.ensureShareOwner(&share.Share{Creator: 1}, 1))
}

func TestShareService_CreateShare(t *testing.T) {
	t.Run("uses provided auto password", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)

		resp, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 11, ResourceType: "dashboard", Exp: time.Now().Add(time.Hour).UnixMilli(), AutoPwd: true}, 7)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotZero(t, resp.ID)
		assert.Len(t, resp.UUID, 32)
		assert.Len(t, resp.Pwd, 4)
		assert.True(t, resp.AutoPwd)
	})

	t.Run("without auto password leaves pwd empty", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)

		resp, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 22, ResourceType: "dashboard", AutoPwd: false}, 8)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.UUID, 32)
		assert.Empty(t, resp.Pwd)
		assert.False(t, resp.AutoPwd)
	})

	t.Run("repository create error propagates", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_share_insert BEFORE INSERT ON core_share BEGIN SELECT RAISE(FAIL, 'deny create'); END;").Error)

		resp, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 23, ResourceType: "dashboard", AutoPwd: false}, 8)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "deny create")
	})
}

func TestShareService_ValidateShare(t *testing.T) {
	t.Run("not found returns invalid", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)

		resp, err := svc.ValidateShare(&share.ShareValidateRequest{UUID: "missing"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.False(t, resp.Valid)
	})

	t.Run("expired returns invalid", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)
		created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 33, ResourceType: "dashboard", Exp: time.Now().Add(-time.Hour).UnixMilli(), AutoPwd: false}, 9)
		require.NoError(t, err)

		resp, validateErr := svc.ValidateShare(&share.ShareValidateRequest{UUID: created.UUID})
		require.NoError(t, validateErr)
		assert.False(t, resp.Valid)
	})

	t.Run("password mismatch returns invalid", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)
		created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 44, ResourceType: "dashboard", Exp: time.Now().Add(time.Hour).UnixMilli(), AutoPwd: true}, 10)
		require.NoError(t, err)

		resp, validateErr := svc.ValidateShare(&share.ShareValidateRequest{UUID: created.UUID, Pwd: "bad"})
		require.NoError(t, validateErr)
		assert.False(t, resp.Valid)
	})

	t.Run("valid returns resource fields", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)
		created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 55, ResourceType: "dashboard", Exp: time.Now().Add(time.Hour).UnixMilli(), AutoPwd: true}, 11)
		require.NoError(t, err)

		resp, validateErr := svc.ValidateShare(&share.ShareValidateRequest{UUID: created.UUID, Pwd: created.Pwd})
		require.NoError(t, validateErr)
		require.NotNil(t, resp)
		assert.True(t, resp.Valid)
		assert.Equal(t, int64(55), resp.ResourceID)
		assert.Equal(t, "dashboard", resp.ResourceType)
		assert.False(t, resp.TicketRequire)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		closeShareDB(t, db)

		resp, err := svc.ValidateShare(&share.ShareValidateRequest{UUID: "repo-error"})
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestShareService_RevokeShare(t *testing.T) {
	t.Run("not found returns success false", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)

		resp, err := svc.RevokeShare(999, 1)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.False(t, resp.Success)
	})

	t.Run("wrong creator returns success false", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)
		created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 66, ResourceType: "dashboard", AutoPwd: false}, 12)
		require.NoError(t, err)

		resp, revokeErr := svc.RevokeShare(created.ID, 99)
		require.NoError(t, revokeErr)
		require.NotNil(t, resp)
		assert.False(t, resp.Success)
	})

	t.Run("owner deletes share", func(t *testing.T) {
		svc, repo, _ := setupShareServiceRepoTest(t)
		created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 77, ResourceType: "dashboard", AutoPwd: false}, 13)
		require.NoError(t, err)

		resp, revokeErr := svc.RevokeShare(created.ID, 13)
		require.NoError(t, revokeErr)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)

		loaded, loadErr := repo.GetByID(created.ID)
		require.Error(t, loadErr)
		assert.Nil(t, loaded)
	})

	t.Run("get by id repository error propagates", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		closeShareDB(t, db)

		resp, err := svc.RevokeShare(1, 13)
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestShareService_GetDetail(t *testing.T) {
	t.Run("not found returns nil", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)

		resp, err := svc.GetDetail(999)
		require.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("returns mapped fields", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)
		created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 88, ResourceType: "dashboard", Exp: time.Now().Add(time.Hour).UnixMilli(), AutoPwd: true}, 14)
		require.NoError(t, err)

		resp, detailErr := svc.GetDetail(88)
		require.NoError(t, detailErr)
		require.NotNil(t, resp)
		assert.Equal(t, created.ID, resp.ID)
		assert.Equal(t, created.UUID, resp.UUID)
		assert.Equal(t, created.Pwd, resp.Pwd)
		assert.True(t, resp.AutoPwd)
	})
}

func TestShareService_SwitchStatus(t *testing.T) {
	t.Run("creates share when missing", func(t *testing.T) {
		svc, repo, _ := setupShareServiceRepoTest(t)

		require.NoError(t, svc.SwitchStatus(101, 15))
		loaded, err := repo.GetByResourceID(101)
		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, int64(15), loaded.Creator)
		assert.Equal(t, "dashboard", loaded.ResourceType)
		assert.True(t, loaded.AutoPwd)
	})

	t.Run("deletes existing share", func(t *testing.T) {
		svc, repo, _ := setupShareServiceRepoTest(t)
		_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 102, ResourceType: "dashboard", AutoPwd: false}, 16)
		require.NoError(t, err)

		require.NoError(t, svc.SwitchStatus(102, 16))
		loaded, loadErr := repo.GetByResourceID(102)
		require.Error(t, loadErr)
		assert.Nil(t, loaded)
	})

	t.Run("delete error propagates", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 103, ResourceType: "dashboard", AutoPwd: false}, 16)
		require.NoError(t, err)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_share_delete_switch BEFORE DELETE ON core_share BEGIN SELECT RAISE(FAIL, 'deny switch delete'); END;").Error)

		err = svc.SwitchStatus(103, 16)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deny switch delete")
	})
}

func TestShareService_CreateTicket(t *testing.T) {
	t.Run("generates when empty", func(t *testing.T) {
		svc, repo, _ := setupShareServiceRepoTest(t)

		created, err := svc.CreateTicket(&share.TicketCreateRequest{Ticket: "", UUID: "uuid-1", Exp: time.Now().Add(time.Hour).UnixMilli(), Args: `{"k":"v"}`})
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Len(t, created.Ticket, 32)

		loaded, loadErr := repo.GetTicketByTicket(created.Ticket)
		require.NoError(t, loadErr)
		require.NotNil(t, loaded)
		assert.Equal(t, "uuid-1", loaded.UUID)
	})

	t.Run("uses provided ticket when no generate flag", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)

		created, err := svc.CreateTicket(&share.TicketCreateRequest{Ticket: "provided-ticket", UUID: "uuid-2", Exp: time.Now().Add(time.Hour).UnixMilli()})
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, "provided-ticket", created.Ticket)
	})
}

func TestShareService_ValidateTicket(t *testing.T) {
	t.Run("uuid mismatch returns invalid", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)
		created, err := svc.CreateTicket(&share.TicketCreateRequest{Ticket: "ticket-1", UUID: "uuid-1", Exp: time.Now().Add(time.Hour).UnixMilli(), Args: `{"a":1}`})
		require.NoError(t, err)

		resp, validateErr := svc.ValidateTicket(&share.TicketValidateRequest{Ticket: created.Ticket, UUID: "wrong-uuid"})
		require.NoError(t, validateErr)
		assert.False(t, resp.TicketValid)
		assert.False(t, resp.TicketExp)
	})

	t.Run("expired returns ticket exp true", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)
		created, err := svc.CreateTicket(&share.TicketCreateRequest{Ticket: "ticket-2", UUID: "uuid-2", Exp: time.Now().Add(-time.Hour).UnixMilli()})
		require.NoError(t, err)

		resp, validateErr := svc.ValidateTicket(&share.TicketValidateRequest{Ticket: created.Ticket, UUID: "uuid-2"})
		require.NoError(t, validateErr)
		assert.True(t, resp.TicketValid)
		assert.True(t, resp.TicketExp)
	})

	t.Run("valid updates access time", func(t *testing.T) {
		svc, repo, _ := setupShareServiceRepoTest(t)
		created, err := svc.CreateTicket(&share.TicketCreateRequest{Ticket: "ticket-3", UUID: "uuid-3", Exp: time.Now().Add(time.Hour).UnixMilli(), Args: `{"ok":true}`})
		require.NoError(t, err)

		resp, validateErr := svc.ValidateTicket(&share.TicketValidateRequest{Ticket: created.Ticket, UUID: "uuid-3"})
		require.NoError(t, validateErr)
		assert.True(t, resp.TicketValid)
		assert.False(t, resp.TicketExp)
		assert.Equal(t, `{"ok":true}`, resp.Args)

		loaded, loadErr := repo.GetTicketByTicket(created.Ticket)
		require.NoError(t, loadErr)
		require.NotNil(t, loaded)
		require.NotNil(t, loaded.AccessTime)
	})
}

func TestShareService_EditUUID(t *testing.T) {
	t.Run("empty invalid duplicate and success", func(t *testing.T) {
		svc, repo, _ := setupShareServiceRepoTest(t)
		first, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 201, ResourceType: "dashboard", AutoPwd: false}, 21)
		require.NoError(t, err)
		_, err = svc.CreateShare(&share.ShareCreateRequest{ResourceID: 202, ResourceType: "dashboard", AutoPwd: false}, 21)
		require.NoError(t, err)
		second, err := repo.GetByResourceID(202)
		require.NoError(t, err)
		second.UUID = "DupUUID12"
		require.NoError(t, repo.Update(second))
		_, err = svc.CreateTicket(&share.TicketCreateRequest{Ticket: "ticket-edit-uuid", UUID: first.UUID, Exp: time.Now().Add(time.Hour).UnixMilli()})
		require.NoError(t, err)

		msg, editErr := svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 201, UUID: "   "}, 21)
		require.NoError(t, editErr)
		assert.Equal(t, "uuid cannot be empty", msg)

		msg, editErr = svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 201, UUID: "bad-uuid!"}, 21)
		require.NoError(t, editErr)
		assert.Equal(t, "invalid uuid format", msg)

		msg, editErr = svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 201, UUID: second.UUID}, 21)
		require.NoError(t, editErr)
		assert.Equal(t, "uuid already exists", msg)

		msg, editErr = svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 201, UUID: "NewUUID12"}, 21)
		require.NoError(t, editErr)
		assert.Equal(t, "", msg)

		loaded, loadErr := repo.GetByResourceID(201)
		require.NoError(t, loadErr)
		assert.Equal(t, "NewUUID12", loaded.UUID)
		ticketLoaded, ticketErr := repo.GetTicketByTicket("ticket-edit-uuid")
		require.NoError(t, ticketErr)
		assert.Equal(t, "NewUUID12", ticketLoaded.UUID)
	})

	t.Run("share not found wrong creator and same uuid no op", func(t *testing.T) {
		svc, repo, _ := setupShareServiceRepoTest(t)

		msg, err := svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 999, UUID: "AnyUUID12"}, 21)
		require.NoError(t, err)
		assert.Equal(t, "share not found", msg)

		created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 206, ResourceType: "dashboard", AutoPwd: false}, 25)
		require.NoError(t, err)

		msg, err = svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 206, UUID: "OtherUUID9"}, 99)
		require.Error(t, err)
		assert.Equal(t, "", msg)
		assert.EqualError(t, err, "forbidden")

		msg, err = svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 206, UUID: created.UUID}, 25)
		require.NoError(t, err)
		assert.Equal(t, "", msg)

		loaded, loadErr := repo.GetByResourceID(206)
		require.NoError(t, loadErr)
		assert.Equal(t, created.UUID, loaded.UUID)
	})

	t.Run("exists by uuid repository error propagates", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 213, ResourceType: "dashboard", AutoPwd: false}, 32)
		require.NoError(t, err)
		closeShareDB(t, db)

		msg, err := svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 213, UUID: "ValidUUID8"}, 32)
		require.Error(t, err)
		assert.Equal(t, "", msg)
	})
}

func TestShareService_EditExp(t *testing.T) {
	t.Run("invalid past expiration", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)
		_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 203, ResourceType: "dashboard", AutoPwd: false}, 22)
		require.NoError(t, err)

		err = svc.EditExp(&share.ShareEditExpRequest{ResourceID: 203, Exp: time.Now().Add(-time.Hour).UnixMilli()}, 22)
		require.Error(t, err)
		assert.EqualError(t, err, "invalid expiration")
	})

	t.Run("share not found wrong creator and valid future expiration", func(t *testing.T) {
		svc, repo, _ := setupShareServiceRepoTest(t)

		err := svc.EditExp(&share.ShareEditExpRequest{ResourceID: 999, Exp: time.Now().Add(time.Hour).UnixMilli()}, 22)
		require.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		_, err = svc.CreateShare(&share.ShareCreateRequest{ResourceID: 207, ResourceType: "dashboard", AutoPwd: false}, 26)
		require.NoError(t, err)

		err = svc.EditExp(&share.ShareEditExpRequest{ResourceID: 207, Exp: time.Now().Add(time.Hour).UnixMilli()}, 99)
		require.Error(t, err)
		assert.EqualError(t, err, "forbidden")

		future := time.Now().Add(2 * time.Hour).UnixMilli()
		require.NoError(t, svc.EditExp(&share.ShareEditExpRequest{ResourceID: 207, Exp: future}, 26))
		loaded, loadErr := repo.GetByResourceID(207)
		require.NoError(t, loadErr)
		assert.Equal(t, future, loaded.Exp)
	})

	t.Run("update error propagates", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 211, ResourceType: "dashboard", AutoPwd: false}, 30)
		require.NoError(t, err)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_share_update BEFORE UPDATE ON core_share BEGIN SELECT RAISE(FAIL, 'deny update'); END;").Error)

		err = svc.EditExp(&share.ShareEditExpRequest{ResourceID: 211, Exp: time.Now().Add(time.Hour).UnixMilli()}, 30)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deny update")
	})
}

func TestShareService_EditPwd(t *testing.T) {
	t.Run("empty clears password", func(t *testing.T) {
		svc, repo, _ := setupShareServiceRepoTest(t)
		created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 204, ResourceType: "dashboard", AutoPwd: true}, 23)
		require.NoError(t, err)
		require.NotEmpty(t, created.Pwd)

		require.NoError(t, svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 204, Pwd: "   ", AutoPwd: false}, 23))
		loaded, loadErr := repo.GetByResourceID(204)
		require.NoError(t, loadErr)
		assert.Empty(t, loaded.Pwd)
		assert.False(t, loaded.AutoPwd)
	})

	t.Run("invalid manual password rejected", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)
		_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 205, ResourceType: "dashboard", AutoPwd: false}, 24)
		require.NoError(t, err)

		err = svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 205, Pwd: "abcdef", AutoPwd: false}, 24)
		require.Error(t, err)
		assert.EqualError(t, err, "invalid password format")
	})

	t.Run("manual valid auto pwd and wrong creator", func(t *testing.T) {
		svc, repo, _ := setupShareServiceRepoTest(t)
		_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 208, ResourceType: "dashboard", AutoPwd: false}, 27)
		require.NoError(t, err)
		_, err = svc.CreateShare(&share.ShareCreateRequest{ResourceID: 209, ResourceType: "dashboard", AutoPwd: false}, 28)
		require.NoError(t, err)

		require.NoError(t, svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 208, Pwd: "Ab1!", AutoPwd: false}, 27))
		loaded, loadErr := repo.GetByResourceID(208)
		require.NoError(t, loadErr)
		assert.Equal(t, "Ab1!", loaded.Pwd)
		assert.False(t, loaded.AutoPwd)

		require.NoError(t, svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 208, Pwd: "simple", AutoPwd: true}, 27))
		loaded, loadErr = repo.GetByResourceID(208)
		require.NoError(t, loadErr)
		assert.Equal(t, "simple", loaded.Pwd)
		assert.True(t, loaded.AutoPwd)

		err = svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 209, Pwd: "Ab1!", AutoPwd: false}, 99)
		require.Error(t, err)
		assert.EqualError(t, err, "forbidden")
	})

	t.Run("share not found returns error", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)

		err := svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 999, Pwd: "Ab1!", AutoPwd: false}, 23)
		require.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("update error propagates", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 214, ResourceType: "dashboard", AutoPwd: false}, 33)
		require.NoError(t, err)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_share_pwd_update BEFORE UPDATE ON core_share BEGIN SELECT RAISE(FAIL, 'deny pwd update'); END;").Error)

		err = svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 214, Pwd: "Ab1!", AutoPwd: false}, 33)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deny pwd update")
	})
}

func TestShareService_ExtraEdgePaths(t *testing.T) {
	t.Run("validate ticket not found returns invalid", func(t *testing.T) {
		svc, _, _ := setupShareServiceRepoTest(t)

		resp, err := svc.ValidateTicket(&share.TicketValidateRequest{Ticket: "missing", UUID: "uuid"})
		require.NoError(t, err)
		assert.False(t, resp.TicketValid)
		assert.False(t, resp.TicketExp)
	})

	t.Run("switch status get by resource error propagates", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		closeShareDB(t, db)

		err := svc.SwitchStatus(300, 1)
		require.Error(t, err)
	})

	t.Run("create ticket and validate ticket repo errors", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		closeShareDB(t, db)

		ticket, err := svc.CreateTicket(&share.TicketCreateRequest{Ticket: "ticket-err", UUID: "uuid", Exp: time.Now().Add(time.Hour).UnixMilli()})
		require.Error(t, err)
		assert.Nil(t, ticket)

		resp, err := svc.ValidateTicket(&share.TicketValidateRequest{Ticket: "ticket-err", UUID: "uuid"})
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("validate ticket ignores access time update failure", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		created, err := svc.CreateTicket(&share.TicketCreateRequest{Ticket: "ticket-ignore-update", UUID: "uuid-ok", Exp: time.Now().Add(time.Hour).UnixMilli(), Args: `{"ok":true}`})
		require.NoError(t, err)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_ticket_update BEFORE UPDATE ON core_share_ticket BEGIN SELECT RAISE(FAIL, 'deny ticket update'); END;").Error)

		resp, err := svc.ValidateTicket(&share.TicketValidateRequest{Ticket: created.Ticket, UUID: "uuid-ok"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.TicketValid)
		assert.False(t, resp.TicketExp)
		assert.Equal(t, `{"ok":true}`, resp.Args)
	})

	t.Run("revoke share delete error propagates", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 210, ResourceType: "dashboard", AutoPwd: false}, 29)
		require.NoError(t, err)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_share_delete BEFORE DELETE ON core_share BEGIN SELECT RAISE(FAIL, 'deny delete'); END;").Error)

		resp, revokeErr := svc.RevokeShare(created.ID, 29)
		require.Error(t, revokeErr)
		assert.Nil(t, resp)
	})

	t.Run("get detail repo error propagates", func(t *testing.T) {
		svc, _, db := setupShareServiceRepoTest(t)
		closeShareDB(t, db)

		resp, err := svc.GetDetail(123)
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestShareService_EditUUID_UpdateUUIDWithTicketsErrorPropagates(t *testing.T) {
	svc, repo, db := setupShareServiceRepoTest(t)
	created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 212, ResourceType: "dashboard", AutoPwd: false}, 31)
	require.NoError(t, err)
	_, err = svc.CreateTicket(&share.TicketCreateRequest{Ticket: "ticket-uuid-fail", UUID: created.UUID, Exp: time.Now().Add(time.Hour).UnixMilli()})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TRIGGER deny_ticket_uuid_update BEFORE UPDATE ON core_share_ticket BEGIN SELECT RAISE(FAIL, 'deny uuid update'); END;").Error)

	msg, err := svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 212, UUID: "ValidUUID9"}, 31)
	require.Error(t, err)
	assert.Equal(t, "", msg)
	assert.Contains(t, err.Error(), "deny uuid update")

	loaded, loadErr := repo.GetByResourceID(212)
	require.NoError(t, loadErr)
	assert.Equal(t, created.UUID, loaded.UUID)
}
