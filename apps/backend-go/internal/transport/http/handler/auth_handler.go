package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"io"
	"math/big"
	"strings"
	"sync"
	"time"

	"dataease/backend/internal/domain/auth"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const pkSeparator = "-pk_separator-"

const (
	defaultAdminCredential   = "admin"
	defaultBuiltInCredential = "dataease"
	loginRateLimitRequests   = 10
)

const loginRateLimitWindow = time.Minute

var (
	cryptoOnce       sync.Once
	rsaPrivateKey    *rsa.PrivateKey
	dekeyPayload     string
	symmetricKeyBase string
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	cryptoOnce.Do(initCryptoMaterials)
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) LocalLogin(c *gin.Context) {
	defer recoverServicePanic(c)
	var dto auth.PwdLoginDTO
	if err := c.ShouldBindBodyWith(&dto, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	dto.Name = decryptCredentialIfNeeded(dto.Name)
	dto.Pwd = decryptCredentialIfNeeded(dto.Pwd)

	tokenVO, err := h.authService.LocalLogin(&dto, c.GetHeader("Accept-Language"))
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, tokenVO)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	defer recoverServicePanic(c)
	h.authService.Logout()
	response.Success(c, nil)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.authService == nil {
		response.Error(c, "500000", "auth service is not configured")
		return
	}

	token := strings.TrimPrefix(currentAuthToken(c), "Bearer ")
	if token == "" {
		response.Unauthorized(c, "missing authorization header")
		return
	}

	refreshedToken, exp, err := h.authService.RefreshToken(token)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token": refreshedToken,
		"exp":   exp,
	})
}

func (h *AuthHandler) Dekey(c *gin.Context) {
	defer recoverServicePanic(c)
	response.Success(c, dekeyPayload)
}

func (h *AuthHandler) SymmetricKey(c *gin.Context) {
	defer recoverServicePanic(c)
	response.Success(c, symmetricKeyBase)
}

func (h *AuthHandler) Model(c *gin.Context) {
	defer recoverServicePanic(c)
	response.Success(c, false)
}

func RegisterAuthRoutes(engine *gin.Engine, h *AuthHandler, rateLimitOpts *middleware.RouteRateLimitOptions) {
	loginMiddleware := middleware.RateLimit("login", loginRateLimitRequests, loginRateLimitWindow, middleware.ClientIPKey)
	if rateLimitOpts != nil {
		enabled, maxRequests, window := middleware.ResolveRouteLimit(rateLimitOpts.Config, "login", loginRateLimitRequests, loginRateLimitWindow)
		if enabled {
			loginMiddleware = middleware.ConfigurableRateLimit("login", maxRequests, window, rateLimitOpts.Backend, middleware.ClientIPKey)
		} else {
			loginMiddleware = nil
		}
	}

	loginGroup := engine.Group("")
	if loginMiddleware != nil {
		loginGroup.Use(loginMiddleware)
	}
	loginGroup.POST("/login/localLogin", h.LocalLogin)
	loginGroup.POST("/api/login/localLogin", h.LocalLogin)
	engine.GET("/login/refresh", h.Refresh)
	engine.GET("/logout", h.Logout)
	engine.GET("/dekey", h.Dekey)
	engine.GET("/symmetricKey", h.SymmetricKey)
	engine.GET("/model", h.Model)
	engine.GET("/api/login/refresh", h.Refresh)
	engine.GET("/api/logout", h.Logout)
	engine.GET("/api/dekey", h.Dekey)
	engine.GET("/api/symmetricKey", h.SymmetricKey)
	engine.GET("/api/model", h.Model)
}

func currentAuthToken(c *gin.Context) string {
	token := c.GetHeader("Authorization")
	if token != "" {
		return token
	}
	return c.GetHeader("X-DE-TOKEN")
}

func initCryptoMaterials() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	rsaPrivateKey = key

	pubASN1, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		panic(err)
	}
	pubBase64 := base64.StdEncoding.EncodeToString(pubASN1)

	aesKey := randomAlphaNum(16)
	encryptedPublicKey, err := aesEncryptCBCPKCS7(pubBase64, aesKey)
	if err != nil {
		panic(err)
	}

	sep := base64.RawURLEncoding.EncodeToString([]byte(pkSeparator)) + "="
	dekeyPayload = encryptedPublicKey + sep + aesKey

	symmetricRaw := make([]byte, 16)
	if _, err = io.ReadFull(rand.Reader, symmetricRaw); err != nil {
		panic(err)
	}
	symmetricKeyBase = base64.StdEncoding.EncodeToString(symmetricRaw)
}

func decryptCredentialIfNeeded(v string) string {
	if v == "" || rsaPrivateKey == nil {
		return v
	}
	decoded, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return v
	}
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, decoded)
	if err != nil {
		if v == defaultAdminCredential || v == defaultBuiltInCredential {
			return v
		}
		return v
	}
	return string(plain)
}

func aesEncryptCBCPKCS7(plainText string, key string) (string, error) {
	if len(key) != 16 {
		return "", errors.New("invalid aes key length")
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	iv := []byte("0000000000000000")
	data := pkcs7Pad([]byte(plainText), block.BlockSize())
	encrypted := make([]byte, len(data))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encrypted, data)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - (len(data) % blockSize)
	padding := make([]byte, padLen)
	for i := range padding {
		padding[i] = byte(padLen)
	}
	return append(data, padding...)
}

func randomAlphaNum(n int) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	max := big.NewInt(int64(len(chars)))
	for i := range n {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}
