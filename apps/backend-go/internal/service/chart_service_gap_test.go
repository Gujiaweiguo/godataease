package service

import (
	"encoding/json"
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

type fakePermissionAwareChartRepo struct {
	*fakeChartRepo
	datasetGroupID int64
	getGroupErr    error
	selectCols     string
	whereClause    string
	whereArgs      []interface{}
	filteredRows   []map[string]interface{}
	filteredTotal  int64
	filterErr      error
}

func (r *fakePermissionAwareChartRepo) GetDatasetGroupIDByChartID(chartID int64) (int64, error) {
	if r.getGroupErr != nil {
		return 0, r.getGroupErr
	}
	return r.datasetGroupID, nil
}

func (r *fakePermissionAwareChartRepo) QueryRowsWithFilter(chartID int64, selectColumns string, whereClause string, whereArgs []interface{}, limit int) ([]map[string]interface{}, int64, error) {
	r.selectCols = selectColumns
	r.whereClause = whereClause
	r.whereArgs = whereArgs
	if r.filterErr != nil {
		return nil, 0, r.filterErr
	}
	return r.filteredRows, r.filteredTotal, nil
}

func TestChartService_QueryDataWithPermission_ActualPermissionPath(t *testing.T) {
	t.Parallel()

	t.Run("uses permission-aware repo when available with valid userID", func(t *testing.T) {
		t.Parallel()

		inner := &fakeChartRepo{
			byID: map[int64]*chart.CoreChartView{10: {ID: 10}},
			data: map[int64]chartRegressionSample{},
		}
		repo := &fakePermissionAwareChartRepo{
			fakeChartRepo:  inner,
			datasetGroupID: 42,
			filteredRows: []map[string]interface{}{
				{"category": "绿茶", "sales_amount": 200.0},
			},
			filteredTotal: 1,
		}

		svc := NewChartService(repo)
		resultCount := 50

		resp, err := svc.QueryDataWithPermission(&chart.ChartDataRequest{
			ID:          10,
			ResultCount: &resultCount,
			Payload: map[string]interface{}{
				"id":   10.0,
				"type": "bar",
				"xAxis": []interface{}{
					map[string]interface{}{"id": "2", "dataeaseName": "category", "originName": "category", "name": "分类"},
				},
				"yAxis": []interface{}{
					map[string]interface{}{"id": "5", "dataeaseName": "sales_amount", "originName": "sales_amount", "name": "销售额", "summary": "sum"},
				},
			},
		}, 1)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(42), repo.datasetGroupID)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "绿茶", resp.Data[0].Category)
		assert.Equal(t, 200.0, resp.Data[0].Value)
	})

	t.Run("propagates GetDatasetGroupIDByChartID error", func(t *testing.T) {
		t.Parallel()

		inner := &fakeChartRepo{
			byID: map[int64]*chart.CoreChartView{10: {ID: 10}},
		}
		repo := &fakePermissionAwareChartRepo{
			fakeChartRepo:  inner,
			datasetGroupID: 42,
			getGroupErr:    errors.New("group lookup failed"),
		}

		svc := NewChartService(repo)
		resultCount := 10
		_, err := svc.QueryDataWithPermission(&chart.ChartDataRequest{ID: 10, ResultCount: &resultCount}, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "group lookup failed")
	})

	t.Run("propagates QueryRowsWithFilter error", func(t *testing.T) {
		t.Parallel()

		inner := &fakeChartRepo{
			byID: map[int64]*chart.CoreChartView{10: {ID: 10}},
		}
		repo := &fakePermissionAwareChartRepo{
			fakeChartRepo:  inner,
			datasetGroupID: 42,
			filterErr:      errors.New("filter query failed"),
		}

		svc := NewChartService(repo)
		resultCount := 10
		_, err := svc.QueryDataWithPermission(&chart.ChartDataRequest{ID: 10, ResultCount: &resultCount}, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "filter query failed")
	})
}

