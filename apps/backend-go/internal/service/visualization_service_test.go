package service

import (
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVisualizationServiceRepoTest(t *testing.T) (*VisualizationService, *repository.VisualizationRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&visualization.DataVisualizationInfo{}))

	repo := repository.NewVisualizationRepository(db)
	return NewVisualizationService(repo), repo, db
}

func setupVisualizationServiceWithPermTest(t *testing.T) (*VisualizationService, *repository.VisualizationRepository, *mockResourcePermRepo, *gorm.DB) {
	t.Helper()

	svc, repo, db := setupVisualizationServiceRepoTest(t)
	permRepo := newMockResourcePermRepo()
	permSvc := NewResourcePermissionService(permRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})
	svc.SetResourcePermissionService(permSvc)
	return svc, repo, permRepo, db
}

func closeVisualizationDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func int64Ptr(v int64) *int64 { return &v }

func intPtrVisualization(v int) *int { return &v }

func TestVisualizationServiceHelpers(t *testing.T) {
	t.Run("resolve interactive visualization types", func(t *testing.T) {
		types, err := resolveInteractiveVisualizationTypes("")
		require.NoError(t, err)
		assert.Equal(t, []string{"dashboard", "dataV"}, types)

		types, err = resolveInteractiveVisualizationTypes("panel")
		require.NoError(t, err)
		assert.Equal(t, []string{"dashboard"}, types)

		types, err = resolveInteractiveVisualizationTypes("screen")
		require.NoError(t, err)
		assert.Equal(t, []string{"dataV"}, types)

		types, err = resolveInteractiveVisualizationTypes("bad")
		require.Error(t, err)
		assert.Nil(t, types)
	})

	t.Run("normalize visualization resource type", func(t *testing.T) {
		assert.Equal(t, permission.ResourceTypeDashboard, normalizeVisualizationResourceType(nil))
		assert.Equal(t, permission.ResourceTypeDashboard, normalizeVisualizationResourceType(strPtr("dashboard")))
		assert.Equal(t, permission.ResourceTypeScreen, normalizeVisualizationResourceType(strPtr("dataV")))
		assert.Equal(t, permission.ResourceTypeScreen, normalizeVisualizationResourceType(strPtr(permission.ResourceTypeScreen)))
	})
}

func TestVisualizationService_Save_Defaults(t *testing.T) {
	t.Run("panel defaults status zero", func(t *testing.T) {
		svc, repo, _ := setupVisualizationServiceRepoTest(t)

		id, err := svc.Save(&visualization.SaveRequest{Name: "Panel Default"}, "tester")
		require.NoError(t, err)

		item, err := repo.GetByID(id)
		require.NoError(t, err)
		require.NotNil(t, item.NodeType)
		require.NotNil(t, item.Status)
		assert.Equal(t, "panel", *item.NodeType)
		assert.Equal(t, 0, *item.Status)
	})

	t.Run("folder defaults status one", func(t *testing.T) {
		svc, repo, _ := setupVisualizationServiceRepoTest(t)
		nodeType := "folder"

		id, err := svc.Save(&visualization.SaveRequest{Name: "Folder Default", NodeType: &nodeType}, "tester")
		require.NoError(t, err)

		item, err := repo.GetByID(id)
		require.NoError(t, err)
		require.NotNil(t, item.Status)
		assert.Equal(t, 1, *item.Status)
	})
}

func TestVisualizationService_Copy_Validation(t *testing.T) {
	svc, _, _ := setupVisualizationServiceRepoTest(t)

	_, err := svc.Copy(nil, "tester")
	assert.EqualError(t, err, "copy request is required")

	_, err = svc.Copy(&visualization.CopyRequest{Name: "copy"}, "tester")
	assert.EqualError(t, err, "source id is required")

	_, err = svc.Copy(&visualization.CopyRequest{ID: 1}, "tester")
	assert.EqualError(t, err, "name is required")
}

