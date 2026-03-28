package service

import (
	"errors"
	"sort"
	"testing"
	"time"

	domainauth "dataease/backend/internal/domain/auth"
	domainorg "dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/user"
	pkgauth "dataease/backend/internal/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type sequentialAuthUserRepository struct {
	byUsername    *user.SysUser
	byUsernameErr error
	byIDResponses []*user.SysUser
	byIDErrors    []error
	byIDCalls     int
}

func (m *sequentialAuthUserRepository) GetByUsername(_ string) (*user.SysUser, error) {
	if m.byUsernameErr != nil {
		return nil, m.byUsernameErr
	}
	if m.byUsername == nil {
		return nil, errors.New("not found")
	}
	return m.byUsername, nil
}

func (m *sequentialAuthUserRepository) GetByID(_ int64) (*user.SysUser, error) {
	idx := m.byIDCalls
	m.byIDCalls++
	if idx < len(m.byIDErrors) && m.byIDErrors[idx] != nil {
		return nil, m.byIDErrors[idx]
	}
	if idx < len(m.byIDResponses) && m.byIDResponses[idx] != nil {
		return m.byIDResponses[idx], nil
	}
	return nil, errors.New("not found")
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

func TestAuthService_UserOrgIDs_DeduplicatesSortsAndSkipsInvalid(t *testing.T) {
	svc := NewAuthService(
		&mockAuthUserRepository{},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 1, OrgID: 3}, nil, {UserID: 1, OrgID: 2}, {UserID: 1, OrgID: 3}, {UserID: 1, OrgID: 0}, {UserID: 1, OrgID: -1}}},
		nil,
		nil,
	)

	orgIDs, err := svc.userOrgIDs(1)
	require.NoError(t, err)
	assert.Equal(t, []int64{2, 3}, orgIDs)
	assert.True(t, sort.SliceIsSorted(orgIDs, func(i, j int) bool { return orgIDs[i] < orgIDs[j] }))
}

func TestAuthService_UserOrgIDs_RepoError(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{}, &mockAuthUserRoleRepository{err: errors.New("role repo failed")}, nil, nil)

	orgIDs, err := svc.userOrgIDs(1)
	require.Error(t, err)
	assert.Nil(t, orgIDs)
	assert.Contains(t, err.Error(), "failed to load user organizations")
}

func TestAuthService_ActiveOrgs_FiltersDisabledAndSorts(t *testing.T) {
	svc := NewAuthService(
		&mockAuthUserRepository{},
		nil,
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 5, OrgName: "Disabled", Status: domainorg.StatusDisabled}, {OrgID: 3, OrgName: "Org Three", Status: domainorg.StatusEnabled}, nil, {OrgID: 1, OrgName: "Org One", Status: domainorg.StatusEnabled}}},
		nil,
	)

	orgs, err := svc.activeOrgs([]int64{5, 3, 1})
	require.NoError(t, err)
	assert.Len(t, orgs, 2)
	assert.Equal(t, int64(1), orgs[0].OrgID)
	assert.Equal(t, int64(3), orgs[1].OrgID)
}

func TestAuthService_ActiveOrgs_RepoError(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{}, nil, &mockAuthOrgRepository{err: errors.New("org repo failed")}, nil)

	orgs, err := svc.activeOrgs([]int64{1})
	require.Error(t, err)
	assert.Nil(t, orgs)
	assert.Contains(t, err.Error(), "failed to load organizations")
}

func TestAuthService_ActiveOrgs_EmptyInputAndNilRepoReturnNil(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{}, nil, nil, nil)

	orgs, err := svc.activeOrgs(nil)
	require.NoError(t, err)
	assert.Nil(t, orgs)

	orgs, err = svc.activeOrgs([]int64{1, 2})
	require.NoError(t, err)
	assert.Nil(t, orgs)
}

func TestResolveBootstrapLanguage_PrefersRequestThenStoredThenDefault(t *testing.T) {
	assert.Equal(t, "en", resolveBootstrapLanguage("en-US", "zh-CN"))
	assert.Equal(t, "tw", resolveBootstrapLanguage("", "zh-TW"))
	assert.Equal(t, defaultLanguageZhCN, resolveBootstrapLanguage("", ""))
}

