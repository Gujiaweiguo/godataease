//go:build integration

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openVisualizationHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := envOrDefaultVisualization("TEST_DB_HOST", "localhost")
	port := envOrDefaultVisualization("TEST_DB_PORT", "3306")
	user := envOrDefaultVisualization("TEST_DB_USER", "root")
	password := envOrDefaultVisualization("TEST_DB_PASSWORD", "Admin168")
	name := envOrDefaultVisualization("TEST_DB_NAME", "dataease_test")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&visualization.DataVisualizationInfo{}))
	require.NoError(t, db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error)
	require.NoError(t, db.Exec("DELETE FROM data_visualization_info").Error)
	require.NoError(t, db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error)

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func envOrDefaultVisualization(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func TestVisualizationHandler_FindByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openVisualizationHandlerTestDB(t)
	repo := repository.NewVisualizationRepository(db)
	h := NewVisualizationHandler(service.NewVisualizationService(repo))

	r := gin.New()
	r.POST("/dataVisualization/findById", h.FindByID)

	req := httptest.NewRequest(http.MethodPost, "/dataVisualization/findById", strings.NewReader(`{"id":999999}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "50001", resp.Code)
	assert.Equal(t, "Visualization not found", resp.Msg)
}
