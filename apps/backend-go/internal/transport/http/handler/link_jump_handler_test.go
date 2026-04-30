package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	linkJumpSourceDvID       int64 = 1001
	linkJumpTargetDvID       int64 = 1002
	linkJumpSourceViewID     int64 = 2001
	linkJumpTargetViewID     int64 = 2002
	linkJumpSourceTableID    int64 = 3001
	linkJumpTargetTableID    int64 = 3002
	linkJumpSourceFieldID    int64 = 4001
	linkJumpTargetFieldID    int64 = 4002
	linkJumpSeedJumpID       int64 = 5001
	linkJumpSeedJumpInfoID   int64 = 6001
	linkJumpSeedTargetInfoID int64 = 7001
)

type linkJumpHandlerTestEnv struct {
	r  *gin.Engine
	db *gorm.DB
}

type linkJumpHandlerResp struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func setupLinkJumpHandlerTestEnv(t *testing.T) *linkJumpHandlerTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	linkJumpMustExec(t, db, `CREATE TABLE core_chart_view (
		id INTEGER PRIMARY KEY,
		title TEXT,
		scene_id INTEGER,
		table_id INTEGER,
		type TEXT
	)`)
	linkJumpMustExec(t, db, `CREATE TABLE snapshot_core_chart_view (
		id INTEGER PRIMARY KEY,
		title TEXT,
		scene_id INTEGER,
		table_id INTEGER,
		type TEXT,
		jump_active INTEGER,
		update_time INTEGER
	)`)
	linkJumpMustExec(t, db, `CREATE TABLE core_dataset_table_field (
		id INTEGER PRIMARY KEY,
		dataset_group_id INTEGER,
		origin_name TEXT,
		name TEXT,
		de_type INTEGER,
		type TEXT
	)`)
	linkJumpMustExec(t, db, `CREATE TABLE data_visualization_info (
		id INTEGER PRIMARY KEY,
		type TEXT,
		component_data TEXT
	)`)
	linkJumpMustExec(t, db, `CREATE TABLE snapshot_visualization_link_jump (
		id INTEGER PRIMARY KEY,
		source_dv_id INTEGER,
		source_view_id INTEGER,
		link_jump_info TEXT,
		checked INTEGER,
		copy_from INTEGER,
		copy_id INTEGER
	)`)
	linkJumpMustExec(t, db, `CREATE TABLE snapshot_visualization_link_jump_info (
		id INTEGER PRIMARY KEY,
		link_jump_id INTEGER,
		link_type TEXT,
		jump_type TEXT,
		target_dv_id INTEGER,
		source_field_id INTEGER,
		content TEXT,
		checked INTEGER,
		attach_params INTEGER,
		copy_from INTEGER,
		copy_id INTEGER,
		window_size TEXT
	)`)
	linkJumpMustExec(t, db, `CREATE TABLE snapshot_visualization_link_jump_target_view_info (
		target_id INTEGER PRIMARY KEY,
		link_jump_info_id INTEGER,
		source_field_active_id INTEGER,
		target_view_id TEXT,
		target_field_id TEXT,
		copy_from INTEGER,
		copy_id INTEGER,
		target_type TEXT
	)`)

	linkJumpSeedBaseData(t, db)

	repo := repository.NewLinkJumpRepository(db)
	svc := service.NewLinkJumpService(repo)
	h := NewLinkJumpHandler(svc)

	r := gin.New()
	RegisterLinkJumpRoutes(r.Group("/api"), h)

	return &linkJumpHandlerTestEnv{r: r, db: db}
}

func linkJumpMustExec(t *testing.T, db *gorm.DB, sql string, args ...any) {
	t.Helper()
	require.NoError(t, db.Exec(sql, args...).Error)
}

