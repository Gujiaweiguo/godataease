package repository

import (
	"context"
	"testing"
	"time"

	"dataease/backend/internal/domain/auto"
	datafillingdomain "dataease/backend/internal/domain/datafilling"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/driver"
	"dataease/backend/internal/domain/embedded"
	"dataease/backend/internal/domain/license"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/ticket"
	"dataease/backend/internal/domain/visualization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ==================== Shared helpers ====================

func round7bOpenDB(t *testing.T, models ...interface{}) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(models...))
	return db
}

// ==================== MockPermRepository (9 functions) ====================

func TestRound7BRepo_MockPerm_New(t *testing.T) {
	repo := NewMockPermRepository()
	require.NotNil(t, repo)
	assert.Empty(t, repo.perms)
}

func TestRound7BRepo_MockPerm_Create(t *testing.T) {
	repo := NewMockPermRepository()
	p := &permission.SysPerm{PermName: "test", PermKey: "test_key", PermType: "menu"}
	require.NoError(t, repo.Create(p))
	assert.Equal(t, int64(1), p.PermID)
}

func TestRound7BRepo_MockPerm_Update(t *testing.T) {
	repo := NewMockPermRepository()
	p := &permission.SysPerm{PermName: "old", PermKey: "old_key", PermType: "menu"}
	require.NoError(t, repo.Create(p))
	p.PermName = "new"
	require.NoError(t, repo.Update(p))
	got, err := repo.GetByID(p.PermID)
	require.NoError(t, err)
	assert.Equal(t, "new", got.PermName)
}

func TestRound7BRepo_MockPerm_Update_NotFound(t *testing.T) {
	repo := NewMockPermRepository()
	p := &permission.SysPerm{PermID: 999, PermName: "ghost", PermKey: "ghost_key"}
	require.NoError(t, repo.Update(p))
	_, err := repo.GetByID(999)
	assert.Equal(t, ErrNotFound, err)
}

func TestRound7BRepo_MockPerm_Delete(t *testing.T) {
	repo := NewMockPermRepository()
	p := &permission.SysPerm{PermName: "del", PermKey: "del_key", PermType: "button"}
	require.NoError(t, repo.Create(p))
	require.NoError(t, repo.Delete(p.PermID))
	_, err := repo.GetByID(p.PermID)
	assert.Equal(t, ErrNotFound, err)
}

func TestRound7BRepo_MockPerm_Delete_NotFound(t *testing.T) {
	repo := NewMockPermRepository()
	require.NoError(t, repo.Delete(999))
}

func TestRound7BRepo_MockPerm_GetByID(t *testing.T) {
	repo := NewMockPermRepository()
	p := &permission.SysPerm{PermName: "a", PermKey: "k_a", PermType: "data"}
	require.NoError(t, repo.Create(p))
	got, err := repo.GetByID(p.PermID)
	require.NoError(t, err)
	assert.Equal(t, "a", got.PermName)
}

func TestRound7BRepo_MockPerm_GetByID_NotFound(t *testing.T) {
	repo := NewMockPermRepository()
	_, err := repo.GetByID(42)
	assert.Equal(t, ErrNotFound, err)
}

func TestRound7BRepo_MockPerm_GetByKey(t *testing.T) {
	repo := NewMockPermRepository()
	p := &permission.SysPerm{PermName: "b", PermKey: "unique_key", PermType: "menu"}
	require.NoError(t, repo.Create(p))
	got, err := repo.GetByKey("unique_key")
	require.NoError(t, err)
	assert.Equal(t, "b", got.PermName)
}

func TestRound7BRepo_MockPerm_GetByKey_NotFound(t *testing.T) {
	repo := NewMockPermRepository()
	_, err := repo.GetByKey("missing")
	assert.Equal(t, ErrNotFound, err)
}