func TestVisualizationService_Detail_DeleteLogic_FindDvType(t *testing.T) {
	t.Run("detail delegates get by id", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)
		id, err := svc.Save(&visualization.SaveRequest{Name: "Detail Item"}, "tester")
		require.NoError(t, err)

		item, err := svc.Detail(&visualization.DetailRequest{ID: id})
		require.NoError(t, err)
		assert.Equal(t, "Detail Item", item.Name)
	})

	t.Run("delete logic delegates", func(t *testing.T) {
		svc, repo, _ := setupVisualizationServiceRepoTest(t)
		id, err := svc.Save(&visualization.SaveRequest{Name: "Delete Item"}, "tester")
		require.NoError(t, err)

		require.NoError(t, svc.DeleteLogic(id, "tester"))
		item, err := repo.GetByID(id)
		require.Error(t, err)
		assert.Nil(t, item)
	})

	t.Run("find dv type nil type returns empty", func(t *testing.T) {
		svc, _, db := setupVisualizationServiceRepoTest(t)
		nodeType := "panel"
		status := 0
		now := int64(1)
		createBy := "tester"
		require.NoError(t, db.Create(&visualization.DataVisualizationInfo{Name: "NilType", NodeType: &nodeType, Status: &status, CreateTime: &now, UpdateTime: &now, CreateBy: &createBy, UpdateBy: &createBy}).Error)

		var item visualization.DataVisualizationInfo
		require.NoError(t, db.Where("name = ?", "NilType").First(&item).Error)

		typ, err := svc.FindDvType(item.ID)
		require.NoError(t, err)
		assert.Equal(t, "", typ)
	})
}

func TestVisualizationService_ListInteractiveNameCanvas(t *testing.T) {
	t.Run("list normalizes paging", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)

		resp, err := svc.List(&visualization.ListRequest{Current: 0, Size: 0})
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Current)
		assert.Equal(t, 10, resp.Size)
		assert.Empty(t, resp.List)
	})

	t.Run("interactive tree default and unsupported", func(t *testing.T) {
		svc, _, db := setupVisualizationServiceRepoTest(t)
		nodeType := "panel"
		status := 0
		now := int64(1)
		createBy := "tester"
		dashboard := "dashboard"
		screen := "dataV"
		require.NoError(t, db.Create(&visualization.DataVisualizationInfo{Name: "Dashboard Item", Type: &dashboard, NodeType: &nodeType, Status: &status, CreateTime: &now, UpdateTime: &now, CreateBy: &createBy, UpdateBy: &createBy}).Error)
		require.NoError(t, db.Create(&visualization.DataVisualizationInfo{Name: "Screen Item", Type: &screen, NodeType: &nodeType, Status: &status, CreateTime: &now, UpdateTime: &now, CreateBy: &createBy, UpdateBy: &createBy}).Error)

		items, err := svc.InteractiveTree("")
		require.NoError(t, err)
		assert.Len(t, items, 2)

		items, err = svc.InteractiveTree("bad")
		require.Error(t, err)
		assert.Nil(t, items)
	})

	t.Run("name check nil request returns success", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)
		status, err := svc.NameCheck(nil)
		require.NoError(t, err)
		assert.Equal(t, "success", status)
	})

	t.Run("check canvas change nil or zero id returns empty", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)

		status, err := svc.CheckCanvasChange(nil)
		require.NoError(t, err)
		assert.Equal(t, "", status)

		status, err = svc.CheckCanvasChange(&visualization.CanvasChangeRequest{ID: 0})
		require.NoError(t, err)
		assert.Equal(t, "", status)
	})
}

