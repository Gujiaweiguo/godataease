package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDatasetHandlerGapDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}))
	require.NoError(t, db.Exec(`CREATE TABLE core_chart_view (id INTEGER PRIMARY KEY AUTOINCREMENT, table_id INTEGER)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE data_perm_row (id INTEGER PRIMARY KEY AUTOINCREMENT, dataset_group_id INTEGER, expression_tree TEXT)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE data_perm_column (id INTEGER PRIMARY KEY AUTOINCREMENT, dataset_group_id INTEGER, field_name TEXT)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE visualization_linkage_field (id INTEGER PRIMARY KEY AUTOINCREMENT, source_field INTEGER, target_field INTEGER)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE visualization_link_jump_info (id INTEGER PRIMARY KEY AUTOINCREMENT, source_field_id INTEGER)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE visualization_outer_params_target_view_info (id INTEGER PRIMARY KEY AUTOINCREMENT, target_field_id TEXT)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE dataset_orders (city TEXT, amount INTEGER)`).Error)

	return db
}

func setupDatasetHandlerGapRouter(t *testing.T, db *gorm.DB, userID int64, register func(*gin.Engine, *DatasetHandler)) *gin.Engine {
	t.Helper()
	repo := repository.NewDatasetRepository(db)
	svc := service.NewDatasetService(repo)
	h := NewDatasetHandler(svc, nil)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("userId", userID)
		c.Set("username", "dataset-gap")
		c.Next()
	})
	register(r, h)
	return r
}

func seedDatasetDetailFixture(t *testing.T, db *gorm.DB, groupID int64) {
	t.Helper()
	groupType := dataset.NodeTypeDataset
	tableName := "dataset_orders"
	originName := "city"
	fieldName := "City"
	groupTypeDim := "d"
	valueType := "VARCHAR"
	deType := 0
	checked := true
	seedDatasetGroupRecord(t, db, &dataset.CoreDatasetGroup{ID: groupID, Name: "orders", PID: int64PtrForDatasetHandler(0), Level: intPtrForDatasetHandler(0), NodeType: &groupType, DelFlag: intPtrForDatasetHandler(0)})
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 501, DatasetGroupID: groupID, PhysicalTable: &tableName}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 601, DatasetGroupID: groupID, DatasetTableID: int64PtrForDatasetHandler(501), OriginName: &originName, Name: &fieldName, GroupType: &groupTypeDim, Type: &valueType, DeType: &deType, Checked: &checked}).Error)
	require.NoError(t, db.Exec(`INSERT INTO dataset_orders (city, amount) VALUES ('Beijing', 10), ('Shanghai', 20)`).Error)
}

func TestDatasetHandler_GetDetail_Details_And_DsDetails(t *testing.T) {
	t.Parallel()

	db := setupDatasetHandlerGapDB(t)
	seedDatasetDetailFixture(t, db, 401)
	router := setupDatasetHandlerGapRouter(t, db, 1001, func(r *gin.Engine, h *DatasetHandler) {
		r.POST("/get/:id", h.GetDetail)
		r.POST("/details/:id", h.Details)
		r.POST("/dsDetails", h.DsDetails)
	})

	getResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/get/401", nil).Body.Bytes())
	assert.Equal(t, "000000", getResp.Code)
	assert.Contains(t, string(getResp.Data), `"total":2`)
	assert.Contains(t, string(getResp.Data), `"Beijing"`)

	detailsResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/details/401", nil).Body.Bytes())
	assert.Equal(t, "000000", detailsResp.Code)
	assert.Contains(t, string(detailsResp.Data), `"allFields"`)

	dsDetailsResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/dsDetails", map[string]any{"ids": []int64{401, 999}}).Body.Bytes())
	assert.Equal(t, "000000", dsDetailsResp.Code)
	assert.Contains(t, string(dsDetailsResp.Data), `"id":401`)
	assert.Contains(t, string(dsDetailsResp.Data), `"id":999`)
}

func TestDatasetHandler_PerDelete_And_GetDatasetTotal(t *testing.T) {
	t.Parallel()

	db := setupDatasetHandlerGapDB(t)
	seedDatasetDetailFixture(t, db, 402)
	require.NoError(t, db.Exec(`INSERT INTO core_chart_view (id, table_id) VALUES (701, 501)`).Error)
	router := setupDatasetHandlerGapRouter(t, db, 1001, func(r *gin.Engine, h *DatasetHandler) {
		r.POST("/perDelete/:id", h.PerDelete)
		r.POST("/getDatasetTotal", h.GetDatasetTotal)
	})

	perDeleteResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/perDelete/402", nil).Body.Bytes())
	assert.Equal(t, "000000", perDeleteResp.Code)
	assert.Equal(t, "true", string(perDeleteResp.Data))

	totalResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/getDatasetTotal", map[string]any{"id": 402}).Body.Bytes())
	assert.Equal(t, "000000", totalResp.Code)
	assert.Equal(t, "2", string(totalResp.Data))
}

