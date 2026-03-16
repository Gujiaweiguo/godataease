package handler

import (
	"strings"

	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

const (
	localeZhCN = "zh-CN"
	localeEn   = "en"
	localeTw   = "tw"
)

type userByIDLoader func(userID int64) (*user.SysUser, error)

func normalizeLocale(input string) string {
	for _, part := range strings.Split(input, ",") {
		normalized := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(part, "_", "-")))
		if normalized == "" {
			continue
		}

		if separator := strings.Index(normalized, ";"); separator >= 0 {
			normalized = strings.TrimSpace(normalized[:separator])
		}

		switch {
		case normalized == localeTw, strings.HasPrefix(normalized, "zh-tw"), strings.HasPrefix(normalized, "zh-hk"):
			return localeTw
		case normalized == localeEn, strings.HasPrefix(normalized, "en"):
			return localeEn
		case normalized == strings.ToLower(localeZhCN), normalized == "zh", strings.HasPrefix(normalized, "zh"):
			return localeZhCN
		}
	}

	return ""
}

func resolveLocale(requestLanguage string, userLanguage string) string {
	if locale := normalizeLocale(requestLanguage); locale != "" {
		return locale
	}
	if locale := normalizeLocale(userLanguage); locale != "" {
		return locale
	}
	return localeZhCN
}

func requestLocale(c *gin.Context, loadUserByID userByIDLoader) string {
	return resolveLocale(c.GetHeader("Accept-Language"), currentUserLanguage(c, loadUserByID))
}

func currentUser(c *gin.Context, loadUserByID userByIDLoader) *user.SysUser {
	if loadUserByID == nil {
		return nil
	}

	userID := int64(middleware.GetUserID(c))
	if userID <= 0 {
		return nil
	}

	u, err := loadUserByID(userID)
	if err != nil {
		return nil
	}

	return u
}

func currentUserLanguage(c *gin.Context, loadUserByID userByIDLoader) string {
	u := currentUser(c, loadUserByID)
	if u == nil || u.Language == nil {
		return ""
	}

	return strings.TrimSpace(*u.Language)
}
