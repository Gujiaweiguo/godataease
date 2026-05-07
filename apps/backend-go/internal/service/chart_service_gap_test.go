package service

import (
	"errors"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestChartService_ViewOption(t *testing.T) {
	t.Parallel()

	t.Run("returns filtered view options present in component data", func(t *testing.T) {
		t.Parallel()

		repo := &fakeChartRepo{
			viewOptions: map[int64][]chart.ViewSelectorVO{
				9: {
					{ID: 11, Title: "Chart A", Type: "bar"},
					{ID: 22, Title: "Chart B", Type: "line"},
				},
			},
			componentData: map[int64]string{9: `[{"id":11},{"id":33}]`},
		}

		got, err := NewChartService(repo).ViewOption(9)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, int64(11), got[0].ID)
	})

	t.Run("returns empty when chart options are not found", func(t *testing.T) {
		t.Parallel()

		repo := &fakeChartRepo{
			viewOptions:   map[int64][]chart.ViewSelectorVO{9: {}},
			componentData: map[int64]string{9: `[]`},
		}

		got, err := NewChartService(repo).ViewOption(9)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		t.Parallel()

		repo := &fakeChartRepo{queryViewOptionErr: errors.New("view option failed")}

		got, err := NewChartService(repo).ViewOption(9)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "view option failed")
	})
}

func TestChartService_ChartBaseInfo(t *testing.T) {
	t.Parallel()

	t.Run("returns chart base info", func(t *testing.T) {
		t.Parallel()

		repo := &fakeChartRepo{chartBaseInfo: map[string]*chart.ChartBaseVO{
			"core:18": {
				ChartID:      18,
				ChartName:    "Revenue",
				ChartType:    "bar",
				ResourceID:   6,
				ResourceName: "Dashboard",
			},
		}}

		got, err := NewChartService(repo).ChartBaseInfo(18, "core")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(18), got.ChartID)
		assert.Equal(t, "Revenue", got.ChartName)
	})

	t.Run("returns nil when chart is not found", func(t *testing.T) {
		t.Parallel()

		got, err := NewChartService(&fakeChartRepo{}).ChartBaseInfo(404, "core")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		t.Parallel()

		repo := &fakeChartRepo{chartBaseInfoErr: errors.New("base info failed")}

		got, err := NewChartService(repo).ChartBaseInfo(18, "core")
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "base info failed")
	})
}

func TestChartService_ListByDQWithPermission(t *testing.T) {
	t.Parallel()

	newColumnPermissionService := func(t *testing.T) *ColumnPermissionService {
		t.Helper()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&permission.DataPermColumn{}))
		t.Cleanup(func() {
			sqlDB, dbErr := db.DB()
			require.NoError(t, dbErr)
			require.NoError(t, sqlDB.Close())
		})

		repo := repository.NewColumnPermissionRepository(db)
		require.NoError(t, repo.Create(&permission.DataPermColumn{DatasetID: 7, FieldName: "secret", PermType: permission.PermTypeDisable}))
		require.NoError(t, repo.Create(&permission.DataPermColumn{DatasetID: 7, FieldName: "phone", PermType: permission.PermTypeMask}))
		return NewColumnPermissionService(repo, nil)
	}

	t.Run("applies column permission filtering", func(t *testing.T) {
		t.Parallel()

		groupTypeD := "d"
		groupTypeQ := "q"
		nameRegion := "region"
		originRegion := "region"
		nameSecret := "secret"
		originSecret := "secret"
		namePhone := "phone"
		originPhone := "phone"
		typeName := "VARCHAR"
		deTypeD := 0
		deTypeQ := 2
		checked := true

		repo := &fakeChartRepo{
			dsFieldsByGroup: map[int64][]*dataset.CoreDatasetTableField{
				7: {
					{ID: 1, DatasetGroupID: 7, Name: &nameRegion, OriginName: &originRegion, GroupType: &groupTypeD, Type: &typeName, DeType: &deTypeD, Checked: &checked},
					{ID: 2, DatasetGroupID: 7, Name: &nameSecret, OriginName: &originSecret, GroupType: &groupTypeD, Type: &typeName, DeType: &deTypeD, Checked: &checked},
					{ID: 3, DatasetGroupID: 7, Name: &namePhone, OriginName: &originPhone, GroupType: &groupTypeQ, Type: &typeName, DeType: &deTypeQ, Checked: &checked},
				},
			},
		}

		svc := NewChartService(repo)
		svc.SetColumnPermissionService(newColumnPermissionService(t))

		got, err := svc.ListByDQWithPermission(7, 0, 2)
		require.NoError(t, err)
		require.Len(t, got.DimensionList, 1)
		assert.Equal(t, "region", got.DimensionList[0].Name)
		require.Len(t, got.QuotaList, 2)
		assert.Equal(t, "phone", got.QuotaList[0].Name)
		assert.True(t, got.QuotaList[0].Desensitized)
		assert.Equal(t, int64(-1), got.QuotaList[1].ID)
	})

	t.Run("returns synthetic count field when dataset has no fields", func(t *testing.T) {
		t.Parallel()

		repo := &fakeChartRepo{dsFieldsByGroup: map[int64][]*dataset.CoreDatasetTableField{7: {}}}
		svc := NewChartService(repo)
		svc.SetColumnPermissionService(newColumnPermissionService(t))

		got, err := svc.ListByDQWithPermission(7, 0, 2)
		require.NoError(t, err)
		assert.Empty(t, got.DimensionList)
		require.Len(t, got.QuotaList, 1)
		assert.Equal(t, int64(-1), got.QuotaList[0].ID)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		t.Parallel()

		repo := &fakeChartRepo{listGroupErr: errors.New("list fields failed")}
		svc := NewChartService(repo)
		svc.SetColumnPermissionService(newColumnPermissionService(t))

		got, err := svc.ListByDQWithPermission(7, 0, 2)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "list fields failed")
	})
}
