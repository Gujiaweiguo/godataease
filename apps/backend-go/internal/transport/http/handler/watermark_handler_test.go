package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWatermarkService struct {
	findFunc func() (*visualization.Watermark, error)
	saveFunc func(req *visualization.WatermarkSaveRequest, createBy string) (*visualization.Watermark, error)
}

func (m *mockWatermarkService) Find() (*visualization.Watermark, error) {
	if m.findFunc != nil {
		return m.findFunc()
	}
	return &visualization.Watermark{ID: "default"}, nil
}

func (m *mockWatermarkService) Save(req *visualization.WatermarkSaveRequest, createBy string) (*visualization.Watermark, error) {
	if m.saveFunc != nil {
		return m.saveFunc(req, createBy)
	}
	return &visualization.Watermark{ID: "default"}, nil
}

func setupWatermarkTestRouter(svc *mockWatermarkService) *gin.Engine {
	return setupWatermarkTestRouterWithContext(svc, nil)
}

func setupWatermarkTestRouterWithContext(svc *mockWatermarkService, contextSetter gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if contextSetter != nil {
		r.Use(contextSetter)
	}

	var watermarkSvc *service.WatermarkService
	if svc != nil {
		watermarkSvc = service.NewWatermarkService(&mockWatermarkRepoAdapter{svc: svc})
	}
	h := NewWatermarkHandler(watermarkSvc)
	RegisterWatermarkRoutes(r, h)
	return r
}

func authenticatedWatermarkContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}
}

type mockWatermarkRepoAdapter struct {
	svc *mockWatermarkService
}

func (m *mockWatermarkRepoAdapter) FindLatest() (*visualization.Watermark, error) {
	return m.svc.Find()
}

func (m *mockWatermarkRepoAdapter) SaveDefault(settingContent string, createBy string, createTime int64) (*visualization.Watermark, error) {
	return m.svc.Save(&visualization.WatermarkSaveRequest{SettingContent: settingContent}, createBy)
}

func TestWatermarkHandler_Find_Success(t *testing.T) {
	mock := &mockWatermarkService{
		findFunc: func() (*visualization.Watermark, error) {
			return &visualization.Watermark{
				ID:             "default",
				Version:        "v1",
				SettingContent: `{"enable":true}`,
			}, nil
		},
	}
	router := setupWatermarkTestRouterWithContext(mock, authenticatedWatermarkContext())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/watermark/find", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "000000", resp["code"])
}

func TestWatermarkHandler_Save_Success(t *testing.T) {
	var savedBy string
	mock := &mockWatermarkService{
		saveFunc: func(req *visualization.WatermarkSaveRequest, createBy string) (*visualization.Watermark, error) {
			savedBy = createBy
			return &visualization.Watermark{
				ID:             "default",
				SettingContent: req.SettingContent,
			}, nil
		},
	}
	router := setupWatermarkTestRouterWithContext(mock, authenticatedWatermarkContext())

	body := map[string]string{"settingContent": `{"enable":true}`}
	jsonBody, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/watermark/save", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "000000", resp["code"])
	assert.Equal(t, "100", savedBy)
}

func TestWatermarkHandler_Save_InvalidJSON(t *testing.T) {
	router := setupWatermarkTestRouterWithContext(nil, authenticatedWatermarkContext())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/watermark/save", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code) // Error response still returns 200 with error code
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEqual(t, "000000", resp["code"])
}

func TestWatermarkHandler_Find_Unauthorized(t *testing.T) {
	router := setupWatermarkTestRouter(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/watermark/find", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "20001", resp["code"])
}
