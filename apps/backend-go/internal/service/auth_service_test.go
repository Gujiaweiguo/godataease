package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"

	"dataease/backend/internal/domain/auth"
	"dataease/backend/internal/domain/user"

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

func TestAuthLocalLogin_Success(t *testing.T) {
	hashedPwd := mustHashPassword(t, "pwd123")
	repo := &mockAuthUserRepository{user: &user.SysUser{
		UserID:   101,
		Username: "alice",
		Password: hashedPwd,
		Status:   user.StatusEnabled,
	}}

	svc := NewAuthService(repo)
	vo, err := svc.LocalLogin(&auth.PwdLoginDTO{Name: "alice", Pwd: "pwd123"})
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
	svc := NewAuthService(&mockAuthUserRepository{err: errors.New("not found")})

	_, err := svc.LocalLogin(&auth.PwdLoginDTO{Name: "ghost", Pwd: "pwd123"})
	if err == nil || err.Error() != "用户名或密码错误" {
		t.Fatalf("expected username/password error, got %v", err)
	}
}

func TestAuthLocalLogin_WrongPassword(t *testing.T) {
	hashedPwd := mustHashPassword(t, "correct")
	svc := NewAuthService(&mockAuthUserRepository{user: &user.SysUser{
		UserID:   1,
		Username: "bob",
		Password: hashedPwd,
		Status:   user.StatusEnabled,
	}})

	_, err := svc.LocalLogin(&auth.PwdLoginDTO{Name: "bob", Pwd: "wrong"})
	if err == nil || err.Error() != "用户名或密码错误" {
		t.Fatalf("expected username/password error, got %v", err)
	}
}

func TestAuthLocalLogin_DisabledUser(t *testing.T) {
	hashedPwd := mustHashPassword(t, "pwd123")
	svc := NewAuthService(&mockAuthUserRepository{user: &user.SysUser{
		UserID:   2,
		Username: "charlie",
		Password: hashedPwd,
		Status:   user.StatusDisabled,
	}})

	_, err := svc.LocalLogin(&auth.PwdLoginDTO{Name: "charlie", Pwd: "pwd123"})
	if err == nil || err.Error() != "账号已被禁用" {
		t.Fatalf("expected disabled-user error, got %v", err)
	}
}

func TestAuthLogout_NoPanic(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{})
	svc.Logout()
}

func TestAuthParseToken_Valid(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{})
	token, err := svc.generateToken(7, 9, svc.jwtSalt)
	if err != nil {
		t.Fatalf("generateToken failed: %v", err)
	}

	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.Uid != 7 || claims.Oid != 9 {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestAuthParseToken_InvalidFormat(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{})

	_, err := svc.ParseToken("only.two")
	if err == nil || err.Error() != "invalid token format" {
		t.Fatalf("expected invalid token format, got %v", err)
	}
}

func TestAuthParseToken_InvalidPayload(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{})

	_, err := svc.ParseToken("h.%%%.s")
	if err == nil || err.Error() != "invalid token payload" {
		t.Fatalf("expected invalid token payload, got %v", err)
	}
}

func TestAuthParseToken_InvalidClaims(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{})
	payload := base64URLEncode([]byte("not-json"))

	_, err := svc.ParseToken("h." + payload + ".s")
	if err == nil || err.Error() != "invalid token claims" {
		t.Fatalf("expected invalid token claims, got %v", err)
	}
}

func TestAuthMd5Hash(t *testing.T) {
	if got, want := md5Hash("abc"), "900150983cd24fb0d6963f7d28e17f72"; got != want {
		t.Fatalf("md5Hash mismatch, got=%s want=%s", got, want)
	}
}

func TestAuthHmacSha256(t *testing.T) {
	secret := []byte("secret-key")
	data := "header.payload"

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	expected := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if got := hmacSha256(secret, data); got != expected {
		t.Fatalf("hmacSha256 mismatch, got=%s want=%s", got, expected)
	}
}

func TestAuthBase64URLEncode(t *testing.T) {
	src := []byte{0xfb, 0xff}
	if got, want := base64URLEncode(src), "-_8"; got != want {
		t.Fatalf("base64URLEncode mismatch, got=%s want=%s", got, want)
	}
}

func TestAuthBase64URLDecode(t *testing.T) {
	decoded, err := base64URLDecode("-_8")
	if err != nil {
		t.Fatalf("base64URLDecode failed: %v", err)
	}
	if !bytes.Equal([]byte(decoded), []byte{0xfb, 0xff}) {
		t.Fatalf("decoded bytes mismatch: %v", []byte(decoded))
	}
}

func TestAuthBase64URLDecode_Error(t *testing.T) {
	_, err := base64URLDecode("%%%")
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestAuthSplitJWTToken(t *testing.T) {
	parts := splitJWTToken("a.b.c")
	if len(parts) != 3 || parts[0] != "a" || parts[1] != "b" || parts[2] != "c" {
		t.Fatalf("unexpected parts: %#v", parts)
	}

	single := splitJWTToken("abc")
	if len(single) != 1 || single[0] != "abc" {
		t.Fatalf("unexpected single part: %#v", single)
	}
}

func TestAuthGenerateToken(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{})
	token, err := svc.generateToken(11, 22, "custom_salt")
	if err != nil {
		t.Fatalf("generateToken failed: %v", err)
	}

	parts := splitJWTToken(token)
	if len(parts) != 3 {
		t.Fatalf("expected 3 token parts, got %d", len(parts))
	}

	header, err := base64URLDecode(parts[0])
	if err != nil {
		t.Fatalf("decode header failed: %v", err)
	}
	payload, err := base64URLDecode(parts[1])
	if err != nil {
		t.Fatalf("decode payload failed: %v", err)
	}

	if header != `{"alg":"HS256","typ":"JWT"}` {
		t.Fatalf("unexpected header: %s", header)
	}
	if payload != `{"uid":11,"oid":22}` {
		t.Fatalf("unexpected payload: %s", payload)
	}

	expectedSig := hmacSha256([]byte(md5Hash("custom_salt")), parts[0]+"."+parts[1])
	if parts[2] != expectedSig {
		t.Fatalf("unexpected signature: %s", parts[2])
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