func linkJumpSeedBaseData(t *testing.T, db *gorm.DB) {
	t.Helper()

	linkJumpMustExec(t, db, `INSERT INTO core_chart_view (id, title, scene_id, table_id, type) VALUES (?, ?, ?, ?, ?)`, linkJumpSourceViewID, "Source Chart", linkJumpSourceDvID, linkJumpSourceTableID, "bar")
	linkJumpMustExec(t, db, `INSERT INTO core_chart_view (id, title, scene_id, table_id, type) VALUES (?, ?, ?, ?, ?)`, linkJumpTargetViewID, "Target Chart", linkJumpTargetDvID, linkJumpTargetTableID, "table")
	linkJumpMustExec(t, db, `INSERT INTO snapshot_core_chart_view (id, title, scene_id, table_id, type, jump_active, update_time) VALUES (?, ?, ?, ?, ?, ?, ?)`, linkJumpSourceViewID, "Source Chart", linkJumpSourceDvID, linkJumpSourceTableID, "bar", 1, 1)
	linkJumpMustExec(t, db, `INSERT INTO snapshot_core_chart_view (id, title, scene_id, table_id, type, jump_active, update_time) VALUES (?, ?, ?, ?, ?, ?, ?)`, linkJumpTargetViewID, "Target Chart", linkJumpTargetDvID, linkJumpTargetTableID, "table", 0, 1)

	linkJumpMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type, type) VALUES (?, ?, ?, ?, ?, ?)`, linkJumpSourceFieldID, linkJumpSourceTableID, "province", "Province", 0, "STRING")
	linkJumpMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type, type) VALUES (?, ?, ?, ?, ?, ?)`, linkJumpTargetFieldID, linkJumpTargetTableID, "city", "City", 0, "STRING")

	linkJumpMustExec(t, db, `INSERT INTO data_visualization_info (id, type, component_data) VALUES (?, ?, ?)`, linkJumpSourceDvID, "dashboard", "[]")
	linkJumpMustExec(t, db, `INSERT INTO data_visualization_info (id, type, component_data) VALUES (?, ?, ?)`, linkJumpTargetDvID, "dashboard", "[]")

	linkJumpSeedSnapshotJump(t, db, "seed-jump", 1)
}