func TestNormalizeBootstrapLanguage_ParsesAcceptLanguageVariants(t *testing.T) {
	assert.Equal(t, "en", normalizeBootstrapLanguage("en-US,en;q=0.9,zh-CN;q=0.8"))
	assert.Equal(t, "tw", normalizeBootstrapLanguage("zh-HK,zh;q=0.8"))
	assert.Equal(t, defaultLanguageZhCN, normalizeBootstrapLanguage("zh_CN"))
	assert.Equal(t, "", normalizeBootstrapLanguage("fr-FR"))
}

func TestAuthService_ParseToken_JWTNotConfigured(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{}, nil, nil, nil)

	claims, err := svc.ParseToken("token")
	require.Error(t, err)
	assert.Nil(t, claims)
	assert.EqualError(t, err, "token generator is not configured")
}

func TestAuthService_RefreshToken_JWTNotConfigured(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{}, nil, nil, nil)

	token, exp, err := svc.RefreshToken("token")
	require.Error(t, err)
	assert.Empty(t, token)
	assert.Zero(t, exp)
	assert.EqualError(t, err, "token generator is not configured")
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	jwt := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "refresh-secret", Expire: 3600})
	svc := NewAuthService(&mockAuthUserRepository{}, nil, nil, jwt)
	original, err := jwt.GenerateTokenWithOrgID(1, "alice", "", 2)
	require.NoError(t, err)

	refreshed, exp, refreshErr := svc.RefreshToken(original)
	require.NoError(t, refreshErr)
	assert.NotEmpty(t, refreshed)
	assert.Positive(t, exp)

	claims, parseErr := svc.ParseToken(refreshed)
	require.NoError(t, parseErr)
	assert.Equal(t, int64(1), claims.Uid)
	assert.Equal(t, int64(2), claims.Oid)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	jwt := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "refresh-secret", Expire: 3600})
	svc := NewAuthService(&mockAuthUserRepository{}, nil, nil, jwt)

	refreshed, exp, err := svc.RefreshToken("invalid-token")
	require.Error(t, err)
	assert.Empty(t, refreshed)
	assert.Zero(t, exp)
}

func TestAuthService_ParseToken_Success(t *testing.T) {
	jwt := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "parse-secret", Expire: 3600})
	svc := NewAuthService(&mockAuthUserRepository{}, nil, nil, jwt)
	token, err := jwt.GenerateTokenWithOrgID(77, "parsed-user", "", 12)
	require.NoError(t, err)

	claims, err := svc.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, int64(77), claims.Uid)
	assert.Equal(t, int64(12), claims.Oid)
}

func TestAuthService_BuildIdentityBootstrapForOrg_SelectedOrgMatched(t *testing.T) {
	lang := "zh-CN"
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 11, Username: "selected", Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 11, OrgID: 3}, {UserID: 11, OrgID: 8}}},
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 8, OrgName: "Org Eight", Status: domainorg.StatusEnabled}, {OrgID: 3, OrgName: "Org Three", Status: domainorg.StatusEnabled}}},
		nil,
	)

	bootstrap, err := svc.BuildIdentityBootstrapForOrg(11, 8, "")
	require.NoError(t, err)
	require.NotNil(t, bootstrap.CurrentOrg)
	assert.Equal(t, int64(8), bootstrap.CurrentOrg.OrgID)
	assert.Equal(t, int64(8), bootstrap.Oid)
}

func TestAuthService_BuildIdentityBootstrapForOrg_SelectedOrgMissingFallsBackToFirst(t *testing.T) {
	lang := "en"
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 12, Username: "fallback", Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 12, OrgID: 7}, {UserID: 12, OrgID: 2}}},
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 7, OrgName: "Org Seven", Status: domainorg.StatusEnabled}, {OrgID: 2, OrgName: "Org Two", Status: domainorg.StatusEnabled}}},
		nil,
	)

	bootstrap, err := svc.BuildIdentityBootstrapForOrg(12, 999, "")
	require.NoError(t, err)
	require.NotNil(t, bootstrap.CurrentOrg)
	assert.Equal(t, int64(2), bootstrap.CurrentOrg.OrgID)
	assert.Equal(t, int64(2), bootstrap.Oid)
}

func TestAuthService_BuildIdentityBootstrapForOrg_UserRepoNotConfigured(t *testing.T) {
	svc := NewAuthService(nil, nil, nil, nil)

	bootstrap, err := svc.BuildIdentityBootstrapForOrg(1, 0, "")
	require.Error(t, err)
	assert.Nil(t, bootstrap)
	assert.EqualError(t, err, "user repository is not configured")
}

