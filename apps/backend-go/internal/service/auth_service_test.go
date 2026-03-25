package service

import (
	"errors"
	"testing"
	"time"

	domainauth "dataease/backend/internal/domain/auth"
	domainorg "dataease/backend/internal/domain/org"
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

func (m *mockAuthUserRepository) GetByID(_ int64) (*user.SysUser, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.user == nil {
		return nil, errors.New("not found")
	}
	return m.user, nil
}

type mockAuthUserRoleRepository struct {
	roles []*user.SysUserRole
	err   error
}

func (m *mockAuthUserRoleRepository) GetByUserID(_ int64) ([]*user.SysUserRole, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.roles, nil
}

type mockAuthOrgRepository struct {
	orgs []*domainorg.SysOrg
	err  error
}

func (m *mockAuthOrgRepository) GetByIDs(_ []int64) ([]*domainorg.SysOrg, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.orgs, nil
}

func newTestAuthService(repo AuthUserRepositoryInterface) *AuthService {
	return NewAuthService(repo, nil, nil, pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "test-secret", Expire: 3600}))
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
	vo, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "alice", Pwd: "pwd123"}, "")
	if err != nil {
		t.Fatalf("LocalLogin failed: %v", err)
	}
	if vo == nil || vo.Token == "" {
		t.Fatalf("expected non-empty token, got %#v", vo)
	}
	if vo.ID != 101 || vo.Name != "alice" {
		t.Fatalf("expected login bootstrap identity fields, got %#v", vo)
	}
	now := time.Now().UnixMilli()
	if vo.Exp <= now {
		t.Fatalf("expected future exp, got %d (now=%d)", vo.Exp, now)
	}
	if vo.Exp > now+3600_000+5_000 {
		t.Fatalf("expected exp close to 1h from now, got %d (now=%d)", vo.Exp, now)
	}

	claims, err := svc.ParseToken(vo.Token)
	if err != nil {
		t.Fatalf("ParseToken failed for login token: %v", err)
	}
	if claims.Uid != 101 || claims.Oid != 0 {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestAuthLocalLogin_UserNotFound(t *testing.T) {
	svc := newTestAuthService(&mockAuthUserRepository{err: errors.New("not found")})

	_, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "ghost", Pwd: "pwd123"}, "")
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

	_, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "bob", Pwd: "wrong"}, "")
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

	_, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "charlie", Pwd: "pwd123"}, "")
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
	}}, nil, nil, jwt)

	vo, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "shared-user", Pwd: "pwd123"}, "")
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

func TestAuthService_BuildIdentityBootstrap_UsesOrgMembershipsAndStoredLanguage(t *testing.T) {
	lang := "zh-TW"
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 9, Username: "alice", Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 9, OrgID: 2}, {UserID: 9, OrgID: 1}, {UserID: 9, OrgID: 2}}},
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 2, OrgName: "Org B", Status: domainorg.StatusEnabled}, {OrgID: 1, OrgName: "Org A", Status: domainorg.StatusEnabled}}},
		nil,
	)

	bootstrap, err := svc.BuildIdentityBootstrap(9, "")
	if err != nil {
		t.Fatalf("BuildIdentityBootstrap failed: %v", err)
	}
	if bootstrap.ID != 9 || bootstrap.Name != "alice" {
		t.Fatalf("unexpected identity bootstrap: %#v", bootstrap)
	}
	if bootstrap.Oid != 1 {
		t.Fatalf("expected Oid 1, got %d", bootstrap.Oid)
	}
	if bootstrap.Language != "tw" {
		t.Fatalf("expected normalized tw language, got %q", bootstrap.Language)
	}
	if bootstrap.CurrentOrg == nil || bootstrap.CurrentOrg.OrgID != 1 {
		t.Fatalf("expected first sorted org as current org, got %#v", bootstrap.CurrentOrg)
	}
	if len(bootstrap.AvailableOrgs) != 2 {
		t.Fatalf("expected 2 available orgs, got %#v", bootstrap.AvailableOrgs)
	}
}

func TestAuthService_BuildIdentityBootstrap_RequestLanguageOverridesStoredLanguage(t *testing.T) {
	lang := "zh-CN"
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 7, Username: "bob", Language: &lang}},
		&mockAuthUserRoleRepository{},
		&mockAuthOrgRepository{},
		nil,
	)

	bootstrap, err := svc.BuildIdentityBootstrap(7, "en-US")
	if err != nil {
		t.Fatalf("BuildIdentityBootstrap failed: %v", err)
	}
	if bootstrap.Language != "en" {
		t.Fatalf("expected request language to win, got %q", bootstrap.Language)
	}
	if bootstrap.Oid != 0 || bootstrap.CurrentOrg != nil || len(bootstrap.AvailableOrgs) != 0 {
		t.Fatalf("expected no org context, got %#v", bootstrap)
	}
}

func TestAuthLocalLogin_IncludesOrganizationAwareBootstrap(t *testing.T) {
	lang := "zh-CN"
	hashedPwd := mustHashPassword(t, "pwd123")
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 18, Username: "org-user", Password: hashedPwd, Status: user.StatusEnabled, Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 18, OrgID: 5}}},
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 5, OrgName: "Org Five", Status: domainorg.StatusEnabled}}},
		pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "bootstrap-secret", Expire: 3600}),
	)

	vo, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "org-user", Pwd: "pwd123"}, "en-US")
	if err != nil {
		t.Fatalf("LocalLogin failed: %v", err)
	}
	if vo.Oid != 5 || vo.CurrentOrg == nil || vo.CurrentOrg.OrgID != 5 {
		t.Fatalf("expected org-aware bootstrap, got %#v", vo)
	}
	if vo.Language != "en" {
		t.Fatalf("expected login bootstrap to honor request language, got %#v", vo)
	}
}

func TestAuthService_SwitchOrg_ReturnsRetargetedToken(t *testing.T) {
	lang := "zh-CN"
	jwt := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "switch-secret", Expire: 3600})
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 21, Username: "switcher", Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 21, OrgID: 3}, {UserID: 21, OrgID: 8}}},
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 3, OrgName: "Org Three", Status: domainorg.StatusEnabled}, {OrgID: 8, OrgName: "Org Eight", Status: domainorg.StatusEnabled}}},
		jwt,
	)

	vo, err := svc.SwitchOrg(21, 8, "en-US")
	if err != nil {
		t.Fatalf("SwitchOrg failed: %v", err)
	}
	if vo.Oid != 8 || vo.CurrentOrg == nil || vo.CurrentOrg.OrgID != 8 {
		t.Fatalf("expected switched org bootstrap, got %#v", vo)
	}
	claims, err := jwt.ParseToken(vo.Token)
	if err != nil {
		t.Fatalf("parse switched token failed: %v", err)
	}
	if claims.OrgID != 8 {
		t.Fatalf("expected switched token org id 8, got %d", claims.OrgID)
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