func TestRound7BRepo_MockPerm_List(t *testing.T) {
	repo := NewMockPermRepository()
	require.NoError(t, repo.Create(&permission.SysPerm{PermName: "p1", PermKey: "k1", PermType: "menu"}))
	require.NoError(t, repo.Create(&permission.SysPerm{PermName: "p2", PermKey: "k2", PermType: "button"}))
	list, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestRound7BRepo_MockPerm_List_Empty(t *testing.T) {
	repo := NewMockPermRepository()
	list, err := repo.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestRound7BRepo_MockPerm_CheckKeyExists(t *testing.T) {
	repo := NewMockPermRepository()
	p := &permission.SysPerm{PermName: "c", PermKey: "exists_key", PermType: "data"}
	require.NoError(t, repo.Create(p))
	count, err := repo.CheckKeyExists("exists_key", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRound7BRepo_MockPerm_CheckKeyExists_ExcludeSelf(t *testing.T) {
	repo := NewMockPermRepository()
	p := &permission.SysPerm{PermName: "d", PermKey: "self_key", PermType: "data"}
	require.NoError(t, repo.Create(p))
	count, err := repo.CheckKeyExists("self_key", p.PermID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRound7BRepo_MockPerm_GetByType(t *testing.T) {
	repo := NewMockPermRepository()
	require.NoError(t, repo.Create(&permission.SysPerm{PermName: "m1", PermKey: "km1", PermType: "menu"}))
	require.NoError(t, repo.Create(&permission.SysPerm{PermName: "b1", PermKey: "kb1", PermType: "button"}))
	require.NoError(t, repo.Create(&permission.SysPerm{PermName: "m2", PermKey: "km2", PermType: "menu"}))
	list, err := repo.GetByType("menu")
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestRound7BRepo_MockPerm_GetByType_Empty(t *testing.T) {
	repo := NewMockPermRepository()
	list, err := repo.GetByType("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, list)
}

// ==================== EmbeddedRepository (9 functions) ====================

func round7bEmbeddedDB(t *testing.T) *gorm.DB {
	return round7bOpenDB(t, &embedded.CoreEmbedded{})
}

func TestRound7BRepo_Embedded_New(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	require.NotNil(t, repo)
}

func TestRound7BRepo_Embedded_Create(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	e := &embedded.CoreEmbedded{
		Name: "app1", AppId: "app_001", AppSecret: "secret123",
		Domain: "http://localhost", SecretLength: 16, CreateTime: time.Now().UnixMilli(),
	}
	require.NoError(t, repo.Create(e))
	assert.Positive(t, e.ID)
}

func TestRound7BRepo_Embedded_Update(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	e := &embedded.CoreEmbedded{
		Name: "old", AppId: "app_002", AppSecret: "s", Domain: "", SecretLength: 8, CreateTime: 1,
	}
	require.NoError(t, repo.Create(e))
	e.Name = "updated"
	require.NoError(t, repo.Update(e))
	got, err := repo.GetByID(e.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated", got.Name)
}

func TestRound7BRepo_Embedded_Delete(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	e := &embedded.CoreEmbedded{Name: "del", AppId: "app_003", CreateTime: 1}
	require.NoError(t, repo.Create(e))
	require.NoError(t, repo.Delete(e.ID))
	_, err := repo.GetByID(e.ID)
	assert.Error(t, err)
}

func TestRound7BRepo_Embedded_DeleteBatch(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	e1 := &embedded.CoreEmbedded{Name: "b1", AppId: "b1", CreateTime: 1}
	e2 := &embedded.CoreEmbedded{Name: "b2", AppId: "b2", CreateTime: 1}
	require.NoError(t, repo.Create(e1))
	require.NoError(t, repo.Create(e2))
	require.NoError(t, repo.DeleteBatch([]int64{e1.ID, e2.ID}))
	_, err := repo.GetByID(e1.ID)
	assert.Error(t, err)
}

func TestRound7BRepo_Embedded_GetByID(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	e := &embedded.CoreEmbedded{Name: "get", AppId: "app_004", CreateTime: 1}
	require.NoError(t, repo.Create(e))
	got, err := repo.GetByID(e.ID)
	require.NoError(t, err)
	assert.Equal(t, "get", got.Name)
}

func TestRound7BRepo_Embedded_GetByID_NotFound(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	_, err := repo.GetByID(999)
	assert.Error(t, err)
}

func TestRound7BRepo_Embedded_GetByAppId(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	e := &embedded.CoreEmbedded{Name: "find", AppId: "app_unique", CreateTime: 1}
	require.NoError(t, repo.Create(e))
	got, err := repo.GetByAppId("app_unique")
	require.NoError(t, err)
	assert.Equal(t, "find", got.Name)
}

func TestRound7BRepo_Embedded_GetByAppId_NotFound(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	_, err := repo.GetByAppId("missing")
	assert.Error(t, err)
}

func TestRound7BRepo_Embedded_Query_NoKeyword(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	for i := 0; i < 3; i++ {
		require.NoError(t, repo.Create(&embedded.CoreEmbedded{
			Name: "item", AppId: "q1", CreateTime: int64(i),
		}))
	}
	list, total, err := repo.Query("", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, list, 3)
}

func TestRound7BRepo_Embedded_Query_WithKeyword(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	require.NoError(t, repo.Create(&embedded.CoreEmbedded{Name: "alpha", AppId: "kw1", CreateTime: 1}))
	require.NoError(t, repo.Create(&embedded.CoreEmbedded{Name: "beta", AppId: "kw2", CreateTime: 2}))
	list, total, err := repo.Query("alpha", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	assert.Equal(t, "alpha", list[0].Name)
}

func TestRound7BRepo_Embedded_Query_DefaultPaging(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	require.NoError(t, repo.Create(&embedded.CoreEmbedded{Name: "p", AppId: "pg1", CreateTime: 1}))
	list, total, err := repo.Query("", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}

func TestRound7BRepo_Embedded_Query_PageSizeCap(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	require.NoError(t, repo.Create(&embedded.CoreEmbedded{Name: "cap", AppId: "cap1", CreateTime: 1}))
	list, _, err := repo.Query("", 1, 200)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestRound7BRepo_Embedded_ListDistinctDomains(t *testing.T) {
	db := round7bEmbeddedDB(t)
	repo := NewEmbeddedRepository(db)
	require.NoError(t, repo.Create(&embedded.CoreEmbedded{Name: "d1", AppId: "dm1", Domain: "http://a.com", CreateTime: 1}))
	require.NoError(t, repo.Create(&embedded.CoreEmbedded{Name: "d2", AppId: "dm2", Domain: "http://b.com", CreateTime: 2}))
	require.NoError(t, repo.Create(&embedded.CoreEmbedded{Name: "d3", AppId: "dm3", Domain: "", CreateTime: 3}))
	domains, err := repo.ListDistinctDomains()
	require.NoError(t, err)
	assert.Len(t, domains, 2)
}

// ==================== TicketRepository (8 functions) ====================

func round7bTicketDB(t *testing.T) *gorm.DB {
	return round7bOpenDB(t, &coreTicket{})
}

func TestRound7BRepo_Ticket_New(t *testing.T) {
	db := round7bTicketDB(t)
	repo := NewTicketRepository(db)
	require.NotNil(t, repo)
}

func TestRound7BRepo_Ticket_Create(t *testing.T) {
	db := round7bTicketDB(t)
	repo := NewTicketRepository(db)
	tk := &ticket.Ticket{UUID: "u1", Ticket: "t1", Exp: 9999, Args: `{"k":"v"}`}
	require.NoError(t, repo.Create(tk))
	assert.Positive(t, tk.ID)
}

func TestRound7BRepo_Ticket_FindByTicket(t *testing.T) {
	db := round7bTicketDB(t)
	repo := NewTicketRepository(db)
	require.NoError(t, repo.Create(&ticket.Ticket{UUID: "u2", Ticket: "findme", Exp: 100}))
	got, err := repo.FindByTicket("findme")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u2", got.UUID)
}

func TestRound7BRepo_Ticket_FindByTicket_NotFound(t *testing.T) {
	db := round7bTicketDB(t)
	repo := NewTicketRepository(db)
	got, err := repo.FindByTicket("missing")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestRound7BRepo_Ticket_FindByID(t *testing.T) {
	db := round7bTicketDB(t)
	repo := NewTicketRepository(db)
	tk := &ticket.Ticket{UUID: "u3", Ticket: "id_find", Exp: 200}
	require.NoError(t, repo.Create(tk))
	got, err := repo.FindByID(tk.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "id_find", got.Ticket)
}

func TestRound7BRepo_Ticket_FindByID_NotFound(t *testing.T) {
	db := round7bTicketDB(t)
	repo := NewTicketRepository(db)
	got, err := repo.FindByID(99999)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestRound7BRepo_Ticket_Delete(t *testing.T) {
	db := round7bTicketDB(t)
	repo := NewTicketRepository(db)
	require.NoError(t, repo.Create(&ticket.Ticket{UUID: "u4", Ticket: "deleteme", Exp: 300}))
	require.NoError(t, repo.Delete("deleteme"))
	got, err := repo.FindByTicket("deleteme")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestRound7BRepo_Ticket_UpdateAccessTime(t *testing.T) {
	db := round7bTicketDB(t)
	repo := NewTicketRepository(db)
	require.NoError(t, repo.Create(&ticket.Ticket{UUID: "u5", Ticket: "uptime", Exp: 400}))
	require.NoError(t, repo.UpdateAccessTime("uptime", 12345))
	got, err := repo.FindByTicket("uptime")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(12345), got.AccessTime)
}

func TestRound7BRepo_Ticket_ListByUUID(t *testing.T) {
	db := round7bTicketDB(t)
	repo := NewTicketRepository(db)
	require.NoError(t, repo.Create(&ticket.Ticket{UUID: "shared", Ticket: "t_a", Exp: 1}))
	require.NoError(t, repo.Create(&ticket.Ticket{UUID: "shared", Ticket: "t_b", Exp: 2}))
	require.NoError(t, repo.Create(&ticket.Ticket{UUID: "other", Ticket: "t_c", Exp: 3}))
	list, total, err := repo.ListByUUID("shared", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
}

func TestRound7BRepo_Ticket_ListByUUID_Empty(t *testing.T) {
	db := round7bTicketDB(t)
	repo := NewTicketRepository(db)
	list, total, err := repo.ListByUUID("nothing", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, list)
}

// ==================== LinkageRepository (8 functions) ====================

func round7bLinkageDB(t *testing.T) *gorm.DB {
	return round7bOpenDB(t,
		&auto.SnapshotVisualizationLinkage{},
		&auto.SnapshotVisualizationLinkageField{},
		&auto.VisualizationLinkage{},
		&auto.VisualizationLinkageField{},
		&visualization.CanvasChartView{},
		&visualization.SnapshotCanvasChartView{},
		&dataset.CoreDatasetTableField{},
	)
}

func TestRound7BRepo_Linkage_New(t *testing.T) {
	db := round7bLinkageDB(t)
	repo := NewLinkageRepository(db)
	require.NotNil(t, repo)
}

func TestRound7BRepo_Linkage_GetViewLinkageGather_EmptyTargets(t *testing.T) {
	db := round7bLinkageDB(t)
	repo := NewLinkageRepository(db)
	rows, err := repo.GetViewLinkageGather(1, 10, nil, false)
	require.NoError(t, err)
	assert.Nil(t, rows)
}

func TestRound7BRepo_Linkage_CreateLinkage(t *testing.T) {
	db := round7bLinkageDB(t)
	repo := NewLinkageRepository(db)
	l := &auto.SnapshotVisualizationLinkage{
		DvID: 1, SourceViewID: 10, TargetViewID: 20, LinkageActive: true,
	}
	require.NoError(t, repo.CreateLinkage(l))
	assert.Positive(t, l.ID)
}

func TestRound7BRepo_Linkage_CreateLinkageField(t *testing.T) {
	db := round7bLinkageDB(t)
	repo := NewLinkageRepository(db)
	l := &auto.SnapshotVisualizationLinkage{
		DvID: 2, SourceViewID: 11, TargetViewID: 21, LinkageActive: true,
	}
	require.NoError(t, repo.CreateLinkage(l))
	f := &auto.SnapshotVisualizationLinkageField{
		LinkageID: l.ID, SourceField: 100, TargetField: 200,
	}
	require.NoError(t, repo.CreateLinkageField(f))
	assert.Positive(t, f.ID)
}

func TestRound7BRepo_Linkage_DeleteLinkageAndFields(t *testing.T) {
	db := round7bLinkageDB(t)
	repo := NewLinkageRepository(db)
	l := &auto.SnapshotVisualizationLinkage{
		DvID: 3, SourceViewID: 12, TargetViewID: 22, LinkageActive: true,
	}
	require.NoError(t, repo.CreateLinkage(l))
	f := &auto.SnapshotVisualizationLinkageField{
		LinkageID: l.ID, SourceField: 101, TargetField: 201,
	}
	require.NoError(t, repo.CreateLinkageField(f))
	require.NoError(t, repo.DeleteLinkageAndFields(3, 12))
}

func TestRound7BRepo_Linkage_GetDatasetFieldsByGroupID(t *testing.T) {
	db := round7bLinkageDB(t)
	repo := NewLinkageRepository(db)
	fields, err := repo.GetDatasetFieldsByGroupID(999)
	require.NoError(t, err)
	assert.Empty(t, fields)
}

func TestRound7BRepo_Linkage_GetAllLinkageInfo_Empty(t *testing.T) {
	db := round7bLinkageDB(t)
	repo := NewLinkageRepository(db)
	result, err := repo.GetAllLinkageInfo(1, false)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRound7BRepo_Linkage_UpdateChartLinkageActive(t *testing.T) {
	db := round7bLinkageDB(t)
	repo := NewLinkageRepository(db)
	require.NoError(t, repo.UpdateChartLinkageActive(42, true))
}

func TestRound7BRepo_Linkage_GetViewLinkageGather_WithTargets(t *testing.T) {
	db := round7bLinkageDB(t)
	repo := NewLinkageRepository(db)
	l := &auto.SnapshotVisualizationLinkage{
		DvID: 100, SourceViewID: 50, TargetViewID: 60, LinkageActive: true,
	}
	require.NoError(t, repo.CreateLinkage(l))
	f := &auto.SnapshotVisualizationLinkageField{
		LinkageID: l.ID, SourceField: 500, TargetField: 600,
	}
	require.NoError(t, repo.CreateLinkageField(f))

	rows, err := repo.GetViewLinkageGather(100, 50, []int64{60}, true)
	require.NoError(t, err)
	_ = rows
}

// ==================== DriverRepository (5 functions) ====================

func round7bDriverDB(t *testing.T) *gorm.DB {
	return round7bOpenDB(t, &driver.Driver{}, &driver.DriverJar{})
}

func TestRound7BRepo_Driver_New(t *testing.T) {
	db := round7bDriverDB(t)
	repo := NewDriverRepository(db)
	require.NotNil(t, repo)
}

func TestRound7BRepo_Driver_List_Empty(t *testing.T) {
	db := round7bDriverDB(t)
	repo := NewDriverRepository(db)
	list, err := repo.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestRound7BRepo_Driver_List_WithData(t *testing.T) {
	db := round7bDriverDB(t)
	repo := NewDriverRepository(db)
	require.NoError(t, db.Create(&driver.Driver{Name: "mysql", Type: "mysql"}).Error)
	require.NoError(t, db.Create(&driver.Driver{Name: "pg", Type: "postgresql"}).Error)
	list, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestRound7BRepo_Driver_ListByType(t *testing.T) {
	db := round7bDriverDB(t)
	repo := NewDriverRepository(db)
	require.NoError(t, db.Create(&driver.Driver{Name: "mysql1", Type: "mysql"}).Error)
	require.NoError(t, db.Create(&driver.Driver{Name: "pg1", Type: "postgresql"}).Error)
	require.NoError(t, db.Create(&driver.Driver{Name: "mysql2", Type: "mysql"}).Error)
	list, err := repo.ListByType("mysql")
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestRound7BRepo_Driver_GetByID(t *testing.T) {
	db := round7bDriverDB(t)
	repo := NewDriverRepository(db)
	d := &driver.Driver{Name: "getme", Type: "oracle"}
	require.NoError(t, db.Create(d).Error)
	got, err := repo.GetByID(d.ID)
	require.NoError(t, err)
	assert.Equal(t, "getme", got.Name)
}

func TestRound7BRepo_Driver_GetByID_NotFound(t *testing.T) {
	db := round7bDriverDB(t)
	repo := NewDriverRepository(db)
	_, err := repo.GetByID(999)
	assert.Error(t, err)
}

func TestRound7BRepo_Driver_ListDriverJars(t *testing.T) {
	db := round7bDriverDB(t)
	repo := NewDriverRepository(db)
	d := &driver.Driver{Name: "jar_test", Type: "mysql"}
	require.NoError(t, db.Create(d).Error)
	require.NoError(t, db.Create(&driver.DriverJar{DriverID: d.ID, FileName: "a.jar", FilePath: "/a.jar"}).Error)
	require.NoError(t, db.Create(&driver.DriverJar{DriverID: d.ID, FileName: "b.jar", FilePath: "/b.jar"}).Error)
	require.NoError(t, db.Create(&driver.DriverJar{DriverID: 9999, FileName: "c.jar", FilePath: "/c.jar"}).Error)
	list, err := repo.ListDriverJars(d.ID)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

// ==================== LicenseRepository (4 functions) ====================

func round7bLicenseDB(t *testing.T) *gorm.DB {
	return round7bOpenDB(t, &coreSysSetting{})
}

func TestRound7BRepo_License_New(t *testing.T) {
	db := round7bLicenseDB(t)
	repo := NewLicenseRepository(db)
	require.NotNil(t, repo)
}

func TestRound7BRepo_License_Load_Empty(t *testing.T) {
	db := round7bLicenseDB(t)
	repo := NewLicenseRepository(db)
	result, raw, err := repo.Load()
	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Empty(t, raw)
}

func TestRound7BRepo_License_SaveAndLoad(t *testing.T) {
	db := round7bLicenseDB(t)
	repo := NewLicenseRepository(db)
	saved := &license.ValidateResult{
		Status:  "valid",
		Message: "ok",
		License: &license.LicenseInfo{
			Corporation: "TestCorp",
			Expired:     "2026-12-31",
			Count:       100,
			Version:     "v2",
			Edition:     "enterprise",
			SerialNo:    "SN001",
			Remark:      "test",
			ISV:         "TestISV",
		},
	}
	require.NoError(t, repo.Save(saved, "raw-license-data"))

	result, raw, err := repo.Load()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "valid", result.Status)
	assert.Equal(t, "ok", result.Message)
	require.NotNil(t, result.License)
	assert.Equal(t, "TestCorp", result.License.Corporation)
	assert.Equal(t, int64(100), result.License.Count)
	assert.Equal(t, "raw-license-data", raw)
}

func TestRound7BRepo_License_Save_Minimal(t *testing.T) {
	db := round7bLicenseDB(t)
	repo := NewLicenseRepository(db)
	saved := &license.ValidateResult{Status: "no_license"}
	require.NoError(t, repo.Save(saved, ""))
	result, _, err := repo.Load()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "no_license", result.Status)
	assert.Nil(t, result.License)
}

func TestRound7BRepo_License_Clear(t *testing.T) {
	db := round7bLicenseDB(t)
	repo := NewLicenseRepository(db)
	require.NoError(t, repo.Save(&license.ValidateResult{Status: "valid"}, "raw"))
	require.NoError(t, repo.Clear())
	result, raw, err := repo.Load()
	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Empty(t, raw)
}

// ==================== CommitLogRepository (4 functions) ====================

func round7bCommitLogDB(t *testing.T) *gorm.DB {
	return round7bOpenDB(t, &datafillingdomain.DfCommitLog{})
}

func TestRound7BRepo_CommitLog_New(t *testing.T) {
	db := round7bCommitLogDB(t)
	repo := NewCommitLogRepository(db)
	require.NotNil(t, repo)
}

func TestRound7BRepo_CommitLog_Create(t *testing.T) {
	db := round7bCommitLogDB(t)
	repo := NewCommitLogRepository(db)
	log := &datafillingdomain.DfCommitLog{
		FormID: 1, DataID: "d1", Operate: 1, CommitBy: 100,
		Committer: "user1", CommitTime: time.Now().UnixMilli(), Count: 5,
	}
	require.NoError(t, repo.Create(context.Background(), log))
	assert.Positive(t, log.ID)
}

func TestRound7BRepo_CommitLog_ListByFormID(t *testing.T) {
	db := round7bCommitLogDB(t)
	repo := NewCommitLogRepository(db)
	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Create(context.Background(), &datafillingdomain.DfCommitLog{
			FormID: 10, DataID: "d10", Operate: 1, CommitBy: 1,
			Committer: "u", CommitTime: int64(i), Count: 1,
		}))
	}
	require.NoError(t, repo.Create(context.Background(), &datafillingdomain.DfCommitLog{
		FormID: 99, DataID: "d99", Operate: 1, CommitBy: 1,
		Committer: "u", CommitTime: 1, Count: 1,
	}))
	rows, total, err := repo.ListByFormID(context.Background(), 10, 1, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, rows, 3)
}

func TestRound7BRepo_CommitLog_ListByFormID_Page2(t *testing.T) {
	db := round7bCommitLogDB(t)
	repo := NewCommitLogRepository(db)
	for i := 0; i < 3; i++ {
		require.NoError(t, repo.Create(context.Background(), &datafillingdomain.DfCommitLog{
			FormID: 20, DataID: "d20", Operate: 1, CommitBy: 1,
			Committer: "u", CommitTime: int64(i), Count: 1,
		}))
	}
	rows, total, err := repo.ListByFormID(context.Background(), 20, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, rows, 1)
}

func TestRound7BRepo_CommitLog_DeleteByFormID(t *testing.T) {
	db := round7bCommitLogDB(t)
	repo := NewCommitLogRepository(db)
	require.NoError(t, repo.Create(context.Background(), &datafillingdomain.DfCommitLog{
		FormID: 30, DataID: "d30", Operate: 1, CommitBy: 1,
		Committer: "u", CommitTime: 1, Count: 1,
	}))
	require.NoError(t, repo.Create(context.Background(), &datafillingdomain.DfCommitLog{
		FormID: 31, DataID: "d31", Operate: 1, CommitBy: 1,
		Committer: "u", CommitTime: 1, Count: 1,
	}))
	require.NoError(t, repo.DeleteByFormID(context.Background(), 30))
	rows, total, err := repo.ListByFormID(context.Background(), 30, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, rows)

	_, total, err = repo.ListByFormID(context.Background(), 31, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
}

// ==================== WatermarkRepository (3 functions) ====================

func round7bWatermarkDB(t *testing.T) *gorm.DB {
	return round7bOpenDB(t, &visualization.Watermark{})
}

func TestRound7BRepo_Watermark_New(t *testing.T) {
	db := round7bWatermarkDB(t)
	repo := NewWatermarkRepository(db)
	require.NotNil(t, repo)
}

func TestRound7BRepo_Watermark_FindLatest_Empty(t *testing.T) {
	db := round7bWatermarkDB(t)
	repo := NewWatermarkRepository(db)
	result, err := repo.FindLatest()
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRound7BRepo_Watermark_SaveDefault_Insert(t *testing.T) {
	db := round7bWatermarkDB(t)
	repo := NewWatermarkRepository(db)
	wm, err := repo.SaveDefault(`{"setting":"val"}`, "admin", 1000)
	require.NoError(t, err)
	assert.Equal(t, "default", wm.ID)
	assert.Equal(t, `{"setting":"val"}`, wm.SettingContent)
}

func TestRound7BRepo_Watermark_FindLatest_AfterSave(t *testing.T) {
	db := round7bWatermarkDB(t)
	repo := NewWatermarkRepository(db)
	_, err := repo.SaveDefault(`{"x":1}`, "admin", 2000)
	require.NoError(t, err)
	result, err := repo.FindLatest()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "default", result.ID)
	assert.Equal(t, `{"x":1}`, result.SettingContent)
}

func TestRound7BRepo_Watermark_SaveDefault_Upsert(t *testing.T) {
	db := round7bWatermarkDB(t)
	repo := NewWatermarkRepository(db)
	_, err := repo.SaveDefault(`{"v":1}`, "admin", 3000)
	require.NoError(t, err)
	wm, err := repo.SaveDefault(`{"v":2}`, "admin2", 4000)
	require.NoError(t, err)
	assert.Equal(t, `{"v":2}`, wm.SettingContent)
}
