package service

import (
	"errors"
	"testing"

	domainauth "dataease/backend/internal/domain/auth"
	"dataease/backend/internal/domain/user"
	pkgauth "dataease/backend/internal/pkg/auth"

	"golang.org/x/crypto/bcrypt"
)

type mockAuthUserRepository struct {
	user *user.SysUser
	err  error
}

func (m *mockAuthUserRepository) GetByUsername(_ string) (*user.SysUser, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.user == nil {
		return nil, errors.New("not found")
	}
	return m.user, nil
}

func newTestAuthService(repo UserRepositoryInterface) *AuthService {
	return NewAuthService(repo, pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "test-secret", Expire: 3600}))
}

func TestAuthLocalLogin_Success(t *testing.T) {
	hashedPwd := mustHashPassword(t, "pwd123")
	repo := &mockAuthUserRepository{user: &user.SysUser{
		UserID:   101,
		Username: "alice",
		Password: hashedPwd,
		Status:   user.StatusEnabled,
	}}

	svc := newTestAuthService(repo)
	vo, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "alice", Pwd: "pwd123"})
	if err != nil {
		t.Fatalf("LocalLogin failed: %v", err)
	}
	if vo == nil || vo.Token == "" {
		t.Fatalf("expected non-empty token, got %#v", vo)
	}
	if vo.Exp != 0 {
		t.Fatalf("expected exp=0, got %d", vo.Exp)
	}

	claims, err := svc.ParseToken(vo.Token)
	if err != nil {
		t.Fatalf("ParseToken failed for login token: %v", err)
	}
	if claims.Uid != 101 || claims.Oid != 1 {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestAuthLocalLogin_UserNotFound(t *testing.T) {
	svc := newTestAuthService(&mockAuthUserRepository{err: errors.New("not found")})

	_, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "ghost", Pwd: "pwd123"})
	if err == nil || err.Error() != "用户名或密码错误" {
		t.Fatalf("expected username/password error, got %v", err)
	}
}

func TestAuthLocalLogin_WrongPassword(t *testing.T) {
	hashedPwd := mustHashPassword(t, "correct")
	svc := newTestAuthService(&mockAuthUserRepository{user: &user.SysUser{
		UserID:   1,
		Username: "bob",
		Password: hashedPwd,
		Status:   user.StatusEnabled,
	}})

	_, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "bob", Pwd: "wrong"})
	if err == nil || err.Error() != "用户名或密码错误" {
		t.Fatalf("expected username/password error, got %v", err)
	}
}

func TestAuthLocalLogin_DisabledUser(t *testing.T) {
	hashedPwd := mustHashPassword(t, "pwd123")
	svc := newTestAuthService(&mockAuthUserRepository{user: &user.SysUser{
		UserID:   2,
		Username: "charlie",
		Password: hashedPwd,
		Status:   user.StatusDisabled,
	}})

	_, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "charlie", Pwd: "pwd123"})
	if err == nil || err.Error() != "账号已被禁用" {
		t.Fatalf("expected disabled-user error, got %v", err)
	}
}

func TestAuthLogout_NoPanic(t *testing.T) {
	svc := newTestAuthService(&mockAuthUserRepository{})
	svc.Logout()
}

func TestAuthParseToken_InvalidToken(t *testing.T) {
	svc := newTestAuthService(&mockAuthUserRepository{})

	_, err := svc.ParseToken("invalid-token")
	if !errors.Is(err, pkgauth.ErrTokenMalformed) && !errors.Is(err, pkgauth.ErrTokenInvalid) {
		t.Fatalf("expected token parse error, got %v", err)
	}
}

func TestAuthLocalLogin_ProducesMiddlewareCompatibleToken(t *testing.T) {
	hashedPwd := mustHashPassword(t, "pwd123")
	jwt := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "shared-secret", Expire: 7200})
	svc := NewAuthService(&mockAuthUserRepository{user: &user.SysUser{
		UserID:   88,
		Username: "shared-user",
		Password: hashedPwd,
		Status:   user.StatusEnabled,
	}}, jwt)

	vo, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "shared-user", Pwd: "pwd123"})
	if err != nil {
		t.Fatalf("LocalLogin failed: %v", err)
	}

	claims, err := jwt.ParseToken(vo.Token)
	if err != nil {
		t.Fatalf("middleware JWT parser rejected login token: %v", err)
	}
	if claims.UserID != 88 || claims.Username != "shared-user" {
		t.Fatalf("unexpected jwt claims: %#v", claims)
	}
}

func mustHashPassword(t *testing.T, raw string) string {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}
	return string(hashed)
}
