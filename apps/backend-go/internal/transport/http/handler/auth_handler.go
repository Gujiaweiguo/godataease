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
	"sync"

	"dataease/backend/internal/domain/auth"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

const pkSeparator = "-pk_separator-"

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
	var dto auth.PwdLoginDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	dto.Name = decryptCredentialIfNeeded(dto.Name)
	dto.Pwd = decryptCredentialIfNeeded(dto.Pwd)

	tokenVO, err := h.authService.LocalLogin(&dto)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, tokenVO)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.authService.Logout()
	response.Success(c, nil)
}

func (h *AuthHandler) Dekey(c *gin.Context) {
	response.Success(c, dekeyPayload)
}

func (h *AuthHandler) SymmetricKey(c *gin.Context) {
	response.Success(c, symmetricKeyBase)
}

func (h *AuthHandler) Model(c *gin.Context) {
	response.Success(c, false)
}

func RegisterAuthRoutes(engine *gin.Engine, h *AuthHandler) {
	engine.POST("/login/localLogin", h.LocalLogin)
	engine.GET("/logout", h.Logout)
	engine.GET("/dekey", h.Dekey)
	engine.GET("/symmetricKey", h.SymmetricKey)
	engine.GET("/model", h.Model)
	engine.POST("/api/login/localLogin", h.LocalLogin)
	engine.GET("/api/logout", h.Logout)
	engine.GET("/api/dekey", h.Dekey)
	engine.GET("/api/symmetricKey", h.SymmetricKey)
	engine.GET("/api/model", h.Model)
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
		if v == "admin" || v == "dataease" {
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
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}
