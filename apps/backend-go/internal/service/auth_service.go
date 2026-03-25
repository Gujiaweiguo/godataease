package service

import (
	"fmt"
	"sort"
	"strings"

	domainauth "dataease/backend/internal/domain/auth"
	domainorg "dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/user"
	pkgauth "dataease/backend/internal/pkg/auth"
	"dataease/backend/internal/pkg/logger"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     AuthUserRepositoryInterface
	userRoleRepo AuthUserRoleRepositoryInterface
	orgRepo      AuthOrgRepositoryInterface
	jwt          *pkgauth.JWT
}

type AuthUserRepositoryInterface interface {
	GetByUsername(username string) (*user.SysUser, error)
	GetByID(userID int64) (*user.SysUser, error)
}

type AuthUserRoleRepositoryInterface interface {
	GetByUserID(userID int64) ([]*user.SysUserRole, error)
}

type AuthOrgRepositoryInterface interface {
	GetByIDs(ids []int64) ([]*domainorg.SysOrg, error)
}

func NewAuthService(
	userRepo AuthUserRepositoryInterface,
	userRoleRepo AuthUserRoleRepositoryInterface,
	orgRepo AuthOrgRepositoryInterface,
	jwt *pkgauth.JWT,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		orgRepo:      orgRepo,
		jwt:          jwt,
	}
}

func (s *AuthService) LocalLogin(dto *domainauth.PwdLoginDTO, requestLanguage string) (*domainauth.TokenVO, error) {
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

	bootstrap, err := s.BuildIdentityBootstrap(u.UserID, requestLanguage)
	if err != nil {
		return nil, err
	}

	token, err := s.jwt.GenerateTokenWithOrgID(uint64(u.UserID), u.Username, "", uint64(bootstrap.Oid))
	if err != nil {
		return nil, err
	}

	logger.Info("User logged in", zap.String("username", dto.Name), zap.Int64("userId", u.UserID))
	return &domainauth.TokenVO{
		Token:         token,
		Exp:           s.jwt.ExpirationUnixMilli(),
		ID:            bootstrap.ID,
		Name:          bootstrap.Name,
		Oid:           bootstrap.Oid,
		Language:      bootstrap.Language,
		CurrentOrg:    bootstrap.CurrentOrg,
		AvailableOrgs: bootstrap.AvailableOrgs,
	}, nil
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

	return &domainauth.TokenClaims{Uid: int64(claims.UserID), Oid: int64(claims.OrgID)}, nil
}

func (s *AuthService) RefreshToken(token string) (string, int64, error) {
	if s.jwt == nil {
		return "", 0, fmt.Errorf("token generator is not configured")
	}

	refreshedToken, err := s.jwt.RefreshToken(token)
	if err != nil {
		return "", 0, err
	}

	return refreshedToken, s.jwt.ExpirationUnixMilli(), nil
}

func (s *AuthService) BuildIdentityBootstrap(userID int64, requestLanguage string) (*domainauth.IdentityBootstrap, error) {
	return s.BuildIdentityBootstrapForOrg(userID, 0, requestLanguage)
}

func (s *AuthService) BuildIdentityBootstrapForOrg(userID int64, selectedOrgID int64, requestLanguage string) (*domainauth.IdentityBootstrap, error) {
	if s.userRepo == nil {
		return nil, fmt.Errorf("user repository is not configured")
	}

	u, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load current user: %w", err)
	}

	bootstrap := &domainauth.IdentityBootstrap{
		ID:            u.UserID,
		Name:          u.Username,
		Oid:           0,
		Language:      resolveBootstrapLanguage(requestLanguage, userLanguage(u)),
		AvailableOrgs: []domainauth.OrgSummary{},
	}

	orgIDs, err := s.userOrgIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(orgIDs) == 0 {
		return bootstrap, nil
	}

	orgs, err := s.activeOrgs(orgIDs)
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		return bootstrap, nil
	}

	available := make([]domainauth.OrgSummary, 0, len(orgs))
	for _, org := range orgs {
		available = append(available, domainauth.OrgSummary{
			OrgID:   org.OrgID,
			OrgName: org.OrgName,
		})
	}

	selectedIndex := 0
	if selectedOrgID > 0 {
		for index, candidate := range available {
			if candidate.OrgID == selectedOrgID {
				selectedIndex = index
				break
			}
		}
	}
	bootstrap.CurrentOrg = &available[selectedIndex]
	bootstrap.AvailableOrgs = available
	bootstrap.Oid = available[selectedIndex].OrgID

	return bootstrap, nil
}