func TestVisualizationService_UpdateMovePublishRecover(t *testing.T) {
	t.Run("update not found", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)
		err := svc.Update(&visualization.UpdateRequest{ID: 999, Name: strPtr("missing")}, "tester")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "visualization not found")
	})

	t.Run("update only provided fields change", func(t *testing.T) {
		svc, repo, _ := setupVisualizationServiceRepoTest(t)
		typ := "dashboard"
		mobile := true
		content := "c1"
		checkVersion := "v1"
		id, err := svc.Save(&visualization.SaveRequest{Name: "Original", Type: &typ, MobileLayout: &mobile, ContentID: &content, CheckVersion: &checkVersion}, "creator")
		require.NoError(t, err)

		newName := "Updated"
		err = svc.Update(&visualization.UpdateRequest{ID: id, Name: &newName}, "updater")
		require.NoError(t, err)

		item, err := repo.GetByID(id)
		require.NoError(t, err)
		assert.Equal(t, "Updated", item.Name)
		require.NotNil(t, item.Type)
		assert.Equal(t, "dashboard", *item.Type)
		require.NotNil(t, item.ContentID)
		assert.Equal(t, "c1", *item.ContentID)
	})

	t.Run("move nil request returns nil and delegates update", func(t *testing.T) {
		svc, repo, _ := setupVisualizationServiceRepoTest(t)
		require.NoError(t, svc.Move(nil, "tester"))

		id, err := svc.Save(&visualization.SaveRequest{Name: "Move Me"}, "creator")
		require.NoError(t, err)
		newPID := int64(123)
		require.NoError(t, svc.Move(&visualization.MoveRequest{ID: id, PID: &newPID}, "mover"))

		item, err := repo.GetByID(id)
		require.NoError(t, err)
		require.NotNil(t, item.PID)
		assert.Equal(t, newPID, *item.PID)
	})

	t.Run("update publish status returns reloaded entity and recover sets status one", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)
		id, err := svc.Save(&visualization.SaveRequest{Name: "Publish Me"}, "creator")
		require.NoError(t, err)

		status := 2
		item, err := svc.UpdatePublishStatus(&visualization.UpdateRequest{ID: id, Status: intPtrVisualization(status)}, "publisher")
		require.NoError(t, err)
		require.NotNil(t, item.Status)
		assert.Equal(t, 2, *item.Status)

		recovered, err := svc.RecoverToPublished(id, "recover")
		require.NoError(t, err)
		require.NotNil(t, recovered.Status)
		assert.Equal(t, 1, *recovered.Status)
	})
}

func TestVisualizationService_CopyNameCheckCanvasAndBackfill(t *testing.T) {
	t.Run("copy success", func(t *testing.T) {
		svc, repo, _ := setupVisualizationServiceRepoTest(t)
		typ := "dashboard"
		nodeType := "panel"
		mobile := true
		content := "cid-1"
		checkVersion := "ver-1"
		sourceID, err := svc.Save(&visualization.SaveRequest{Name: "Source", Type: &typ, NodeType: &nodeType, MobileLayout: &mobile, ContentID: &content, CheckVersion: &checkVersion}, "creator")
		require.NoError(t, err)

		newPID := int64(99)
		copyID, err := svc.Copy(&visualization.CopyRequest{ID: sourceID, Name: "Copy", PID: &newPID}, "copier")
		require.NoError(t, err)
		assert.NotEqual(t, sourceID, copyID)

		item, err := repo.GetByID(copyID)
		require.NoError(t, err)
		assert.Equal(t, "Copy", item.Name)
		require.NotNil(t, item.PID)
		assert.Equal(t, newPID, *item.PID)
		require.NotNil(t, item.Type)
		assert.Equal(t, "dashboard", *item.Type)
		require.NotNil(t, item.ContentID)
		assert.Equal(t, "cid-1", *item.ContentID)
	})

	t.Run("name check repeat and success", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)
		pid := int64(7)
		id, err := svc.Save(&visualization.SaveRequest{Name: "Existing", PID: &pid}, "creator")
		require.NoError(t, err)

		status, err := svc.NameCheck(&visualization.NameCheckRequest{Name: "Existing", PID: &pid})
		require.NoError(t, err)
		assert.Equal(t, "repeat", status)

		status, err = svc.NameCheck(&visualization.NameCheckRequest{ID: id, Name: "Existing", PID: &pid})
		require.NoError(t, err)
		assert.Equal(t, "success", status)

		status, err = svc.NameCheck(&visualization.NameCheckRequest{Name: "Other", PID: &pid})
		require.NoError(t, err)
		assert.Equal(t, "success", status)
	})

	t.Run("check canvas change content and version repeat", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)
		content := "cid-1"
		checkVersion := "ver-1"
		id, err := svc.Save(&visualization.SaveRequest{Name: "Canvas", ContentID: &content, CheckVersion: &checkVersion}, "creator")
		require.NoError(t, err)

		status, err := svc.CheckCanvasChange(&visualization.CanvasChangeRequest{ID: id, ContentID: strPtr("cid-2")})
		require.NoError(t, err)
		assert.Equal(t, "Repeat", status)

		status, err = svc.CheckCanvasChange(&visualization.CanvasChangeRequest{ID: id, CheckVersion: strPtr("ver-2")})
		require.NoError(t, err)
		assert.Equal(t, "Repeat", status)

		status, err = svc.CheckCanvasChange(&visualization.CanvasChangeRequest{ID: id, ContentID: &content, CheckVersion: &checkVersion})
		require.NoError(t, err)
		assert.Equal(t, "", status)
	})

	t.Run("find dv type returns type", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)
		typ := "dataV"
		id, err := svc.Save(&visualization.SaveRequest{Name: "Type Item", Type: &typ}, "creator")
		require.NoError(t, err)

		found, err := svc.FindDvType(id)
		require.NoError(t, err)
		assert.Equal(t, "dataV", found)
	})

	t.Run("backfill guards", func(t *testing.T) {
		svc := &VisualizationService{}
		report, err := svc.BackfillGovernedVisualizationResourcesWithOptions(nil)
		require.Error(t, err)
		assert.Nil(t, report)
		assert.EqualError(t, err, "visualization repository not initialized")

		svc, _, _ = setupVisualizationServiceRepoTest(t)
		report, err = svc.BackfillGovernedResources()
		require.Error(t, err)
		assert.Nil(t, report)
		assert.EqualError(t, err, "resource permission service not initialized")
	})
}