func TestDatasetHandler_GetSQLParams_And_EnumShortCircuits(t *testing.T) {
	t.Parallel()

	db := setupDatasetHandlerGapDB(t)
	groupType := dataset.NodeTypeDataset
	tableName := "dataset_orders"
	sqlVars := `[{"variableName":"region","type":["string"],"params":["north"]}]`
	seedDatasetGroupRecord(t, db, &dataset.CoreDatasetGroup{ID: 403, Name: "orders", PID: int64PtrForDatasetHandler(0), Level: intPtrForDatasetHandler(0), NodeType: &groupType, DelFlag: intPtrForDatasetHandler(0)})
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 502, DatasetGroupID: 403, PhysicalTable: &tableName, SQLVariables: &sqlVars}).Error)
	router := setupDatasetHandlerGapRouter(t, db, 1001, func(r *gin.Engine, h *DatasetHandler) {
		r.POST("/getSqlParams", h.GetSQLParams)
		r.POST("/enumValueObj", h.EnumValueObj)
		r.POST("/enumValueDs", h.EnumValueDs)
		r.POST("/enumValue", h.EnumValue)
	})

	sqlResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/getSqlParams", map[string]any{"ids": []int64{403}}).Body.Bytes())
	assert.Equal(t, "000000", sqlResp.Code)
	assert.Contains(t, string(sqlResp.Data), `"variableName":"region"`)

	enumObjResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/enumValueObj", map[string]any{}).Body.Bytes())
	assert.Equal(t, "000000", enumObjResp.Code)
	assert.Equal(t, "[]", string(enumObjResp.Data))

	enumDsResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/enumValueDs", map[string]any{}).Body.Bytes())
	assert.Equal(t, "000000", enumDsResp.Code)
	assert.Equal(t, "[]", string(enumDsResp.Data))

	enumResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/enumValue", map[string]any{"fieldIds": []int64{0}}).Body.Bytes())
	assert.Equal(t, "000000", enumResp.Code)
	assert.Equal(t, "[]", string(enumResp.Data))
}

func TestDatasetHandler_ListFieldsByDsIds_DeleteFieldErrors_And_ListWithPermissions(t *testing.T) {
	t.Parallel()

	db := setupDatasetHandlerGapDB(t)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 802, DatasourceID: int64PtrForDatasetHandler(77), DatasetGroupID: 500}).Error)
	repo := repository.NewDatasetRepository(db)
	svc := service.NewDatasetService(repo)
	h := &DatasetHandler{service: svc, chartService: nil}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1001))
		c.Set("userId", int64(1001))
		c.Next()
	})
	router.POST("/listFieldsByDsIds", h.ListFieldsByDsIds)
	router.POST("/datasetField/delete/:id", h.DeleteDatasetField)
	router.POST("/datasetField/deleteByChartId/:id", h.DeleteDatasetFieldByChart)
	router.POST("/listWithPermissions/:datasetId", h.ListWithPermissions)

	listResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/listFieldsByDsIds", map[string]any{"dsIds": []int64{77}}).Body.Bytes())
	assert.Equal(t, "000000", listResp.Code)
	assert.Contains(t, string(listResp.Data), `"datasourceId":77`)

	deleteResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/datasetField/delete/801", nil).Body.Bytes())
	assert.Equal(t, "500000", deleteResp.Code)
	assert.Contains(t, deleteResp.Msg, "dataset field not found")

	deleteByChartResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/datasetField/deleteByChartId/900", nil).Body.Bytes())
	assert.Equal(t, "000000", deleteByChartResp.Code)

	listWithPermResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/listWithPermissions/500", nil).Body.Bytes())
	assert.Equal(t, "000000", listWithPermResp.Code)
	assert.Equal(t, "[]", string(listWithPermResp.Data))
}

func TestDatasetHandler_SaveField_ServiceUnavailable_And_GetFieldFunctions(t *testing.T) {
	t.Parallel()

	router := gin.New()
	h := &DatasetHandler{service: nil, chartService: nil}
	router.POST("/saveField", h.SaveField)
	router.GET("/fieldFunctions", h.GetFieldFunctions)

	saveFieldResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/saveField", map[string]any{"datasetGroupId": 500}).Body.Bytes())
	assert.Equal(t, "500000", saveFieldResp.Code)
	assert.Contains(t, saveFieldResp.Msg, "dataset service unavailable")

	fieldFunctionsResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodGet, "/fieldFunctions", nil).Body.Bytes())
	assert.Equal(t, "000000", fieldFunctionsResp.Code)
	assert.Equal(t, "[]", string(fieldFunctionsResp.Data))
}

func TestDatasetHandler_MultFieldValuesForPermissions_And_ListByDatasetGroup_NilChartService(t *testing.T) {
	t.Parallel()

	db := setupDatasetHandlerGapDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := service.NewDatasetService(repo)
	h := &DatasetHandler{service: svc, chartService: nil}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1001))
		c.Set("userId", int64(1001))
		c.Next()
	})
	router.POST("/multFieldValuesForPermissions", h.MultFieldValuesForPermissions)
	router.GET("/listByDatasetGroup/:datasetId", h.ListByDatasetGroup)

	multResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodPost, "/multFieldValuesForPermissions", map[string]any{"fieldIds": []int64{0}}).Body.Bytes())
	assert.Equal(t, "000000", multResp.Code)
	assert.Equal(t, "[]", string(multResp.Data))

	listResp := decodeDatasetResp(t, performDatasetJSONRequest(t, router, http.MethodGet, "/listByDatasetGroup/1", nil).Body.Bytes())
	assert.Equal(t, "000000", listResp.Code)
	var fields []chart.ChartField
	require.NoError(t, json.Unmarshal(listResp.Data, &fields))
	assert.Empty(t, fields)
}
