package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type visualizationHandlerResponse struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func setupVisualizationHandlerUnitDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&visualization.DataVisualizationInfo{}))

	statements := []string{
		`CREATE TABLE core_opt_recent (uid INTEGER, resource_id INTEGER, time INTEGER)`,
		`CREATE TABLE core_store (uid INTEGER, resource_id INTEGER)`,
		`CREATE TABLE core_dataset_group (id INTEGER PRIMARY KEY, name TEXT, node_type TEXT, create_by TEXT)`,
		`CREATE TABLE core_datasource (id INTEGER PRIMARY KEY, name TEXT, type TEXT, create_by TEXT)`,
		`CREATE TABLE core_chart_view (id INTEGER PRIMARY KEY, title TEXT, scene_id INTEGER, table_id INTEGER, type TEXT, render TEXT, result_count INTEGER, result_mode TEXT, x_axis TEXT, x_axis_ext TEXT, y_axis TEXT, y_axis_ext TEXT, ext_stack TEXT, ext_bubble TEXT, ext_label TEXT, ext_tooltip TEXT, custom_attr TEXT, custom_attr_mobile TEXT, custom_style TEXT, custom_style_mobile TEXT, custom_filter TEXT, drill_fields TEXT, senior TEXT, create_by TEXT, create_time INTEGER, update_time INTEGER, snapshot TEXT, style_priority INTEGER, chart_type TEXT, is_plugin INTEGER, data_from TEXT, view_fields TEXT, refresh_view_enable INTEGER, refresh_unit TEXT, refresh_time INTEGER, linkage_active INTEGER, jump_active INTEGER, copy_from INTEGER, copy_id INTEGER, aggregate INTEGER, flow_map_start_name TEXT, flow_map_end_name TEXT, ext_color TEXT, sort_priority TEXT)`,
		`CREATE TABLE snapshot_core_chart_view (id INTEGER PRIMARY KEY, title TEXT, scene_id INTEGER, table_id INTEGER, type TEXT, render TEXT, result_count INTEGER, result_mode TEXT, x_axis TEXT, x_axis_ext TEXT, y_axis TEXT, y_axis_ext TEXT, ext_stack TEXT, ext_bubble TEXT, ext_label TEXT, ext_tooltip TEXT, custom_attr TEXT, custom_attr_mobile TEXT, custom_style TEXT, custom_style_mobile TEXT, custom_filter TEXT, drill_fields TEXT, senior TEXT, create_by TEXT, create_time INTEGER, update_time INTEGER, snapshot TEXT, style_priority INTEGER, chart_type TEXT, is_plugin INTEGER, data_from TEXT, view_fields TEXT, refresh_view_enable INTEGER, refresh_unit TEXT, refresh_time INTEGER, linkage_active INTEGER, jump_active INTEGER, copy_from INTEGER, copy_id INTEGER, aggregate INTEGER, flow_map_start_name TEXT, flow_map_end_name TEXT, ext_color TEXT, sort_priority TEXT)`,
		`CREATE TABLE visualization_linkage (id INTEGER PRIMARY KEY, dv_id INTEGER, source_view_id INTEGER, target_view_id INTEGER, update_time INTEGER, update_people TEXT, linkage_active INTEGER, ext1 TEXT, ext2 TEXT, copy_from INTEGER, copy_id INTEGER)`,
		`CREATE TABLE visualization_linkage_field (id INTEGER PRIMARY KEY, linkage_id INTEGER, source_field TEXT, target_field TEXT, update_time INTEGER, copy_from INTEGER, copy_id INTEGER)`,
		`CREATE TABLE visualization_link_jump (id INTEGER PRIMARY KEY, source_dv_id INTEGER, source_view_id INTEGER, link_jump_info TEXT, checked INTEGER, copy_from INTEGER, copy_id INTEGER)`,
		`CREATE TABLE visualization_link_jump_info (id INTEGER PRIMARY KEY, link_jump_id INTEGER, link_type TEXT, jump_type TEXT, target_dv_id INTEGER, source_field_id INTEGER, content TEXT, checked INTEGER, attach_params TEXT, copy_from INTEGER, copy_id INTEGER)`,
		`CREATE TABLE visualization_link_jump_target_view_info (target_id INTEGER PRIMARY KEY, link_jump_info_id INTEGER, source_field_active_id INTEGER, target_view_id INTEGER, target_field_id INTEGER, copy_from INTEGER, copy_id INTEGER)`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}

	t.Cleanup(func() {
		sqlDB, dbErr := db.DB()
		require.NoError(t, dbErr)
		require.NoError(t, sqlDB.Close())
	})

	return db
}

func setupVisualizationHandlerRouter(db *gorm.DB, register func(*gin.Engine, *VisualizationHandler)) *gin.Engine {
	repo := repository.NewVisualizationRepository(db)
	h := NewVisualizationHandler(service.NewVisualizationService(repo))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(7))
		c.Set("userId", int64(7))
		c.Next()
	})
	register(r, h)
	return r
}

