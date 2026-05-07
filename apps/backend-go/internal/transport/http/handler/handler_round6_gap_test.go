package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRound6VisDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:round6_%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func setupRound6DsEnv(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:round6_%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))

	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Set("userId", int64(42))
		c.Set("username", "round6tester")
		c.Next()
	})
	registerDatasourceCompatRoutes(r, h, nil, nil)
	return r, db
}

func setupRound6FontDB(t *testing.T) (*gorm.DB, *repository.TypefaceRepository) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.CoreFont{}))
	repo := repository.NewTypefaceRepository(db)
	return db, repo
}

func setupRound6OrgRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:round6_%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&org.SysOrg{}))

	orgRepo := repository.NewOrgRepository(db)
	orgSvc := service.NewOrgService(orgRepo, nil, nil, nil, nil, nil)
	h := NewOrgHandler(orgSvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", uint64(1))
		c.Set("user_id", uint64(7))
		c.Set("username", "round6admin")
		c.Next()
	})
	RegisterOrgRoutes(r.Group("/api"), h)
	return r, db
}

func round6Request(t *testing.T, r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeRound6Resp(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func TestRound6_Vis_ParseJSONStrings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := map[string]interface{}{
		"jsonObj":  `{"nested":"val"}`,
		"jsonArr":  `[1,2,3]`,
		"plainStr": "hello",
		"number":   float64(42),
	}
	parseJSONStrings(m)
	assert.Equal(t, map[string]interface{}{"nested": "val"}, m["jsonObj"])
	assert.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, m["jsonArr"])
	assert.Equal(t, "hello", m["plainStr"])
	assert.Equal(t, float64(42), m["number"])
}

func TestRound6_Vis_DeleteLogicWithBusiFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound6VisDB(t)
	router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) {
		r.DELETE("/delete/:id/:busiFlag", h.DeleteLogic)
	})

	resp := performVisualizationRequest(t, router, http.MethodDelete, "/delete/999/dashboard", "")
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6_Vis_FindCopyResourceNotTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupVisualizationHandlerUnitDB(t)
	require.NoError(t, db.Exec(`INSERT INTO data_visualization_info (id, name, type, node_type, status, pid, create_by) VALUES (601, 'NotTemplate', 'dashboard', 'leaf', 1, 5, 'tester')`).Error)

	router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) {
		r.GET("/findCopyResource/:dvId/:busiFlag", h.FindCopyResource)
	})

	resp := performVisualizationRequest(t, router, http.MethodGet, "/findCopyResource/601/dashboard", "")
	require.Equal(t, "000000", resp.Code)
	assert.Equal(t, "", strings.Trim(string(resp.Data), `"`))
}

func TestRound6_Vis_RecordExportLogWithUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound6VisDB(t)
	router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) {
		r.POST("/exportLogApp", h.ExportLogApp)
		r.POST("/exportLogTemplate", h.ExportLogTemplate)
	})

	for _, path := range []string{"/exportLogApp", "/exportLogTemplate"} {
		resp := performVisualizationRequest(t, router, http.MethodPost, path, `{"id":1,"type":"dashboard"}`)
		require.Equal(t, "000000", resp.Code)
	}
}

func TestRound6_Vis_ResolveBusiTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		input   string
		want    []string
		wantErr bool
	}{
		{"", []string{"dashboard", "dataV"}, false},
		{"dashboard-dataV", []string{"dashboard", "dataV"}, false},
		{"panel", []string{"dashboard"}, false},
		{"screen", []string{"dataV"}, false},
		{"dashboard", []string{"dashboard"}, false},
		{"dataV", []string{"dataV"}, false},
		{"bad", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := resolveBusiTypes(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRound6_BridgeDs_MoveCompat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db := setupRound6DsEnv(t)
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 10, PID: int64PtrForDatasourceHandler(0), Name: "moveme", Type: "MySQL"})

	w := round6Request(t, r, http.MethodPost, "/datasource/move", `{"id":10,"pid":5}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_BridgeDs_MoveCompatInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupRound6DsEnv(t)

	w := round6Request(t, r, http.MethodPost, "/datasource/move", `{"id":"bad"}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp["code"])
}

func TestRound6_BridgeDs_RenameCompat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db := setupRound6DsEnv(t)
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 20, PID: int64PtrForDatasourceHandler(0), Name: "oldname", Type: "MySQL"})

	w := round6Request(t, r, http.MethodPost, "/datasource/reName", `{"id":20,"name":"newname"}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_BridgeDs_RenameCompatInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupRound6DsEnv(t)

	w := round6Request(t, r, http.MethodPost, "/datasource/reName", `{"id":-1,"name":"x"}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp["code"])
}

