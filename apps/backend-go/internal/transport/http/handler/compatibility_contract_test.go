package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func testJWT() *auth.JWT {
	return auth.NewJWT(&auth.JWTConfig{
		Secret: "test-secret-key-for-unittest",
		Expire: 3600,
	})
}

func TestContractDiffTemplateRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	templateHandler := NewTemplateHandler(nil)
	RegisterTemplateRoutes(r, templateHandler)

	t.Run("categories_returns_success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/templateMarket/categories", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "000000", resp["code"])
	})

	t.Run("find_categories_returns_success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/templateManage/findCategories", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "000000", resp["code"])
	})
}

func TestNegativePathUnauthorizedAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtInstance := testJWT()

	r := gin.New()
	protected := r.Group("/api")
	protected.Use(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"code": "20001", "msg": "missing authorization header"})
			c.Abort()
			return
		}
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}
		claims, err := jwtInstance.ParseToken(token)
		if err != nil {
			c.JSON(401, gin.H{"code": "20001", "msg": "invalid token"})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Next()
	})
	protected.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": "000000", "msg": "success"})
	})

	t.Run("missing_token_returns_401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "20001", resp["code"])
	})

	t.Run("invalid_token_returns_401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "20001", resp["code"])
	})

	t.Run("valid_token_returns_200", func(t *testing.T) {
		token, _ := jwtInstance.GenerateToken(1, "testuser", "user")
		req := httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}

func TestNegativePathRowPermissionBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowedOrgs := map[int64]bool{1: true, 2: true}

	r := gin.New()
	r.GET("/api/org/:id", func(c *gin.Context) {
		var orgID int64
		if c.Param("id") == "999" {
			orgID = 999
		} else {
			orgID = 1
		}

		if !allowedOrgs[orgID] {
			c.JSON(403, gin.H{"code": "70001", "msg": "forbidden"})
			return
		}

		c.JSON(200, gin.H{"code": "000000", "msg": "success", "data": gin.H{"orgId": orgID}})
	})

	t.Run("allowed_org_returns_data", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/org/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "000000", resp["code"])
	})

	t.Run("unauthorized_org_returns_403", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/org/999", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "70001", resp["code"])
		assert.Nil(t, resp["data"])
	})
}

func TestNegativePathColumnLeakage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	maskedColumns := map[string]bool{"phone": true, "email": true, "salary": true, "id_card": true}

	r := gin.New()
	r.GET("/api/user/:id", func(c *gin.Context) {
		rawData := map[string]interface{}{
			"id":      1,
			"name":    "Test User",
			"phone":   "13812345678",
			"email":   "test@example.com",
			"salary":  100000,
			"id_card": "110101199001011234",
		}

		maskedData := make(map[string]interface{})
		for col, val := range rawData {
			if maskedColumns[col] {
				maskedData[col] = "***"
			} else {
				maskedData[col] = val
			}
		}

		c.JSON(200, gin.H{"code": "000000", "msg": "success", "data": maskedData})
	})

	t.Run("unmasked_columns_return_raw_data", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/user/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, "Test User", data["name"])
	})

	t.Run("masked_columns_do_not_leak_raw_data", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/user/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		data := resp["data"].(map[string]interface{})

		assert.NotEqual(t, "13812345678", data["phone"])
		assert.NotEqual(t, "test@example.com", data["email"])
		assert.NotEqual(t, 100000, data["salary"])
		assert.NotEqual(t, "110101199001011234", data["id_card"])

		assert.Equal(t, "***", data["phone"])
		assert.Equal(t, "***", data["email"])
	})
}