func performVisualizationRequest(t *testing.T, router *gin.Engine, method string, path string, body string) visualizationHandlerResponse {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp visualizationHandlerResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func seedVisualizationHandlerListTreeData(t *testing.T, db *gorm.DB) {
	t.Helper()

	dashboardType := "dashboard"
	leafType := "leaf"
	folderType := "folder"
	status := 1
	pid := int64(11)
	createBy := "tester"

	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{ID: 11, Name: "Dash Folder", Type: &dashboardType, NodeType: &folderType, Status: &status, CreateBy: &createBy}).Error)
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{ID: 12, Name: "Dash Panel", PID: &pid, Type: &dashboardType, NodeType: &leafType, Status: &status, CreateBy: &createBy}).Error)
}

func seedVisualizationHandlerFindRecentData(t *testing.T, db *gorm.DB) {
	t.Helper()

	dashboardType := "dashboard"
	leafType := "leaf"
	status := 1
	createBy := "tester"
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{ID: 31, Name: "Recent Panel", Type: &dashboardType, NodeType: &leafType, Status: &status, CreateBy: &createBy}).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_opt_recent (uid, resource_id, time) VALUES (7, 31, 1000)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_store (uid, resource_id) VALUES (7, 31)`).Error)
}

func seedVisualizationHandlerRecord(t *testing.T, db *gorm.DB, id int64, name string) {
	t.Helper()

	dashboardType := "dashboard"
	leafType := "leaf"
	status := 0
	createBy := "tester"
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{ID: id, Name: name, Type: &dashboardType, NodeType: &leafType, Status: &status, CreateBy: &createBy, UpdateBy: &createBy}).Error)
}

func TestVisualizationHandler_List(t *testing.T) {
	t.Parallel()

	t.Run("returns results", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		seedVisualizationHandlerListTreeData(t, db)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/list", h.List) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/list", `{"type":"dashboard","current":1,"size":10}`)
		require.Equal(t, "000000", resp.Code)

		var data visualization.ListResponse
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		assert.Len(t, data.List, 2)
		assert.Equal(t, int64(2), data.Total)
	})

	t.Run("returns empty results", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/list", h.List) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/list", `{"type":"dashboard","current":1,"size":10}`)
		require.Equal(t, "000000", resp.Code)

		var data visualization.ListResponse
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		assert.Empty(t, data.List)
		assert.Zero(t, data.Total)
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/list", h.List) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/list", `{"type":"dashboard","current":1,"size":10}`)
		assert.Equal(t, "500000", resp.Code)
	})
}

func TestVisualizationHandler_FindRecent(t *testing.T) {
	t.Parallel()

	t.Run("returns recent results", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		seedVisualizationHandlerFindRecentData(t, db)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/recent", h.FindRecent) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/recent", `{"type":"panel"}`)
		require.Equal(t, "000000", resp.Code)

		var data []visualization.VisualizationResourceVO
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		require.Len(t, data, 1)
		assert.Equal(t, "Recent Panel", data[0].Name)
	})

	t.Run("returns empty recent results", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/recent", h.FindRecent) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/recent", `{"type":"panel"}`)
		require.Equal(t, "000000", resp.Code)

		var data []visualization.VisualizationResourceVO
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		assert.Empty(t, data)
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/recent", h.FindRecent) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/recent", `{"type":"panel"}`)
		assert.Equal(t, "500000", resp.Code)
	})
}

func TestVisualizationHandler_Tree(t *testing.T) {
	t.Parallel()

	t.Run("returns tree data", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		seedVisualizationHandlerListTreeData(t, db)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/tree", h.Tree) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/tree", `{"busiFlag":"dashboard"}`)
		require.Equal(t, "000000", resp.Code)

		var data []map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		require.Len(t, data, 1)
		children := data[0]["children"].([]interface{})
		require.Len(t, children, 1)
		folder := children[0].(map[string]interface{})
		assert.Equal(t, "Dash Folder", folder["name"])
	})

	t.Run("returns empty tree", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/tree", h.Tree) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/tree", `{"busiFlag":"dashboard"}`)
		require.Equal(t, "000000", resp.Code)

		var data []map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		assert.Nil(t, data[0]["children"])
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/tree", h.Tree) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/tree", `{"busiFlag":"dashboard"}`)
		assert.Equal(t, "500000", resp.Code)
	})
}

func TestVisualizationHandler_SaveCanvas(t *testing.T) {
	t.Parallel()

	t.Run("saves canvas", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/save", h.SaveCanvas) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/save", `{"name":"Saved Canvas","type":"dashboard"}`)
		require.Equal(t, "000000", resp.Code)
		assert.NotEmpty(t, strings.Trim(string(resp.Data), `"`))
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/save", h.SaveCanvas) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/save", `{`)
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/save", h.SaveCanvas) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/save", `{"name":"Saved Canvas"}`)
		assert.Equal(t, "500000", resp.Code)
	})
}