func TestAuthService_BuildIdentityBootstrapForOrg_UserLoadError(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{err: errors.New("load failed")}, nil, nil, nil)

	bootstrap, err := svc.BuildIdentityBootstrapForOrg(1, 0, "")
	require.Error(t, err)
	assert.Nil(t, bootstrap)
	assert.Contains(t, err.Error(), "failed to load current user")
}

func TestAuthService_BuildIdentityBootstrapForOrg_NoUserRolesReturnsBootstrapWithoutOrg(t *testing.T) {
	lang := "en"
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 31, Username: "noroles", Language: &lang}},
		&mockAuthUserRoleRepository{roles: nil},
		&mockAuthOrgRepository{},
		nil,
	)

	bootstrap, err := svc.BuildIdentityBootstrapForOrg(31, 0, "")
	require.NoError(t, err)
	assert.Equal(t, int64(31), bootstrap.ID)
	assert.Equal(t, int64(0), bootstrap.Oid)
	assert.Nil(t, bootstrap.CurrentOrg)
	assert.Empty(t, bootstrap.AvailableOrgs)
	assert.Equal(t, "en", bootstrap.Language)
}

func TestAuthService_BuildIdentityBootstrapForOrg_NoActiveOrgsReturnsBootstrapWithoutOrg(t *testing.T) {
	lang := "zh-CN"
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 32, Username: "inactive-orgs", Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 32, OrgID: 9}}},
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 9, OrgName: "Disabled Org", Status: domainorg.StatusDisabled}}},
		nil,
	)

	bootstrap, err := svc.BuildIdentityBootstrapForOrg(32, 9, "")
	require.NoError(t, err)
	assert.Equal(t, int64(0), bootstrap.Oid)
	assert.Nil(t, bootstrap.CurrentOrg)
	assert.Empty(t, bootstrap.AvailableOrgs)
	assert.Equal(t, defaultLanguageZhCN, bootstrap.Language)
}

func TestAuthService_BuildIdentityBootstrapForOrg_NilOrgRepoReturnsBootstrapWithoutOrg(t *testing.T) {
	lang := "en"
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 33, Username: "nil-org-repo", Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 33, OrgID: 8}}},
		nil,
		nil,
	)

	bootstrap, err := svc.BuildIdentityBootstrapForOrg(33, 8, "")
	require.NoError(t, err)
	assert.Equal(t, int64(0), bootstrap.Oid)
	assert.Nil(t, bootstrap.CurrentOrg)
	assert.Empty(t, bootstrap.AvailableOrgs)
	assert.Equal(t, "en", bootstrap.Language)
}

func TestAuthService_BuildIdentityBootstrapForOrg_OrgRepoError(t *testing.T) {
	lang := "en"
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 34, Username: "org-error", Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 34, OrgID: 5}}},
		&mockAuthOrgRepository{err: errors.New("org load failed")},
		nil,
	)

	bootstrap, err := svc.BuildIdentityBootstrapForOrg(34, 5, "")
	require.Error(t, err)
	assert.Nil(t, bootstrap)
	assert.Contains(t, err.Error(), "failed to load organizations")
}

func TestAuthService_BuildIdentityBootstrap_DelegatesToForOrg(t *testing.T) {
	lang := "zh-CN"
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 35, Username: "delegated", Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 35, OrgID: 4}}},
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 4, OrgName: "Org Four", Status: domainorg.StatusEnabled}}},
		nil,
	)

	bootstrap, err := svc.BuildIdentityBootstrap(35, "en-US")
	require.NoError(t, err)
	assert.Equal(t, int64(35), bootstrap.ID)
	assert.Equal(t, int64(4), bootstrap.Oid)
	require.NotNil(t, bootstrap.CurrentOrg)
	assert.Equal(t, int64(4), bootstrap.CurrentOrg.OrgID)
	assert.Equal(t, "en", bootstrap.Language)
}

func TestAuthService_SwitchOrg_TargetOrgRequired(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{}, nil, nil, nil)

	vo, err := svc.SwitchOrg(1, 0, "")
	require.Error(t, err)
	assert.Nil(t, vo)
	assert.EqualError(t, err, "target organization is required")
}

