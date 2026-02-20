package auth

type PwdLoginDTO struct {
	Name   string `json:"name" binding:"required"`
	Pwd    string `json:"pwd" binding:"required"`
	Origin int    `json:"origin"`
}

type TokenVO struct {
	Token string `json:"token"`
	Exp   int64  `json:"exp"`
}

type TokenClaims struct {
	Uid int64 `json:"uid"`
	Oid int64 `json:"oid"`
}

type LoginConfig struct {
	AdminUsername        string
	AdminPasswordEnv     string
	DefaultAdminPassword string
}

func DefaultLoginConfig() *LoginConfig {
	return &LoginConfig{
		AdminUsername:        "admin",
		AdminPasswordEnv:     "ADMIN_PASSWORD",
		DefaultAdminPassword: "dataease",
	}
}