func TestVisualizationHandler_UpdateCanvas(t *testing.T) {
	t.Parallel()

	t.Run("updates canvas", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		seedVisualizationHandlerRecord(t, db, 41, "Before Update")
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/update", h.UpdateCanvas) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/update", `{"id":41,"name":"After Update"}`)
		require.Equal(t, "000000", resp.Code)

		var item visualization.DataVisualizationInfo
		require.NoError(t, db.First(&item, 41).Error)
		assert.Equal(t, "After Update", item.Name)
	})

	t.Run("rejects invalid body", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/update", h.UpdateCanvas) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/update", `{`)
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/update", h.UpdateCanvas) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/update", `{"id":41,"name":"After Update"}`)
		assert.Equal(t, "500000", resp.Code)
	})
}

func TestVisualizationHandler_DeleteLogic(t *testing.T) {
	t.Parallel()

	t.Run("deletes visualization logically", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		seedVisualizationHandlerRecord(t, db, 51, "Delete Me")
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.DELETE("/delete/:id", h.DeleteLogic) })

		resp := performVisualizationRequest(t, router, http.MethodDelete, "/delete/51", ``)
		require.Equal(t, "000000", resp.Code)

		var item visualization.DataVisualizationInfo
		require.NoError(t, db.Unscoped().First(&item, 51).Error)
		require.NotNil(t, item.DeleteFlag)
		assert.True(t, *item.DeleteFlag)
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.DELETE("/delete/:id", h.DeleteLogic) })

		resp := performVisualizationRequest(t, router, http.MethodDelete, "/delete/bad", ``)
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.DELETE("/delete/:id", h.DeleteLogic) })

		resp := performVisualizationRequest(t, router, http.MethodDelete, "/delete/51", ``)
		assert.Equal(t, "500000", resp.Code)
	})
}

func TestVisualizationHandler_Copy(t *testing.T) {
	t.Parallel()

	t.Run("copies visualization", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		seedVisualizationHandlerRecord(t, db, 61, "Source Viz")
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/copy", h.Copy) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/copy", `{"id":61,"name":"Copied Viz"}`)
		require.Equal(t, "000000", resp.Code)

		newID, err := strconv.ParseInt(strings.Trim(string(resp.Data), `"`), 10, 64)
		require.NoError(t, err)
		var item visualization.DataVisualizationInfo
		require.NoError(t, db.First(&item, newID).Error)
		assert.Equal(t, "Copied Viz", item.Name)
	})

	t.Run("rejects invalid body", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/copy", h.Copy) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/copy", `{`)
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/copy", h.Copy) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/copy", `{"id":61,"name":"Copied Viz"}`)
		assert.Equal(t, "500000", resp.Code)
	})
}

func TestVisualizationHandler_Move(t *testing.T) {
	t.Parallel()

	t.Run("moves visualization", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		seedVisualizationHandlerRecord(t, db, 71, "Movable Viz")
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/move", h.Move) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/move", `{"id":71,"pid":99}`)
		require.Equal(t, "000000", resp.Code)

		var item visualization.DataVisualizationInfo
		require.NoError(t, db.First(&item, 71).Error)
		require.NotNil(t, item.PID)
		assert.Equal(t, int64(99), *item.PID)
	})

	t.Run("rejects invalid body", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/move", h.Move) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/move", `{`)
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/move", h.Move) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/move", `{"id":71,"pid":99}`)
		assert.Equal(t, "500000", resp.Code)
	})
}

func TestVisualizationHandler_NameCheck(t *testing.T) {
	t.Parallel()

	t.Run("checks duplicate name", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		seedVisualizationHandlerRecord(t, db, 81, "Taken Name")
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/name-check", h.NameCheck) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/name-check", `{"name":"Taken Name"}`)
		require.Equal(t, "000000", resp.Code)
		assert.Equal(t, `"repeat"`, string(resp.Data))
	})

	t.Run("rejects invalid body", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/name-check", h.NameCheck) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/name-check", `{`)
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/name-check", h.NameCheck) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/name-check", `{"name":"Taken Name"}`)
		assert.Equal(t, "500000", resp.Code)
	})
}

func TestVisualizationHandler_UpdatePublishStatus(t *testing.T) {
	t.Parallel()

	t.Run("updates publish status", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		seedVisualizationHandlerRecord(t, db, 91, "Publishable Viz")
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/publish", h.UpdatePublishStatus) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/publish", `{"id":91,"status":1}`)
		require.Equal(t, "000000", resp.Code)

		var data visualization.DataVisualizationInfo
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		require.NotNil(t, data.Status)
		assert.Equal(t, 1, *data.Status)
	})

	t.Run("rejects invalid body", func(t *testing.T) {
		t.Parallel()
		db := setupVisualizationHandlerUnitDB(t)
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/publish", h.UpdatePublishStatus) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/publish", `{`)
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() { sqlDB, _ := db.DB(); _ = sqlDB.Close() })
		router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/publish", h.UpdatePublishStatus) })

		resp := performVisualizationRequest(t, router, http.MethodPost, "/publish", `{"id":91,"status":1}`)
		assert.Equal(t, "500000", resp.Code)
	})
}