func TestRound6_BridgeDs_CreateFolderCompat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupRound6DsEnv(t)

	w := round6Request(t, r, http.MethodPost, "/datasource/createFolder", `{"name":"my-folder","pid":0}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_BridgeDs_GetCompat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db := setupRound6DsEnv(t)
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 30, PID: int64PtrForDatasourceHandler(0), Name: "getme", Type: "MySQL"})

	w := round6Request(t, r, http.MethodGet, "/datasource/get/30", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_BridgeDs_GetCompatInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupRound6DsEnv(t)

	w := round6Request(t, r, http.MethodGet, "/datasource/get/abc", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp["code"])
}

func TestRound6_BridgeDs_HidePwCompat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db := setupRound6DsEnv(t)
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 40, PID: int64PtrForDatasourceHandler(0), Name: "hideme", Type: "MySQL"})

	w := round6Request(t, r, http.MethodGet, "/datasource/hidePw/40", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_BridgeDs_GetSimpleDsCompat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db := setupRound6DsEnv(t)
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 50, PID: int64PtrForDatasourceHandler(0), Name: "simple", Type: "MySQL"})

	w := round6Request(t, r, http.MethodGet, "/datasource/getSimpleDs/50", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "50", data["id"])
	assert.Equal(t, "simple", data["name"])
	assert.Equal(t, "MySQL", data["type"])
}

func TestRound6_BridgeDs_MoveCompatMissingPID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db := setupRound6DsEnv(t)
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 60, PID: int64PtrForDatasourceHandler(0), Name: "nopid", Type: "MySQL"})

	w := round6Request(t, r, http.MethodPost, "/datasource/move", `{"id":60}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_BridgeDs_CreateFolderCompatNoPID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupRound6DsEnv(t)

	w := round6Request(t, r, http.MethodPost, "/datasource/createFolder", `{"name":"no-pid-folder"}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_Font_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, repo := setupRound6FontDB(t)
	require.NoError(t, repo.CreateFont(&auto.CoreFont{ID: 1, Name: "Arial", IsDefault: true, UpdateTime: 1}))
	require.NoError(t, repo.CreateFont(&auto.CoreFont{ID: 2, Name: "Courier", IsBuiltIn: true, UpdateTime: 2}))

	h := NewFontHandler(repo)
	r := gin.New()
	r.GET("/typeface/listFont", h.List)

	w := round6Request(t, r, http.MethodGet, "/typeface/listFont", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
}

func TestRound6_Font_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, repo := setupRound6FontDB(t)
	h := NewFontHandler(repo)
	r := gin.New()
	r.POST("/typeface/create", h.Create)

	w := round6Request(t, r, http.MethodPost, "/typeface/create", `{"name":"Helvetica","fileName":"helv.ttf","isDefault":false,"size":12.5,"sizeType":"KB"}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Helvetica", data["name"])
	assert.Equal(t, false, data["isBuiltin"])
}

func TestRound6_Font_CreateDuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, repo := setupRound6FontDB(t)
	require.NoError(t, repo.CreateFont(&auto.CoreFont{ID: 1, Name: "Exists", UpdateTime: 1}))

	h := NewFontHandler(repo)
	r := gin.New()
	r.POST("/typeface/create", h.Create)

	w := round6Request(t, r, http.MethodPost, "/typeface/create", `{"name":"Exists"}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp["code"])
}

func TestRound6_Font_EditExisting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, repo := setupRound6FontDB(t)
	require.NoError(t, repo.CreateFont(&auto.CoreFont{ID: 100, Name: "OldName", UpdateTime: 1}))

	h := NewFontHandler(repo)
	r := gin.New()
	r.POST("/typeface/edit", h.Edit)

	w := round6Request(t, r, http.MethodPost, "/typeface/edit", `{"id":100,"name":"NewName","isDefault":true}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NewName", data["name"])
}

func TestRound6_Font_EditCreateWhenZeroID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, repo := setupRound6FontDB(t)
	h := NewFontHandler(repo)
	r := gin.New()
	r.POST("/typeface/edit", h.Edit)

	w := round6Request(t, r, http.MethodPost, "/typeface/edit", `{"id":0,"name":"ViaEditCreate"}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_Font_DefaultFont(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, repo := setupRound6FontDB(t)
	require.NoError(t, repo.CreateFont(&auto.CoreFont{ID: 200, Name: "Default1", IsDefault: true, UpdateTime: 1}))
	require.NoError(t, repo.CreateFont(&auto.CoreFont{ID: 201, Name: "NotDefault", IsDefault: false, UpdateTime: 2}))

	h := NewFontHandler(repo)
	r := gin.New()
	r.GET("/typeface/defaultFont", h.DefaultFont)

	w := round6Request(t, r, http.MethodGet, "/typeface/defaultFont", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 1)
}

func TestRound6_Font_IsAllowedExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert.True(t, isAllowedFontDownloadExtension("demo.ttf"))
	assert.True(t, isAllowedFontDownloadExtension("demo.otf"))
	assert.True(t, isAllowedFontDownloadExtension("demo.woff"))
	assert.True(t, isAllowedFontDownloadExtension("demo.woff2"))
	assert.True(t, isAllowedFontDownloadExtension("demo.TTF"))
	assert.False(t, isAllowedFontDownloadExtension("demo.exe"))
	assert.False(t, isAllowedFontDownloadExtension("demo.txt"))
	assert.False(t, isAllowedFontDownloadExtension("demo"))
}

func TestRound6_Font_ResolveSafePath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	safe, ok := resolveSafeFontFilePath("/fonts", "arial.ttf")
	assert.True(t, ok)
	assert.Equal(t, "/fonts/arial.ttf", safe)

	_, ok = resolveSafeFontFilePath("/fonts", "")
	assert.False(t, ok)

	_, ok = resolveSafeFontFilePath("/fonts", "../etc/passwd")
	assert.False(t, ok)

	_, ok = resolveSafeFontFilePath("/fonts", "/absolute/path.ttf")
	assert.False(t, ok)

	_, ok = resolveSafeFontFilePath("/fonts", "evil.exe")
	assert.False(t, ok)
}

func TestRound6_Font_FontToDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)
	font := &auto.CoreFont{
		ID:            42,
		Name:          "TestFont",
		FileName:      "test.ttf",
		FileTransName: "uuid.ttf",
		IsDefault:     true,
		IsBuiltIn:     false,
		Size:          1024.0,
		SizeType:      "KB",
	}
	dto := fontToDTO(font)
	assert.Equal(t, int64(42), dto.ID)
	assert.Equal(t, "TestFont", dto.Name)
	assert.Equal(t, "test.ttf", dto.FileName)
	assert.Equal(t, "uuid.ttf", dto.FileTransName)
	assert.True(t, dto.IsDefault)
	assert.False(t, dto.IsBuiltIn)
	assert.Equal(t, 1024.0, dto.Size)
	assert.Equal(t, "KB", dto.SizeType)
}

func TestRound6_Org_ListOrgs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db := setupRound6OrgRouter(t)
	require.NoError(t, db.Create(&org.SysOrg{OrgName: "ListOrg1", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}).Error)
	require.NoError(t, db.Create(&org.SysOrg{OrgName: "ListOrg2", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}).Error)

	w := round6Request(t, r, http.MethodGet, "/api/system/organization/list", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
}

func TestRound6_Org_GetOrgByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db := setupRound6OrgRouter(t)
	require.NoError(t, db.Create(&org.SysOrg{OrgName: "GetOrg", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}).Error)

	var created org.SysOrg
	require.NoError(t, db.Where("org_name = ?", "GetOrg").First(&created).Error)

	w := round6Request(t, r, http.MethodGet, fmt.Sprintf("/api/system/organization/info/%d", created.OrgID), "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_Org_GetOrgTree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupRound6OrgRouter(t)

	w := round6Request(t, r, http.MethodGet, "/api/system/organization/tree", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_Org_CheckOrgName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupRound6OrgRouter(t)

	w := round6Request(t, r, http.MethodGet, "/api/system/organization/checkName?orgName=TestOrg", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_Org_CheckOrgNameMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupRound6OrgRouter(t)

	w := round6Request(t, r, http.MethodGet, "/api/system/organization/checkName", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp["code"])
}

func TestRound6_Org_GetChildOrgs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db := setupRound6OrgRouter(t)
	require.NoError(t, db.Create(&org.SysOrg{OrgName: "Parent4Child", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}).Error)

	var parent org.SysOrg
	require.NoError(t, db.Where("org_name = ?", "Parent4Child").First(&parent).Error)
	require.NoError(t, db.Create(&org.SysOrg{OrgName: "ChildOrg", ParentID: parent.OrgID, Level: 2, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}).Error)

	w := round6Request(t, r, http.MethodGet, fmt.Sprintf("/api/system/organization/children/%d", parent.OrgID), "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 1)
}

func TestRound6_Org_MountedAlias(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupRound6OrgRouter(t)

	w := round6Request(t, r, http.MethodPost, "/api/org/mounted", "")
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp["code"])
}

func TestRound6_Org_TransferUserAlias(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupRound6OrgRouter(t)

	w := round6Request(t, r, http.MethodPost, "/api/organization/transfer-user", `{"sourceOrgId":1,"targetOrgId":2,"userId":99}`)
	resp := decodeRound6Resp(t, w.Body.Bytes())
	assert.Contains(t, []string{"000000", "500000"}, resp["code"])
}