func (s *AuthService) SwitchOrg(userID int64, targetOrgID int64, requestLanguage string) (*domainauth.TokenVO, error) {
	if targetOrgID <= 0 {
		return nil, fmt.Errorf("target organization is required")
	}

	bootstrap, err := s.BuildIdentityBootstrapForOrg(userID, targetOrgID, requestLanguage)
	if err != nil {
		return nil, err
	}
	if bootstrap.CurrentOrg == nil || bootstrap.CurrentOrg.OrgID != targetOrgID {
		return nil, fmt.Errorf("user is not a member of the target organization")
	}

	u, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load current user: %w", err)
	}

	token, err := s.jwt.GenerateTokenWithOrgID(uint64(userID), u.Username, "", uint64(targetOrgID))
	if err != nil {
		return nil, err
	}

	return &domainauth.TokenVO{
		Token:         token,
		Exp:           s.jwt.ExpirationUnixMilli(),
		ID:            bootstrap.ID,
		Name:          bootstrap.Name,
		Oid:           bootstrap.Oid,
		Language:      bootstrap.Language,
		CurrentOrg:    bootstrap.CurrentOrg,
		AvailableOrgs: bootstrap.AvailableOrgs,
	}, nil
}

func (s *AuthService) userOrgIDs(userID int64) ([]int64, error) {
	if s.userRoleRepo == nil {
		return nil, nil
	}

	userRoles, err := s.userRoleRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user organizations: %w", err)
	}

	orgIDSet := make(map[int64]struct{}, len(userRoles))
	orgIDs := make([]int64, 0, len(userRoles))
	for _, userRole := range userRoles {
		if userRole == nil || userRole.OrgID <= 0 {
			continue
		}
		if _, exists := orgIDSet[userRole.OrgID]; exists {
			continue
		}
		orgIDSet[userRole.OrgID] = struct{}{}
		orgIDs = append(orgIDs, userRole.OrgID)
	}

	sort.Slice(orgIDs, func(i, j int) bool {
		return orgIDs[i] < orgIDs[j]
	})

	return orgIDs, nil
}

func (s *AuthService) activeOrgs(orgIDs []int64) ([]*domainorg.SysOrg, error) {
	if len(orgIDs) == 0 || s.orgRepo == nil {
		return nil, nil
	}

	orgs, err := s.orgRepo.GetByIDs(orgIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load organizations: %w", err)
	}

	active := make([]*domainorg.SysOrg, 0, len(orgs))
	for _, org := range orgs {
		if org == nil || org.Status != domainorg.StatusEnabled {
			continue
		}
		active = append(active, org)
	}

	sort.Slice(active, func(i, j int) bool {
		return active[i].OrgID < active[j].OrgID
	})

	return active, nil
}

func userLanguage(u *user.SysUser) string {
	if u == nil || u.Language == nil {
		return ""
	}
	return strings.TrimSpace(*u.Language)
}

func resolveBootstrapLanguage(requestLanguage string, storedLanguage string) string {
	if locale := normalizeBootstrapLanguage(requestLanguage); locale != "" {
		return locale
	}
	if locale := normalizeBootstrapLanguage(storedLanguage); locale != "" {
		return locale
	}
	return "zh-CN"
}

func normalizeBootstrapLanguage(input string) string {
	for _, part := range strings.Split(input, ",") {
		normalized := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(part, "_", "-")))
		if normalized == "" {
			continue
		}

		if separator := strings.Index(normalized, ";"); separator >= 0 {
			normalized = strings.TrimSpace(normalized[:separator])
		}

		switch {
		case normalized == "tw", strings.HasPrefix(normalized, "zh-tw"), strings.HasPrefix(normalized, "zh-hk"):
			return "tw"
		case normalized == "en", strings.HasPrefix(normalized, "en"):
			return "en"
		case normalized == "zh-cn", normalized == "zh", strings.HasPrefix(normalized, "zh"):
			return "zh-CN"
		}
	}

	return ""
}