func TestVisualizationService_ListInteractiveAndErrorBranches(t *testing.T) {
	t.Run("list preserves explicit paging", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)

		resp, err := svc.List(&visualization.ListRequest{Current: 2, Size: 20})
		require.NoError(t, err)
		assert.Equal(t, 2, resp.Current)
		assert.Equal(t, 20, resp.Size)
	})

	t.Run("list repo error", func(t *testing.T) {
		svc, _, db := setupVisualizationServiceRepoTest(t)
		closeVisualizationDB(t, db)

		resp, err := svc.List(&visualization.ListRequest{Current: 1, Size: 10})
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("interactive tree dashboard only and screen only", func(t *testing.T) {
		svc, _, db := setupVisualizationServiceRepoTest(t)
		nodeType := "panel"
		status := 0
		now := int64(1)
		createBy := "tester"
		dashboard := "dashboard"
		screen := "dataV"
		require.NoError(t, db.Create(&visualization.DataVisualizationInfo{Name: "Dashboard Item", Type: &dashboard, NodeType: &nodeType, Status: &status, CreateTime: &now, UpdateTime: &now, CreateBy: &createBy, UpdateBy: &createBy}).Error)
		require.NoError(t, db.Create(&visualization.DataVisualizationInfo{Name: "Screen Item", Type: &screen, NodeType: &nodeType, Status: &status, CreateTime: &now, UpdateTime: &now, CreateBy: &createBy, UpdateBy: &createBy}).Error)

		items, err := svc.InteractiveTree("dashboard")
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.NotNil(t, items[0].Type)
		assert.Equal(t, "dashboard", *items[0].Type)

		items, err = svc.InteractiveTree("dataV")
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.NotNil(t, items[0].Type)
		assert.Equal(t, "dataV", *items[0].Type)
	})
}

func TestVisualizationService_NameCanvasAndDelegationErrors(t *testing.T) {
	t.Run("name check repo error", func(t *testing.T) {
		svc, _, db := setupVisualizationServiceRepoTest(t)
		closeVisualizationDB(t, db)

		status, err := svc.NameCheck(&visualization.NameCheckRequest{Name: "Any"})
		require.Error(t, err)
		assert.Equal(t, "", status)
	})

	t.Run("check canvas get by id error", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)

		status, err := svc.CheckCanvasChange(&visualization.CanvasChangeRequest{ID: 999, ContentID: strPtr("x")})
		require.Error(t, err)
		assert.Equal(t, "", status)
	})

	t.Run("check canvas blank incoming and missing stored values do not repeat", func(t *testing.T) {
		svc, _, db := setupVisualizationServiceRepoTest(t)
		nodeType := "panel"
		status := 0
		now := int64(1)
		createBy := "tester"
		require.NoError(t, db.Create(&visualization.DataVisualizationInfo{Name: "Canvas Nil", NodeType: &nodeType, Status: &status, CreateTime: &now, UpdateTime: &now, CreateBy: &createBy, UpdateBy: &createBy}).Error)

		var item visualization.DataVisualizationInfo
		require.NoError(t, db.Where("name = ?", "Canvas Nil").First(&item).Error)

		result, err := svc.CheckCanvasChange(&visualization.CanvasChangeRequest{ID: item.ID, ContentID: strPtr(""), CheckVersion: strPtr("")})
		require.NoError(t, err)
		assert.Equal(t, "", result)

		content := "cid-1"
		version := "ver-1"
		secondID, saveErr := svc.Save(&visualization.SaveRequest{Name: "Canvas Some", ContentID: &content, CheckVersion: &version}, "creator")
		require.NoError(t, saveErr)

		result, err = svc.CheckCanvasChange(&visualization.CanvasChangeRequest{ID: secondID, ContentID: strPtr(""), CheckVersion: &version})
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("update base delegates to update", func(t *testing.T) {
		svc, repo, _ := setupVisualizationServiceRepoTest(t)
		id, err := svc.Save(&visualization.SaveRequest{Name: "Base Update"}, "creator")
		require.NoError(t, err)

		name := "Base Updated"
		require.NoError(t, svc.UpdateBase(&visualization.UpdateRequest{ID: id, Name: &name}, "updater"))
		item, err := repo.GetByID(id)
		require.NoError(t, err)
		assert.Equal(t, "Base Updated", item.Name)
	})

	t.Run("move publish recover propagate update errors", func(t *testing.T) {
		svc, _, db := setupVisualizationServiceRepoTest(t)
		closeVisualizationDB(t, db)
		pid := int64(10)
		statusVal := 1

		err := svc.Move(&visualization.MoveRequest{ID: 1, PID: &pid}, "tester")
		require.Error(t, err)

		item, err := svc.UpdatePublishStatus(&visualization.UpdateRequest{ID: 1, Status: intPtrVisualization(statusVal)}, "tester")
		require.Error(t, err)
		assert.Nil(t, item)

		item, err = svc.RecoverToPublished(1, "tester")
		require.Error(t, err)
		assert.Nil(t, item)
	})
}

func TestVisualizationService_InheritanceAndBackfill(t *testing.T) {
	t.Run("save inherits parent permissions when parent governed", func(t *testing.T) {
		svc, _, permRepo, _ := setupVisualizationServiceWithPermTest(t)
		parentType := "dashboard"
		parentID, err := svc.Save(&visualization.SaveRequest{Name: "Governed Parent", Type: &parentType}, "creator")
		require.NoError(t, err)
		permRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDashboard, parentID)] = []int64{11, 12}

		childID, err := svc.Save(&visualization.SaveRequest{Name: "Governed Child", Type: &parentType, PID: int64Ptr(parentID)}, "creator")
		require.NoError(t, err)
		permIDs, exists, permErr := permRepo.GetResourcePermissionIDs(childID, permission.ResourceTypeDashboard)
		require.NoError(t, permErr)
		assert.True(t, exists)
		assert.Equal(t, []int64{11, 12}, permIDs)
		assert.Equal(t, 1, permRepo.registerCalls)
		assert.Equal(t, 1, permRepo.replaceCalls)
	})

	t.Run("save skips inheritance when pid nil or zero", func(t *testing.T) {
		svc, _, permRepo, _ := setupVisualizationServiceWithPermTest(t)
		parentType := "dashboard"

		_, err := svc.Save(&visualization.SaveRequest{Name: "No Parent", Type: &parentType}, "creator")
		require.NoError(t, err)
		zero := int64(0)
		_, err = svc.Save(&visualization.SaveRequest{Name: "Zero Parent", Type: &parentType, PID: &zero}, "creator")
		require.NoError(t, err)
		assert.Zero(t, permRepo.registerCalls)
		assert.Zero(t, permRepo.replaceCalls)
	})

	t.Run("save rolls back when inheritance fails", func(t *testing.T) {
		svc, _, permRepo, db := setupVisualizationServiceWithPermTest(t)
		parentType := "dashboard"
		parentID, err := svc.Save(&visualization.SaveRequest{Name: "Rollback Parent", Type: &parentType}, "creator")
		require.NoError(t, err)
		permRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDashboard, parentID)] = []int64{21}
		permRepo.registerErr = assert.AnError

		childID, err := svc.Save(&visualization.SaveRequest{Name: "Rollback Child", Type: &parentType, PID: int64Ptr(parentID)}, "creator")
		require.Error(t, err)
		assert.Zero(t, childID)

		var count int64
		require.NoError(t, db.Model(&visualization.DataVisualizationInfo{}).Where("name = ? AND COALESCE(delete_flag, 0) = 0", "Rollback Child").Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("backfill skips missing parent and parent not governed and governs inherited items", func(t *testing.T) {
		svc, repo, permRepo, db := setupVisualizationServiceWithPermTest(t)
		parentType := "dashboard"
		orphanPID := int64(9999)

		parentID, err := svc.Save(&visualization.SaveRequest{Name: "Backfill Parent", Type: &parentType}, "creator")
		require.NoError(t, err)
		ungovernedParentID, err := svc.Save(&visualization.SaveRequest{Name: "Ungoverned Parent", Type: &parentType}, "creator")
		require.NoError(t, err)
		permRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDashboard, parentID)] = []int64{31, 32}

		require.NoError(t, repo.Create(&visualization.DataVisualizationInfo{Name: "Orphan Child", PID: &orphanPID, Type: &parentType}))
		require.NoError(t, repo.Create(&visualization.DataVisualizationInfo{Name: "Ungoverned Child", PID: int64Ptr(ungovernedParentID), Type: &parentType}))
		require.NoError(t, repo.Create(&visualization.DataVisualizationInfo{Name: "Governed Child", PID: int64Ptr(parentID), Type: &parentType}))

		report, err := svc.BackfillGovernedVisualizationResourcesWithOptions(&GovernanceBackfillOptions{Limit: 10})
		require.NoError(t, err)
		require.NotNil(t, report)
		assert.Equal(t, 5, report.Scanned)
		assert.Equal(t, 1, report.Governed)
		assert.Equal(t, 4, report.Skipped)
		assert.Len(t, report.ResourceIDs, 1)
		assert.Len(t, report.SkippedItems, 4)

		var governedChild visualization.DataVisualizationInfo
		require.NoError(t, db.Where("name = ?", "Governed Child").First(&governedChild).Error)
		permIDs, exists, permErr := permRepo.GetResourcePermissionIDs(governedChild.ID, permission.ResourceTypeDashboard)
		require.NoError(t, permErr)
		assert.True(t, exists)
		assert.Equal(t, []int64{31, 32}, permIDs)

		reasons := make(map[GovernanceBackfillSkipReason]bool)
		for _, item := range report.SkippedItems {
			reasons[item.Reason] = true
		}
		assert.True(t, reasons[GovernanceBackfillSkipReasonMissingParent])
		assert.True(t, reasons[GovernanceBackfillSkipReasonParentNotGoverned])
	})

	t.Run("backfill repo and inherit errors propagate", func(t *testing.T) {
		svc, _, _, db := setupVisualizationServiceWithPermTest(t)
		closeVisualizationDB(t, db)

		report, err := svc.BackfillGovernedVisualizationResourcesWithOptions(&GovernanceBackfillOptions{Limit: 1})
		require.Error(t, err)
		assert.Nil(t, report)

		svc2, _, permRepo, db2 := setupVisualizationServiceWithPermTest(t)
		parentType := "dashboard"
		parentID, saveErr := svc2.Save(&visualization.SaveRequest{Name: "Err Parent", Type: &parentType}, "creator")
		require.NoError(t, saveErr)
		permRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDashboard, parentID)] = []int64{41}
		permRepo.replaceErr = assert.AnError
		var child visualization.DataVisualizationInfo
		require.NoError(t, svc2.repo.Create(&visualization.DataVisualizationInfo{Name: "Err Child", PID: int64Ptr(parentID), Type: &parentType}))
		require.NoError(t, db2.Where("name = ?", "Err Child").First(&child).Error)

		report, err = svc2.BackfillGovernedVisualizationResourcesWithOptions(&GovernanceBackfillOptions{Limit: 10})
		require.Error(t, err)
		assert.Nil(t, report)
	})
}

