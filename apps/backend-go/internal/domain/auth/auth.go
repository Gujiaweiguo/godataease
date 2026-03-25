package auth

type PwdLoginDTO struct {
	Name   string `json:"name" binding:"required"`
	Pwd    string `json:"pwd" binding:"required"`
	Origin int    `json:"origin"`
}

type TokenVO struct {
	Token         string       `json:"token"`
	Exp           int64        `json:"exp"`
	ID            int64        `json:"id,omitempty"`
	Name          string       `json:"name,omitempty"`
	Oid           int64        `json:"oid,omitempty"`
	Language      string       `json:"language,omitempty"`
	CurrentOrg    *OrgSummary  `json:"currentOrg,omitempty"`
	AvailableOrgs []OrgSummary `json:"availableOrgs,omitempty"`
}

type TokenClaims struct {
	Uid int64 `json:"uid"`
	Oid int64 `json:"oid"`
}

type OrgSummary struct {
	OrgID   int64  `json:"orgId"`
	OrgName string `json:"orgName"`
}

type IdentityBootstrap struct {
	ID            int64        `json:"id"`
	Name          string       `json:"name"`
	Oid           int64        `json:"oid"`
	Language      string       `json:"language"`
	CurrentOrg    *OrgSummary  `json:"currentOrg,omitempty"`
	AvailableOrgs []OrgSummary `json:"availableOrgs"`
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
