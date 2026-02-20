package embedded

import (
	"math/rand"
	"strings"
	"time"
)

const (
	DefaultSecretLength = 16
	TokenExpireTime     = 86400000 // 24 hours in ms
)

// CoreEmbedded 嵌入式应用实体 - 映射 core_embedded 表
type CoreEmbedded struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string `gorm:"column:name;size:255" json:"name"`
	AppId        string `gorm:"column:app_id;size:100" json:"appId"`
	AppSecret    string `gorm:"column:app_secret;size:255" json:"appSecret"`
	Domain       string `gorm:"column:domain;type:text" json:"domain"`
	SecretLength int    `gorm:"column:secret_length" json:"secretLength"`
	CreateTime   int64  `gorm:"column:create_time" json:"createTime"`
	UpdateBy     string `gorm:"column:update_by;size:100" json:"updateBy"`
	UpdateTime   int64  `gorm:"column:update_time" json:"updateTime"`
}

func (CoreEmbedded) TableName() string {
	return "core_embedded"
}

// EmbeddedCreator 创建嵌入式应用请求
type EmbeddedCreator struct {
	Name         string `json:"name" binding:"required"`
	Domain       string `json:"domain"`
	SecretLength *int   `json:"secretLength"`
}

// EmbeddedEditor 编辑嵌入式应用请求
type EmbeddedEditor struct {
	ID           int64   `json:"id" binding:"required"`
	Name         string  `json:"name"`
	Domain       *string `json:"domain"`
	SecretLength *int    `json:"secretLength"`
}

// EmbeddedResetRequest 重置密钥请求
type EmbeddedResetRequest struct {
	ID        int64   `json:"id" binding:"required"`
	AppSecret *string `json:"appSecret"`
}

// EmbeddedOrigin iframe 初始化请求
type EmbeddedOrigin struct {
	Token  string `json:"token" binding:"required"`
	Origin string `json:"origin"`
}

// EmbeddedGridVO 列表显示 VO
type EmbeddedGridVO struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	AppId        string `json:"appId"`
	AppSecret    string `json:"appSecret"`
	Domain       string `json:"domain"`
	SecretLength int    `json:"secretLength"`
}

// EmbeddedPagerResponse 分页响应
type EmbeddedPagerResponse struct {
	List    []EmbeddedGridVO `json:"list"`
	Total   int64            `json:"total"`
	Current int              `json:"current"`
	Size    int              `json:"size"`
}

// KeywordRequest 关键字查询请求
type KeywordRequest struct {
	Keyword *string `json:"keyword"`
}

// TokenArgsResponse Token 参数响应
type TokenArgsResponse struct {
	UserId int64 `json:"userId"`
	OrgId  int64 `json:"orgId"`
}

// GenerateAppId 生成应用 ID
func GenerateAppId() string {
	return "app_" + generateSnowflakeId()
}

// GenerateAppSecret 生成应用密钥
func GenerateAppSecret(length int) string {
	if length <= 0 {
		length = DefaultSecretLength
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// MaskAppSecret 密钥脱敏显示
func MaskAppSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return "********"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// ParseDomains 解析域名列表
func ParseDomains(domainList string) []string {
	if domainList == "" {
		return nil
	}
	parts := strings.FieldsFunc(domainList, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\n'
	})
	var domains []string
	for _, p := range parts {
		normalized := NormalizeOrigin(p)
		if normalized != "" {
			domains = append(domains, normalized)
		}
	}
	return domains
}

// NormalizeOrigin 标准化 origin
func NormalizeOrigin(origin string) string {
	if origin == "" {
		return ""
	}
	normalized := strings.TrimSpace(origin)
	for strings.HasSuffix(normalized, "/") {
		normalized = strings.TrimSuffix(normalized, "/")
	}
	return normalized
}

// IsOriginAllowed 检查 origin 是否在白名单中
func IsOriginAllowed(origin, domainList string) bool {
	if domainList == "" {
		return false
	}
	normalizedOrigin := NormalizeOrigin(origin)
	if normalizedOrigin == "" {
		return false
	}
	originHost := extractHost(normalizedOrigin)
	domains := ParseDomains(domainList)
	for _, allowed := range domains {
		if strings.EqualFold(allowed, normalizedOrigin) {
			return true
		}
		if originHost != "" && strings.EqualFold(allowed, originHost) {
			return true
		}
	}
	return false
}

// extractHost 从 URL 中提取主机名
func extractHost(origin string) string {
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(strings.ToLower(origin), prefix) {
			origin = origin[len(prefix):]
			break
		}
	}
	if idx := strings.Index(origin, "/"); idx > 0 {
		origin = origin[:idx]
	}
	if idx := strings.Index(origin, ":"); idx > 0 {
		origin = origin[:idx]
	}
	return origin
}

// generateSnowflakeId 生成类雪花 ID
func generateSnowflakeId() string {
	ts := time.Now().UnixNano() / 1000000
	randPart := rand.Int63n(10000)
	return formatInt64(ts*10000 + randPart)
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
