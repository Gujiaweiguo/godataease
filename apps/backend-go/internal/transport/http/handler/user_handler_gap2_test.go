package handler

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserHandlerGap2(t *testing.T) (*UserHandler, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&org.SysOrg{}, &role.SysRole{}, &user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{}))
	require.NoError(t, db.Create(&user.SysUser{UserID: 1, Username: "admin", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	userSvc := service.NewUserService(
		repository.NewUserRepository(db),
		repository.NewUserRoleRepository(db),
		repository.NewUserPermRepository(db),
	)
	userSvc.SetOrgRepository(repository.NewOrgRepository(db))
	userSvc.SetRoleRepository(repository.NewRoleRepository(db))
	return NewUserHandler(userSvc, service.NewUserImportService(userSvc)), db
}

func TestUserHandlerGap2_SetAuthService(t *testing.T) {
	h := &UserHandler{}
	h.SetAuthService(nil)
	assert.Nil(t, h.buildBootstrap)
	assert.Nil(t, h.switchOrg)
}

func TestUserHandlerGap2_CreateUpdateDelete(t *testing.T) {
	h, db := setupUserHandlerGap2(t)
	seedEnabledOrgAndRole(t, db, 9)
	r := gin.New()
	r.POST("/user/create", func(c *gin.Context) {
		c.Set("org_id", int64(9))
		h.CreateUser(c)
	})
	r.POST("/user/update", func(c *gin.Context) {
		c.Set("org_id", int64(9))
		h.UpdateUser(c)
	})
	r.POST("/user/delete/:id", func(c *gin.Context) {
		c.Set("org_id", int64(9))
		h.DeleteUser(c)
	})

	createReq := httptest.NewRequest(http.MethodPost, "/user/create", bytes.NewBufferString(`{"username":"gap2-user","password":"secret","realName":"Gap Two"}`))
	createReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, createReq)
	assert.Equal(t, http.StatusOK, w.Code)
	var createResp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	assert.Equal(t, "000000", createResp["code"])
	createdID := int64(createResp["data"].(float64))

	var role user.SysUserRole
	require.NoError(t, db.Where("user_id = ?", createdID).First(&role).Error)
	assert.Equal(t, int64(9), role.OrgID)

	updateReq := httptest.NewRequest(http.MethodPost, "/user/update", bytes.NewBufferString(`{"id":`+strconv.FormatInt(createdID, 10)+`,"realName":"Renamed","status":0}`))
	updateReq.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, updateReq)
	assert.Equal(t, http.StatusOK, w.Code)
	var updateResp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updateResp))
	assert.Equal(t, "000000", updateResp["code"])

	var updated user.SysUser
	require.NoError(t, db.Where("user_id = ?", createdID).First(&updated).Error)
	assert.Equal(t, "Renamed", updated.NickName)
	assert.Equal(t, user.StatusDisabled, updated.Status)

	deleteReq := httptest.NewRequest(http.MethodPost, "/user/delete/"+strconv.FormatInt(createdID, 10), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, deleteReq)
	assert.Equal(t, http.StatusOK, w.Code)
	var deleteResp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &deleteResp))
	assert.Equal(t, "000000", deleteResp["code"])
	require.NoError(t, db.Where("user_id = ?", createdID).First(&updated).Error)
	assert.Equal(t, user.DelFlagDeleted, updated.DelFlag)
}

func TestUserHandlerGap2_SwitchLanguage(t *testing.T) {
	h, db := setupUserHandlerGap2(t)
	require.NoError(t, db.Create(&user.SysUser{UserID: 2, Username: "lang-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	r := gin.New()
	r.POST("/user/lang", func(c *gin.Context) {
		c.Set("user_id", uint64(2))
		h.SwitchLanguage(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/user/lang", bytes.NewBufferString(`{"lang":"en_US"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
	var updated user.SysUser
	require.NoError(t, db.Where("user_id = ?", 2).First(&updated).Error)
	if assert.NotNil(t, updated.Language) {
		assert.Equal(t, "en", *updated.Language)
	}
}

func TestUserHandlerGap2_BatchImportAndErrorRecords(t *testing.T) {
	h, db := setupUserHandlerGap2(t)
	seedEnabledOrgAndRole(t, db, 7)
	r := gin.New()
	r.POST("/user/batchImport", func(c *gin.Context) {
		c.Set("org_id", int64(7))
		c.Set("username", "importer")
		h.BatchImportUsers(c)
	})
	r.GET("/user/errorRecord/:key", h.DownloadErrorRecord)
	r.GET("/user/clearErrorRecord/:key", h.ClearErrorRecord)

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	part, err := writer.CreateFormFile("file", "users.csv")
	require.NoError(t, err)
	csvWriter := csv.NewWriter(part)
	require.NoError(t, csvWriter.Write([]string{"username", "realName", "email", "phone"}))
	require.NoError(t, csvWriter.Write([]string{"import-ok", "Import Ok", "ok@example.com", "123"}))
	require.NoError(t, csvWriter.Write([]string{"import-bad", "Import Bad", "bad-email", "456"}))
	csvWriter.Flush()
	require.NoError(t, csvWriter.Error())
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/user/batchImport", buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(2), data["totalRows"])
	assert.Equal(t, float64(1), data["successRows"])
	assert.Equal(t, float64(1), data["failedRows"])
	errorKey := data["errorKey"].(string)

	var imported user.SysUser
	require.NoError(t, db.Where("username = ?", "import-ok").First(&imported).Error)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/user/errorRecord/"+errorKey, nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "user_import_error_"+errorKey+".xlsx")
	assert.NotEmpty(t, w.Body.Bytes())

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/user/clearErrorRecord/"+errorKey, nil))
	assert.Equal(t, http.StatusOK, w.Code)
	var clearResp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &clearResp))
	assert.Equal(t, "000000", clearResp["code"])
	_, statErr := os.Stat(filepath.Join(os.TempDir(), "dataease", "user-import-errors", "user_import_error_"+errorKey+".xlsx"))
	assert.Error(t, statErr)
}

func seedEnabledOrgAndRole(t *testing.T, db *gorm.DB, orgID int64) {
	t.Helper()
	require.NoError(t, db.Create(&org.SysOrg{OrgID: orgID, OrgName: "org-" + strconv.FormatInt(orgID, 10), ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}).Error)
	var count int64
	require.NoError(t, db.Model(&role.SysRole{}).Where("role_code = ?", role.BuiltInOrgUserRoleCode).Count(&count).Error)
	if count == 0 {
		roleType := role.RoleTypeOrganization
		dataScope := role.DataScopeSelf
		createBy := "test"
		require.NoError(t, db.Create(&role.SysRole{RoleName: role.BuiltInOrgUserRoleName, RoleCode: role.BuiltInOrgUserRoleCode, RoleType: &roleType, DataScope: &dataScope, Status: role.StatusEnabled, CreateBy: &createBy}).Error)
	}
}
