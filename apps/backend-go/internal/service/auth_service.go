package service

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"dataease/backend/internal/domain/auth"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/pkg/logger"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo UserRepositoryInterface
	jwtSalt  string
}

type UserRepositoryInterface interface {
	GetByUsername(username string) (*user.SysUser, error)
}

func NewAuthService(userRepo UserRepositoryInterface) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtSalt:  "dataease_jwt_salt",
	}
}

func (s *AuthService) LocalLogin(dto *auth.PwdLoginDTO) (*auth.TokenVO, error) {
	u, err := s.userRepo.GetByUsername(dto.Name)
	if err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	if u.Status != user.StatusEnabled {
		return nil, fmt.Errorf("账号已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(dto.Pwd)); err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	token, err := s.generateToken(u.UserID, 1, s.jwtSalt)
	if err != nil {
		return nil, err
	}

	logger.Info("User logged in", zap.String("username", dto.Name), zap.Int64("userId", u.UserID))
	return &auth.TokenVO{Token: token, Exp: 0}, nil
}

func (s *AuthService) Logout() {
	logger.Info("User logged out")
}

func (s *AuthService) generateToken(userId, orgId int64, salt string) (string, error) {
	secret := []byte(md5Hash(salt))

	header := base64URLEncode([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64URLEncode([]byte(fmt.Sprintf(`{"uid":%d,"oid":%d}`, userId, orgId)))

	signature := hmacSha256(secret, header+"."+payload)
	return header + "." + payload + "." + signature, nil
}

func md5Hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func hmacSha256(secret []byte, data string) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return base64URLEncode(h.Sum(nil))
}

func base64URLEncode(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	result := ""
	for _, c := range encoded {
		switch c {
		case '+':
			result += "-"
		case '/':
			result += "_"
		case '=':
		default:
			result += string(c)
		}
	}
	return result
}

func (s *AuthService) ParseToken(token string) (*auth.TokenClaims, error) {
	parts := splitJWTToken(token)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token payload")
	}

	var claims struct {
		Uid int64 `json:"uid"`
		Oid int64 `json:"oid"`
	}
	if err := json.Unmarshal([]byte(payload), &claims); err != nil {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &auth.TokenClaims{Uid: claims.Uid, Oid: claims.Oid}, nil
}

func splitJWTToken(token string) []string {
	var parts []string
	start := 0
	for i, c := range token {
		if c == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	if start < len(token) {
		parts = append(parts, token[start:])
	}
	return parts
}

func base64URLDecode(data string) (string, error) {
	replaced := ""
	for _, c := range data {
		switch c {
		case '-':
			replaced += "+"
		case '_':
			replaced += "/"
		default:
			replaced += string(c)
		}
	}
	for len(replaced)%4 != 0 {
		replaced += "="
	}
	decoded, err := base64.StdEncoding.DecodeString(replaced)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