func TestVisualizationService_CopyAndUpdateExtraBranches(t *testing.T) {
	t.Run("copy propagates source lookup error and applies overrides", func(t *testing.T) {
		svc, repo, _ := setupVisualizationServiceRepoTest(t)

		_, err := svc.Copy(&visualization.CopyRequest{ID: 999, Name: "missing"}, "tester")
		require.Error(t, err)

		typ := "dashboard"
		nodeType := "panel"
		mobile := false
		content := "cid-origin"
		checkVersion := "ver-origin"
		sourceID, err := svc.Save(&visualization.SaveRequest{Name: "Source Override", Type: &typ, NodeType: &nodeType, MobileLayout: &mobile, ContentID: &content, CheckVersion: &checkVersion}, "creator")
		require.NoError(t, err)

		newType := "dataV"
		newNodeType := "screen"
		newMobile := true
		copyID, err := svc.Copy(&visualization.CopyRequest{ID: sourceID, Name: "Copy Override", Type: &newType, NodeType: &newNodeType, MobileLayout: &newMobile}, "copier")
		require.NoError(t, err)

		copied, err := repo.GetByID(copyID)
		require.NoError(t, err)
		require.NotNil(t, copied.Type)
		assert.Equal(t, "dataV", *copied.Type)
		require.NotNil(t, copied.NodeType)
		assert.Equal(t, "screen", *copied.NodeType)
		require.NotNil(t, copied.MobileLayout)
		assert.True(t, *copied.MobileLayout)
	})

	t.Run("update changes all optional fields and publish helpers propagate errors", func(t *testing.T) {
		svc, repo, _ := setupVisualizationServiceRepoTest(t)
		id, err := svc.Save(&visualization.SaveRequest{Name: "Update All"}, "creator")
		require.NoError(t, err)

		name := "Updated All"
		pid := int64(88)
		typ := "dataV"
		canvas := `{"bg":"black"}`
		component := `[{"type":"text"}]`
		mobile := true
		content := "cid-new"
		checkVersion := "ver-new"
		status := 2

		err = svc.Update(&visualization.UpdateRequest{
			ID:              id,
			Name:            &name,
			PID:             &pid,
			Type:            &typ,
			CanvasStyleData: &canvas,
			ComponentData:   &component,
			MobileLayout:    &mobile,
			ContentID:       &content,
			CheckVersion:    &checkVersion,
			Status:          &status,
		}, "updater")
		require.NoError(t, err)

		item, err := repo.GetByID(id)
		require.NoError(t, err)
		assert.Equal(t, "Updated All", item.Name)
		require.NotNil(t, item.PID)
		assert.Equal(t, int64(88), *item.PID)
		require.NotNil(t, item.Type)
		assert.Equal(t, "dataV", *item.Type)
		require.NotNil(t, item.CanvasStyleData)
		assert.Equal(t, canvas, *item.CanvasStyleData)
		require.NotNil(t, item.ComponentData)
		assert.Equal(t, component, *item.ComponentData)
		require.NotNil(t, item.MobileLayout)
		assert.True(t, *item.MobileLayout)
		require.NotNil(t, item.ContentID)
		assert.Equal(t, content, *item.ContentID)
		require.NotNil(t, item.CheckVersion)
		assert.Equal(t, checkVersion, *item.CheckVersion)
		require.NotNil(t, item.Status)
		assert.Equal(t, 2, *item.Status)
		require.NotNil(t, item.UpdateBy)
		assert.Equal(t, "updater", *item.UpdateBy)

		published, err := svc.UpdatePublishStatus(&visualization.UpdateRequest{ID: 999, Status: intPtrVisualization(1)}, "publisher")
		require.Error(t, err)
		assert.Nil(t, published)

		recovered, err := svc.RecoverToPublished(999, "recover")
		require.Error(t, err)
		assert.Nil(t, recovered)
	})
}

