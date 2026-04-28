package handler

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	domainauth "dataease/backend/internal/domain/auth"
	pkgauth "dataease/backend/internal/pkg/auth"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_DekeyAndSymmetricKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(nil)
	r := gin.New()
	r.GET("/dekey", h.Dekey)
	r.GET("/symmetricKey", h.SymmetricKey)

	t.Run("dekey", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/dekey", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp struct {
			Code string `json:"code"`
			Data string `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp.Code)
		assert.NotEmpty(t, resp.Data)
	})

	t.Run("symmetric key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/symmetricKey", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp struct {
			Code string `json:"code"`
			Data string `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp.Code)
		assert.NotEmpty(t, resp.Data)
	})
}

func TestAuthHandler_ModelAndRouteAliases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(service.NewAuthService(nil, nil, nil, nil))
	r := gin.New()
	RegisterAuthRoutes(r, h)

	for _, path := range []string{"/model", "/api/model"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			var resp struct {
				Code string `json:"code"`
				Data bool   `json:"data"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, "000000", resp.Code)
			assert.False(t, resp.Data)
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwt := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "test-secret", Expire: 3600})
	h := NewAuthHandler(service.NewAuthService(nil, nil, nil, jwt))
	r := gin.New()
	r.GET("/login/refresh", h.Refresh)

	t.Run("missing token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/login/refresh", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
		var resp bridgeCodeResp
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "20001", resp.Code)
	})

	t.Run("success", func(t *testing.T) {
		token, err := jwt.GenerateTokenWithOrgID(7, "alice", "", 3)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/login/refresh", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp struct {
			Code string                 `json:"code"`
			Data map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp.Code)
		refreshed, ok := resp.Data["token"].(string)
		require.True(t, ok)
		assert.NotEmpty(t, refreshed)
		claims, err := jwt.ParseToken(refreshed)
		require.NoError(t, err)
		assert.Equal(t, uint64(7), claims.UserID)
		assert.Equal(t, uint64(3), claims.OrgID)
	})
}

func TestAuthHandler_LocalLogin_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(service.NewAuthService(nil, nil, nil, nil))
	r := gin.New()
	r.POST("/login/localLogin", h.LocalLogin)

	req := httptest.NewRequest("POST", "/login/localLogin", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp bridgeCodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
}

func TestRegisterAuthRoutes_RateLimitsLocalLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(service.NewAuthService(nil, nil, nil, nil))
	r := gin.New()
	RegisterAuthRoutes(r, h)

	for i := 0; i < loginRateLimitRequests; i++ {
		req := httptest.NewRequest("POST", "/login/localLogin", strings.NewReader("{"))
		req.RemoteAddr = "192.0.2.25:8080"
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"500000"`)
	}

	req := httptest.NewRequest("POST", "/api/login/localLogin", strings.NewReader("{"))
	req.RemoteAddr = "192.0.2.25:8080"
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 429, w.Code)
	var resp bridgeCodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "429001", resp.Code)
}

func TestDecryptCredentialIfNeeded(t *testing.T) {
	NewAuthHandler(nil)

	t.Run("plain text passthrough", func(t *testing.T) {
		assert.Equal(t, defaultAdminCredential, decryptCredentialIfNeeded(defaultAdminCredential))
		assert.Equal(t, "plain-text", decryptCredentialIfNeeded("plain-text"))
	})

	t.Run("encrypted roundtrip", func(t *testing.T) {
		plain := domainauth.PwdLoginDTO{Name: "encrypted-user", Pwd: "secret-password"}
		encryptedName := encryptCredentialForTest(t, plain.Name)
		encryptedPwd := encryptCredentialForTest(t, plain.Pwd)

		assert.Equal(t, plain.Name, decryptCredentialIfNeeded(encryptedName))
		assert.Equal(t, plain.Pwd, decryptCredentialIfNeeded(encryptedPwd))
	})
}

func TestAuthCryptoHelpers(t *testing.T) {
	t.Run("randomAlphaNum length and charset", func(t *testing.T) {
		value := randomAlphaNum(24)
		assert.Len(t, value, 24)
		for _, ch := range value {
			assert.Contains(t, "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", string(ch))
		}
	})

	t.Run("pkcs7Pad uses full block", func(t *testing.T) {
		padded := pkcs7Pad([]byte("abc"), 8)
		require.Len(t, padded, 8)
		assert.Equal(t, []byte{5, 5, 5, 5, 5}, padded[3:])
	})

	t.Run("aesEncrypt rejects wrong key length", func(t *testing.T) {
		_, err := aesEncryptCBCPKCS7("payload", "short")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid aes key length")
	})
}

func encryptCredentialForTest(t *testing.T, plain string) string {
	t.Helper()
	require.NotNil(t, rsaPrivateKey)
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, &rsaPrivateKey.PublicKey, []byte(plain))
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(cipherText)
}
