package service

import (
	"fmt"

	domainauth "dataease/backend/internal/domain/auth"
	"dataease/backend/internal/domain/user"
	pkgauth "dataease/backend/internal/pkg/auth"
	"dataease/backend/internal/pkg/logger"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo UserRepositoryInterface
	jwt      *pkgauth.JWT
}

type UserRepositoryInterface interface {
	GetByUsername(username string) (*user.SysUser, error)
}

func NewAuthService(userRepo UserRepositoryInterface, jwt *pkgauth.JWT) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwt:      jwt,
	}
}

func (s *AuthService) LocalLogin(dto *domainauth.PwdLoginDTO) (*domainauth.TokenVO, error) {
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

	if s.jwt == nil {
		return nil, fmt.Errorf("token generator is not configured")
	}

	token, err := s.jwt.GenerateToken(uint64(u.UserID), u.Username, "")
	if err != nil {
		return nil, err
	}

	logger.Info("User logged in", zap.String("username", dto.Name), zap.Int64("userId", u.UserID))
	return &domainauth.TokenVO{Token: token, Exp: 0}, nil
}

func (s *AuthService) Logout() {
	logger.Info("User logged out")
}

func (s *AuthService) ParseToken(token string) (*domainauth.TokenClaims, error) {
	if s.jwt == nil {
		return nil, fmt.Errorf("token generator is not configured")
	}

	claims, err := s.jwt.ParseToken(token)
	if err != nil {
		return nil, err
	}

	return &domainauth.TokenClaims{Uid: int64(claims.UserID), Oid: 1}, nil
}