func TestVisualizationService_Decompression(t *testing.T) {
	t.Run("unsupported newFrom returns error", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)
		_, err := svc.Decompression(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")

		_, err = svc.Decompression(&visualization.DecompressionRequest{NewFrom: "bad"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported newFrom")
	})

	t.Run("new_market_template returns clear error", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)
		_, err := svc.Decompression(&visualization.DecompressionRequest{NewFrom: "new_market_template", TemplateURL: "http://example.com/t.json"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not yet supported")
	})

	t.Run("new_inner_template without templateId returns error", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)
		_, err := svc.Decompression(&visualization.DecompressionRequest{NewFrom: "new_inner_template"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "templateId is required")
	})

	t.Run("new_outer_template returns correct response shape", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)

		dynamicData := `{"view_100": {"title": "Chart1", "type": "bar", "tableId": 5}}`
		componentData := `[{"id":"view_100","type":"bar"}]`
		canvasStyleData := `{"scale":100}`

		resp, err := svc.Decompression(&visualization.DecompressionRequest{
			NewFrom:         "new_outer_template",
			Name:            "Test Panel",
			Type:            "dashboard",
			CanvasStyleData: canvasStyleData,
			ComponentData:   componentData,
			DynamicData:     dynamicData,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, "Test Panel", resp.Name)
		assert.Equal(t, "dashboard", resp.Type)
		assert.Equal(t, 3, resp.Version)
		assert.Equal(t, canvasStyleData, resp.CanvasStyleData)
		assert.NotEmpty(t, resp.ID)
		assert.NotNil(t, resp.CanvasViewInfo)
		assert.Len(t, resp.CanvasViewInfo, 1)

		for viewIDStr, cv := range resp.CanvasViewInfo {
			assert.NotEmpty(t, viewIDStr)
			assert.Equal(t, "Chart1", cv["title"])
			assert.Equal(t, "bar", cv["type"])
			assert.Equal(t, "template", cv["dataFrom"])
			assert.Equal(t, int64(5), cv["sourceTableId"])
			assert.Nil(t, cv["tableId"])
			assert.Contains(t, resp.ComponentData, viewIDStr)
			assert.NotContains(t, resp.ComponentData, "view_100")
		}
	})

	t.Run("appData keeps tableId for imported app templates", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)

		resp, err := svc.Decompression(&visualization.DecompressionRequest{
			NewFrom:       "new_outer_template",
			Name:          "App Import",
			Type:          "dashboard",
			ComponentData: `[{"id":"v1"}]`,
			DynamicData:   `{"v1":{"title":"C","type":"line","tableId":7}}`,
			AppData:       `{"visualizationInfo":{"id":1}}`,
		})
		require.NoError(t, err)
		require.Len(t, resp.CanvasViewInfo, 1)
		for _, view := range resp.CanvasViewInfo {
			assert.Equal(t, int64(7), view["sourceTableId"])
			assert.Equal(t, int64(7), view["tableId"])
		}
	})

	t.Run("new_outer_template with empty dynamicData returns empty canvasViewInfo", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)

		resp, err := svc.Decompression(&visualization.DecompressionRequest{
			NewFrom:       "new_outer_template",
			Name:          "Empty Dynamic",
			Type:          "dataV",
			ComponentData: `[]`,
		})
		require.NoError(t, err)
		assert.Empty(t, resp.CanvasViewInfo)
		assert.Equal(t, "dataV", resp.Type)
	})

	t.Run("customFilter array is adapted to object", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)

		dynamicData := `{"v1": {"title":"C","type":"line","customFilter":[1,2,3]}}`
		resp, err := svc.Decompression(&visualization.DecompressionRequest{
			NewFrom:       "new_outer_template",
			Name:          "Filter Adapt",
			Type:          "dashboard",
			ComponentData: `[{"id":"v1"}]`,
			DynamicData:   dynamicData,
		})
		require.NoError(t, err)
		require.Len(t, resp.CanvasViewInfo, 1)
	})

	t.Run("dynamicData with string values handles correctly", func(t *testing.T) {
		svc, _, _ := setupVisualizationServiceRepoTest(t)

		dynamicData := `{"view_200": "{\"title\":\"Embedded\",\"type\":\"pie\"}"}`
		resp, err := svc.Decompression(&visualization.DecompressionRequest{
			NewFrom:       "new_outer_template",
			Name:          "String Values",
			Type:          "dashboard",
			ComponentData: `[{"id":"view_200"}]`,
			DynamicData:   dynamicData,
		})
		require.NoError(t, err)
		require.Len(t, resp.CanvasViewInfo, 1)
	})
}