func linkJumpSeedSnapshotJump(t *testing.T, db *gorm.DB, linkJumpInfo string, checked int) {
	t.Helper()

	linkJumpMustExec(t, db, `INSERT INTO snapshot_visualization_link_jump (id, source_dv_id, source_view_id, link_jump_info, checked, copy_from, copy_id) VALUES (?, ?, ?, ?, ?, ?, ?)`, linkJumpSeedJumpID, linkJumpSourceDvID, linkJumpSourceViewID, linkJumpInfo, checked, 0, 0)
	linkJumpMustExec(t, db, `INSERT INTO snapshot_visualization_link_jump_info (id, link_jump_id, link_type, jump_type, target_dv_id, source_field_id, content, checked, attach_params, copy_from, copy_id, window_size) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, linkJumpSeedJumpInfoID, linkJumpSeedJumpID, "inner", "_blank", linkJumpTargetDvID, linkJumpSourceFieldID, "", 1, 1, 0, 0, "middle")
	linkJumpMustExec(t, db, `INSERT INTO snapshot_visualization_link_jump_target_view_info (target_id, link_jump_info_id, source_field_active_id, target_view_id, target_field_id, copy_from, copy_id, target_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, linkJumpSeedTargetInfoID, linkJumpSeedJumpInfoID, linkJumpSourceFieldID, fmt.Sprintf("%d", linkJumpTargetViewID), fmt.Sprintf("%d", linkJumpTargetFieldID), 0, 0, "view")
}

func performLinkJumpJSONRequest(t *testing.T, r *gin.Engine, method string, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody []byte
	switch v := body.(type) {
	case nil:
		reqBody = nil
	case []byte:
		reqBody = v
	default:
		var err error
		reqBody, err = json.Marshal(v)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func decodeLinkJumpResp(t *testing.T, body []byte) linkJumpHandlerResp {
	t.Helper()
	var resp linkJumpHandlerResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func TestLinkJumpHandler_GetTableFieldWithViewID(t *testing.T) {
	env := setupLinkJumpHandlerTestEnv(t)

	w := performLinkJumpJSONRequest(t, env.r, http.MethodGet, fmt.Sprintf("/api/linkJump/getTableFieldWithViewId/%d", linkJumpSourceViewID), nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeLinkJumpResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var fields []repository.DatasetFieldDTO
	require.NoError(t, json.Unmarshal(resp.Data, &fields))
	require.Len(t, fields, 1)
	assert.Equal(t, linkJumpSourceFieldID, fields[0].ID)
	assert.Equal(t, linkJumpSourceTableID, fields[0].DatasetTableID)
	assert.Equal(t, "province", fields[0].OriginName)
	assert.Equal(t, "Province", fields[0].Name)
	assert.Equal(t, 0, fields[0].DeType)
}

func TestLinkJumpHandler_QueryWithViewId(t *testing.T) {
	env := setupLinkJumpHandlerTestEnv(t)

	path := fmt.Sprintf("/api/linkJump/queryWithViewId/%d/%d", linkJumpSourceDvID, linkJumpSourceViewID)
	w := performLinkJumpJSONRequest(t, env.r, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeLinkJumpResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data service.LinkJumpDTO
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, linkJumpSeedJumpID, data.ID)
	assert.Equal(t, linkJumpSourceDvID, data.SourceDvID)
	assert.Equal(t, linkJumpSourceViewID, data.SourceViewID)
	assert.Equal(t, "seed-jump", data.LinkJumpInfo)
	require.Len(t, data.LinkJumpInfoArray, 1)
	info := data.LinkJumpInfoArray[0]
	assert.Equal(t, linkJumpSeedJumpID, info.LinkJumpID)
	assert.Equal(t, "inner", info.LinkType)
	assert.Equal(t, "_blank", info.JumpType)
	assert.Equal(t, "middle", info.WindowSize)
	assert.Equal(t, linkJumpTargetDvID, info.TargetDvID)
	assert.Equal(t, linkJumpSourceFieldID, info.SourceFieldID)
	assert.Equal(t, "Province", info.SourceFieldName)
	assert.Equal(t, "dashboard", info.TargetDvType)
	require.Len(t, info.TargetViewInfoList, 1)
	assert.Equal(t, fmt.Sprintf("%d", linkJumpTargetViewID), info.TargetViewInfoList[0].TargetViewID)
	assert.Equal(t, fmt.Sprintf("%d", linkJumpTargetFieldID), info.TargetViewInfoList[0].TargetFieldID)
	assert.Equal(t, "view", info.TargetViewInfoList[0].TargetType)
}

func TestLinkJumpHandler_UpdateJumpSet(t *testing.T) {
	env := setupLinkJumpHandlerTestEnv(t)
	body := service.LinkJumpDTO{
		SourceDvID:   linkJumpSourceDvID,
		SourceViewID: linkJumpSourceViewID,
		LinkJumpInfo: "updated-jump",
		Checked:      true,
		LinkJumpInfoArray: []service.LinkJumpInfoDTO{{
			LinkType:      "outer",
			JumpType:      "_self",
			WindowSize:    "large",
			TargetDvID:    linkJumpTargetDvID,
			SourceFieldID: linkJumpSourceFieldID,
			Content:       "https://example.com",
			Checked:       true,
			AttachParams:  true,
			TargetViewInfoList: []service.LinkJumpTargetViewInfoDTO{{
				SourceFieldActiveID: linkJumpSourceFieldID,
				TargetViewID:        fmt.Sprintf("%d", linkJumpTargetViewID),
				TargetFieldID:       fmt.Sprintf("%d", linkJumpTargetFieldID),
				TargetType:          "filter",
			}},
		}},
	}

	w := performLinkJumpJSONRequest(t, env.r, http.MethodPost, "/api/linkJump/updateJumpSet", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeLinkJumpResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var jumps []map[string]any
	require.NoError(t, env.db.Raw(`SELECT id, source_dv_id, source_view_id, link_jump_info, checked FROM snapshot_visualization_link_jump`).Scan(&jumps).Error)
	require.Len(t, jumps, 1)
	assert.NotEqualValues(t, linkJumpSeedJumpID, jumps[0]["id"])
	assert.EqualValues(t, linkJumpSourceDvID, jumps[0]["source_dv_id"])
	assert.EqualValues(t, linkJumpSourceViewID, jumps[0]["source_view_id"])
	assert.Equal(t, "updated-jump", jumps[0]["link_jump_info"])
	assert.EqualValues(t, 1, jumps[0]["checked"])

	var infos []map[string]any
	require.NoError(t, env.db.Raw(`SELECT link_type, jump_type, target_dv_id, source_field_id, content, checked, attach_params, window_size FROM snapshot_visualization_link_jump_info`).Scan(&infos).Error)
	require.Len(t, infos, 1)
	assert.Equal(t, "outer", infos[0]["link_type"])
	assert.Equal(t, "_self", infos[0]["jump_type"])
	assert.EqualValues(t, linkJumpTargetDvID, infos[0]["target_dv_id"])
	assert.EqualValues(t, linkJumpSourceFieldID, infos[0]["source_field_id"])
	assert.Equal(t, "https://example.com", infos[0]["content"])
	assert.EqualValues(t, 1, infos[0]["checked"])
	assert.EqualValues(t, 1, infos[0]["attach_params"])
	assert.Equal(t, "large", infos[0]["window_size"])

	var targets []map[string]any
	require.NoError(t, env.db.Raw(`SELECT source_field_active_id, target_view_id, target_field_id, target_type FROM snapshot_visualization_link_jump_target_view_info`).Scan(&targets).Error)
	require.Len(t, targets, 1)
	assert.EqualValues(t, linkJumpSourceFieldID, targets[0]["source_field_active_id"])
	assert.Equal(t, fmt.Sprintf("%d", linkJumpTargetViewID), targets[0]["target_view_id"])
	assert.Equal(t, fmt.Sprintf("%d", linkJumpTargetFieldID), targets[0]["target_field_id"])
	assert.Equal(t, "filter", targets[0]["target_type"])

	var oldCount int64
	require.NoError(t, env.db.Raw(`SELECT COUNT(*) FROM snapshot_visualization_link_jump WHERE id = ?`, linkJumpSeedJumpID).Scan(&oldCount).Error)
	assert.Zero(t, oldCount)
}

func TestLinkJumpHandler_QueryVisualizationJumpInfo(t *testing.T) {
	env := setupLinkJumpHandlerTestEnv(t)

	path := fmt.Sprintf("/api/linkJump/queryVisualizationJumpInfo/%d/snapshot", linkJumpSourceDvID)
	w := performLinkJumpJSONRequest(t, env.r, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeLinkJumpResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data service.LinkJumpBaseResponse
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.Len(t, data.BaseJumpInfoMap, 1)
	key := fmt.Sprintf("%d#%d", linkJumpSourceViewID, linkJumpSourceFieldID)
	info, ok := data.BaseJumpInfoMap[key]
	require.True(t, ok)
	assert.Equal(t, linkJumpSeedJumpID, info.LinkJumpID)
	assert.Equal(t, "inner", info.LinkType)
	assert.Equal(t, linkJumpTargetDvID, info.TargetDvID)
	require.Len(t, info.TargetViewInfoList, 1)
	assert.Equal(t, fmt.Sprintf("%d", linkJumpTargetViewID), info.TargetViewInfoList[0].TargetViewID)
}

func TestLinkJumpHandler_UpdateJumpSetActive(t *testing.T) {
	env := setupLinkJumpHandlerTestEnv(t)
	body := service.LinkJumpRequest{
		SourceDvID:   linkJumpSourceDvID,
		SourceViewID: linkJumpSourceViewID,
		ActiveStatus: false,
	}

	w := performLinkJumpJSONRequest(t, env.r, http.MethodPost, "/api/linkJump/updateJumpSetActive", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeLinkJumpResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data service.LinkJumpBaseResponse
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Empty(t, data.BaseJumpInfoMap)

	var jumpActive int
	require.NoError(t, env.db.Raw(`SELECT jump_active FROM snapshot_core_chart_view WHERE id = ?`, linkJumpSourceViewID).Scan(&jumpActive).Error)
	assert.Zero(t, jumpActive)
}

func TestLinkJumpHandler_RemoveJumpSet(t *testing.T) {
	env := setupLinkJumpHandlerTestEnv(t)
	body := service.LinkJumpDTO{
		SourceDvID:   linkJumpSourceDvID,
		SourceViewID: linkJumpSourceViewID,
	}

	w := performLinkJumpJSONRequest(t, env.r, http.MethodPost, "/api/linkJump/removeJumpSet", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeLinkJumpResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var jumpCount int64
	require.NoError(t, env.db.Raw(`SELECT COUNT(*) FROM snapshot_visualization_link_jump WHERE source_dv_id = ? AND source_view_id = ?`, linkJumpSourceDvID, linkJumpSourceViewID).Scan(&jumpCount).Error)
	assert.Zero(t, jumpCount)

	var infoCount int64
	require.NoError(t, env.db.Raw(`SELECT COUNT(*) FROM snapshot_visualization_link_jump_info`).Scan(&infoCount).Error)
	assert.Zero(t, infoCount)

	var targetCount int64
	require.NoError(t, env.db.Raw(`SELECT COUNT(*) FROM snapshot_visualization_link_jump_target_view_info`).Scan(&targetCount).Error)
	assert.Zero(t, targetCount)
}