func TestAuthService_SwitchOrg_UserNotMemberOfTargetOrg(t *testing.T) {
	lang := "zh-CN"
	jwt := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "switch-secret", Expire: 3600})
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 41, Username: "switcher", Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 41, OrgID: 3}}},
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 3, OrgName: "Org Three", Status: domainorg.StatusEnabled}}},
		jwt,
	)

	vo, err := svc.SwitchOrg(41, 8, "")
	require.Error(t, err)
	assert.Nil(t, vo)
	assert.EqualError(t, err, "user is not a member of the target organization")
}

func TestAuthLocalLogin_JWTNotConfigured(t *testing.T) {
	hashedPwd := mustHashPassword(t, "pwd123")
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 51, Username: "nojwt", Password: hashedPwd, Status: user.StatusEnabled}},
		nil,
		nil,
		nil,
	)

	vo, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "nojwt", Pwd: "pwd123"}, "")
	require.Error(t, err)
	assert.Nil(t, vo)
	assert.EqualError(t, err, "token generator is not configured")
}

func TestAuthLocalLogin_NoActiveOrgsStillReturnsToken(t *testing.T) {
	hashedPwd := mustHashPassword(t, "pwd123")
	lang := "zh-CN"
	jwt := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "no-org-secret", Expire: 3600})
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 53, Username: "no-org", Password: hashedPwd, Status: user.StatusEnabled, Language: &lang}},
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 53, OrgID: 9}}},
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 9, OrgName: "Disabled Org", Status: domainorg.StatusDisabled}}},
		jwt,
	)

	vo, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "no-org", Pwd: "pwd123"}, "")
	require.NoError(t, err)
	assert.NotEmpty(t, vo.Token)
	assert.Equal(t, int64(0), vo.Oid)
	assert.Nil(t, vo.CurrentOrg)
	assert.Empty(t, vo.AvailableOrgs)
}

func TestAuthLocalLogin_BuildIdentityBootstrapError(t *testing.T) {
	hashedPwd := mustHashPassword(t, "pwd123")
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 52, Username: "bootstrap-fail", Password: hashedPwd, Status: user.StatusEnabled}},
		&mockAuthUserRoleRepository{err: errors.New("role repo failed")},
		nil,
		pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "bootstrap-secret", Expire: 3600}),
	)

	vo, err := svc.LocalLogin(&domainauth.PwdLoginDTO{Name: "bootstrap-fail", Pwd: "pwd123"}, "")
	require.Error(t, err)
	assert.Nil(t, vo)
	assert.Contains(t, err.Error(), "failed to load user organizations")
}

func TestAuthService_SwitchOrg_UserReloadError(t *testing.T) {
	lang := "en"
	seqRepo := &sequentialAuthUserRepository{
		byIDResponses: []*user.SysUser{{UserID: 61, Username: "switcher", Language: &lang}},
		byIDErrors:    []error{nil, errors.New("reload failed")},
	}
	jwt := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "switch-secret", Expire: 3600})
	svc := NewAuthService(
		seqRepo,
		&mockAuthUserRoleRepository{roles: []*user.SysUserRole{{UserID: 61, OrgID: 7}}},
		&mockAuthOrgRepository{orgs: []*domainorg.SysOrg{{OrgID: 7, OrgName: "Org Seven", Status: domainorg.StatusEnabled}}},
		jwt,
	)

	vo, err := svc.SwitchOrg(61, 7, "")
	require.Error(t, err)
	assert.Nil(t, vo)
	assert.Contains(t, err.Error(), "failed to load current user")
}

func TestAuthService_SwitchOrg_BuildBootstrapError(t *testing.T) {
	lang := "en"
	jwt := pkgauth.NewJWT(&pkgauth.JWTConfig{Secret: "switch-secret", Expire: 3600})
	svc := NewAuthService(
		&mockAuthUserRepository{user: &user.SysUser{UserID: 62, Username: "switcher", Language: &lang}},
		&mockAuthUserRoleRepository{err: errors.New("role repo failed")},
		&mockAuthOrgRepository{},
		jwt,
	)

	vo, err := svc.SwitchOrg(62, 7, "")
	require.Error(t, err)
	assert.Nil(t, vo)
	assert.Contains(t, err.Error(), "failed to load user organizations")
}

func TestAuthService_UserOrgIDs_NilRepoReturnsNil(t *testing.T) {
	svc := NewAuthService(&mockAuthUserRepository{}, nil, nil, nil)

	orgIDs, err := svc.userOrgIDs(88)
	require.NoError(t, err)
	assert.Nil(t, orgIDs)
}

func mustHashPassword(t *testing.T, raw string) string {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}
	return string(hashed)
}