func TestChartService_QueryData_TableNormal(t *testing.T) {
	t.Parallel()

	repo := &fakeChartRepo{
		byID: map[int64]*chart.CoreChartView{20: {ID: 20}},
		data: map[int64]chartRegressionSample{
			20: {
				ChartID: 20,
				Rows: []map[string]interface{}{
					{"product": "茶杯", "amount": 50.0, "category": "日用品"},
				},
				Total: 1,
			},
		},
	}
	svc := NewChartService(repo)

	resp, err := svc.QueryData(&chart.ChartDataRequest{
		ID: 20,
		Payload: map[string]interface{}{
			"id":   20.0,
			"type": "table-normal",
			"xAxis": []interface{}{
				map[string]interface{}{"id": "1", "dataeaseName": "f_product", "originName": "product", "name": "产品"},
			},
			"yAxis": []interface{}{
				map[string]interface{}{"id": "2", "dataeaseName": "f_amount", "originName": "amount", "name": "金额"},
				map[string]interface{}{"id": "3", "dataeaseName": "f_category", "originName": "category", "name": "分类"},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.TableRow, 1)
	assert.Equal(t, "茶杯", resp.TableRow[0]["f_product"])
	assert.Equal(t, 50.0, resp.TableRow[0]["f_amount"])
	assert.Equal(t, "日用品", resp.TableRow[0]["f_category"])
	assert.Empty(t, resp.Data)
}

func TestChartService_QueryData_NoXAxis_SeriesPath(t *testing.T) {
	t.Parallel()

	repo := &fakeChartRepo{
		byID: map[int64]*chart.CoreChartView{30: {ID: 30}},
		data: map[int64]chartRegressionSample{
			30: {
				ChartID: 30,
				Rows: []map[string]interface{}{
					{"sales_amount": 100.0},
					{"sales_amount": 200.0},
				},
				Total: 2,
			},
		},
	}
	svc := NewChartService(repo)

	resp, err := svc.QueryData(&chart.ChartDataRequest{
		ID: 30,
		Payload: map[string]interface{}{
			"id":   30.0,
			"type": "bar",
			"yAxis": []interface{}{
				map[string]interface{}{"id": "5", "dataeaseName": "sales_amount", "originName": "sales_amount", "name": "销售额", "summary": "sum"},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "全部", resp.Data[0].Category)
	assert.Equal(t, 300.0, resp.Data[0].Value)
	assert.Empty(t, resp.Data[0].DimensionList)
}

func TestChartService_MetricAccumulator_CountOnlyEdge(t *testing.T) {
	t.Parallel()

	t.Run("countOnly=true with non-count summary still counts", func(t *testing.T) {
		acc := &metricAccumulator{summary: "sum"}
		acc.add(42.0, true)
		assert.Equal(t, 1, acc.count)
		assert.Equal(t, 0.0, acc.sum)
		assert.Equal(t, 0.0, acc.value())
	})

	t.Run("countOnly=true ignores non-numeric value", func(t *testing.T) {
		acc := &metricAccumulator{summary: "avg"}
		acc.add("not-a-number", true)
		assert.Equal(t, 1, acc.count)
		assert.Equal(t, 0.0, acc.sum)
	})

	t.Run("empty accumulator default summary returns sum", func(t *testing.T) {
		acc := &metricAccumulator{summary: "unknown_summary"}
		acc.add(10.0, false)
		acc.add(20.0, false)
		assert.Equal(t, 30.0, acc.value())
	})
}

func TestChartService_IsCountField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		field map[string]interface{}
		want  bool
	}{
		{
			name:  "dataeaseName is star",
			field: map[string]interface{}{"dataeaseName": "*"},
			want:  true,
		},
		{
			name:  "summary count and preferred key is star",
			field: map[string]interface{}{"summary": "count", "dataeaseName": "*"},
			want:  true,
		},
		{
			name:  "summary count but preferred key is not star",
			field: map[string]interface{}{"summary": "count", "dataeaseName": "amount"},
			want:  false,
		},
		{
			name:  "regular field",
			field: map[string]interface{}{"dataeaseName": "sales", "summary": "sum"},
			want:  false,
		},
		{
			name:  "empty field",
			field: map[string]interface{}{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isCountField(tt.field))
		})
	}
}

func TestChartService_SaveFromMap_AdditionalFields(t *testing.T) {
	t.Parallel()

	t.Run("bool fields are applied", func(t *testing.T) {
		title := "Test"
		repo := &fakeChartRepo{
			byID: map[int64]*chart.CoreChartView{
				7: {ID: 7, Title: &title},
			},
		}
		svc := NewChartService(repo)

		view, err := svc.SaveFromMap(map[string]interface{}{
			"id":                int64(7),
			"isPlugin":          true,
			"refreshViewEnable": "false",
			"linkageActive":     1,
			"jumpActive":        "true",
			"aggregate":         json.Number("0"),
		})
		require.NoError(t, err)
		require.NotNil(t, view)
		require.NotNil(t, view.IsPlugin)
		assert.True(t, *view.IsPlugin)
		require.NotNil(t, view.RefreshViewEnable)
		assert.False(t, *view.RefreshViewEnable)
		require.NotNil(t, view.LinkageActive)
		assert.True(t, *view.LinkageActive)
		require.NotNil(t, view.JumpActive)
		assert.True(t, *view.JumpActive)
		require.NotNil(t, view.Aggregate)
		assert.False(t, *view.Aggregate)
	})

	t.Run("int fields refreshTime is applied", func(t *testing.T) {
		title := "Test"
		repo := &fakeChartRepo{
			byID: map[int64]*chart.CoreChartView{
				7: {ID: 7, Title: &title},
			},
		}
		svc := NewChartService(repo)

		view, err := svc.SaveFromMap(map[string]interface{}{
			"id":          int64(7),
			"resultCount": 42,
			"refreshTime": int64(3600),
		})
		require.NoError(t, err)
		require.NotNil(t, view)
		require.NotNil(t, view.ResultCount)
		assert.Equal(t, 42, *view.ResultCount)
		require.NotNil(t, view.RefreshTime)
		assert.Equal(t, 3600, *view.RefreshTime)
	})

	t.Run("additional json fields are applied", func(t *testing.T) {
		title := "Test"
		repo := &fakeChartRepo{
			byID: map[int64]*chart.CoreChartView{
				7: {ID: 7, Title: &title},
			},
		}
		svc := NewChartService(repo)

		view, err := svc.SaveFromMap(map[string]interface{}{
			"id":           int64(7),
			"xAxisExt":     []string{"ext1"},
			"yAxisExt":     map[string]any{"k": "v"},
			"extStack":     []interface{}{1, 2},
			"extBubble":    "bubble",
			"extLabel":     "label",
			"extTooltip":   "tooltip",
			"drillFields":  []string{"d1"},
			"senior":       map[string]any{"s": 1},
			"snapshot":     "snap",
			"viewFields":   []string{"vf1"},
			"extColor":     "color",
			"sortPriority": []string{"sort1"},
		})
		require.NoError(t, err)
		require.NotNil(t, view)
		require.NotNil(t, view.XAxisExt)
		assert.JSONEq(t, `["ext1"]`, *view.XAxisExt)
		require.NotNil(t, view.YAxisExt)
		assert.JSONEq(t, `{"k":"v"}`, *view.YAxisExt)
		require.NotNil(t, view.ExtStack)
		assert.JSONEq(t, `[1,2]`, *view.ExtStack)
		require.NotNil(t, view.ExtBubble)
		assert.Equal(t, `"bubble"`, *view.ExtBubble)
		require.NotNil(t, view.ExtLabel)
		assert.Equal(t, `"label"`, *view.ExtLabel)
		require.NotNil(t, view.ExtTooltip)
		assert.Equal(t, `"tooltip"`, *view.ExtTooltip)
		require.NotNil(t, view.DrillFields)
		assert.JSONEq(t, `["d1"]`, *view.DrillFields)
		require.NotNil(t, view.Senior)
		assert.JSONEq(t, `{"s":1}`, *view.Senior)
		require.NotNil(t, view.Snapshot)
		assert.Equal(t, `"snap"`, *view.Snapshot)
		require.NotNil(t, view.ViewFields)
		assert.JSONEq(t, `["vf1"]`, *view.ViewFields)
		require.NotNil(t, view.ExtColor)
		assert.Equal(t, `"color"`, *view.ExtColor)
		require.NotNil(t, view.SortPriority)
		assert.JSONEq(t, `["sort1"]`, *view.SortPriority)
	})

	t.Run("additional string fields are applied", func(t *testing.T) {
		title := "Test"
		repo := &fakeChartRepo{
			byID: map[int64]*chart.CoreChartView{
				7: {ID: 7, Title: &title},
			},
		}
		svc := NewChartService(repo)

		view, err := svc.SaveFromMap(map[string]interface{}{
			"id":               int64(7),
			"stylePriority":    "view",
			"chartType":        "waterfall",
			"refreshUnit":      "minute",
			"flowMapStartName": "Beijing",
			"flowMapEndName":   "Shanghai",
		})
		require.NoError(t, err)
		require.NotNil(t, view)
		require.NotNil(t, view.StylePriority)
		assert.Equal(t, "view", *view.StylePriority)
		require.NotNil(t, view.ChartType)
		assert.Equal(t, "waterfall", *view.ChartType)
		require.NotNil(t, view.RefreshUnit)
		assert.Equal(t, "minute", *view.RefreshUnit)
		require.NotNil(t, view.FlowMapStartName)
		assert.Equal(t, "Beijing", *view.FlowMapStartName)
		require.NotNil(t, view.FlowMapEndName)
		assert.Equal(t, "Shanghai", *view.FlowMapEndName)
	})
}

func TestChartService_BuildTableRows_EmptyFields(t *testing.T) {
	t.Parallel()

	t.Run("empty sourceFields clones rows directly", func(t *testing.T) {
		rows := []map[string]interface{}{{"a": 1, "b": 2}}
		result := buildTableRows(rows, nil)
		require.Len(t, result, 1)
		assert.Equal(t, 1, result[0]["a"])
		assert.Equal(t, 2, result[0]["b"])
	})

	t.Run("empty rows returns empty slice", func(t *testing.T) {
		result := buildTableRows(nil, []map[string]interface{}{{"id": "1"}})
		assert.Empty(t, result)
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
