package service

import (
	"fmt"
	"time"

	"dataease/backend/internal/domain/embedded"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
)

type EmbeddedService struct {
	repo *repository.EmbeddedRepository
}

func NewEmbeddedService(repo *repository.EmbeddedRepository) *EmbeddedService {
	return &EmbeddedService{repo: repo}
}

func (s *EmbeddedService) Create(req *embedded.EmbeddedCreator, updateBy string) (int64, error) {
	secretLength := embedded.DefaultSecretLength
	if req.SecretLength != nil && *req.SecretLength > 0 {
		secretLength = *req.SecretLength
	}

	e := &embedded.CoreEmbedded{
		Name:         req.Name,
		AppId:        embedded.GenerateAppId(),
		AppSecret:    embedded.GenerateAppSecret(secretLength),
		Domain:       req.Domain,
		SecretLength: secretLength,
		CreateTime:   time.Now().UnixMilli(),
		UpdateBy:     updateBy,
		UpdateTime:   time.Now().UnixMilli(),
	}

	if err := s.repo.Create(e); err != nil {
		logger.Error("Failed to create embedded app", zap.Error(err))
		return 0, fmt.Errorf("failed to create embedded app: %w", err)
	}

	logger.Info("Embedded app created", zap.Int64("id", e.ID), zap.String("appId", e.AppId))
	return e.ID, nil
}

func (s *EmbeddedService) Edit(req *embedded.EmbeddedEditor, updateBy string) error {
	e, err := s.repo.GetByID(req.ID)
	if err != nil {
		return fmt.Errorf("embedded app not found: %w", err)
	}

	if req.Name != "" {
		e.Name = req.Name
	}
	if req.Domain != nil {
		e.Domain = *req.Domain
	}
	if req.SecretLength != nil {
		e.SecretLength = *req.SecretLength
	}
	e.UpdateBy = updateBy
	e.UpdateTime = time.Now().UnixMilli()

	if err := s.repo.Update(e); err != nil {
		logger.Error("Failed to update embedded app", zap.Error(err))
		return fmt.Errorf("failed to update embedded app: %w", err)
	}

	logger.Info("Embedded app updated", zap.Int64("id", req.ID))
	return nil
}

func (s *EmbeddedService) Delete(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		logger.Error("Failed to delete embedded app", zap.Error(err))
		return fmt.Errorf("failed to delete embedded app: %w", err)
	}
	logger.Info("Embedded app deleted", zap.Int64("id", id))
	return nil
}

func (s *EmbeddedService) BatchDelete(ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("ids list cannot be empty")
	}
	if err := s.repo.DeleteBatch(ids); err != nil {
		logger.Error("Failed to batch delete embedded apps", zap.Error(err))
		return fmt.Errorf("failed to batch delete embedded apps: %w", err)
	}
	logger.Info("Embedded apps batch deleted", zap.Int("count", len(ids)))
	return nil
}

func (s *EmbeddedService) ResetSecret(req *embedded.EmbeddedResetRequest, updateBy string) error {
	e, err := s.repo.GetByID(req.ID)
	if err != nil {
		return fmt.Errorf("embedded app not found: %w", err)
	}

	var newSecret string
	if req.AppSecret != nil && *req.AppSecret != "" {
		newSecret = *req.AppSecret
	} else {
		newSecret = embedded.GenerateAppSecret(e.SecretLength)
	}

	e.AppSecret = newSecret
	e.UpdateBy = updateBy
	e.UpdateTime = time.Now().UnixMilli()

	if err := s.repo.Update(e); err != nil {
		logger.Error("Failed to reset embedded app secret", zap.Error(err))
		return fmt.Errorf("failed to reset secret: %w", err)
	}

	logger.Info("Embedded app secret reset", zap.Int64("id", req.ID))
	return nil
}

func (s *EmbeddedService) QueryGrid(keyword string, page, pageSize int) (*embedded.EmbeddedPagerResponse, error) {
	items, total, err := s.repo.Query(keyword, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to query embedded apps: %w", err)
	}

	list := make([]embedded.EmbeddedGridVO, 0, len(items))
	for _, e := range items {
		list = append(list, embedded.EmbeddedGridVO{
			ID:           e.ID,
			Name:         e.Name,
			AppId:        e.AppId,
			AppSecret:    embedded.MaskAppSecret(e.AppSecret),
			Domain:       e.Domain,
			SecretLength: e.SecretLength,
		})
	}

	return &embedded.EmbeddedPagerResponse{
		List:    list,
		Total:   total,
		Current: page,
		Size:    pageSize,
	}, nil
}

func (s *EmbeddedService) GetDomainList() ([]string, error) {
	rawDomains, err := s.repo.ListDistinctDomains()
	if err != nil {
		return nil, fmt.Errorf("failed to get domain list: %w", err)
	}

	domainSet := make(map[string]bool)
	var result []string
	for _, d := range rawDomains {
		domains := embedded.ParseDomains(d)
		for _, domain := range domains {
			if domain != "" && !domainSet[domain] {
				domainSet[domain] = true
				result = append(result, domain)
			}
		}
	}
	return result, nil
}

func (s *EmbeddedService) InitIframe(token, origin string) ([]string, error) {
	if token == "" {
		return nil, fmt.Errorf("embedded token cannot be empty")
	}

	appId, err := s.extractAppIdFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid embedded token: %w", err)
	}

	e, err := s.repo.GetByAppId(appId)
	if err != nil {
		return nil, fmt.Errorf("embedded app not found: %w", err)
	}

	if !embedded.IsOriginAllowed(origin, e.Domain) {
		return nil, fmt.Errorf("embedded origin not allowed")
	}

	return embedded.ParseDomains(e.Domain), nil
}

func (s *EmbeddedService) extractAppIdFromToken(token string) (string, error) {
	parts := splitToken(token)
	for _, part := range parts {
		if key, val := parseClaim(part); key == "appId" {
			return val, nil
		}
	}
	return "", fmt.Errorf("appId claim not found")
}

func splitToken(token string) []string {
	parts := make([]string, 0)
	for _, segment := range splitBy(token, '.') {
		parts = append(parts, segment)
	}
	if len(parts) >= 2 {
		claims := decodeBase64(parts[1])
		return splitBy(claims, ',')
	}
	return nil
}

func splitBy(s string, sep rune) []string {
	var result []string
	var current string
	for _, r := range s {
		if r == sep {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func parseClaim(s string) (string, string) {
	for i, r := range s {
		if r == ':' || r == '=' {
			key := trimQuotes(s[:i])
			val := trimQuotes(s[i+1:])
			return key, val
		}
	}
	return "", ""
}

func trimQuotes(s string) string {
	s = trim(s, '"')
	s = trim(s, '\'')
	return s
}

func trim(s string, c rune) string {
	for len(s) > 0 && rune(s[0]) == c {
		s = s[1:]
	}
	for len(s) > 0 && rune(s[len(s)-1]) == c {
		s = s[:len(s)-1]
	}
	return s
}

func decodeBase64(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '-' {
			result = append(result, '+')
		} else if c == '_' {
			result = append(result, '/')
		} else {
			result = append(result, c)
		}
	}
	return string(result)
}

func (s *EmbeddedService) GetTokenArgs(userId, orgId int64) *embedded.TokenArgsResponse {
	return &embedded.TokenArgsResponse{
		UserId: userId,
		OrgId:  orgId,
	}
}

func (s *EmbeddedService) GetLimitCount() int {
	return 5
}
